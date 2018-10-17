package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dtylman/scp"
	"golang.org/x/crypto/ssh"
)

func connect(host, user, password string) (*ssh.Client, error) {

	fmt.Printf("Opening tcp to %v\n", host)
	conn, err := net.DialTimeout("tcp", host, time.Second*30)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}
	fmt.Printf("Establishing ssh session %v...\n", host)
	sshconn, chans, reqs, err := ssh.NewClientConn(conn, host, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(sshconn, chans, reqs), nil
}

func doScp(host, user, remotepath string) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	sc, err := connect(host, user, strings.TrimSpace(password))
	if err != nil {
		return err
	}
	start := time.Now()
	n, err := scp.CopyFrom(sc, remotepath, filepath.Base(remotepath))
	if err != nil {
		return err
	}
	fmt.Printf("Copied %v bytes in %v\n", n, time.Now().Sub(start))
	return nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println(os.Args[0] + " will scp a remote file here.")
		fmt.Println("Usage [host:port] [user] [remote_path]")
		return
	}
	err := doScp(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		fmt.Println(err.Error())
	}
}
