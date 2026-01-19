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
	if len(cfg.Specification.Tools) != 1 {
		t.Errorf("expected 1 item, got %d", len(cfg.Specification.Tools))
	}
	if cfg.Specification.Tools[0].Name != "TestTool" {
		t.Errorf("expected tool TestTool, got %s", cfg.Specification.Tools[0].Name)
	}
}

func TestLoadConfig_Tools(t *testing.T) {
	// Create a temporary config file using 'tools' instead of 'contextItems'
	content := `
apiVersion: v1
kind: DynamicContextSource
metadata:
  name: test-mcp
spec:
  tools:
    - name: TestTool
      command: echo test
`
	tmpfile, err := os.CreateTemp("", "config-tools-*.yaml")
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

	if len(cfg.Specification.Tools) != 1 {
		t.Errorf("expected 1 item, got %d", len(cfg.Specification.Tools))
	}
	if cfg.Specification.Tools[0].Name != "TestTool" {
		t.Errorf("expected tool TestTool, got %s", cfg.Specification.Tools[0].Name)
	}
}

func TestLoadConfig_BothItemsAndTools(t *testing.T) {
	content := `
apiVersion: v1
kind: DynamicContextSource
metadata:
  name: test-mcp
spec:
  contextItems:
    - name: Tool1
      command: echo 1
  tools:
    - name: Tool2
      command: echo 2
`
	tmpfile, err := os.CreateTemp("", "config-both-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.Write([]byte(content))
	tmpfile.Close()

	_, err = LoadConfig(tmpfile.Name())
	if err == nil {
		t.Fatal("expected error when both contextItems and tools are present, got nil")
	}
	if !strings.Contains(err.Error(), "both 'contextItems' and 'tools' are defined") {
		t.Errorf("expected specific error message, got: %v", err)
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

	if len(cfg.Specification.Tools) == 0 {
		t.Error("Default config should define at least one tool")
	}

	// Verify that the contentFile was loaded correctly for the overview resource
	overviewResourceFound := false
	for _, resource := range cfg.Specification.Resources {
		if resource.URI == "simple-mcp://system/overview" {
			overviewResourceFound = true
			expectedContent := "This is a detailed overview of the system, loaded from an external file.\n"
			if resource.Content != expectedContent {
				t.Errorf("Expected overview resource content to be '%s', got '%s'", expectedContent, resource.Content)
			}
			break
		}
	}
	if !overviewResourceFound {
		t.Error("Did not find the 'simple-mcp://system/overview' resource in the default config")
	}
}

func TestLoadConfig_UnifiedCore(t *testing.T) {
	filename := "unified-core.yaml"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("%s not found, skipping integration test", filename)
	}

	cfg, err := LoadConfig(filename)
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", filename, err)
	}

	if len(cfg.Specification.Tools) == 0 {
		t.Errorf("%s should define at least one tool", filename)
	}
}
