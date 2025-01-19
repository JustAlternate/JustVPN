package terraform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"bytes"
	"strings"
	"golang.org/x/crypto/ssh"
)

type TerraformService struct {
}
func NewTerraformService() *TerraformService {
	return &TerraformService{
	}
}

func (ts *TerraformService) Apply(endpoint string, region string) error {
	args := []string{
			"apply",
			"-var=endpoint=" + endpoint,
			"-var=region="+ region,
			"-var-file=secrets.tfvars",
			"--auto-approve",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "terraform", args...)

	absDir, err := filepath.Abs(".")
	if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %s", err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
			return fmt.Errorf("specified terraform directory does not exist: %s", absDir)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
			return fmt.Errorf("Error running terraform apply: %v", err)
	}
	return nil
}

func (ts *TerraformService) GetOutput() (string, error) {
    args := []string{
        "output",
        "instance_ip",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
    defer cancel()

    cmd := exec.CommandContext(ctx, "terraform", args...)
    cmd.Dir = "."

    outputBytes, err := cmd.Output()
    if err != nil {
        if exitError, ok := err.(*exec.ExitError); ok {
            stderr := string(exitError.Stderr)
            return "", fmt.Errorf("error running terraform output: %v, stderr: %s", err, stderr)
        }
        return "", fmt.Errorf("error running terraform output: %w", err)
    }

    output := string(outputBytes)
		output = strings.ReplaceAll(output, " ", "")
		return output[1 : len(output)-2], nil
}

func (ts *TerraformService) GetPubKey(hostIp string) (string, error) {
	key, err := os.ReadFile("/home/justalternate/.ssh/id_ed25519")
	if err != nil {
		return "", fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", hostIp+":22", config)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to SSH server: %v\n", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("Failed to create session: %v\n", err)
	}
	defer session.Close()

	var output bytes.Buffer
	session.Stdout = &output

	err = session.Run("cat wg-public.key")
	if err != nil {
		return "", fmt.Errorf("Failed to run command: %v\n", err)
	}
	return output.String(), nil
}

func (ts *TerraformService) Destroy(endpoint string, timeBeforeDeletion int) error {

	time.Sleep(time.Second * time.Duration(timeBeforeDeletion))

	args := []string{
			"destroy",
			"-var=endpoint=" + endpoint,
			"-var-file=secrets.tfvars",
			"--auto-approve",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "terraform", args...)

	absDir, err := filepath.Abs(".")
	if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %s", err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
			return fmt.Errorf("specified terraform directory does not exist: %s", absDir)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
			return fmt.Errorf("Error running terraform destroy: %v", err)
	}
	return nil

}
