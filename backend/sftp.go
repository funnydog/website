package backend

import (
	"io"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/funnydog/website/config"
)

type SFTPBackend struct {
	conn   *ssh.Client
	client *sftp.Client
}

func (s *SFTPBackend) Fetch(name string) (io.ReadCloser, error) {
	return s.client.Open(name)
}

func (s *SFTPBackend) Store(name string, src io.Reader) error {
	dst, err := s.client.Create(name)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	return err
}

func (s *SFTPBackend) Delete(name string) error {
	return s.client.Remove(name)
}

func (s *SFTPBackend) MakeDir(name string) error {
	return s.client.MkdirAll(name)
}

func (s *SFTPBackend) RemoveDir(name string) error {
	return s.client.RemoveDirectory(name)
}

func (s *SFTPBackend) Close() error {
	err := s.client.Close()
	s.conn.Close()
	return err
}

func newSFTPBackend(c *config.Configuration) (Backend, error) {
	var auths []ssh.AuthMethod

	aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}

	if c.Password != "" {
		auths = append(auths, ssh.Password(c.Password))
	}

	config := ssh.ClientConfig{
		User:            c.Username,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", c.Hostname, &config)
	if err != nil {
		return nil, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &SFTPBackend{
		conn,
		client,
	}, nil
}
