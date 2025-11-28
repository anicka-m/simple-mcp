// Copyright (c) 2025 Vojtech Pavlik <vojtech@suse.com>
//
// Created using AI tools
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScratchLogic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scratch-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("CreateDirectory", func(t *testing.T) {
		res, err := createDirectory(tmpDir, "test-dir")
		require.NoError(t, err)
		assert.Equal(t, "Directory created successfully.", res.Content[0].(mcp.TextContent).Text)
		_, err = os.Stat(filepath.Join(tmpDir, "test-dir"))
		assert.NoError(t, err)
	})

	t.Run("CreateFile", func(t *testing.T) {
		res, err := createFile(tmpDir, "test-file.txt", "hello world\n")
		require.NoError(t, err)
		assert.Equal(t, "File created successfully.", res.Content[0].(mcp.TextContent).Text)
		content, err := os.ReadFile(filepath.Join(tmpDir, "test-file.txt"))
		assert.NoError(t, err)
		assert.Equal(t, "hello world\n", string(content))
	})

	t.Run("CreateFile_WithSubdir", func(t *testing.T) {
		res, err := createFile(tmpDir, "subdir/test-file.txt", "hello subdir\n")
		require.NoError(t, err)
		assert.Equal(t, "File created successfully.", res.Content[0].(mcp.TextContent).Text)
		content, err := os.ReadFile(filepath.Join(tmpDir, "subdir/test-file.txt"))
		assert.NoError(t, err)
		assert.Equal(t, "hello subdir\n", string(content))
	})

	t.Run("CopyResourceToFile", func(t *testing.T) {
		resourceMap := map[string]ResourceItem{
			"simple-mcp://content": {
				URI:     "simple-mcp://content",
				Content: "resource content",
			},
			"simple-mcp://command": {
				URI:     "simple-mcp://command",
				Command: "echo command content",
			},
		}

		res, err := copyResourceToFile(resourceMap, tmpDir, false, "simple-mcp://content", "resource-file.txt")
		require.NoError(t, err)
		assert.Equal(t, "File created successfully.", res.Content[0].(mcp.TextContent).Text)
		content, err := os.ReadFile(filepath.Join(tmpDir, "resource-file.txt"))
		assert.NoError(t, err)
		assert.Equal(t, "resource content", string(content))

		res, err = copyResourceToFile(resourceMap, tmpDir, false, "simple-mcp://command", "command-file.txt")
		require.NoError(t, err)
		assert.Equal(t, "File created successfully.", res.Content[0].(mcp.TextContent).Text)
		content, err = os.ReadFile(filepath.Join(tmpDir, "command-file.txt"))
		assert.NoError(t, err)
		assert.Equal(t, "command content\n", string(content))
	})

	t.Run("CopyResourceToFile_Combined", func(t *testing.T) {
		resourceMap := map[string]ResourceItem{
			"simple-mcp://combined": {
				URI:     "simple-mcp://combined",
				Content: "static content\n",
				Command: "echo dynamic content",
			},
		}

		res, err := copyResourceToFile(resourceMap, tmpDir, false, "simple-mcp://combined", "combined-file.txt")
		require.NoError(t, err)
		assert.Equal(t, "File created successfully.", res.Content[0].(mcp.TextContent).Text)
		content, err := os.ReadFile(filepath.Join(tmpDir, "combined-file.txt"))
		assert.NoError(t, err)
		assert.Equal(t, "static content\ndynamic content\n", string(content))
	})

	t.Run("ReadFile", func(t *testing.T) {
		_, err := createFile(tmpDir, "test-file-for-read.txt", "hello read\n")
		require.NoError(t, err)
		res, err := readFile(tmpDir, "test-file-for-read.txt")
		require.NoError(t, err)
		assert.Equal(t, "hello read\n", res.Content[0].(mcp.TextContent).Text)
	})

	t.Run("ModifyFile", func(t *testing.T) {
		_, err := createFile(tmpDir, "test-file-for-modify.txt", "hello world\n")
		require.NoError(t, err)

		patch := `--- a/test-file-for-modify.txt
+++ b/test-file-for-modify.txt
@@ -1 +1 @@
-hello world
+hello gopher
`
		res, err := modifyFile(tmpDir, "test-file-for-modify.txt", patch)
		require.NoError(t, err)
		assert.Equal(t, "File modified successfully.", res.Content[0].(mcp.TextContent).Text)
		content, err := os.ReadFile(filepath.Join(tmpDir, "test-file-for-modify.txt"))
		assert.NoError(t, err)
		assert.Equal(t, "hello gopher\n", string(content))
	})

	t.Run("ModifyFile_NonExistent", func(t *testing.T) {
		patch := `--- a/non-existent-file.txt
+++ b/non-existent-file.txt
@@ -1 +1 @@
-hello world
+hello gopher
`
		_, err := modifyFile(tmpDir, "non-existent-file.txt", patch)
		assert.Error(t, err)
	})

	t.Run("ListDirectory", func(t *testing.T) {
		listDir := filepath.Join(tmpDir, "list-test")
		require.NoError(t, os.Mkdir(listDir, 0755))
		_, err := createFile(listDir, "file1.txt", "content1\n")
		require.NoError(t, err)
		_, err = createDirectory(listDir, "subdir")
		require.NoError(t, err)

		res, err := listDirectory(tmpDir, "list-test")
		require.NoError(t, err)

		expectedContent := "file1.txt\nsubdir/\n"
		assert.Equal(t, expectedContent, res.Content[0].(mcp.TextContent).Text)
	})

	t.Run("DeleteFile", func(t *testing.T) {
		_, err := createFile(tmpDir, "test-file-for-delete.txt", "content\n")
		require.NoError(t, err)
		res, err := deleteFile(tmpDir, "test-file-for-delete.txt")
		require.NoError(t, err)
		assert.Equal(t, "File deleted successfully.", res.Content[0].(mcp.TextContent).Text)
		_, err = os.Stat(filepath.Join(tmpDir, "test-file-for-delete.txt"))
		assert.Error(t, err)
	})

	t.Run("RemoveDirectory", func(t *testing.T) {
		_, err := createDirectory(tmpDir, "dir-for-remove")
		require.NoError(t, err)
		res, err := removeDirectory(tmpDir, "dir-for-remove")
		require.NoError(t, err)
		assert.Equal(t, "Directory removed successfully.", res.Content[0].(mcp.TextContent).Text)
		_, err = os.Stat(filepath.Join(tmpDir, "dir-for-remove"))
		assert.Error(t, err)
	})

	t.Run("PathSecurity", func(t *testing.T) {
		paths := []string{
			"/etc/passwd",
			"../",
			"test/../../",
		}
		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				_, err := resolvePath(tmpDir, path)
				assert.Error(t, err)
			})
		}
	})
}
