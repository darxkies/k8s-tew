package servers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
)

type VIPManager struct {
	id                 string
	session            *concurrency.Session
	election           *concurrency.Election
	endpoints          []string
	caCertificate      string
	clientCertificate  string
	clientKey          string
	stopSignal         chan struct{}
	stopWaitGroup      sync.WaitGroup
	ip                 string
	_interface         string
	electionKey        string
	campaignContext    context.Context
	campaignCancel     context.CancelFunc
	observationContext context.Context
	observationCancel  context.CancelFunc
	stop               bool
}

func NewVIPManager(electionKey string, id string, ip, _interface string, endpoints []string, caCertificate string, clientCertificate string, clientKey string) *VIPManager {
	result := &VIPManager{}

	result.id = id
	result.endpoints = endpoints
	result.caCertificate = caCertificate
	result.clientCertificate = clientCertificate
	result.clientKey = clientKey
	result.stopSignal = make(chan struct{})
	result.ip = ip
	result._interface = _interface
	result.electionKey = electionKey
	result.campaignContext, result.campaignCancel = context.WithCancel(context.Background())
	result.observationContext, result.observationCancel = context.WithCancel(context.Background())
	result.stop = false

	return result
}

func (manager *VIPManager) getSession() (*concurrency.Session, error) {
	tlsInfo := transport.TLSInfo{
		CertFile:      manager.clientCertificate,
		KeyFile:       manager.clientKey,
		TrustedCAFile: manager.caCertificate,
	}

	tlsConfig, error := tlsInfo.ClientConfig()
	if error != nil {
		return nil, error
	}

	config := clientv3.Config{
		Endpoints:   manager.endpoints,
		DialTimeout: time.Second,
		TLS:         tlsConfig,
	}

	_client, error := clientv3.New(config)
	if error != nil {
		return nil, error
	}

	return concurrency.NewSession(_client, concurrency.WithTTL(2))
}

func (manager *VIPManager) Name() string {
	return "vip-manager" + manager.electionKey
}

func (manager *VIPManager) updateNetworkConfiguration(action string) error {
	command := fmt.Sprintf("ip addr %s %s/32 dev %s", action, manager.ip, manager._interface)

	return utils.RunCommand(command)
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
	manager.stopWaitGroup.Add(2)

	manager.deleteIP()

	// Campaign
	go func() {
		for !manager.stop {
			session, error := manager.getSession()

			if error != nil {
				log.WithFields(log.Fields{"name": manager.Name(), "error": error}).Error("Starting campaign failed")

				continue
			}

			// Register session close
			defer session.Close()

			// Initiate an election
			election := concurrency.NewElection(session, utils.ELECTION_NAMESPACE+manager.electionKey)

			log.WithFields(log.Fields{"name": manager.Name()}).Info("Start campaign")

			// Enter campaign
			if error := election.Campaign(manager.campaignContext, manager.id); error != nil {
				if error.Error() == "context canceled" {
					log.WithFields(log.Fields{"name": manager.Name()}).Info("Campaign canceled")

					break
				}

				log.WithFields(log.Fields{"name": manager.Name(), "error": error}).Error("Campaign terminated")

				continue
			}

			// Resign on exiting
			defer election.Resign(context.Background())

			// Elected
			log.WithFields(log.Fields{"name": manager.Name()}).Info("Elected leader")
		}

		manager.stopWaitGroup.Done()
	}()

	// Observation
	go func() {
		for !manager.stop {

			session, error := manager.getSession()

			if error != nil {
				log.WithFields(log.Fields{"name": manager.Name(), "error": error}).Error("Starting observation failed")

				continue
			}

			// Register session close
			defer session.Close()

			// Initiate observation
			election := concurrency.NewElection(session, utils.ELECTION_NAMESPACE+manager.electionKey)

			log.WithFields(log.Fields{"name": manager.Name()}).Info("Start observation")

			for observation := range election.Observe(manager.observationContext) {
				value := string(observation.Kvs[0].Value)

				log.WithFields(log.Fields{"name": manager.Name(), "value": value, "own": manager.id}).Info("Observation notification")

				if value == manager.id {
					manager.addIP()
				} else {
					manager.deleteIP()
				}
			}

			break
		}

		manager.deleteIP()

		manager.stopWaitGroup.Done()
	}()

	log.WithFields(log.Fields{"name": manager.Name()}).Info("Started server")

	return nil
}

func (manager *VIPManager) Stop() {
	manager.stop = true

	manager.campaignCancel()
	manager.observationCancel()

	close(manager.stopSignal)

	manager.stopWaitGroup.Wait()
}
