# Global Middleware System for go-mcp

This document describes the global middleware system implemented in go-mcp server framework.

## Overview

The global middleware system allows developers to register middleware functions that are automatically applied to **all** tool handlers. This is different from the existing per-tool middleware system, which only applies to specific tools.

## API Reference

### MiddlewareFunc

```go
type MiddlewareFunc func(ctx context.Context, req *protocol.CallToolRequest, next ToolHandlerFunc) (*protocol.CallToolResult, error)
```

Middleware functions receive:
- `ctx context.Context`: The request context
- `req *protocol.CallToolRequest`: The tool call request
- `next ToolHandlerFunc`: The next middleware or the actual tool handler

They can:
- Perform pre-processing before calling `next`
- Perform post-processing after calling `next`
- Return early without calling `next` (e.g., for authentication failures)
- Modify the request before passing it to `next`
- Modify the result before returning it

### Server.Use Method

```go
func (server *Server) Use(middlewares ...MiddlewareFunc)
```

Registers one or more global middleware functions. Middlewares are executed in the order they are registered.

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/ThinkInAIXYZ/go-mcp/protocol"
    "github.com/ThinkInAIXYZ/go-mcp/server"
    "github.com/ThinkInAIXYZ/go-mcp/transport"
)

// Logging middleware logs all tool calls
func LoggingMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    start := time.Now()
    log.Printf("[Middleware] Calling tool: %s", req.Name)
    
    result, err := next(ctx, req)
    
    duration := time.Since(start)
    if err != nil {
        log.Printf("[Middleware] Tool %s failed after %v: %v", req.Name, duration, err)
    } else {
        log.Printf("[Middleware] Tool %s succeeded after %v", req.Name, duration)
    }
    
    return result, err
}

// Authentication middleware
func AuthMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    // Check authentication
    if !isAuthorized(ctx) {
        return nil, errors.New("unauthorized")
    }
    
    return next(ctx, req)
}

// Metrics middleware
func MetricsMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    // Record metrics
    recordMetric("tool_called", req.Name)
    
    result, err := next(ctx, req)
    
    if err != nil {
        recordMetric("tool_error", req.Name)
    } else {
        recordMetric("tool_success", req.Name)
    }
    
    return result, err
}

// Panic recovery middleware
func PanicRecoveryMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("[Middleware] Recovered from panic in tool %s: %v", req.Name, r)
        }
    }()
    
    return next(ctx, req)
}

func main() {
    // Create server
    s, err := server.NewServer(transport.NewStdioServerTransport())
    if err != nil {
        log.Fatal(err)
    }

    // Register global middlewares
    s.Use(
        PanicRecoveryMiddleware, // Should be first
        LoggingMiddleware,       // Log all calls
        AuthMiddleware,          // Check authentication
        MetricsMiddleware,       // Record metrics
    )

    // Register tools (these will automatically use the middlewares)
    s.RegisterTool(
        &protocol.Tool{
            Name:        "hello",
            Description: "Say hello",
        },
        func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
            return &protocol.CallToolResult{
                Content: []protocol.Content{
                    {Type: "text", Text: "Hello, World!"},
                },
            }, nil
        },
    )

    // Start server
    if err := s.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Middleware with Early Return

```go
// Rate limiting middleware
func RateLimitMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    if !rateLimiter.Allow(req.Name) {
        return nil, errors.New("rate limit exceeded")
    }
    
    return next(ctx, req)
}

// Validation middleware
func ValidationMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    if err := validateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    return next(ctx, req)
}
```

### Combining Global and Per-tool Middleware

```go
// Global middleware applies to all tools
server.Use(LoggingMiddleware, AuthMiddleware)

// Per-tool middleware applies only to specific tools
server.RegisterTool(
    &protocol.Tool{Name: "database-query", Description: "Query database"},
    databaseHandler,
    RateLimitMiddleware(rateLimiter), // Only for this tool
    CacheMiddleware(cache),           // Only for this tool
)

// This tool will only use the global middlewares
server.RegisterTool(
    &protocol.Tool{Name: "simple-tool", Description: "Simple operation"},
    simpleHandler,
)
```

## Execution Order

When both global and per-tool middlewares are used, the execution order is:

1. Global middlewares (in registration order)
2. Per-tool middlewares (in registration order)
3. Tool handler

Example:
```go
// Global middlewares
server.Use(Global1, Global2)

// Per-tool middlewares
server.RegisterTool(tool, handler, Tool1, Tool2)

// Execution order:
// Global1 → Global2 → Tool1 → Tool2 → Handler
```

## Best Practices

1. **Order Matters**: Register middlewares in the correct order. For example, authentication should come before logging.

2. **Early Return**: Use early return for authentication failures, rate limiting, etc.

3. **Error Handling**: Always handle errors properly in middleware.

4. **Performance**: Keep middleware lightweight to avoid performance bottlenecks.

5. **Context Usage**: Use context for request-scoped data like user information, request IDs, etc.

6. **Panic Recovery**: Always include panic recovery middleware as the first middleware.

## Migration Guide

### From Per-tool Middleware to Global Middleware

If you have existing per-tool middleware that should apply to all tools:

**Before:**
```go
// Duplicated for each tool
server.RegisterTool(tool1, handler1, LoggingMiddleware, AuthMiddleware)
server.RegisterTool(tool2, handler2, LoggingMiddleware, AuthMiddleware)
server.RegisterTool(tool3, handler3, LoggingMiddleware, AuthMiddleware)
```

**After:**
```go
// Register once for all tools
server.Use(LoggingMiddleware, AuthMiddleware)

server.RegisterTool(tool1, handler1)
server.RegisterTool(tool2, handler2)
server.RegisterTool(tool3, handler3)
```

## Common Middleware Examples

### Request ID Middleware
```go
func RequestIDMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    requestID := generateRequestID()
    ctx = context.WithValue(ctx, "request_id", requestID)
    
    log.Printf("[RequestID] %s: %s", requestID, req.Name)
    
    result, err := next(ctx, req)
    
    log.Printf("[RequestID] %s completed", requestID)
    
    return result, err
}
```

### Caching Middleware
```go
func CacheMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    cacheKey := generateCacheKey(req)
    
    if cached, ok := cache.Get(cacheKey); ok {
        return cached.(*protocol.CallToolResult), nil
    }
    
    result, err := next(ctx, req)
    if err == nil {
        cache.Set(cacheKey, result, cacheDuration)
    }
    
    return result, err
}
```

### Validation Middleware
```go
func ValidationMiddleware(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
    schema := getSchemaForTool(req.Name)
    if err := validateAgainstSchema(req.Arguments, schema); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    return next(ctx, req)
}
```

This middleware system provides a powerful and flexible way to implement cross-cutting concerns in your MCP server while maintaining clean separation of concerns and good code organization.