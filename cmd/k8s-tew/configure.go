package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set configuration settings",
	Long:  "Set configuration settings",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config and check the rights
		if error := Bootstrap(false); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Configure failed")

			os.Exit(-1)
		}

		utils.SetProgressSteps(1)

		cmd.Flags().Visit(func(flag *pflag.Flag) {
			for key, handler := range setterHandlers {
				if flag.Name != key {
					continue
				}

				handler(flag.Value.String())

				log.WithFields(log.Fields{"parameter": flag.Name, "value": flag.Value}).Info("Configuration changed")

				break
			}
		})

		if error := _config.Save(); error != nil {
			log.WithFields(log.Fields{"error": error}).Error("Configure failed")

			os.Exit(-1)
		}
	},
}

type stringSetter func(value string)
type uint16Setter func(value uint16)

var setterHandlers map[string]stringSetter

func addStringOption(name string, value string, description string, handler stringSetter) {
	configureCmd.Flags().String(name, value, description)

	setterHandlers[name] = handler
}

func addUint16Option(name string, value uint16, description string, handler uint16Setter) {
	configureCmd.Flags().Uint16(name, value, description)

	setterHandlers[name] = func(value string) {
		fmt.Printf("%s\n", value)

		_value, _ := strconv.ParseUint(value, 10, 16)

		fmt.Printf("%d\n", _value)

		handler(uint16(_value))
	}
}

func init() {
	setterHandlers = map[string]stringSetter{}

	addUint16Option("rsa-key-size", utils.RSA_SIZE, "RSA Key Size", func(value uint16) {
		_config.Config.RSASize = value
	})

	addUint16Option("ca-certificate-validity-period", utils.CA_VALIDITY_PERIOD, "CA Certificate Validity Period", func(value uint16) {
		_config.Config.CAValidityPeriod = uint(value)
	})

	addUint16Option("client-certificate-validity-period", utils.CLIENT_VALIDITY_PERIOD, "Client Certificate Validity Period", func(value uint16) {
		_config.Config.ClientValidityPeriod = uint(value)
	})

	addUint16Option("apiserver-port", utils.API_SERVER_PORT, "API Server Port", func(value uint16) {
		_config.Config.APIServerPort = value
	})

	addUint16Option("load-balancer-port", utils.LOAD_BALANCER_PORT, "Load Balancer Port", func(value uint16) {
		_config.Config.LoadBalancerPort = value
	})

	addUint16Option("dashboard-port", utils.DASHBOARD_PORT, "Dashboard Port", func(value uint16) {
		_config.Config.DashboardPort = value
	})

	addStringOption("controller-virtual-ip", "", "Controller Virtual/Floating IP for the cluster", func(value string) {
		_config.Config.ControllerVirtualIP = value
	})

	addStringOption("controller-virtual-ip-interface", "", "Controller Virtual/Floating IP interface for the cluster", func(value string) {
		_config.Config.ControllerVirtualIPInterface = value
	})

	addStringOption("worker-virtual-ip", "", "Worker Virtual/Floating IP for the cluster", func(value string) {
		_config.Config.WorkerVirtualIP = value
	})

	addStringOption("worker-virtual-ip-interface", "", "Worker Virtual/Floating IP interface for the cluster", func(value string) {
		_config.Config.WorkerVirtualIPInterface = value
	})

	addStringOption("cluster-domain", utils.CLUSTER_DOMAIN, "Cluster domain", func(value string) {
		_config.Config.ClusterDomain = value
	})

	addStringOption("cluster-ip-range", utils.CLUSTER_IP_RANGE, "Cluster IP range", func(value string) {
		_config.Config.ClusterIPRange = value
	})

	addStringOption("cluster-dns-ip", utils.CLUSTER_DNS_IP, "Cluster DNS IP", func(value string) {
		_config.Config.ClusterDNSIP = value
	})

	addStringOption("cluster-cidr", utils.CLUSTER_CIDR, "Cluster CIDR", func(value string) {
		_config.Config.ClusterCIDR = value
	})

	addStringOption("resolv-conf", utils.RESOLV_CONF, "Custom resolv.conf", func(value string) {
		_config.Config.ResolvConf = value
	})

	addStringOption("public-network", utils.PUBLIC_NETWORK, "Public Network", func(value string) {
		_config.Config.PublicNetwork = value
	})

	addStringOption("email", utils.EMAIL, "Email address used for example for Let's Encrypt", func(value string) {
		_config.Config.Email = value
	})

	addStringOption("ingress-domain", utils.INGRESS_DOMAIN, "Ingress domain name", func(value string) {
		_config.Config.IngressDomain = value
	})

	addStringOption("deployment-directory", utils.DEPLOYMENT_DIRECTORY, "Deployment directory", func(value string) {
		_config.Config.DeploymentDirectory = value
	})

	addStringOption("version-etcd", utils.ETCD_VERSION, "Etcd version", func(value string) {
		_config.Config.Versions.Etcd = value
	})

	addStringOption("version-k8s", utils.K8S_VERSION, "Kubernetes version", func(value string) {
		_config.Config.Versions.K8S = value
	})

	addStringOption("version-helm", utils.HELM_VERSION, "Helm version", func(value string) {
		_config.Config.Versions.Helm = value
	})

	addStringOption("version-containerd", utils.CONTAINERD_VERSION, "Containerd version", func(value string) {
		_config.Config.Versions.Containerd = value
	})

	addStringOption("version-runc", utils.RUNC_VERSION, "Runc version", func(value string) {
		_config.Config.Versions.Runc = value
	})

	addStringOption("version-crictl", utils.CRICTL_VERSION, "CriCtl version", func(value string) {
		_config.Config.Versions.CriCtl = value
	})

	addStringOption("version-gobetween", utils.GOBETWEEN_VERSION, "Gobetween version", func(value string) {
		_config.Config.Versions.Gobetween = value
	})

	addStringOption("version-ark", utils.ARK_VERSION, "Ark version", func(value string) {
		_config.Config.Versions.Ark = value
	})

	addStringOption("version-minio-server", utils.MINIO_SERVER_VERSION, "Minio server version", func(value string) {
		_config.Config.Versions.MinioServer = value
	})

	addStringOption("version-minio-client", utils.MINIO_CLIENT_VERSION, "Minio client version", func(value string) {
		_config.Config.Versions.MinioClient = value
	})

	RootCmd.AddCommand(configureCmd)
}
