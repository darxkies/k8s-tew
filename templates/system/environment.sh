export PATH={{.K8STEWPath}}:{{.K8SPath}}:{{.EtcdPath}}:{{.CRIPath}}:{{.CNIPath}}:{{.ArkPath}}:{{.CurrentPath}}
export KUBECONFIG={{.KubeConfig}}
export CONTAINER_RUNTIME_ENDPOINT=unix://{{.ContainerdSock}}
