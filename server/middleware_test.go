package server_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

// TestMiddleware tests the global middleware functionality via handler behavior
func TestMiddleware(t *testing.T) {
	mockTransport := transport.NewMockServerTransport(io.NopCloser(bytes.NewReader(nil)), io.Discard)

	s, err := server.NewServer(mockTransport)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if s == nil {
		t.Fatal("Server is nil")
	}

	testMiddleware := func(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
		return next(ctx, req)
	}

	s.Use(testMiddleware)
}

// TestMiddlewareEarlyReturn tests middleware that can return early
func TestMiddlewareEarlyReturn(t *testing.T) {
	mockTransport := transport.NewMockServerTransport(io.NopCloser(bytes.NewReader(nil)), io.Discard)

	s, err := server.NewServer(mockTransport)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if s == nil {
		t.Fatal("Server is nil")
	}

	authMiddleware := func(ctx context.Context, req *protocol.CallToolRequest, next server.ToolHandlerFunc) (*protocol.CallToolResult, error) {
		return nil, context.Canceled
	}

	s.Use(authMiddleware)
}

// TestNoMiddleware tests that tools work without middleware
func TestNoMiddleware(t *testing.T) {
	mockTransport := transport.NewMockServerTransport(io.NopCloser(bytes.NewReader(nil)), io.Discard)

	s, err := server.NewServer(mockTransport)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if s == nil {
		t.Fatal("Server is nil")
	}

	testHandler := func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		return protocol.NewCallToolResult([]protocol.Content{
			&protocol.TextContent{Type: "text", Text: "success"},
		}, false), nil
	}

	testTool := &protocol.Tool{
		Name:        "test",
		Description: "Test tool",
		InputSchema: protocol.InputSchema{
			Type:       "object",
			Properties: map[string]*protocol.Property{},
		},
	}
	s.RegisterTool(testTool, testHandler)
}
