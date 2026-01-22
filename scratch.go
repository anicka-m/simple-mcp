// Copyright (c) 2025 Vojtech Pavlik <vojtech@suse.com>
//
// Created using AI tools
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"regexp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerScratchTools registers the file and directory manipulation tools.
func registerScratchTools(mcpServer *server.MCPServer, resourceMap map[string]ResourceItem, tmpDir string, verbose bool) {
	createFileTool := mcp.NewTool("CreateFile",
		mcp.WithDescription("Creates a new file in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file within the scratch space.")),
		mcp.WithString("content", mcp.Required(), mcp.Description("The content of the file. Do not forget to include a newline character on the last line of a text file.")))
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
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file within the scratch space.")))
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
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file within the scratch space.")))
	mcpServer.AddTool(deleteFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		if verbose {
			log.Printf("Handling DeleteFile request for path: %s", path)
		}
		return deleteFile(tmpDir, path)
	})
	log.Printf("Registered built-in scratch tool: %s", deleteFileTool.Name)

	replaceInFileTool := mcp.NewTool("ReplaceInFile",
		mcp.WithDescription("Replaces a pattern in a file in the scratch space using a regular expression."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the file within the scratch space.")),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("The regular expression pattern to search for.")),
		mcp.WithString("replacement", mcp.Required(), mcp.Description("The replacement string. Supports capture groups (e.g., $1).")),
		mcp.WithBoolean("replaceAll", mcp.Description("If true, replace all occurrences. If false (default), replace only the first occurrence.")))
	mcpServer.AddTool(replaceInFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		pattern, _ := request.RequireString("pattern")
		replacement, _ := request.RequireString("replacement")
		replaceAll := request.GetBool("replaceAll", false)
		if verbose {
			log.Printf("Handling ReplaceInFile request for path: %s", path)
		}
		return replaceInFile(tmpDir, path, pattern, replacement, replaceAll)
	})
	log.Printf("Registered built-in scratch tool: %s", replaceInFileTool.Name)

	listDirectoryTool := mcp.NewTool("ListDirectory",
		mcp.WithDescription("Lists the contents of a directory in the scratch space."),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory within the scratch space. Absolute paths are not allowed.")))
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
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory within the scratch space.")))
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
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the directory within the scratch space.")))
	mcpServer.AddTool(removeDirectoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, _ := request.RequireString("path")
		if verbose {
			log.Printf("Handling RemoveDirectory request for path: %s", path)
		}
		return removeDirectory(tmpDir, path)
	})
	log.Printf("Registered built-in scratch tool: %s", removeDirectoryTool.Name)

	copyResourceToFileTool := mcp.NewTool("CopyResourceToFile",
		mcp.WithDescription("Copies the content of a resource to a file in the scratch space."),
		mcp.WithString("resourceURI", mcp.Required(), mcp.Description("The URI of the resource to copy.")),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path to the destination file within the scratch space.")))
	mcpServer.AddTool(copyResourceToFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resourceURI, _ := request.RequireString("resourceURI")
		path, _ := request.RequireString("path")
		if verbose {
			log.Printf("Handling CopyResourceToFile request for resourceURI: %s, path: %s", resourceURI, path)
		}
		return copyResourceToFile(resourceMap, tmpDir, verbose, resourceURI, path)
	})
	log.Printf("Registered built-in scratch tool: %s", copyResourceToFileTool.Name)

	copyResourceTreeTool := mcp.NewTool("CopyResourceTree",
		mcp.WithDescription("Recursively copies all resources whose URIs start with a given prefix into a directory in the scratch space."),
		mcp.WithString("resourcePrefix", mcp.Required(), mcp.Description("The prefix of the resource URIs to copy.")),
		mcp.WithString("destinationPath", mcp.Required(), mcp.Description("The destination directory path within the scratch space.")))
	mcpServer.AddTool(copyResourceTreeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resourcePrefix, _ := request.RequireString("resourcePrefix")
		destinationPath, _ := request.RequireString("destinationPath")
		if verbose {
			log.Printf("Handling CopyResourceTree request for resourcePrefix: %s, destinationPath: %s", resourcePrefix, destinationPath)
		}
		return copyResourceTree(resourceMap, tmpDir, verbose, resourcePrefix, destinationPath)
	})
	log.Printf("Registered built-in scratch tool: %s", copyResourceTreeTool.Name)
}

func copyResourceToFile(resourceMap map[string]ResourceItem, tmpDir string, verbose bool, resourceURI, path string) (*mcp.CallToolResult, error) {
	item, ok := resourceMap[resourceURI]
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("resource not found: %s", resourceURI)), nil
	}

	content, err := getResourceContent(item, tmpDir, verbose)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get resource content for %s: %v", resourceURI, err)), nil
	}

	if content == "" {
		return mcp.NewToolResultError(fmt.Sprintf("resource %s has no content or command", resourceURI)), nil
	}

	return createFile(tmpDir, path, content)
}

func copyResourceTree(resourceMap map[string]ResourceItem, tmpDir string, verbose bool, resourcePrefix, destinationPath string) (*mcp.CallToolResult, error) {
	var matchedURIs []string
	for uri := range resourceMap {
		if uri == resourcePrefix {
			matchedURIs = append(matchedURIs, uri)
			continue
		}
		if strings.HasPrefix(uri, resourcePrefix) {
			rest := uri[len(resourcePrefix):]
			if strings.HasPrefix(rest, "/") || strings.HasSuffix(resourcePrefix, "/") {
				matchedURIs = append(matchedURIs, uri)
			}
		}
	}

	if len(matchedURIs) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("no resources found matching prefix: %s", resourcePrefix)), nil
	}

	for _, uri := range matchedURIs {
		item := resourceMap[uri]
		relPath := strings.TrimPrefix(uri, resourcePrefix)
		relPath = strings.TrimPrefix(relPath, "/")

		targetPath := destinationPath
		if relPath != "" {
			targetPath = filepath.Join(destinationPath, relPath)
		}

		content, err := getResourceContent(item, tmpDir, verbose)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get resource content for %s: %v", uri, err)), nil
		}

		if content == "" {
			return mcp.NewToolResultError(fmt.Sprintf("resource %s has no content or command", uri)), nil
		}

		res, err := createFile(tmpDir, targetPath, content)
		if err != nil {
			return nil, err
		}
		if res.IsError {
			return res, nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully copied %d resources to %s.", len(matchedURIs), destinationPath)), nil
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
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create parent directories: %v", err)), nil
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

func replaceInFile(tmpDir, path, pattern, replacement string, replaceAll bool) (*mcp.CallToolResult, error) {
	fullPath, err := resolvePath(tmpDir, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}
	content := string(contentBytes)

	re, err := regexp.Compile("(?s)" + pattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid regular expression: %v", err)), nil
	}

	var newContent string
	if replaceAll {
		if !re.MatchString(content) {
			return mcp.NewToolResultError("pattern not found in file"), nil
		}
		newContent = re.ReplaceAllString(content, replacement)
	} else {
		indices := re.FindStringSubmatchIndex(content)
		if indices == nil {
			return mcp.NewToolResultError("pattern not found in file"), nil
		}
		result := []byte{}
		result = re.ExpandString(result, replacement, content, indices)
		newContent = content[:indices[0]] + string(result) + content[indices[1]:]
	}

	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
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
