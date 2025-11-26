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
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerScratchTools registers the file and directory manipulation tools.
func registerScratchTools(mcpServer *server.MCPServer, tmpDir string, verbose bool) {
	createFileTool := mcp.NewTool("CreateFile",
		mcp.WithDescription("Creates a new file in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file.")),
		mcp.WithString("content", mcp.Required(), mcp.Description("The content of the file.")))
	mcpServer.AddTool(createFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		content, _ := request.RequireString("content")
		if verbose {
			log.Printf("Handling CreateFile request for path: %s", path)
		}
		return createFile(tmpDir, path, content)
	})
	log.Printf("Registered built-in scratch tool: %s", createFileTool.Name)

	readFileTool := mcp.NewTool("ReadFile",
		mcp.WithDescription("Reads the content of a file in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file.")))
	mcpServer.AddTool(readFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		if verbose {
			log.Printf("Handling ReadFile request for path: %s", path)
		}
		return readFile(tmpDir, path)
	})
	log.Printf("Registered built-in scratch tool: %s", readFileTool.Name)

	deleteFileTool := mcp.NewTool("DeleteFile",
		mcp.WithDescription("Deletes a file in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file.")))
	mcpServer.AddTool(deleteFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		if verbose {
			log.Printf("Handling DeleteFile request for path: %s", path)
		}
		return deleteFile(tmpDir, path)
	})
	log.Printf("Registered built-in scratch tool: %s", deleteFileTool.Name)

	modifyFileTool := mcp.NewTool("ModifyFile",
		mcp.WithDescription("Modifies a file in the scratch space using a unified diff."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file.")),
		mcp.WithString("patch", mcp.Required(), mcp.Description("The unified diff patch to apply.")))
	mcpServer.AddTool(modifyFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		patch, _ := request.RequireString("patch")
		if verbose {
			log.Printf("Handling ModifyFile request for path: %s", path)
		}
		return modifyFile(tmpDir, path, patch)
	})
	log.Printf("Registered built-in scratch tool: %s", modifyFileTool.Name)

	listDirectoryTool := mcp.NewTool("ListDirectory",
		mcp.WithDescription("Lists the contents of a directory in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory.")))
	mcpServer.AddTool(listDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		if verbose {
			log.Printf("Handling ListDirectory request for path: %s", path)
		}
		return listDirectory(tmpDir, path)
	})
	log.Printf("Registered built-in scratch tool: %s", listDirectoryTool.Name)

	createDirectoryTool := mcp.NewTool("CreateDirectory",
		mcp.WithDescription("Creates a new directory in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory.")))
	mcpServer.AddTool(createDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		if verbose {
			log.Printf("Handling CreateDirectory request for path: %s", path)
		}
		return createDirectory(tmpDir, path)
	})
	log.Printf("Registered built-in scratch tool: %s", createDirectoryTool.Name)

	removeDirectoryTool := mcp.NewTool("RemoveDirectory",
		mcp.WithDescription("Removes an empty directory in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory.")))
	mcpServer.AddTool(removeDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		if verbose {
			log.Printf("Handling RemoveDirectory request for path: %s", path)
		}
		return removeDirectory(tmpDir, path)
	})
	log.Printf("Registered built-in scratch tool: %s", removeDirectoryTool.Name)
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
