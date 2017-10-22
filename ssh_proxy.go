// Package sshproxy provides an ssh proxy dialer
package sshproxy

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/user"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/kevinburke/ssh_config"
)

// SSHProxy is a dialer for tunneling tcp connections through an ssh connection
type SSHProxy struct {
	Addr string
	rc   net.Conn
	from string
	cfg  *ssh.ClientConfig
}

// NewSSHProxy returns a new SSHProxy initialized from the default ssh config
func NewSSHProxy(addr string) *SSHProxy {
	return &SSHProxy{
		from: net.JoinHostPort(addr, ssh_config.Get(addr, "Port")),
		cfg: &ssh.ClientConfig{
			Config:          ssh.Config{},
			User:            ssh_config.Get(addr, "User"),
			Auth:            authMethods(addr),
			HostKeyCallback: HostKeyAcceptAll,
			Timeout:         getTimeout(addr),
		},
	}
}

// Dial  dials a connection to addr through the proxy. The function prototype
// is usable with grpc.WithDialer (https://godoc.org/google.golang.org/grpc#WithDialer)
func (sp *SSHProxy) Dial(addr string, retry time.Duration) (net.Conn, error) {
	client, err := ssh.Dial("tcp", sp.from, sp.cfg)
	if err != nil {
		return nil, err
	}

	sp.rc, err = client.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return sp.rc, nil
}

// Close closes the previous connection
func (sp *SSHProxy) Close() {
	sp.rc.Close()
}

// getList returns the given ssh config key containing a comma separated list
// as a []string
func getList(host, key string) []string {
	s := ssh_config.Get(host, key)
	return strings.Split(s, ",")
}

// authMethods builds a list of ssh.AuthMethod from the configuration for the
// specified host.
func authMethods(host string) []ssh.AuthMethod {
	am := []ssh.AuthMethod{}

	if ssh_config.Get(host, "PubkeyAuthentication") == "yes" {
		var signers []ssh.Signer
		// Process IdentityFile and add ssh keyss
		keyFiles := GetList(host, "IdentityFile")

		for _, keyFile := range keyFiles {
			// open, read, append
			key, err := ioutil.ReadFile(expandHomeDir(keyFile))
			if err != nil {
				continue
			}
			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				continue
			}
			signers = append(signers, signer)
		}
		am = append(am, ssh.PublicKeys(signers...))
	}

	// TODO: PasswordAuthentication
	// TODO: KbdInteractiveAuthentication

	return am
}

func HostKeyAcceptAll(_ string, _ net.Addr, _ ssh.PublicKey) error { return nil }
func HostKeyDenyAll(_ string, _ net.Addr, _ ssh.PublicKey) error   { return fmt.Errorf("denied") }

func getTimeout(host string) time.Duration {
	s := ssh_config.Get(host, "ConnectTimeout")
	d, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		d = 0
	}
	return time.Duration(d) * time.Second
}

func expandHomeDir(in string) string {
	// TODO: Support paths with other users' home dir
	usr, err := user.Current()
	if err != nil {
		return in
	}
	return strings.Replace(in, "~", usr.HomeDir, -1)
}
