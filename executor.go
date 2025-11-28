// Copyright (c) 2025 Vojtech Pavlik <vojtech@suse.com>
//
// Created using AI tools
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Package main provides the command execution logic for the server.
// It handles secure template rendering of commands and their execution with
// strictly enforced timeouts.
package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"text/template"
	"time"
)

// executeCommand renders the command template with the provided parameters
// and executes it in a shell. It returns the combined stdout/stderr,
// the exit code, and any Go-level error that occurred.
func executeCommand(item ContextItem, params map[string]interface{}, workDir string) (string, int, time.Duration, error) {
	startTime := time.Now()
	tmpl, err := template.New("command").Parse(item.Command)
	if err != nil {
		return "", -1, 0, fmt.Errorf("invalid command template in config: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", -1, 0, fmt.Errorf("failed to build command from template: %w", err)
	}
	finalCommand := buf.String()

	const defaultTimeout = 30
	timeout := item.TimeoutSeconds
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", finalCommand)

	// Set the working directory for the command.
	if workDir != "" {
		cmd.Dir = workDir
	} else {
		cmd.Dir = "/tmp"
	}

	output, err := cmd.CombinedOutput()

	// Default exit code to 0 on success, -1 for Go-level errors (e.g., timeout).
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1 // Indicates a non-execution error (e.g., context deadline).
		}
	}

	duration := time.Since(startTime)

	if ctx.Err() == context.DeadlineExceeded {
		return "", -1, duration, fmt.Errorf("command timed out after %d seconds", timeout)
	}

	if err != nil {
		// Return the output (likely stderr) along with the error to aid debugging.
		return string(output), exitCode, duration, fmt.Errorf("command failed: %w", err)
	}

	return string(output), exitCode, duration, nil
}
