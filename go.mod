module github.com/darxkies/k8s-tew

go 1.15

replace github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5
	github.com/Microsoft/hcsshim v0.8.7
	github.com/briandowns/spinner v0.0.0-20181029155426-195c31b675a7
	github.com/cavaliercoder/grab v2.0.0+incompatible
	github.com/cespare/reflex v0.3.0 // indirect
	github.com/containerd/aufs v0.0.0-20200106064538-76944a95669d
	github.com/containerd/btrfs v0.0.0-20200117014249-153935315f4a
	github.com/containerd/cgroups v0.0.0-20200407151229-7fc7a507c04c
	github.com/containerd/console v1.0.0
	github.com/containerd/containerd v1.3.4
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb
	github.com/containerd/cri v1.11.1
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b
	github.com/containerd/go-cni v0.0.0-20200107172653-c154a49e2c75 // indirect
	github.com/containerd/go-runc v0.0.0-20200220073739-7016d3ce2328
	github.com/containerd/ttrpc v1.0.0
	github.com/containerd/typeurl v1.0.0
	github.com/containerd/zfs v0.0.0-20200115132605-fdbd9435326f
	github.com/containernetworking/plugins v0.8.5 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c
	github.com/docker/go-metrics v0.0.1
	github.com/docker/go-units v0.4.0
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gobuffalo/genny v0.0.0-20190315121735-8b38fb089e88 // indirect
	github.com/gobuffalo/gogen v0.0.0-20190315121717-8f38393713f5 // indirect
	github.com/gobuffalo/mapi v1.0.1 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/gobuffalo/packr/v2 v2.8.1 // indirect
	github.com/gobuffalo/syncx v0.0.0-20190224160051-33c29581e754 // indirect
	github.com/gogo/googleapis v1.3.2
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-cmp v0.4.0
	github.com/google/uuid v1.1.1
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/imdario/mergo v0.3.6
	github.com/jsonnet-bundler/jsonnet-bundler v0.4.0 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc9
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/opencontainers/selinux v1.5.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/seccomp/libseccomp-golang v0.9.1 // indirect
	github.com/sethvargo/go-password v0.1.3
	github.com/sirupsen/logrus v1.7.0
	github.com/smallnest/goreq v0.0.0-20180727030113-2e3372c80388
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2
	github.com/tchap/go-patricia v2.3.0+incompatible // indirect
	github.com/tmc/scp v0.0.0-20170824174625-f7b48647feef
	github.com/urfave/cli v1.22.4
	github.com/wille/osutil v0.0.0-20190526221756-e91b8656e290
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68
	google.golang.org/grpc v1.27.0
	gopkg.in/ini.v1 v1.58.0
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/apiserver v0.19.3
	k8s.io/cli-runtime v0.19.3
	k8s.io/client-go v0.19.3
	k8s.io/cri-api v0.19.3
	k8s.io/kubectl v0.19.3
)
