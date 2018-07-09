package backend

import (
	"fmt"
	"io"

	"github.com/funnydog/website/config"
)

type ErrUnknownBackend string

func (e ErrUnknownBackend) Error() string {
	return fmt.Sprintf("Unknown backend '%s'", e)
}

type Backend interface {
	Fetch(name string) (io.ReadCloser, error)
	Store(name string, src io.Reader) error
	Delete(name string) error
	MakeDir(name string) error
	RemoveDir(name string) error
	Close() error
}

func Create(c *config.Configuration) (Backend, error) {
	switch c.BackendType {
	case "ftp":
		return newFTPBackend(c)
	default:
		return nil, ErrUnknownBackend(c.BackendType)
	}
}
