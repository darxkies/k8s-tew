package config

import (
	"fmt"

	"github.com/darxkies/k8s-tew/utils"
	"github.com/satori/go.uuid"
)

type Config struct {
	Version                      string      `yaml:"version"`
	VIPRaftControllerPort        uint16      `yaml:"vip-raft-controller-port"`
	VIPRaftWorkerPort            uint16      `yaml:"vip-raft-worker-port"`
	ClusterID                    string      `yaml:"cluster-id"`
	Email                        string      `ỳaml:"email"`
	IngressDomain                string      `yaml:"ingress-domain"`
	LoadBalancerPort             uint16      `yaml:"load-balancer-port"`
	DashboardPort                uint16      `yaml:"dashboard-port"`
	APIServerPort                uint16      `yaml:"apiserver-port,omitempty"`
	PublicNetwork                string      `yaml:"public-network"`
	ControllerVirtualIP          string      `yaml:"controller-virtual-ip,omitempty"`
	ControllerVirtualIPInterface string      `yaml:"controller-virtual-ip-interface,omitempty"`
	WorkerVirtualIP              string      `yaml:"worker-virtual-ip,omitempty"`
	WorkerVirtualIPInterface     string      `yaml:"worker-virtual-ip-interface,omitempty"`
	ClusterDomain                string      `yaml:"cluster-domain"`
	ClusterIPRange               string      `yaml:"cluster-ip-range"`
	ClusterDNSIP                 string      `yaml:"cluster-dns-ip"`
	ClusterCIDR                  string      `yaml:"cluster-cidr"`
	ResolvConf                   string      `yaml:"resolv-conf"`
	DeploymentDirectory          string      `yaml:"deployment-directory,omitempty"`
	RSASize                      uint16      `yaml:"rsa-size"`
	CAValidityPeriod             uint        `yaml:"ca-validity-period"`
	ClientValidityPeriod         uint        `yaml:"client-validity-period"`
	Versions                     Versions    `yaml:"versions"`
	Assets                       AssetConfig `yaml:"assets,omitempty"`
	Nodes                        Nodes       `yaml:"nodes"`
	Commands                     Commands    `yaml:"commands,omitempty"`
	Servers                      Servers     `yaml:"servers,omitempty"`
}

func NewConfig() *Config {
	config := &Config{Version: utils.VERSION_CONFIG}

	config.VIPRaftControllerPort = utils.VIP_RAFT_CONTROLLER_PORT
	config.VIPRaftWorkerPort = utils.VIP_RAFT_WORKER_PORT
	config.ClusterID = fmt.Sprintf("%s", uuid.NewV4())
	config.Email = utils.EMAIL
	config.IngressDomain = utils.INGRESS_DOMAIN
	config.LoadBalancerPort = utils.LOAD_BALANCER_PORT
	config.DashboardPort = utils.DASHBOARD_PORT
	config.APIServerPort = utils.API_SERVER_PORT
	config.PublicNetwork = utils.PUBLIC_NETWORK
	config.ClusterDomain = utils.CLUSTER_DOMAIN
	config.ClusterIPRange = utils.CLUSTER_IP_RANGE
	config.ClusterDNSIP = utils.CLUSTER_DNS_IP
	config.ClusterCIDR = utils.CLUSTER_CIDR
	config.ResolvConf = utils.RESOLV_CONF
	config.DeploymentDirectory = utils.DEPLOYMENT_DIRECTORY
	config.RSASize = utils.RSA_SIZE
	config.CAValidityPeriod = utils.CA_VALIDITY_PERIOD
	config.ClientValidityPeriod = utils.CLIENT_VALIDITY_PERIOD
	config.Versions = NewVersions()
	config.Assets = AssetConfig{Directories: map[string]*AssetDirectory{}, Files: map[string]*AssetFile{}}
	config.Nodes = Nodes{}
	config.Commands = Commands{}
	config.Servers = Servers{}

	return config
}
