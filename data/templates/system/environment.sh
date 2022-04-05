export PATH="{{.K8STEWPath}}":"{{.K8SPath}}":"{{.EtcdPath}}":"{{.CRIPath}}":"{{.CNIPath}}":"{{.VeleroPath}}":"{{.HostPath}}":{{.CurrentPath}}
{{- if .KubeConfig }}
export KUBECONFIG="{{.KubeConfig}}"
{{- end }}
export VELERO_NAMESPACE=backup
export CONTAINER_RUNTIME_ENDPOINT=unix://{{.ContainerdSock}}
export CONTAINERD_NAMESPACE=k8s.io
export ETCDCTL_API=3
