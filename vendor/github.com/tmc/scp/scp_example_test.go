package scp_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/tmc/scp"
)

func getAgent() (agent.Agent, error) {
	agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	return agent.NewClient(agentConn), err
}

func ExampleCopyPath() {
	f, _ := ioutil.TempFile("", "")
	fmt.Fprintln(f, "hello world")
	f.Close()
	defer os.Remove(f.Name())
	defer os.Remove(f.Name() + "-copy")

	agent, err := getAgent()
	if err != nil {
		log.Fatalln("Failed to connect to SSH_AUTH_SOCK:", err)
	}

	client, err := ssh.Dial("tcp", "127.0.0.1:22", &ssh.ClientConfig{
		User: os.Getenv("USER"),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agent.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // FIXME: please be more secure in checking host keys
	})
	if err != nil {
		log.Fatalln("Failed to dial:", err)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Fatalln("Failed to create session: " + err.Error())
	}

	dest := f.Name() + "-copy"
	err = scp.CopyPath(f.Name(), dest, session)
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		fmt.Printf("no such file or directory: %s", dest)
	} else {
		fmt.Println("success")
	}
	// output:
	// success
}
