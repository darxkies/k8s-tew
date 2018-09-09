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

	addUint16Option("vip-raft-controller-port", utils.VIP_RAFT_CONTROLLER_PORT, "VIP Raft Controller Port", func(value uint16) {
		_config.Config.VIPRaftControllerPort = value
	})

	addUint16Option("vip-raft-worker-port", utils.VIP_RAFT_WORKER_PORT, "VIP Raft Worker Port", func(value uint16) {
		_config.Config.VIPRaftWorkerPort = value
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

	addStringOption("calico-typha-ip", utils.CALICO_TYPHA_IP, "Calico Typha IP", func(value string) {
		_config.Config.CalicoTyphaIP = value
	})

	addStringOption("resolv-conf", utils.RESOLV_CONF, "Custom resolv.conf", func(value string) {
		_config.Config.ResolvConf = value
	})

	addStringOption("public-network", utils.PUBLIC_NETWORK, "Public Network", func(value string) {
		_config.Config.PublicNetwork = value
	})

	addStringOption("cluster-name", utils.CLUSTER_NAME, "Cluster Name used for Kubernetes Dashboard", func(value string) {
		_config.Config.ClusterName = value
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

	addStringOption("version-etcd", utils.VERSION_ETCD, "Etcd version", func(value string) {
		_config.Config.Versions.Etcd = value
	})

	addStringOption("version-k8s", utils.VERSION_K8S, "Kubernetes version", func(value string) {
		_config.Config.Versions.K8S = value
	})

	addStringOption("version-helm", utils.VERSION_HELM, "Helm version", func(value string) {
		_config.Config.Versions.Helm = value
	})

	addStringOption("version-containerd", utils.VERSION_CONTAINERD, "Containerd version", func(value string) {
		_config.Config.Versions.Containerd = value
	})

	addStringOption("version-runc", utils.VERSION_RUNC, "Runc version", func(value string) {
		_config.Config.Versions.Runc = value
	})

	addStringOption("version-crictl", utils.VERSION_CRICTL, "CriCtl version", func(value string) {
		_config.Config.Versions.CriCtl = value
	})

	addStringOption("version-gobetween", utils.VERSION_GOBETWEEN, "Gobetween version", func(value string) {
		_config.Config.Versions.Gobetween = value
	})

	addStringOption("version-ark", utils.VERSION_ARK, "Ark version", func(value string) {
		_config.Config.Versions.Ark = value
	})

	addStringOption("version-minio-server", utils.VERSION_MINIO_SERVER, "Minio server version", func(value string) {
		_config.Config.Versions.MinioServer = value
	})

	addStringOption("version-minio-client", utils.VERSION_MINIO_CLIENT, "Minio client version", func(value string) {
		_config.Config.Versions.MinioClient = value
	})

	addStringOption("version-pause", utils.VERSION_PAUSE, "Pause version", func(value string) {
		_config.Config.Versions.Pause = value
	})

	addStringOption("version-coredns", utils.VERSION_COREDNS, "CoreDNS version", func(value string) {
		_config.Config.Versions.CoreDNS = value
	})

	addStringOption("version-elasticsearch", utils.VERSION_ELASTICSEARCH, "Elasticsearch version", func(value string) {
		_config.Config.Versions.Elasticsearch = value
	})

	addStringOption("version-elasticsearch-cron", utils.VERSION_ELASTICSEARCH_CRON, "Elasticsearch Cron version", func(value string) {
		_config.Config.Versions.ElasticsearchCron = value
	})

	addStringOption("version-elasticsearch-operator", utils.VERSION_ELASTICSEARCH_OPERATOR, "Elasticsearch Operator version", func(value string) {
		_config.Config.Versions.ElasticsearchOperator = value
	})

	addStringOption("version-kibana", utils.VERSION_KIBANA, "Kibana version", func(value string) {
		_config.Config.Versions.Kibana = value
	})

	addStringOption("version-cerebro", utils.VERSION_CEREBRO, "Cerebro version", func(value string) {
		_config.Config.Versions.Cerebro = value
	})

	addStringOption("version-fluent-bit", utils.VERSION_FLUENT_BIT, "Fluent-Bit version", func(value string) {
		_config.Config.Versions.FluentBit = value
	})

	addStringOption("version-calico-typha", utils.VERSION_CALICO_TYPHA, "Calico Typha version", func(value string) {
		_config.Config.Versions.CalicoTypha = value
	})

	addStringOption("version-calico-node", utils.VERSION_CALICO_NODE, "Calico Node version", func(value string) {
		_config.Config.Versions.CalicoNode = value
	})

	addStringOption("version-calico-cni", utils.VERSION_CALICO_CNI, "Calico CNI version", func(value string) {
		_config.Config.Versions.CalicoCNI = value
	})

	addStringOption("version-rbd-provisioner", utils.VERSION_RBD_PROVISIONER, "RBD-Provisioner version", func(value string) {
		_config.Config.Versions.RBDProvisioner = value
	})

	addStringOption("version-ceph", utils.VERSION_CEPH, "Ceph version", func(value string) {
		_config.Config.Versions.Ceph = value
	})

	addStringOption("version-cert-manager", utils.VERSION_CERT_MANAGER, "Cert Manager version", func(value string) {
		_config.Config.Versions.CertManager = value
	})

	addStringOption("version-heapster", utils.VERSION_HEAPSTER, "Heapster version", func(value string) {
		_config.Config.Versions.Heapster = value
	})

	addStringOption("version-addon-resizer", utils.VERSION_ADDON_RESIZER, "Addon-Resizer version", func(value string) {
		_config.Config.Versions.AddonResizer = value
	})

	addStringOption("version-kubernetes-dashboard", utils.VERSION_KUBERNETES_DASHBOARD, "Kubernetes Dashboard version", func(value string) {
		_config.Config.Versions.KubernetesDashboard = value
	})

	addStringOption("version-cert-manager-controller", utils.VERSION_CERT_MANAGER_CONTROLLER, "Cert Manager Controller version", func(value string) {
		_config.Config.Versions.CertManagerController = value
	})

	addStringOption("version-nginx-ingress-controller", utils.VERSION_NGINX_INGRESS_CONTROLLER, "Nginx Ingress Controller version", func(value string) {
		_config.Config.Versions.NginxIngressController = value
	})

	addStringOption("version-nginx-ingress-default-backend", utils.VERSION_NGINX_INGRESS_DEFAULT_BACKEND, "Nginx Ingress Default Backend version", func(value string) {
		_config.Config.Versions.NginxIngressDefaultBackend = value
	})

	addStringOption("version-metrics-server", utils.VERSION_METRICS_SERVER, "Metrics Server version", func(value string) {
		_config.Config.Versions.MetricsServer = value
	})

	addStringOption("version-prometheus-operator", utils.VERSION_PROMETHEUS_OPERATOR, "Prometheus Operator version", func(value string) {
		_config.Config.Versions.PrometheusOperator = value
	})

	addStringOption("version-prometheus-config-reloader", utils.VERSION_PROMETHEUS_CONFIG_RELOADER, "Prometheus Config Reloader version", func(value string) {
		_config.Config.Versions.PrometheusConfigReloader = value
	})

	addStringOption("version-configmap-reload", utils.VERSION_CONFIGMAP_RELOAD, "ConfigMap Reload version", func(value string) {
		_config.Config.Versions.ConfigMapReload = value
	})

	addStringOption("version-kube-state-metrics", utils.VERSION_KUBE_STATE_METRICS, "Kube State Metrics version", func(value string) {
		_config.Config.Versions.KubeStateMetrics = value
	})

	addStringOption("version-grafana", utils.VERSION_GRAFANA, "Grafana version", func(value string) {
		_config.Config.Versions.Grafana = value
	})

	addStringOption("version-grafana-watcher", utils.VERSION_GRAFANA_WATCHER, "Grafana Watcher version", func(value string) {
		_config.Config.Versions.GrafanaWatcher = value
	})

	addStringOption("version-prometheus-node-exporter", utils.VERSION_PROMETHEUS_NODE_EXPORTER, "Prometheus Node Exporter version", func(value string) {
		_config.Config.Versions.PrometheusNodeExporter = value
	})

	addStringOption("version-prometheus-alert-manager", utils.VERSION_PROMETHEUS_ALERT_MANAGER, "Prometheus Alert Manager version", func(value string) {
		_config.Config.Versions.PrometheusAlertManager = value
	})

	addStringOption("version-prometheus", utils.VERSION_PROMETHEUS, "Prometheus version", func(value string) {
		_config.Config.Versions.Prometheus = value
	})

	RootCmd.AddCommand(configureCmd)
}
