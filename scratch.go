// Copyright (c) 2025 Vojtech Pavlik <vojtech@suse.com>
//
// Created using AI tools
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerScratchTools registers the file and directory manipulation tools.
func registerScratchTools(mcpServer *server.MCPServer, tmpDir string) {
	// CreateFile
	mcpServer.AddTool(mcp.NewTool("CreateFile",
		mcp.WithDescription("Creates a new file in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file.")),
		mcp.WithString("content", mcp.Required(), mcp.Description("The content of the file."))),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, _ := request.RequireString("path")
			content, _ := request.RequireString("content")
			return createFile(tmpDir, path, content)
		})

	// ReadFile
	mcpServer.AddTool(mcp.NewTool("ReadFile",
		mcp.WithDescription("Reads the content of a file in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file."))),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, _ := request.RequireString("path")
			return readFile(tmpDir, path)
		})

	// DeleteFile
	mcpServer.AddTool(mcp.NewTool("DeleteFile",
		mcp.WithDescription("Deletes a file in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file."))),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, _ := request.RequireString("path")
			return deleteFile(tmpDir, path)
		})

	// ModifyFile
	mcpServer.AddTool(mcp.NewTool("ModifyFile",
		mcp.WithDescription("Modifies a file in the scratch space using a unified diff."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file.")),
		mcp.WithString("patch", mcp.Required(), mcp.Description("The unified diff patch to apply."))),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, _ := request.RequireString("path")
			patch, _ := request.RequireString("patch")
			return modifyFile(tmpDir, path, patch)
		})

	// ListDirectory
	mcpServer.AddTool(mcp.NewTool("ListDirectory",
		mcp.WithDescription("Lists the contents of a directory in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory."))),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, _ := request.RequireString("path")
			return listDirectory(tmpDir, path)
		})

	// CreateDirectory
	mcpServer.AddTool(mcp.NewTool("CreateDirectory",
		mcp.WithDescription("Creates a new directory in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory."))),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, _ := request.RequireString("path")
			return createDirectory(tmpDir, path)
		})

	// RemoveDirectory
	mcpServer.AddTool(mcp.NewTool("RemoveDirectory",
		mcp.WithDescription("Removes an empty directory in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory."))),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path, _ := request.RequireString("path")
			return removeDirectory(tmpDir, path)
		})
}

func resolvePath(base, path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	cleanedPath := filepath.Clean(path)
	if strings.Contains(cleanedPath, "..") {
		return "", fmt.Errorf("path must not contain '..'")
	}
	fullPath := filepath.Join(base, cleanedPath)
	if !strings.HasPrefix(fullPath, base) {
		return "", fmt.Errorf("path escapes the scratch directory")
	}
	return fullPath, nil
}

func createFile(tmpDir, path, content string) (*mcp.CallToolResult, error) {
	fullPath, err := resolvePath(tmpDir, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create file: %v", err)), nil
	}
	return mcp.NewToolResultText("File created successfully."), nil
}

func readFile(tmpDir, path string) (*mcp.CallToolResult, error) {
	fullPath, err := resolvePath(tmpDir, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}
	return mcp.NewToolResultText(string(content)), nil
}

func deleteFile(tmpDir, path string) (*mcp.CallToolResult, error) {
	fullPath, err := resolvePath(tmpDir, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.Remove(fullPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete file: %v", err)), nil
	}
	return mcp.NewToolResultText("File deleted successfully."), nil
}

func modifyFile(tmpDir, path, patch string) (*mcp.CallToolResult, error) {
	fullPath, err := resolvePath(tmpDir, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	original, err := os.Open(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open file: %v", err)), nil
	}
	defer original.Close()
	files, _, err := gitdiff.Parse(strings.NewReader(patch))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse patch: %v", err)), nil
	}
	if len(files) != 1 {
		return mcp.NewToolResultError("patch must contain exactly one file"), nil
	}
	var output bytes.Buffer
	if err := gitdiff.Apply(&output, original, files[0]); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to apply patch: %v", err)), nil
	}
	if err := os.WriteFile(fullPath, output.Bytes(), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write modified file: %v", err)), nil
	}
	return mcp.NewToolResultText("File modified successfully."), nil
}

func listDirectory(tmpDir, path string) (*mcp.CallToolResult, error) {
	fullPath, err := resolvePath(tmpDir, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list directory: %v", err)), nil
	}
	var out strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Fprintf(&out, "%s/\n", entry.Name())
		} else {
			fmt.Fprintf(&out, "%s\n", entry.Name())
		}
	}
	return mcp.NewToolResultText(out.String()), nil
}

func createDirectory(tmpDir, path string) (*mcp.CallToolResult, error) {
	fullPath, err := resolvePath(tmpDir, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create directory: %v", err)), nil
	}
	return mcp.NewToolResultText("Directory created successfully."), nil
}

func removeDirectory(tmpDir, path string) (*mcp.CallToolResult, error) {
	fullPath, err := resolvePath(tmpDir, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.Remove(fullPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to remove directory: %v", err)), nil
	}
	return mcp.NewToolResultText("Directory removed successfully."), nil
}
