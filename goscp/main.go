package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/dtylman/scp"
	"golang.org/x/crypto/ssh"
)

var Options struct {
	From        string
	Match       string
	To          string
	Host        string
	Port        int
	Username    string
	Password    string
	DialTimeout time.Duration
}

func connect() (*ssh.Client, error) {
	if Options.Host == "" {
		return nil, errors.New("host is empty")
	}
	address := fmt.Sprintf("%v:%v", Options.Host, Options.Port)
	log.Printf("opening tcp to %v", Options.Host)
	conn, err := net.DialTimeout("tcp", address, Options.DialTimeout)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            Options.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(Options.Password)},
	}
	log.Printf("establishing ssh session %v...", address)
	sshConn, sshChan, reqChan, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(sshConn, sshChan, reqChan), nil
}

func goSCP() error {
	walker, err := NewFilesWalker(Options.From, Options.Match)
	if err != nil {
		return err
	}
	err = walker.Walk()
	if err != nil {
		return err
	}
	if len(walker.Matches) == 0 {
		log.Printf("pattern '%v' yields no files to copy", Options.From)
		return nil
	}
	sc, err := connect()
	if err != nil {
		return err
	}
	defer sc.Close()
	start := time.Now()
	total := int64(0)
	for _, path := range walker.Matches {
		remotePath := fmt.Sprintf("%v/%v", Options.To, filepath.Base(path))
		log.Printf("copying %v to %v", path, remotePath)
		n, err := scp.CopyTo(sc, path, remotePath)
		if err != nil {
			log.Printf("error: %v", err)
		}
		total += n
	}
	log.Printf("copied %v bytes in %v", total, time.Since(start))
	return nil
}

func main() {
	log.SetFlags(0)
	flag.StringVar(&Options.From, "from", "", "copy from (path)")
	flag.StringVar(&Options.Match, "match", ".*", "if [from] is a folder, scan recursively and match against this regular expression")
	flag.StringVar(&Options.To, "to", "", "path on target machine")
	flag.StringVar(&Options.Host, "host", "", "host machine")
	flag.IntVar(&Options.Port, "port", 22, "port")
	flag.StringVar(&Options.Username, "username", "", "user name")
	flag.StringVar(&Options.Password, "password", "", "user password")
	flag.DurationVar(&Options.DialTimeout, "dial-timeout", time.Minute/2, "dial timeout")

	flag.Parse()

	if Options.Password == "" {
		Options.Password = os.Getenv("GOSCP_PASSWORD")
	}
	err := goSCP()
	if err != nil {
		log.Printf("error: %v", err)
	}
}
