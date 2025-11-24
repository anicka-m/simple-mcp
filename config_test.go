package main

import (
	"os"
	"strings"
	"testing"
)

func TestLoadConfig_Valid(t *testing.T) {
	// Create a temporary config file
	content := `
apiVersion: v1
kind: DynamicContextSource
metadata:
  name: test-mcp
spec:
  contextItems:
    - name: TestTool
      command: echo test
      parameters: ["arg1"]
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Metadata.Name != "test-mcp" {
		t.Errorf("expected name test-mcp, got %s", cfg.Metadata.Name)
	}
	if len(cfg.Specification.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(cfg.Specification.Items))
	}
	if cfg.Specification.Items[0].Name != "TestTool" {
		t.Errorf("expected tool TestTool, got %s", cfg.Specification.Items[0].Name)
	}
}

func TestLoadConfig_InvalidYaml(t *testing.T) {
	content := `
apiVersion: v1
metadata:
  name: broken
  - indentation_error: yes
`
	tmpfile, err := os.CreateTemp("", "bad-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.Write([]byte(content))
	tmpfile.Close()

	_, err = LoadConfig(tmpfile.Name())
	if err == nil {
		t.Fatal("expected YAML parse error, got nil")
	}

	// Verify our custom error formatter is working (checking for line number info)
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected error message to contain line number info, got: %s", err.Error())
	}
}

// TestLoadConfig_DefaultFile validates the actual simple-mcp.yaml shipped with the repo.
// This ensures that the default configuration is always valid and parsable.
func TestLoadConfig_DefaultFile(t *testing.T) {
	filename := "simple-mcp.yaml"
	
	// Skip if the file is not found (e.g. running tests in isolation/different dir)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found, skipping integration test", filename)
	}

	cfg, err := LoadConfig(filename)
	if err != nil {
		t.Fatalf("Failed to parse default config %s: %v", filename, err)
	}

	// Basic validation of the default config content
	expectedName := "dynamic-mcp-context"
	if cfg.Metadata.Name != expectedName {
		t.Errorf("Default config name mismatch. Expected '%s', got '%s'", expectedName, cfg.Metadata.Name)
	}

	if len(cfg.Specification.Resources) == 0 {
		t.Error("Default config should define at least one resource")
	}

	if len(cfg.Specification.Items) == 0 {
		t.Error("Default config should define at least one tool")
	}
}
