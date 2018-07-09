package backend

import (
	"io"

	"github.com/funnydog/website/config"
	"github.com/jlaffaye/ftp"
)

type FTPBackend struct {
	client *ftp.ServerConn
}

func (f *FTPBackend) Fetch(name string) (io.ReadCloser, error) {
	return f.client.Retr(name)
}

func (f *FTPBackend) Store(name string, src io.Reader) error {
	return f.client.Stor(name, src)
}

func (f *FTPBackend) Delete(name string) error {
	return f.client.Delete(name)
}

func (f *FTPBackend) MakeDir(name string) error {
	return f.client.MakeDir(name)
}

func (f *FTPBackend) RemoveDir(name string) error {
	return f.client.RemoveDirRecur(name)
}

func (f *FTPBackend) Close() error {
	return f.client.Quit()
}

func newFTPBackend(c *config.Configuration) (Backend, error) {
	client, err := ftp.Dial(c.Hostname)
	if err != nil {
		return nil, err
	}

	err = client.Login(c.Username, c.Password)
	if err != nil {
		return nil, err
	}

	return &FTPBackend{client}, nil
}
