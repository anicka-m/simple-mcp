// Copyright (c) 2025 Vojtech Pavlik <vojtech@suse.com>
//
// Created using AI tools
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Package main is the entry point for the simple-mcp server. It initializes the
// MCP server, sets up the async task store, parses the configuration, and
// registers all tools and resources before starting the HTTP listener.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	configFile := flag.String("config", "./simple-mcp.yaml", "Path to the YAML configuration file.")
	listenAddr := flag.String("listen-addr", ":8080", "Address to listen on for HTTP requests.")
	tmpDir := flag.String("tmpdir", "", "Path to a directory for scratch space.")
	flag.Parse()

	if *tmpDir != "" {
		log.Printf("Scratch space enabled at: %s", *tmpDir)
		if err := checkTmpDir(*tmpDir); err != nil {
			log.Fatalf("FATAL: Invalid --tmpdir: %v", err)
		}
	}

	cfg, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("FATAL: Error loading configuration: %v", err)
	}
	log.Printf("Configuration loaded successfully from %s", *configFile)

	taskStore := NewTaskStore()
	log.Printf("Task store initialized.")

	// Pre-cache resource definitions for efficient lookup by the GetResource tool.
	resourceMap := make(map[string]ResourceItem)
	for _, item := range cfg.Specification.Resources {
		resourceMap[item.URI] = item
	}
	log.Printf("Cached %d resource definitions.", len(resourceMap))

	mcpServer := server.NewMCPServer(
		cfg.Metadata.Name,
		cfg.APIVersion,
		server.WithToolCapabilities(false),
		server.WithRecovery(),                       // Gracefully handle panics in handlers
		server.WithResourceCapabilities(true, true), // Advertise resource support
	)
	log.Printf("MCP Server %s with API %s created.", cfg.Metadata.Name, cfg.APIVersion)

	registerBuiltinTools(mcpServer, taskStore, resourceMap, *tmpDir)
	registerConfigTools(mcpServer, cfg, taskStore, *tmpDir)
	registerResources(mcpServer, cfg, *tmpDir)

	if *tmpDir != "" {
		registerScratchTools(mcpServer, *tmpDir)
	}

	log.Printf("Creating Streamable HTTP server...")
	httpOpts := []server.StreamableHTTPOption{}
	httpServer := server.NewStreamableHTTPServer(mcpServer, httpOpts...)

	log.Printf("MCP server starting, listening on %s/mcp ...", *listenAddr)
	if err := httpServer.Start(*listenAddr); err != nil {
		log.Fatalf("FATAL: Could not start HTTP server: %v", err)
	}
}

func checkTmpDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("could not stat path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check for write permissions by creating a temporary file.
	tmpFile, err := os.CreateTemp(path, "simple-mcp-write-test-")
	if err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	os.Remove(tmpFile.Name()) // Clean up the temporary file.

	return nil
}


// registerBuiltinTools adds the core infrastructure tools required for
// mcphost compatibility and async task management.
func registerBuiltinTools(mcpServer *server.MCPServer, taskStore *TaskStore, resourceMap map[string]ResourceItem, tmpDir string) {
	mcpServer.AddTool(mcp.NewTool(
		"ping",
		mcp.WithDescription("Responds with 'pong' to keep the connection alive."),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("pong"), nil
	})

	// Helps the LLM recover context if it forgets a task ID.
	listTasksTool := mcp.NewTool(
		"ListPendingTasks",
		mcp.WithDescription("Lists all asynchronous tasks that are currently 'pending' or 'running'."),
	)
	listTasksHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Println("Handling ListPendingTasks request.")
		activeTasks := taskStore.ListActiveTasks()
		if len(activeTasks) == 0 {
			return mcp.NewToolResultText("No active (pending or running) tasks found."), nil
		}

		var b strings.Builder
		b.WriteString(fmt.Sprintf("Found %d active tasks:\n\n", len(activeTasks)))
		for _, task := range activeTasks {
			b.WriteString(fmt.Sprintf("Tool: %s\nTaskID: %s\nStatus: %s\nRunning For: %s\n\n",
				task.ToolName, task.ID, task.Status, time.Since(task.StartTime).Truncate(time.Second)))
		}
		return mcp.NewToolResultText(b.String()), nil
	}
	mcpServer.AddTool(listTasksTool, listTasksHandler)

	// Polling mechanism for clients that don't support async resource subscriptions.
	taskStatusTool := mcp.NewTool(
		"TaskStatus",
		mcp.WithDescription("Gets the status of a long-running async task from its Task ID or URI."),
		mcp.WithString(
			"taskID",
			mcp.Required(),
			mcp.Description("The Task ID (UUID) or full Task URI (e.g., simple-mcp://tasks/...)"),
		),
	)
	taskStatusHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, _ := request.RequireString("taskID")

		if strings.HasPrefix(taskID, "simple-mcp://tasks/") {
			taskID = strings.TrimPrefix(taskID, "simple-mcp://tasks/")
		}

		task, ok := taskStore.Get(taskID)
		if !ok {
			log.Printf("TaskStatus request for non-existent ID: %s", taskID)
			return mcp.NewToolResultText(fmt.Sprintf("Status: not_found\nMessage: No task found with ID: %s", taskID)), nil
		}

		log.Printf("Handling TaskStatus request for: %s", taskID)
		return mcp.NewToolResultText(task.FormatStatus()), nil
	}
	mcpServer.AddTool(taskStatusTool, taskStatusHandler)

	// Provides a discoverable list of system context resources.
	listResourcesTool := mcp.NewTool(
		"ListResources",
		mcp.WithDescription("Lists all available system resources (context) provided by this server."),
	)
	listResourcesHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Println("Handling ListResources request.")
		var b strings.Builder
		b.WriteString(fmt.Sprintf("Found %d resources:\n\n", len(resourceMap)))
		for uri, item := range resourceMap {
			b.WriteString(fmt.Sprintf("URI: %s\nDescription: %s\n\n", uri, item.Description))
		}
		return mcp.NewToolResultText(b.String()), nil
	}
	mcpServer.AddTool(listResourcesTool, listResourcesHandler)

	// Allows retrieving resource content via a tool call, bypassing client-side restrictions.
	getResourceTool := mcp.NewTool(
		"GetResource",
		mcp.WithDescription("Gets the current content of a specific resource by its URI."),
		mcp.WithString(
			"resourceURI",
			mcp.Required(),
			mcp.Description("The full URI of the resource (e.g., simple-mcp://system/uptime)."),
		),
	)
	getResourceHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resourceURI, _ := request.RequireString("resourceURI")
		log.Printf("Handling GetResource request for: %s", resourceURI)

		item, ok := resourceMap[resourceURI]
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("Resource not found: %s. Call ListResources to see available URIs.", resourceURI)), nil
		}

		if item.Command != "" {
			cmdItem := ContextItem{Command: item.Command}
			output, err := executeCommand(cmdItem, nil, tmpDir)
			if err != nil {
				log.Printf("Error executing command for resource %s: %v", resourceURI, err)
				return mcp.NewToolResultError(fmt.Sprintf("Error executing command for %s: %v", resourceURI, err)), nil
			}
			return mcp.NewToolResultText(output), nil
		} else if item.Content != "" {
			return mcp.NewToolResultText(item.Content), nil
		}

		return mcp.NewToolResultError(fmt.Sprintf("Resource %s is invalid (no content or command).", resourceURI)), nil
	}
	mcpServer.AddTool(getResourceTool, getResourceHandler)
}

// registerConfigTools iterates through the configuration and registers
// declared tools, routing them to sync or async handlers.
func registerConfigTools(mcpServer *server.MCPServer, cfg *Config, taskStore *TaskStore, tmpDir string) {
	for _, item := range cfg.Specification.Items {
		currentItem := item
		var toolOptions []mcp.ToolOption
		toolOptions = append(toolOptions, mcp.WithDescription(item.Description))

		for _, paramName := range item.Parameters {
			toolOptions = append(toolOptions, mcp.WithString(
				paramName,
				mcp.Required(),
				mcp.Description(fmt.Sprintf("Parameter: %s", paramName)),
			))
		}

		tool := mcp.NewTool(item.Name, toolOptions...)

		handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("Handling request for tool: %s", currentItem.Name)

			params := make(map[string]interface{})
			for _, paramName := range currentItem.Parameters {
				val, err := request.RequireString(paramName)
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				params[paramName] = val
			}

			if currentItem.Async {
				return handleAsyncTask(ctx, currentItem, params, taskStore, tmpDir)
			}
			return handleSyncTask(ctx, currentItem, params, tmpDir)
		}

		mcpServer.AddTool(tool, handler)
		log.Printf("Registered tool: %s (Async: %v)", item.Name, item.Async)
	}
}

func handleSyncTask(ctx context.Context, currentItem ContextItem, params map[string]interface{}, tmpDir string) (*mcp.CallToolResult, error) {
	output, err := executeCommand(currentItem, params, tmpDir)
	if err != nil {
		log.Printf("Error executing command '%s': %v", currentItem.Name, err)
		// Return stderr output to the LLM to help with diagnosing the failure.
		return mcp.NewToolResultError(fmt.Sprintf("Command failed: %v. Output: %s", err, output)), nil
	}

	log.Printf("Successfully executed tool '%s', output size: %d", currentItem.Name, len(output))
	return mcp.NewToolResultText(output), nil
}

func handleAsyncTask(ctx context.Context, currentItem ContextItem, params map[string]interface{}, taskStore *TaskStore, tmpDir string) (*mcp.CallToolResult, error) {
	// Enforce concurrency lock: prevent multiple instances of the same long-running task.
	if taskStore.HasActiveTask(currentItem.Name) {
		log.Printf("Rejected async task %s: task is already running.", currentItem.Name)
		return mcp.NewToolResultError(fmt.Sprintf("Task '%s' is already in progress. Call 'ListPendingTasks' or 'TaskStatus' to monitor it.", currentItem.Name)), nil
	}

	srv := server.ServerFromContext(ctx)
	if srv == nil {
		log.Println("Error: could not get server from context for async task")
		return mcp.NewToolResultError("could not get server from context"), nil
	}

	jobID := uuid.NewString()
	taskURI := fmt.Sprintf("simple-mcp://tasks/%s", jobID)

	task := taskStore.Create(jobID, currentItem.Name)

	// Create a dynamic resource for this specific task ID. This follows the
	// standard MCP pattern where a task becomes a subscribable resource.
	taskResource := mcp.NewResource(
		taskURI,
		fmt.Sprintf("Status of async job: %s (Job ID: %s)", currentItem.Name, jobID),
		mcp.WithMIMEType("text/plain"),
	)
	taskResourceHandler := func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		log.Printf("Handling standard MCP resource read for task: %s", jobID)
		task, ok := taskStore.Get(jobID)
		if !ok {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      taskURI,
					MIMEType: "text/plain",
					Text:     "Status: unknown\nMessage: Task ID not found.",
				},
			}, nil
		}

		contents := []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      taskURI,
				MIMEType: "text/plain",
				Text:     task.FormatStatus(),
			},
		}
		return contents, nil
	}

	srv.AddResource(taskResource, taskResourceHandler)

	go func() {
		// Ensure this goroutine does not crash the main server.
		defer func() {
			if r := recover(); r != nil {
				log.Printf("FATAL PANIC in async job %s: %v", jobID, r)
				errMsg := fmt.Sprintf("Async job %s failed with an internal server panic: %v", jobID, r)
				taskStore.SetStatus(jobID, "failed", errMsg)
			}
		}()

		log.Printf("Starting async job %s: %s", jobID, currentItem.Name)
		taskStore.SetStatus(jobID, "running", "Job is executing...")

		output, err := executeCommand(currentItem, params, tmpDir)

		if err != nil {
			log.Printf("Async job %s finished with status: failed", jobID)
			errMsg := fmt.Sprintf("%v. Output: %s", err, output)
			taskStore.SetStatus(jobID, "failed", errMsg)
		} else {
			log.Printf("Async job %s finished with status: completed", jobID)
			taskStore.SetStatus(jobID, "completed", output)
		}
	}()

	log.Printf("Async tool %s started. Task URI: %s", currentItem.Name, taskURI)
	initialContents := mcp.TextResourceContents{
		URI:      taskURI,
		MIMEType: "text/plain",
		Text:     task.FormatStatus(),
	}
	return mcp.NewToolResultResource(taskURI, initialContents), nil
}

// registerResources registers the static or dynamic resources defined in the
// config file. These are separate from the ephemeral task resources.
func registerResources(mcpServer *server.MCPServer, cfg *Config, tmpDir string) {
	for _, item := range cfg.Specification.Resources {
		currentItem := item

		resource := mcp.NewResource(
			currentItem.URI,
			currentItem.Description,
			mcp.WithResourceDescription(currentItem.Description),
			mcp.WithMIMEType("text/plain"),
		)

		var handler server.ResourceHandlerFunc

		// Combined handler for content, contentFile, and command
		handler = func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			log.Printf("Handling resource read request for: %s", currentItem.URI)
			var combinedContent strings.Builder

			// Append static content first
			if currentItem.Content != "" {
				combinedContent.WriteString(currentItem.Content)
			}

			// Then, append command output if a command is defined
			if currentItem.Command != "" {
				cmdItem := ContextItem{Command: currentItem.Command}
				output, err := executeCommand(cmdItem, nil, tmpDir)
				if err != nil {
					log.Printf("Error executing command for resource %s: %v", currentItem.URI, err)
					// Append error message to content for visibility
					output = fmt.Sprintf("\nError executing command: %v. Output: %s", err, output)
				}
				combinedContent.WriteString(output)
			}

			contents := []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      currentItem.URI,
					MIMEType: "text/plain",
					Text:     combinedContent.String(),
				},
			}
			return contents, nil
		}
		log.Printf("Registered resource: %s (dynamic: %v)", currentItem.URI, currentItem.Command != "")

		mcpServer.AddResource(resource, handler)
	}
}
