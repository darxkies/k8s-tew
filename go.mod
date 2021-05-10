module github.com/darxkies/k8s-tew

go 1.15

replace github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible

require (
	github.com/briandowns/spinner v0.0.0-20181029155426-195c31b675a7
	github.com/cavaliercoder/grab v2.0.0+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gobuffalo/packd v1.0.0 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.6 // indirect
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/onsi/gomega v1.7.1 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sethvargo/go-password v0.1.3
	github.com/sirupsen/logrus v1.8.1
	github.com/smallnest/goreq v0.0.0-20180727030113-2e3372c80388
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/tmc/scp v0.0.0-20170824174625-f7b48647feef
	github.com/wille/osutil v0.0.0-20190526221756-e91b8656e290
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	google.golang.org/grpc v1.27.0
	gopkg.in/ini.v1 v1.58.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/cli-runtime v0.19.3
	k8s.io/client-go v0.19.3
	k8s.io/cri-api v0.19.3
	k8s.io/kubectl v0.19.3
)
