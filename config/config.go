package config

import (
	"github.com/darxkies/k8s-tew/utils"
	uuid "github.com/satori/go.uuid"
)

type Config struct {
	Version                      string      `yaml:"version"`
	ClusterID                    string      `yaml:"cluster-id"`
	ClusterName                  string      `yaml:"cluster-name"`
	Email                        string      `yaml:"email"`
	IngressDomain                string      `yaml:"ingress-domain"`
	LoadBalancerPort             uint16      `yaml:"load-balancer-port"`
	VIPRaftControllerPort        uint16      `yaml:"vip-raft-controller-port"`
	VIPRaftWorkerPort            uint16      `yaml:"vip-raft-worker-port"`
	KubernetesDashboardPort      uint16      `yaml:"kubernetes-dashboard-port"`
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
	CalicoTyphaIP                string      `yaml:"calico-typha-ip"`
	MetalLBAddresses             string      `yaml:"metallb-addresses"`
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
	config := &Config{Version: utils.VersionConfig}

	config.VIPRaftControllerPort = utils.PortVipRaftController
	config.VIPRaftWorkerPort = utils.PortVipRaftWorker
	config.ClusterID = uuid.NewV4().String()
	config.ClusterName = utils.ClusterName
	config.Email = utils.Email
	config.IngressDomain = utils.IngressDomain
	config.LoadBalancerPort = utils.PortLoadBalancer
	config.KubernetesDashboardPort = utils.PortKubernetesDashboard
	config.APIServerPort = utils.PortApiServer
	config.PublicNetwork = utils.PublicNetwork
	config.ClusterDomain = utils.ClusterDomain
	config.ClusterIPRange = utils.ClusterIpRange
	config.ClusterDNSIP = utils.ClusterDnsIp
	config.ClusterCIDR = utils.ClusterCidr
	config.CalicoTyphaIP = utils.CalicoTyphaIp
	config.MetalLBAddresses = utils.MetalLBAddresses
	config.ResolvConf = utils.ResolvConf
	config.DeploymentDirectory = utils.DeploymentDirectory
	config.RSASize = utils.RsaSize
	config.CAValidityPeriod = utils.CaValidityPeriod
	config.ClientValidityPeriod = utils.ClientValidityPeriod
	config.Versions = NewVersions()
	config.Assets = AssetConfig{Directories: map[string]*AssetDirectory{}, Files: map[string]*AssetFile{}}
	config.Nodes = Nodes{}
	config.Commands = Commands{}
	config.Servers = Servers{}

	return config
}
