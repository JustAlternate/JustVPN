package terraform

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"bytes"
	"strings"
	"golang.org/x/crypto/ssh"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type TerraformService struct {
	execPath string
	workingDir string
	iacDirPath string
}

func NewTerraformService() *TerraformService {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.9.8")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}

	// Get working directory from environment variable or use default
	workingDir := os.Getenv("TERRAFORM_WORKING_DIR")
	if workingDir == "" {
		workingDir = "."
	}

	// Get IAC directory path from environment variable or use default
	iacDirPath := os.Getenv("IAC_DIR_PATH")
	if iacDirPath == "" {
		iacDirPath = "../iac"
	}

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("failed to create Terraform object: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running terraform init: %s", err)
	}

	return &TerraformService{
		execPath: execPath,
		workingDir: workingDir,
		iacDirPath: iacDirPath,
	}
}

func (ts *TerraformService) Apply(endpoint string, region string) error {
	tf, err := tfexec.NewTerraform(ts.workingDir, ts.execPath)
	if err != nil {
		return fmt.Errorf("failed to create Terraform object: %w", err)
	}

	varFilePath := fmt.Sprintf("%s/secrets.tfvars", ts.iacDirPath)

	applyOptions := []tfexec.ApplyOption{
		tfexec.Var(fmt.Sprintf("endpoint=%s", endpoint)),
		tfexec.Var(fmt.Sprintf("region=%s", region)),
		tfexec.VarFile(varFilePath),
	}

	err = tf.Apply(context.Background(), applyOptions...)
	if err != nil {
		return fmt.Errorf("error running terraform apply: %w", err)
	}

	return nil
}

func (ts *TerraformService) GetOutput() (string, error) {
	tf, err := tfexec.NewTerraform(ts.workingDir, ts.execPath)
	if err != nil {
		return "", fmt.Errorf("failed to create Terraform object: %w", err)
	}

	outputs, err := tf.Output(context.Background())
	if err != nil {
		return "", fmt.Errorf("error running terraform output: %w", err)
	}

	output_ip, exists := outputs["instance_ip"]
	if !exists {
		return "", fmt.Errorf("output 'instance_ip' not found in the output")
	}

	output := strings.TrimSpace(strings.ReplaceAll(string(output_ip.Value), " ", ""))
	return output[1 : len(output)-1], nil
}

func (ts *TerraformService) GetPubKey(hostIp string) (string, error) {
	password := os.Getenv("SSH_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("No env var given for ssh password")
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
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

	tf, err := tfexec.NewTerraform(ts.workingDir, ts.execPath)
	if err != nil {
		return fmt.Errorf("failed to create Terraform object: %w", err)
	}

	varFilePath := fmt.Sprintf("%s/secrets.tfvars", ts.iacDirPath)

	destroyOptions := []tfexec.DestroyOption{
		tfexec.Var(fmt.Sprintf("endpoint=%s", endpoint)),
		tfexec.VarFile(varFilePath),
	}

	err = tf.Destroy(context.Background(), destroyOptions...)
	if err != nil {
		return fmt.Errorf("error running terraform destroy: %w", err)
	}

	return nil
}
