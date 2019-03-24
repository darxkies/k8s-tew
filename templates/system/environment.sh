export PATH={{.K8STEWPath}}:{{.K8SPath}}:{{.EtcdPath}}:{{.CRIPath}}:{{.CNIPath}}:{{.VeleroPath}}:{{.HostPath}}:{{.CurrentPath}}
export KUBECONFIG={{.KubeConfig}}
export CONTAINER_RUNTIME_ENDPOINT=unix://{{.ContainerdSock}}
export CONTAINERD_NAMESPACE=k8s.io
export ETCDCTL_API=3
