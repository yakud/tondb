package liteclient

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	gocmd "github.com/go-cmd/cmd"
)

type CmdClientConfig struct {
	ExecPath    string
	WorkingDir  string
	PubCertPath string
	Host        string
	Port        int
}

type CmdClient struct {
	config *CmdClientConfig
}

func (c *CmdClient) Exec2(command string) ([]string, error) {
	cmdOptions := gocmd.Options{
		Buffered:  false,
		Streaming: true,
	}

	findCmd := gocmd.NewCmdOptions(
		cmdOptions,
		c.config.ExecPath,
		"-a", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		"-p", c.config.PubCertPath,
		"-v", "0",
		"-c", "'"+command+"'",
	)

	// Print STDOUT and STDERR lines streaming from Cmd
	go func() {
		for {
			select {
			case line := <-findCmd.Stdout:
				fmt.Println(line)
			case line := <-findCmd.Stderr:
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	// Run and wait for Cmd to return, discard Status
	finalStatus := <-findCmd.Start()

	// Cmd has finished but wait for goroutine to print all lines
	for len(findCmd.Stdout) > 0 || len(findCmd.Stderr) > 0 {
		time.Sleep(1000 * time.Millisecond)
	}

	return finalStatus.Stdout, finalStatus.Error
}

func (c *CmdClient) Exec(command string) ([]string, error) {

	var out bytes.Buffer
	var stderr bytes.Buffer

	params := []string{
		"-a", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		"-p", c.config.PubCertPath,
		"-v", "0",
		"-c", "'" + command + "'",
	}
	//params = append(params, strings.Fields()...)

	cmd := exec.Command(
		c.config.ExecPath,
		params...,
	)
	//cmd.Dir = c.config.WorkingDir
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	//cmd.SysProcAttr

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}

	fmt.Println("RUN CMD: ", cmd.String())

	if err := cmd.Wait(); err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}

	return []string{out.String()}, nil
}

func NewCmdClient(config *CmdClientConfig) *CmdClient {
	return &CmdClient{
		config: config,
	}
}
