package manifest

type Config struct {
	ArgsEscaped  bool
	AttachStdin  bool
	AttachStderr bool
	AttachStdout bool
	Tty          bool
	OpenStdin    bool
	StdinOnce    bool
	Hostname     string
	Domainname   string
	Image        string
	Cmd          []string
	Env          []string
	User         string
	ExposedPorts map[string]struct{}
	Volumes      map[string]struct{}
	WorkingDir   string
	Entrypoint   []string
	OnBuild      []string
	Labels       map[string]string
}
