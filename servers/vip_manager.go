package servers

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/darxkies/k8s-tew/utils"
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

type Peer struct {
	ID   string
	Bind string
}

type Peers map[string]string

type Logger struct {
}

func (logger Logger) Write(data []byte) (count int, error error) {
	return len(data), nil
}

type FSM struct {
}

func (fsm FSM) Apply(log *raft.Log) interface{} {
	return nil
}

func (fsm FSM) Restore(snap io.ReadCloser) error {
	return nil
}

func (fsm FSM) Snapshot() (raft.FSMSnapshot, error) {
	return Snapshot{}, nil
}

type Snapshot struct {
}

func (snapshot Snapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (snapshot Snapshot) Release() {
}

type VIPManager struct {
	_type      string
	id         string
	bind       string
	virtualIP  string
	fsm        FSM
	peers      Peers
	logger     Logger
	_interface string
	stop       chan bool
}

func NewVIPManager(_type, id, bind string, virtualIP string, peers Peers, logger Logger, _interface string) *VIPManager {
	return &VIPManager{_type: _type, id: id, peers: peers, bind: bind, virtualIP: virtualIP, fsm: FSM{}, logger: logger, _interface: _interface}
}

func (manager *VIPManager) Name() string {
	return "vip-manager-" + manager._type
}

func (manager *VIPManager) updateNetworkConfiguration(action string) error {
	command := fmt.Sprintf("ip addr %s %s/32 dev %s", action, manager.virtualIP, manager._interface)

	if error := utils.RunCommand(command); error != nil {
		log.WithFields(log.Fields{"action": action, "name": manager.Name(), "error": error}).Error("Network update failed")

		return error
	}

	return nil
}

func (manager *VIPManager) addIP() error {
	log.WithFields(log.Fields{"name": manager.Name()}).Info("Add virtual ip")

	return manager.updateNetworkConfiguration("add")
}

func (manager *VIPManager) deleteIP() error {
	log.WithFields(log.Fields{"name": manager.Name()}).Info("Delete virtual ip")

	return manager.updateNetworkConfiguration("delete")
}

func (manager *VIPManager) Start() error {
	// Create configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(manager.id)
	config.LogOutput = manager.logger

	// Initialize communication
	address, error := net.ResolveTCPAddr("tcp", manager.bind)
	if error != nil {
		return error
	}

	// Create transport
	transport, error := raft.NewTCPTransport(manager.bind, address, 3, 10*time.Second, manager.logger)
	if error != nil {
		return error
	}

	// Create Raft structures
	snapshots := raft.NewInmemSnapshotStore()
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Cluster configuration
	configuration := raft.Configuration{}

	for id, ip := range manager.peers {
		configuration.Servers = append(configuration.Servers, raft.Server{ID: raft.ServerID(id), Address: raft.ServerAddress(ip)})
	}

	// Bootstrap cluster
	if error := raft.BootstrapCluster(config, logStore, stableStore, snapshots, transport, configuration); error != nil {
		return error
	}

	// Create RAFT instance
	raftServer, error := raft.NewRaft(config, manager.fsm, logStore, stableStore, snapshots, transport)
	if error != nil {
		return error
	}

	manager.stop = make(chan bool, 1)

	manager.deleteIP()

	go func() {
		for {
			select {
			case leader := <-raftServer.LeaderCh():
				if leader {
					manager.addIP()
				} else {
					manager.deleteIP()
				}

			case <-manager.stop:
				manager.deleteIP()
			}
		}
	}()

	return nil
}

func (manager *VIPManager) Stop() {
	close(manager.stop)
}
