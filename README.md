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
  * Resources can be defined with:
    * `content`: Static text content.
    * `contentFile`: Path to a file whose content will be loaded (relative to the config file).
    * `directory`: Path to a directory whose entire subtree of files will be added as individual resources. The file's relative path will be appended to the resource URI.
    * `command`: A shell command to execute to generate dynamic content.
* **Tools:** Executable commands exposed to the LLM (e.g., GetKubernetesPods, SystemUpgrade).

## **Security & Remote Access**

**Important:** This tool provides shell execution capabilities.

By default, simple-mcp binds to localhost:8080, allowing only local connections. If you need to access simple-mcp from a remote machine (e.g., an LLM running on a different server), **do not** expose this port directly to the internet.

Instead, use a reverse proxy like Nginx or Caddy to handle authentication and TLS.

### **Example: Nginx with Basic Auth**

1. Install Nginx and apache2-utils.  
2. Create a password file: sudo htpasswd \-c /etc/nginx/.htpasswd myuser  
3. Configure Nginx:

server {  
    listen 443 ssl;  
    server\_name mcp.example.com;

    \# ... ssl config ...

    location / {  
        proxy\_pass \[http://localhost:8080\](http://localhost:8080);
          
        \# Enable Basic Authentication  
        auth\_basic "Restricted MCP Access";  
        auth\_basic\_user\_file /etc/nginx/.htpasswd;

        \# WebSocket support (required for MCP)  
        proxy\_http\_version 1.1;  
        proxy\_set\_header Upgrade $http\_upgrade;  
        proxy\_set\_header Connection "upgrade";  
    }  
}

## **Usage with mcphost**

This repository includes an example configuration for mcphost. To use it

1. Ensure ollama is running and you have the qwen3-coder:30b (or similar) model pulled.  
2. Edit mcphost.yaml if you need to change the model name or provider URL.  
3. Run mcphost: mcphost \--config mcphost.yaml

**Note:** The mcphost.yaml configuration references systemprompt.txt using a relative path. You must run mcphost from the directory containing systemprompt.txt (or update the yaml to provide the full path), otherwise mcphost will fail to load the system prompt.

## **License**

This project is licensed under the MIT License \- see the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.