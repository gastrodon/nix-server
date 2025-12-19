package build

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ShellCommand wraps external process spawning and returns output as lines
func ShellCommand(command string, args ...string) ([]string, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	return lines, nil
}

// ShellScript executes a shell script and returns output as lines
func ShellScript(script string) ([]string, error) {
	cmd := exec.Command("bash", "-c", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("script failed: %w\nstderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	return lines, nil
}

// ShellInteractive runs a command interactively (inheriting stdin/stdout/stderr)
func ShellInteractive(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Run()
}
