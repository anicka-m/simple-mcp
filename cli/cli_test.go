package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestListTools(t *testing.T) {
	// Get the project root directory
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	rootDir := filepath.Dir(dir)

	// Build the server and client from the project root
	buildCmd := exec.Command("make")
	buildCmd.Dir = rootDir
	err = buildCmd.Run()
	if err != nil {
		t.Fatalf("Failed to build server and client: %v", err)
	}

	// Start the server from the project root
	serverCmd := exec.Command("./simple-mcp")
	serverCmd.Dir = rootDir
	err = serverCmd.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer serverCmd.Process.Kill()

	// Give the server a moment to start up
	time.Sleep(3 * time.Second)

	// Run the client from the project root and capture the output
	clientCmd := exec.Command("./simple-mcp-cli", "list-tools")
	clientCmd.Dir = rootDir
	var out bytes.Buffer
	clientCmd.Stdout = &out
	err = clientCmd.Run()
	if err != nil {
		t.Fatalf("Client command failed: %v", err)
	}

	// Verify that the client output contains the expected tool names
	expectedTools := []string{"ListPendingTasks", "TaskStatus", "ListResources", "GetResource"}
	output := out.String()
	for _, tool := range expectedTools {
		if !strings.Contains(output, tool) {
			t.Errorf("Expected to find tool '%s' in the client output, but it was not found.", tool)
		}
	}
}
