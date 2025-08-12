# Go-MCP

<div align="center">
<img src="docs/images/img_2.png" height="250" alt="Statusphere logo">
</div>
<br/>

<p align="center">
  <a href="https://github.com/hhfgeg/go-mcp/releases"><img src="https://img.shields.io/github/v/release/hhfgeg/go-mcp?style=flat" alt="Release"></a>
  <a href="https://github.com/hhfgeg/go-mcp/stargazers"><img src="https://img.shields.io/github/stars/hhfgeg/go-mcp?style=flat" alt="Stars"></a>
  <a href="https://github.com/hhfgeg/go-mcp/network/members"><img src="https://img.shields.io/github/forks/hhfgeg/go-mcp?style=flat" alt="Forks"></a>
  <a href="https://github.com/hhfgeg/go-mcp/issues"><img src="https://img.shields.io/github/issues/hhfgeg/go-mcp?color=gold&style=flat" alt="Issues"></a>
  <a href="https://github.com/hhfgeg/go-mcp/pulls"><img src="https://img.shields.io/github/issues-pr/hhfgeg/go-mcp?color=gold&style=flat" alt="Pull Requests"></a>
  <a href="https://github.com/hhfgeg/go-mcp/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License"></a>
  <a href="https://github.com/hhfgeg/go-mcp/graphs/contributors"><img src="https://img.shields.io/github/contributors/hhfgeg/go-mcp?color=green&style=flat" alt="Contributors"></a>
  <a href="https://github.com/hhfgeg/go-mcp/commits"><img src="https://img.shields.io/github/last-commit/hhfgeg/go-mcp?color=green&style=flat" alt="Last Commit"></a>
</p>
<p align="center">
  <a href="https://pkg.go.dev/github.com/hhfgeg/go-mcp"><img src="https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&style=flat" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/hhfgeg/go-mcp"><img src="https://img.shields.io/badge/go%20report-A+-brightgreen?style=flat" alt="Go Report"></a>
  <a href="https://github.com/hhfgeg/go-mcp/actions"><img src="https://img.shields.io/badge/Go%20Tests-passing-brightgreen?style=flat" alt="Go Tests"></a>
</p>

<p align="center">
  <a href="README_CN.md">简体中文</a> | <a href="README_TW.md">繁體中文</a> | <a href="README_VI.md">Tiếng Việt</a>
</p>

## 🚀 Overview

Go-MCP is a powerful Go version of the MCP SDK that implements the Model Context Protocol (MCP) to facilitate seamless communication between external systems and AI applications. Based on the strong typing and performance advantages of the Go language, it provides a concise and idiomatic API to facilitate your integration of external systems into AI applications.

### ✨ Key Features

- 🔄 **Complete Protocol Implementation**: Full implementation of the MCP specification, ensuring seamless integration with all compatible services
- 🏗️ **Elegant Architecture Design**: Adopts a clear three-layer architecture, supports bidirectional communication, ensuring code modularity and extensibility
- 🔌 **Seamless Integration with Web Frameworks**: Provides MCP protocol-compliant http.Handler, allowing developers to integrate MCP into their service frameworks
- 🛡️ **Type Safety**: Leverages Go's strong type system for clear, highly maintainable code
- 📦 **Simple Deployment**: Benefits from Go's static compilation, eliminating complex dependency management
- ⚡ **High-Performance Design**: Fully utilizes Go's concurrency capabilities, maintaining excellent performance and low resource overhead across various scenarios

## 🛠️ Installation

```bash
go get github.com/hhfgeg/go-mcp
```

Requires Go 1.18 or higher.

## 🎯 Quick Start

### Client Example

```go
package main

import (
	"context"
	"log"

	"github.com/hhfgeg/go-mcp/client"
	"github.com/hhfgeg/go-mcp/transport"
)

func main() {
	// Create SSE transport client
	transportClient, err := transport.NewSSEClientTransport("http://127.0.0.1:8080/sse")
	if err != nil {
		log.Fatalf("Failed to create transport client: %v", err)
	}

	// Initialize MCP client
	mcpClient, err := client.NewClient(transportClient)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	defer mcpClient.Close()

	// Get available tools
	tools, err := mcpClient.ListTools(context.Background())
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}
	log.Printf("Available tools: %+v", tools)
}
```

### Server Example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hhfgeg/go-mcp/protocol"
	"github.com/hhfgeg/go-mcp/server"
	"github.com/hhfgeg/go-mcp/transport"
)

type TimeRequest struct {
	Timezone string `json:"timezone" description:"timezone" required:"true"` // Use field tag to describe input schema
}

func main() {
	// Create SSE transport server
	transportServer, err := transport.NewSSEServerTransport("127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Failed to create transport server: %v", err)
	}

	// Initialize MCP server
	mcpServer, err := server.NewServer(transportServer)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}
	
	// Optional: Global middleware (variadic - multiple middleware supported)
	// mcpServer.Use(
	// 	LoggingMiddleware,
	// 	AuthMiddleware,
	// 	MetricsMiddleware,
	// )

	// Register time query tool
	tool, err := protocol.NewTool("current_time", "Get current time for specified timezone", TimeRequest{})
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
		return
	}
	mcpServer.RegisterTool(tool, handleTimeRequest)

	// Start server
	if err = mcpServer.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleTimeRequest(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	var timeReq TimeRequest
	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &timeReq); err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation(timeReq.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %v", err)
	}

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: time.Now().In(loc).String(),
			},
		},
	}, nil
}
```

### Integration With Gin Server

```go
package main

import (
	"context"
	"log"

	"github.com/hhfgeg/go-mcp/protocol"
	"github.com/hhfgeg/go-mcp/server"
	"github.com/hhfgeg/go-mcp/transport"
	"github.com/gin-gonic/gin"
)

func main() {
	messageEndpointURL := "/message"

	sseTransport, mcpHandler, err := transport.NewSSEServerTransportAndHandler(messageEndpointURL)
	if err != nil {
		log.Panicf("new sse transport and hander with error: %v", err)
	}

	// new mcp server
	mcpServer, _ := server.NewServer(sseTransport)

	// register tool with mcpServer
	// mcpServer.RegisterTool(tool, toolHandler)

	// start mcp Server
	go func() {
		mcpServer.Run()
	}()

	defer mcpServer.Shutdown(context.Background())

	r := gin.Default()
	r.GET("/sse", func(ctx *gin.Context) {
		mcpHandler.HandleSSE().ServeHTTP(ctx.Writer, ctx.Request)
	})
	r.POST(messageEndpointURL, func(ctx *gin.Context) {
		mcpHandler.HandleMessage().ServeHTTP(ctx.Writer, ctx.Request)
	})

	if err = r.Run(":8080"); err != nil {
		return
	}
}
```

[Reference：A more complete example](https://github.com/hhfgeg/go-mcp/blob/main/examples/http_handler/main.go)

## 🏗️ Architecture Design

Go-MCP adopts an elegant three-layer architecture:

![Architecture Overview](docs/images/img.png)

1. **Transport Layer**: Handles underlying communication implementation, supporting multiple transport protocols
2. **Protocol Layer**: Handles MCP protocol encoding/decoding and data structure definitions
3. **User Layer**: Provides friendly client and server APIs

Currently supported transport methods:

![Transport Methods](docs/images/img_1.png)

- **HTTP SSE/POST**: HTTP-based server push and client requests, suitable for web scenarios
- **Streamable HTTP**: Supports HTTP POST/GET requests with both stateless and stateful modes, where stateful mode utilizes SSE for multi-message streaming to enable server-to-client notifications and requests
- **Stdio**: Standard input/output stream-based, suitable for local inter-process communication

The transport layer uses a unified interface abstraction, making it simple to add new transport methods (like Streamable HTTP, WebSocket, gRPC) without affecting upper-layer code.

## 🤝 Contributing

We welcome all forms of contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License

本项目基于 [Original Project](https://github.com/ThinkInAIXYZ/go-mcp) 修改，
原始项目遵循 MIT License。

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Contact Us

- **GitHub Issues**: [Submit an issue](https://github.com/hhfgeg/go-mcp/issues)

## ✨ Contributors

Thanks to all developers who have contributed to this project!

<a href="https://github.com/hhfgeg/go-mcp/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=hhfgeg/go-mcp" alt="Contributors" />
</a>

## 📈 Project Trends

[![Star History](https://api.star-history.com/svg?repos=hhfgeg/go-mcp&type=Date)](https://www.star-history.com/#hhfgeg/go-mcp&Date)
