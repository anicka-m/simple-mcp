package main

import (
	"strings"
	"testing"
	"time"
)

func TestExecuteCommand_Templating(t *testing.T) {
	item := ContextItem{
		Name:    "Greet",
		Command: "echo Hello {{.name}}",
	}
	params := map[string]interface{}{
		"name": "World",
	}

	output, err := executeCommand(item, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(output) != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", output)
	}
}

func TestExecuteCommand_Timeout(t *testing.T) {
	// This command sleeps for 2 seconds, but we set timeout to 1 second
	item := ContextItem{
		Name:           "Sleepy",
		Command:        "sleep 2",
		TimeoutSeconds: 1,
	}

	start := time.Now()
	_, err := executeCommand(item, nil)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// We verify that it failed reasonably close to the timeout
	if duration.Seconds() > 1.5 {
		t.Errorf("test took too long (%v), timeout logic might be broken", duration)
	}
}

func TestExecuteCommand_InvalidTemplate(t *testing.T) {
	item := ContextItem{
		Command: "echo {{.missing_end_brace",
	}
	_, err := executeCommand(item, nil)
	if err == nil {
		t.Error("expected template parse error, got nil")
	}
}
