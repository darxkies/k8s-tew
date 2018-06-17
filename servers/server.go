package servers

type Server interface {
	Start() error
	Stop()
	Name() string
}
