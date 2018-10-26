package models


// TODO use JSON from server, not ENV
type Configurator interface {
	Configure()
}

type Renderer interface {
	Render(string) (string, error)
}

type Closer interface {
	Close()
}