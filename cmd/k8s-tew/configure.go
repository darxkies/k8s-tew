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
		if error := bootstrap(false); error != nil {
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

	addUint16Option("rsa-key-size", utils.RsaSize, "RSA Key Size", func(value uint16) {
		_config.Config.RSASize = value
	})

	addUint16Option("ca-certificate-validity-period", utils.CaValidityPeriod, "CA Certificate Validity Period", func(value uint16) {
		_config.Config.CAValidityPeriod = uint(value)
	})

	addUint16Option("client-certificate-validity-period", utils.ClientValidityPeriod, "Client Certificate Validity Period", func(value uint16) {
		_config.Config.ClientValidityPeriod = uint(value)
	})

	addUint16Option("apiserver-port", utils.PortApiServer, "API Server Port", func(value uint16) {
		_config.Config.APIServerPort = value
	})

	addUint16Option("vip-raft-controller-port", utils.PortVipRaftController, "VIP Raft Controller Port", func(value uint16) {
		_config.Config.VIPRaftControllerPort = value
	})

	addUint16Option("vip-raft-worker-port", utils.PortVipRaftWorker, "VIP Raft Worker Port", func(value uint16) {
		_config.Config.VIPRaftWorkerPort = value
	})

	addUint16Option("load-balancer-port", utils.PortKubernetesDashboard, "Load Balancer Port", func(value uint16) {
		_config.Config.LoadBalancerPort = value
	})

	addUint16Option("kubernetes-dashboard-port", utils.PortKubernetesDashboard, "Kubernetes Dashboard Port", func(value uint16) {
		_config.Config.KubernetesDashboardPort = value
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

	addStringOption("cluster-domain", utils.ClusterDomain, "Cluster domain", func(value string) {
		_config.Config.ClusterDomain = value
	})

	addStringOption("cluster-ip-range", utils.ClusterIpRange, "Cluster IP range", func(value string) {
		_config.Config.ClusterIPRange = value
	})

	addStringOption("cluster-dns-ip", utils.ClusterDnsIp, "Cluster DNS IP", func(value string) {
		_config.Config.ClusterDNSIP = value
	})

	addStringOption("cluster-cidr", utils.ClusterCidr, "Cluster CIDR", func(value string) {
		_config.Config.ClusterCIDR = value
	})

	addStringOption("calico-typha-ip", utils.CalicoTyphaIp, "Calico Typha IP", func(value string) {
		_config.Config.CalicoTyphaIP = value
	})

	addStringOption("metallb-addresses", utils.MetalLBAddresses, "Comma separated MetalLB address ranges and CIDR (e.g 192.168.0.16/28,192.168.0.75-192.168.0.100)", func(value string) {
		_config.Config.MetalLBAddresses = value
	})

	addStringOption("resolv-conf", utils.ResolvConf, "Custom resolv.conf", func(value string) {
		_config.Config.ResolvConf = value
	})

	addStringOption("public-network", utils.PublicNetwork, "Public Network", func(value string) {
		_config.Config.PublicNetwork = value
	})

	addStringOption("cluster-name", utils.ClusterName, "Cluster Name used for Kubernetes Dashboard", func(value string) {
		_config.Config.ClusterName = value
	})

	addStringOption("email", utils.Email, "Email address used for example for Let's Encrypt", func(value string) {
		_config.Config.Email = value
	})

	addStringOption("ingress-domain", utils.IngressDomain, "Ingress domain name", func(value string) {
		_config.Config.IngressDomain = value
	})

	addStringOption("deployment-directory", utils.DeploymentDirectory, "Deployment directory", func(value string) {
		_config.Config.DeploymentDirectory = value
	})

	addStringOption("version-etcd", utils.VersionEtcd, "Etcd version", func(value string) {
		_config.Config.Versions.Etcd = value
	})

	addStringOption("version-k8s", utils.VersionK8s, "Kubernetes version", func(value string) {
		_config.Config.Versions.K8S = value
	})

	addStringOption("version-helm", utils.VersionHelm, "Helm version", func(value string) {
		_config.Config.Versions.Helm = value
	})

	addStringOption("version-containerd", utils.VersionContainerd, "Containerd version", func(value string) {
		_config.Config.Versions.Containerd = value
	})

	addStringOption("version-runc", utils.VersionRunc, "Runc version", func(value string) {
		_config.Config.Versions.Runc = value
	})

	addStringOption("version-crictl", utils.VersionCrictl, "CriCtl version", func(value string) {
		_config.Config.Versions.CriCtl = value
	})

	addStringOption("version-gobetween", utils.VersionGobetween, "Gobetween version", func(value string) {
		_config.Config.Versions.Gobetween = value
	})

	addStringOption("version-virtual-ip", utils.VersionVirtualIP, "Virtual-IP version", func(value string) {
		_config.Config.Versions.VirtualIP = value
	})

	addStringOption("version-busybox", utils.VersionBusybox, "Busybox version", func(value string) {
		_config.Config.Versions.Busybox = value
	})

	addStringOption("version-ark", utils.VersionArk, "Ark version", func(value string) {
		_config.Config.Versions.Ark = value
	})

	addStringOption("version-minio-server", utils.VersionMinioServer, "Minio server version", func(value string) {
		_config.Config.Versions.MinioServer = value
	})

	addStringOption("version-minio-client", utils.VersionMinioClient, "Minio client version", func(value string) {
		_config.Config.Versions.MinioClient = value
	})

	addStringOption("version-pause", utils.VersionPause, "Pause version", func(value string) {
		_config.Config.Versions.Pause = value
	})

	addStringOption("version-coredns", utils.VersionCoredns, "CoreDNS version", func(value string) {
		_config.Config.Versions.CoreDNS = value
	})

	addStringOption("version-elasticsearch", utils.VersionElasticsearch, "Elasticsearch version", func(value string) {
		_config.Config.Versions.Elasticsearch = value
	})

	addStringOption("version-elasticsearch-cron", utils.VersionElasticsearchCron, "Elasticsearch Cron version", func(value string) {
		_config.Config.Versions.ElasticsearchCron = value
	})

	addStringOption("version-elasticsearch-operator", utils.VersionElasticsearchOperator, "Elasticsearch Operator version", func(value string) {
		_config.Config.Versions.ElasticsearchOperator = value
	})

	addStringOption("version-kibana", utils.VersionKibana, "Kibana version", func(value string) {
		_config.Config.Versions.Kibana = value
	})

	addStringOption("version-cerebro", utils.VersionCerebro, "Cerebro version", func(value string) {
		_config.Config.Versions.Cerebro = value
	})

	addStringOption("version-fluent-bit", utils.VersionFluentBit, "Fluent-Bit version", func(value string) {
		_config.Config.Versions.FluentBit = value
	})

	addStringOption("version-calico-typha", utils.VersionCalicoTypha, "Calico Typha version", func(value string) {
		_config.Config.Versions.CalicoTypha = value
	})

	addStringOption("version-calico-node", utils.VersionCalicoNode, "Calico Node version", func(value string) {
		_config.Config.Versions.CalicoNode = value
	})

	addStringOption("version-calico-cni", utils.VersionCalicoCni, "Calico CNI version", func(value string) {
		_config.Config.Versions.CalicoCNI = value
	})

	addStringOption("version-calico-kube-controllers", utils.VersionCalicoKubeControllers, "Calico Kube Controllers  version", func(value string) {
		_config.Config.Versions.CalicoKubeControllers = value
	})

	addStringOption("version-metallb-controller", utils.VersionMetalLBController, "MetalLB Controller version", func(value string) {
		_config.Config.Versions.MetalLBController = value
	})

	addStringOption("version-metallb-speaker", utils.VersionMetalLBSpeaker, "MetalLB Speaker version", func(value string) {
		_config.Config.Versions.MetalLBSpeaker = value
	})

	addStringOption("version-ceph", utils.VersionCeph, "Ceph version", func(value string) {
		_config.Config.Versions.Ceph = value
	})

	addStringOption("version-heapster", utils.VersionHeapster, "Heapster version", func(value string) {
		_config.Config.Versions.Heapster = value
	})

	addStringOption("version-addon-resizer", utils.VersionAddonResizer, "Addon-Resizer version", func(value string) {
		_config.Config.Versions.AddonResizer = value
	})

	addStringOption("version-kubernetes-dashboard", utils.VersionKubernetesDashboard, "Kubernetes Dashboard version", func(value string) {
		_config.Config.Versions.KubernetesDashboard = value
	})

	addStringOption("version-cert-manager-controller", utils.VersionCertManagerController, "Cert Manager Controller version", func(value string) {
		_config.Config.Versions.CertManagerController = value
	})

	addStringOption("version-nginx-ingress-controller", utils.VersionNginxIngressController, "Nginx Ingress Controller version", func(value string) {
		_config.Config.Versions.NginxIngressController = value
	})

	addStringOption("version-nginx-ingress-default-backend", utils.VersionNginxIngressDefaultBackend, "Nginx Ingress Default Backend version", func(value string) {
		_config.Config.Versions.NginxIngressDefaultBackend = value
	})

	addStringOption("version-metrics-server", utils.VersionMetricsServer, "Metrics Server version", func(value string) {
		_config.Config.Versions.MetricsServer = value
	})

	addStringOption("version-prometheus-operator", utils.VersionPrometheusOperator, "Prometheus Operator version", func(value string) {
		_config.Config.Versions.PrometheusOperator = value
	})

	addStringOption("version-prometheus-config-reloader", utils.VersionPrometheusConfigReloader, "Prometheus Config Reloader version", func(value string) {
		_config.Config.Versions.PrometheusConfigReloader = value
	})

	addStringOption("version-configmap-reload", utils.VersionConfigmapReload, "ConfigMap Reload version", func(value string) {
		_config.Config.Versions.ConfigMapReload = value
	})

	addStringOption("version-kube-state-metrics", utils.VersionKubeStateMetrics, "Kube State Metrics version", func(value string) {
		_config.Config.Versions.KubeStateMetrics = value
	})

	addStringOption("version-grafana", utils.VersionGrafana, "Grafana version", func(value string) {
		_config.Config.Versions.Grafana = value
	})

	addStringOption("version-grafana-watcher", utils.VersionGrafanaWatcher, "Grafana Watcher version", func(value string) {
		_config.Config.Versions.GrafanaWatcher = value
	})

	addStringOption("version-prometheus-node-exporter", utils.VersionPrometheusNodeExporter, "Prometheus Node Exporter version", func(value string) {
		_config.Config.Versions.PrometheusNodeExporter = value
	})

	addStringOption("version-prometheus-alert-manager", utils.VersionPrometheusAlertManager, "Prometheus Alert Manager version", func(value string) {
		_config.Config.Versions.PrometheusAlertManager = value
	})

	addStringOption("version-prometheus", utils.VersionPrometheus, "Prometheus version", func(value string) {
		_config.Config.Versions.Prometheus = value
	})

	addStringOption("version-csi-attacher", utils.VersionCsiAttacher, "CSI Attacher version", func(value string) {
		_config.Config.Versions.CSIAttacher = value
	})

	addStringOption("version-csi-provisioner", utils.VersionCsiProvisioner, "CSI Provisioner version", func(value string) {
		_config.Config.Versions.CSIProvisioner = value
	})

	addStringOption("version-csi-driver-registrar", utils.VersionCsiDriverRegistrar, "CSI Driver Registrar version", func(value string) {
		_config.Config.Versions.CSIDriverRegistrar = value
	})

	addStringOption("version-csi-ceph-rbd-plugin", utils.VersionCsiCephRbdPlugin, "CSI Ceph RBD Plugin version", func(value string) {
		_config.Config.Versions.CSICephRBDPlugin = value
	})

	addStringOption("version-csi-ceph-fs-plugin", utils.VersionCsiCephFsPlugin, "CSI Ceph FS Plugin version", func(value string) {
		_config.Config.Versions.CSICephFSPlugin = value
	})

	addStringOption("version-csi-ceph-snapshotter", utils.VersionCsiCephSnapshotter, "CSI Ceph Snapshotter version", func(value string) {
		_config.Config.Versions.CSICephSnapshotter = value
	})

	addStringOption("version-wordpress", utils.VersionWordpress, "WordPress version", func(value string) {
		_config.Config.Versions.WordPress = value
	})

	addStringOption("version-mysql", utils.VersionMysql, "MySQL version", func(value string) {
		_config.Config.Versions.MySQL = value
	})

	RootCmd.AddCommand(configureCmd)
}
