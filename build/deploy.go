package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DeployContext represents the context for deploying to a host
type DeployContext struct {
	Host     string
	User     string
	SourceDir string
}

// NewDeployContext creates a new deploy context
func NewDeployContext(host, user string) (*DeployContext, error) {
	sourceDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return &DeployContext{
		Host:     host,
		User:     user,
		SourceDir: sourceDir,
	}, nil
}

// Execute runs a series of shell commands in the deploy context
func (ctx *DeployContext) Execute(commands [][]string) error {
	for _, cmd := range commands {
		if len(cmd) == 0 {
			continue
		}
		
		_, err := ShellCommand(cmd[0], cmd[1:]...)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deploy deploys the configuration to the specified host
func Deploy(host, user string) error {
	ctx, err := NewDeployContext(host, user)
	if err != nil {
		return err
	}

	fmt.Printf("========================================\n")
	fmt.Printf("Deploying to %s...\n", host)
	fmt.Printf("========================================\n")

	// Build rsync exclude arguments from .gitignore
	excludes, err := buildRsyncExcludes(ctx.SourceDir)
	if err != nil {
		return fmt.Errorf("failed to build rsync excludes: %w", err)
	}

	// Copy entire directory structure to remote host
	fmt.Println("Copying directory structure...")
	rsyncArgs := []string{"-avz"}
	rsyncArgs = append(rsyncArgs, excludes...)
	rsyncArgs = append(rsyncArgs, "-e", "ssh -o StrictHostKeyChecking=no")
	rsyncArgs = append(rsyncArgs, ctx.SourceDir+"/")
	rsyncArgs = append(rsyncArgs, fmt.Sprintf("%s@%s:/tmp/nixos-deploy/", user, host))

	_, err = ShellCommand("rsync", rsyncArgs...)
	if err != nil {
		return fmt.Errorf("failed to copy files to %s: %w", host, err)
	}

	// Move files to /etc/nixos/ and set permissions
	fmt.Println("Installing files to /etc/nixos/...")
	sshCmd := fmt.Sprintf("sudo mkdir -p /etc/nixos && sudo rsync -a /tmp/nixos-deploy/ /etc/nixos/ && sudo rm -rf /tmp/nixos-deploy")
	_, err = ShellCommand("ssh", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s@%s", user, host), sshCmd)
	if err != nil {
		return fmt.Errorf("failed to install files on %s: %w", host, err)
	}

	// Execute rebuild and reboot on remote host
	fmt.Println("Executing rebuild and reboot on", host)
	fmt.Println("(SSH connection will be terminated during reboot - this is normal)")

	rebootCmd := "cd /etc/nixos && sudo nixos-rebuild boot && sudo reboot"
	// Use ShellCommand but ignore connection drop errors
	_, err = ShellCommand("ssh", "-o", "StrictHostKeyChecking=no", "-o", "ServerAliveInterval=5", 
		fmt.Sprintf("%s@%s", user, host), rebootCmd)
	if err != nil {
		// Connection drop is expected during reboot, so we just log it
		fmt.Printf("Note: SSH connection terminated (expected during reboot): %v\n", err)
	}

	fmt.Printf("Deployment to %s initiated.\n", host)
	fmt.Println()

	return nil
}

// buildRsyncExcludes reads .gitignore and builds rsync exclude arguments
func buildRsyncExcludes(sourceDir string) ([]string, error) {
	gitignorePath := filepath.Join(sourceDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var excludes []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		excludes = append(excludes, "--exclude="+line)
	}

	return excludes, nil
}
