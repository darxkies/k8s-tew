package main

import (
	"os"

	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var controllerVirtualIP string
var controllerVirtualIPInterface string
var workerVirtualIP string
var workerVirtualIPInterface string
var clusterIPRange string
var clusterDNSIP string
var clusterCIDR string
var resolvConf string
var apiServerPort uint16
var loadBalancerPort uint16

const CONTROLLER_VIRTUAL_IP = "controller-virtual-ip"
const CONTROLLER_VIRTUAL_IP_INTERFACE = "controller-virtual-ip-interface"
const WORKER_VIRTUAL_IP = "worker-virtual-ip"
const WORKER_VIRTUAL_IP_INTERFACE = "worker-virtual-ip-interface"
const CLUSTER_IP_RANGE = "cluster-ip-range"
const CLUSTER_DNS_IP = "cluster-dns-ip"
const CLUSTER_CIDR = "cluster-cidr"
const RESOLV_CONF = "resolv-conf"
const API_SERVER_PORT = "api-server-port"
const LOAD_BALANCER_PORT = "load-balancer-port"

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set configuration settings",
	Long:  "Set configuration settings",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("configure failed")

			os.Exit(-1)
		}

		cmd.Flags().Visit(func(flag *pflag.Flag) {
			if flag.Name == CONTROLLER_VIRTUAL_IP {
				_config.Config.ControllerVirtualIP = controllerVirtualIP
			}

			if flag.Name == CONTROLLER_VIRTUAL_IP_INTERFACE {
				_config.Config.ControllerVirtualIPInterface = controllerVirtualIPInterface
			}

			if flag.Name == WORKER_VIRTUAL_IP {
				_config.Config.WorkerVirtualIP = workerVirtualIP
			}

			if flag.Name == WORKER_VIRTUAL_IP_INTERFACE {
				_config.Config.WorkerVirtualIPInterface = workerVirtualIPInterface
			}

			if flag.Name == CLUSTER_IP_RANGE {
				_config.Config.ClusterIPRange = clusterIPRange
			}

			if flag.Name == CLUSTER_DNS_IP {
				_config.Config.ClusterDNSIP = clusterDNSIP
			}

			if flag.Name == CLUSTER_CIDR {
				_config.Config.ClusterCIDR = clusterCIDR
			}

			if flag.Name == RESOLV_CONF {
				_config.Config.ResolvConf = resolvConf
			}

			if flag.Name == API_SERVER_PORT {
				_config.Config.APIServerPort = apiServerPort
			}

			if flag.Name == LOAD_BALANCER_PORT {
				_config.Config.LoadBalancerPort = loadBalancerPort
			}
		})

		if error := _config.Save(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("configure failed")

			os.Exit(-1)
		}
	},
}

func init() {
	configureCmd.Flags().StringVar(&controllerVirtualIP, CONTROLLER_VIRTUAL_IP, "", "Controller Virtual IP for the cluster")
	configureCmd.Flags().StringVar(&controllerVirtualIPInterface, CONTROLLER_VIRTUAL_IP_INTERFACE, "", "Controller Virtual IP interface for the cluster")
	configureCmd.Flags().StringVar(&workerVirtualIP, WORKER_VIRTUAL_IP, "", "Worker Virtual IP for the cluster")
	configureCmd.Flags().StringVar(&workerVirtualIPInterface, WORKER_VIRTUAL_IP_INTERFACE, "", "Worker Virtual IP interface for the cluster")
	configureCmd.Flags().StringVar(&clusterIPRange, CLUSTER_IP_RANGE, utils.CLUSTER_IP_RANGE, "Cluster IP range")
	configureCmd.Flags().StringVar(&clusterDNSIP, CLUSTER_DNS_IP, utils.CLUSTER_DNS_IP, "Cluster DNS IP")
	configureCmd.Flags().StringVar(&clusterCIDR, CLUSTER_CIDR, utils.CLUSTER_CIDR, "Cluster CIDR")
	configureCmd.Flags().StringVar(&resolvConf, RESOLV_CONF, utils.RESOLV_CONF, "Custom resolv.conf")
	configureCmd.Flags().Uint16Var(&apiServerPort, API_SERVER_PORT, utils.API_SERVER_PORT, "API Server Port")
	configureCmd.Flags().Uint16Var(&loadBalancerPort, LOAD_BALANCER_PORT, utils.LOAD_BALANCER_PORT, "Load Balancer Port")
	RootCmd.AddCommand(configureCmd)
}
