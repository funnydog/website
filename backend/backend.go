package backend

import (
	"fmt"
	"io"

	"github.com/funnydog/website/config"
	"github.com/jlaffaye/ftp"
)

type ErrUnknownBackend string

func (e ErrUnknownBackend) Error() string {
	return fmt.Sprintf("Unknown backend '%s'", e)
}

type Backend interface {
	Fetch(name string) (io.ReadCloser, error)
	Store(name string, src io.Reader) error
	MakeDir(name string) error
	Delete(name string) error
	Close() error
}

func Create(c *config.Configuration) (Backend, error) {
	switch c.BackendType {
	case "ftp":
		client, err := ftp.Dial(c.Hostname)
		if err != nil {
			return nil, err
		}

		err = client.Login(c.Username, c.Password)
		if err != nil {
			return nil, err
		}

		return &FTPBackend{client}, nil
	default:
		return nil, ErrUnknownBackend(c.BackendType)
	}
}
