# **Simple shell-interface MCP Server (simple-mcp)**

simple-mcp is a Model Context Protocol (MCP) server designed interface basic Linux shell command line commands to LLM-based agents.It is built using Go and the [mcp-go](https://github.com/mark3labs/mcp-go) library.

## **Prerequisites**

* **Go 1.23+** (for building)  
* **mcphost** (or another MCP-compliant client)  
* **Ollama** (or another LLM provider) running a capable model (e.g., qwen3-coder, llama3).

## **Building**

To build the server binary:

make

This will produce the simple-mcp binary in the current directory.

## **Configuration**

The server is configured via simple-mcp.yaml. This file defines:

* **Resources:** Static or dynamic system information (e.g., simple-mcp://system/uptime).  
* **Tools:** Executable commands exposed to the LLM (e.g., GetKubernetesPods, SystemUpgrade).

## **Usage with mcphost**

This repository includes an example configuration for mcphost. To use it

1. Ensure ollama is running and you have the qwen3-coder:30b (or similar) model pulled.  
2. Edit mcphost.yaml if you need to change the model name or provider URL.  
3. Run mcphost: mcphost --config mcphost.yaml

## **License**

This project is licensed under the MIT License \- see the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.
