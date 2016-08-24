package scp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"errors"
	"strconv"

	"bytes"
	"io/ioutil"
	"strings"
)

const (
	fileMode = "0644"
	buffSize = 1024 * 256
)

//CopyTo copy from local to remote
func CopyTo(sshClient *ssh.Client, local string, remote string) (int64, error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return 0, err
	}
	defer session.Close()
	stderr := &bytes.Buffer{}
	session.Stderr = stderr
	stdout := &bytes.Buffer{}
	session.Stdout = stdout
	writer, err := session.StdinPipe()
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	err = session.Start("scp -t " + filepath.Dir(remote))
	if err != nil {
		return 0, err
	}

	localFile, err := os.Open(local)
	if err != nil {
		return 0, err
	}
	fileInfo, err := localFile.Stat()
	if err != nil {
		return 0, err
	}
	_, err = fmt.Fprintf(writer, "C%s %d %s\n", fileMode, fileInfo.Size(), filepath.Base(remote))
	if err != nil {
		return 0, err
	}
	n, err := copyN(writer, localFile, fileInfo.Size())
	if err != nil {
		return 0, err
	}
	err = ack(writer)
	if err != nil {
		return 0, err
	}

	err = session.Wait()
	log.Debugf("Copied %v bytes out of %v. err: %v stdout:%v. stderr:%v", n, fileInfo.Size(), err, stdout, stderr)
	//NOTE: Process exited with status 1 is not an error, it just how scp work. (waiting for the next control message and we send EOF)
	return n, nil
}

//CopyFrom copy from remote to local
func CopyFrom(sshClient *ssh.Client, remote string, local string) (int64, error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return 0, err
	}
	defer session.Close()
	stderr := &bytes.Buffer{}
	session.Stderr = stderr
	writer, err := session.StdinPipe()
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	reader, err := session.StdoutPipe()
	if err != nil {
		return 0, err
	}
	err = session.Start("scp -f " + remote)
	if err != nil {
		return 0, err
	}
	err = ack(writer)
	if err != nil {
		return 0, err
	}
	opCode := make([]byte, 1)
	_, err = reader.Read(opCode)
	switch opCode[0] {
	case 'C':
	case 0x1, 0x2:
		msg, err := ioutil.ReadAll(reader)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("scp source error: %v. read error: %v", opCode[0], err))
		}
		return 0, errors.New(strings.TrimSpace(string(msg)))
	default:
		return 0, errors.New(fmt.Sprintf("Unsupported opcode: %v", opCode[0]))
	}
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	mode, err := nextWord(scanner)
	if err != nil {
		return 0, err
	}
	size, err := nextNumber(scanner)
	if err != nil {
		return 0, err
	}
	fileName, err := nextWord(scanner)
	if err != nil {
		return 0, err
	}
	log.Debugf("Receiving %v %v %v", mode, size, fileName)

	err = ack(writer)

	if err != nil {
		return 0, err
	}
	outFile, err := os.Create(local)
	if err != nil {
		return 0, err
	}
	defer outFile.Close()
	n, err := copyN(outFile, reader, size)
	if err != nil {
		return 0, err
	}
	err = outFile.Sync()
	if err != nil {
		return 0, err
	}
	err = outFile.Close()
	if err != nil {
		return 0, err
	}
	err = session.Wait()
	log.Debugf("Copied %v bytes out of %v. err: %v stderr:%v", n, size, err, stderr)
	return n, nil
}

func ack(writer io.Writer) error {
	var msg = []byte{0, 0, 10, 13}
	n, err := writer.Write(msg)
	if err != nil {
		return err
	}
	if n < len(msg) {
		return errors.New("Failed to write ack buffer")
	}
	return nil
}

func nextWord(scanner *bufio.Scanner) (string, error) {
	if !scanner.Scan() {
		return "", scanner.Err()
	}
	return scanner.Text(), nil
}

func nextNumber(scanner *bufio.Scanner) (int64, error) {
	tok, err := nextWord(scanner)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(tok, 10, 0)
}

func copyN(writer io.Writer, src io.Reader, size int64) (int64, error) {
	reader := io.LimitReader(src, size)
	var total int64
	for total < size {
		n, err := io.CopyBuffer(writer, reader, make([]byte, buffSize))
		if err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}
