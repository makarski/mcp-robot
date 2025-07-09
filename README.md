MCP Robot - Go Server Library for Model Context Protocol
===

A Go server library for implementing [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) servers, built following Go conventions and inspired by the standard `net/http` package.

> **⚠️ Work in Progress**: This library currently supports **tools only**. Resources, prompts, and other MCP features are planned for future releases.

## 🚀 Features

- **Type-Safe Tool Definitions**: Leverage Go generics for compile-time type safety
- **Fluent Builder API**: Intuitive schema building with method chaining
- **Automatic Input Validation**: Input arguments are automatically validated against input schemas
- **Multiple Server Types**: HTTP and Stdio transport support
- **Rich Result Types**: Support for text, media, structured data, and resource links
- **Go Conventions**: Follows standard Go patterns, inspired by `net/http

## 📦 Installation

```bash
go get github.com/makarski/mcp-robot
```

## 🔧 Quick Start

```go
package main

import (
    "github.com/makarski/mcp-robot/server"
    "github.com/makarski/mcp-robot/tools"
)

func main() {
    // Define a tool with input/output schemas
    weatherTool := tools.NewTool("get_weather").
        Title("Weather Information").
        Description("Get current weather for a location").
        Input().
            WithString("location", "City name or coordinates", true).
            WithBoolean("celsius", "Use celsius temperature", false).
            Done().
        Output().
            WithString("status", "Response status", true).
            WithNumber("temperature", "Temperature value", true).
            WithString("conditions", "Weather conditions", true).
            Done().
        MarkReadOnly(true).
        Build()

    // Create tool function
    weatherFunc := tools.ToolFunc[tools.ToolResultStructured](func(params map[string]any) (tools.ToolResultStructured, error) {
        location := params["location"].(string)
        return tools.ToolResultStructured{
            "status": "success",
            "temperature": 22.5,
            "conditions": "Sunny",
            "location": location,
        }, nil
    })

    // Build and start server
    server := server.NewServerBuilder("weather-server", "1.0.0").
        WithTool(weatherTool, weatherFunc).
        BuildHTTPServer()

    server.ListenAndServe("mcp")(":8080", nil)
}
```

#### Stdio Server

```go
server := server.NewServerBuilder("weather-server", "1.0.0").
    WithTool(weatherTool, weatherFunc).
    BuildStdioServer()

// Start server
server.ListenAndServe()
```

## 🏗️ Tool Definition Builder

_The library provides validation for both input and output schemas_

### Basic Types

```go
tool := tools.NewTool("example").
    Input().
        WithString("name", "User name", true).           // required
        WithNumber("age", "User age", false).            // optional
        WithBoolean("active", "Is active", true).
        Done().
    Build()
```

### Arrays

```go
tool := tools.NewTool("process_data").
    Input().
        WithArray("items", "List of items to process", true).
            Of("object", "Individual item").
                WithString("id", "Item ID", true).
                WithString("name", "Item name", true).
                Done().
        Done().
    Build()
```

### Nested Objects

Arrays

```go
tool := tools.NewTool("process_data").
    Input().
        WithArray("items", "List of items to process", true).
            Of("object", "Individual item").
                WithString("id", "Item ID", true).
                WithString("name", "Item name", true).
                Done().
        Done().
    Build()
```

## 🎯 Tool Result Types

### Text Results

```go
func textTool(params map[string]any) (tools.ToolResultText, error) {
    return tools.NewToolResultText("Hello, World!"), nil
}
```

### Structured Results

```go
func structuredTool(params map[string]any) (tools.ToolResultStructured, error) {
    return tools.ToolResultStructured{
        "status": "success",
        "data": map[string]any{
            "count": 42,
            "items": []string{"item1", "item2"},
        },
    }, nil
}
```

### Rich Media Results

```go
func mediaTool(params map[string]any) (tools.ToolResultUnion, error) {
    imageData, _ := os.ReadFile("chart.png")

    return *tools.NewToolResultUnion().
        AddText("Analysis complete").
        AddImage(imageData, "image/png").
        AddResourceLink("https://example.com/report", "Full Report", "Detailed analysis", "text/html"), nil
}
```

## 🔧 Tool Annotations

```go
tool := tools.NewTool("database_query").
    Description("Query database").
    MarkReadOnly(true).              // Doesn't modify data
    MarkAsIdempotent(true).          // Safe to retry
    MarkAsCallingOpenWorld(false).   // Closed system
    Build()
```

## 📝 Error Handling

The library provides structured error handling with MCP protocol errors:

```go
func toolWithValidation(params map[string]any) (tools.ToolResultText, error) {
    email, ok := params["email"].(string)
    if !ok {
        return tools.ToolResultText{}, spec.NewProtocolError(
            spec.ErrorCodeInvalidParams,
            "email parameter is required and must be a string",
        )
    }

    // Tool logic...
    return tools.NewToolResultText("Success"), nil
}
```

## 🚀 What's Next

- **Resources**: File and data resource management
- **Prompts**: Dynamic prompt templates
- **Enhanced Validation**: Full JSON Schema support for arrays and nested objects
- **Progress Reporting**: Long-running operation support
- **Streaming**: Real-time data streaming

---

**Built with Go conventions in mind** • Inspired by `net/http` • Type-safe • Protocol compliant
