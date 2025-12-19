package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Rebuild rebuilds the NixOS VM
func Rebuild() error {
	fmt.Println("Removing previous builds...")

	// Remove result/* files
	resultPath := "result"
	if entries, err := os.ReadDir(resultPath); err == nil && len(entries) > 0 {
		for _, entry := range entries {
			path := filepath.Join(resultPath, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", path, err)
			}
		}
	}

	// Remove *.qcow2 files
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".qcow2") {
			path := filepath.Join(cwd, entry.Name())
			if err := os.Remove(path); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", path, err)
			}
		}
	}

	fmt.Println("Rebuilding NixOS VM...")
	cmd := exec.Command("sudo", "nix-build", "<nixpkgs/nixos>", "-A", "vm",
		"-I", "nixpkgs=channel:nixos-25.05",
		"-I", "nixos-config=./configuration.nix")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nix-build failed: %w", err)
	}

	return nil
}

// Boot starts the QEMU VM
func Boot() error {
	fmt.Println("Starting QEMU VM...")

	cmd := exec.Command("./result/bin/run-nixos-vm", "-nographic")
	cmd.Env = append(os.Environ(), "QEMU_KERNEL_PARAMS=console=ttyS0")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	
	// Reset terminal after VM exits
	resetCmd := exec.Command("reset")
	resetCmd.Stdin = os.Stdin
	resetCmd.Stdout = os.Stdout
	resetCmd.Stderr = os.Stderr
	_ = resetCmd.Run()

	if err != nil {
		return fmt.Errorf("VM execution failed: %w", err)
	}

	return nil
}
