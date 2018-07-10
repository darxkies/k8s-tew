package config

import (
	"fmt"

	"github.com/darxkies/k8s-tew/utils"
	"github.com/satori/go.uuid"
)

type Config struct {
	Version                      string      `yaml:"version"`
	ClusterID                    string      `yaml:"cluster-id"`
	LoadBalancerPort             uint16      `yaml:"load-balancer-port"`
	APIServerPort                uint16      `yaml:"apiserver-port,omitempty"`
	PublicNetwork                string      `yaml:"public-network"`
	ControllerVirtualIP          string      `yaml:"controller-virtual-ip,omitempty"`
	ControllerVirtualIPInterface string      `yaml:"controller-virtual-ip-interface,omitempty"`
	WorkerVirtualIP              string      `yaml:"worker-virtual-ip,omitempty"`
	WorkerVirtualIPInterface     string      `yaml:"worker-virtual-ip-interface,omitempty"`
	ClusterIPRange               string      `yaml:"cluster-ip-range"`
	ClusterDNSIP                 string      `yaml:"cluster-dns-ip"`
	ClusterCIDR                  string      `yaml:"cluster-cidr"`
	ResolvConf                   string      `yaml:"resolv-conf"`
	DeploymentDirectory          string      `yaml:"deployment-directory,omitempty"`
	Assets                       AssetConfig `yaml:"assets,omitempty"`
	Nodes                        Nodes       `yaml:"nodes"`
	Commands                     Commands    `yaml:"commands,omitempty"`
	Servers                      Servers     `yaml:"servers,omitempty"`
}

func NewConfig() *Config {
	config := &Config{Version: utils.CONFIG_VERSION}

	config.ClusterID = fmt.Sprintf("%s", uuid.NewV4())
	config.LoadBalancerPort = utils.LOAD_BALANCER_PORT
	config.APIServerPort = utils.API_SERVER_PORT
	config.PublicNetwork = utils.PUBLIC_NETWORK
	config.ClusterIPRange = utils.CLUSTER_IP_RANGE
	config.ClusterDNSIP = utils.CLUSTER_DNS_IP
	config.ClusterCIDR = utils.CLUSTER_CIDR
	config.ResolvConf = utils.RESOLV_CONF
	config.Assets = AssetConfig{Directories: map[string]*AssetDirectory{}, Files: map[string]*AssetFile{}}
	config.Nodes = Nodes{}
	config.Commands = Commands{}
	config.Servers = Servers{}

	return config
}
