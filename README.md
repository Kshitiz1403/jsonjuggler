# JSONJuggler

JSONJuggler is a powerful workflow engine that implements the [Serverless Workflow Specification](https://serverlessworkflow.io/). It enables you to orchestrate complex workflows using JSON definitions and execute them with a rich set of built-in and custom activities.

## 🌟 Key Features

- **Flexible Workflow States**: Support for Operation and Switch states with powerful data conditions
- **JQ Integration**: Leverage JQ expressions for sophisticated data manipulation and conditional logic
- **Extensible Activities**: Plugin your own custom activities or use the built-in ones
- **Robust Error Handling**: Comprehensive error management with customizable transitions
- **Debug Superpowers**: Rich debugging capabilities with detailed execution tracing
- **Structured Logging**: Context-aware logging for better observability
- **State Management**: Efficient state data handling with current, states, and globals scopes

### Built-in Activities
- 🔄 **JQ Transform**: Transform your data using powerful JQ expressions
- 🌐 **HTTP Request**: Make configurable RESTful API calls
- 🔐 **JWE Encrypt**: Built-in JSON Web Encryption support

## 📦 Installation

```bash
go get github.com/kshitiz1403/jsonjuggler
```

## 🚀 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/kshitiz1403/jsonjuggler/config"
    "github.com/kshitiz1403/jsonjuggler/logger"
    "github.com/kshitiz1403/jsonjuggler/parser"
)

func main() {
    // Initialize engine with debug mode
    engine := config.Initialize(
        config.WithDebug(true),
        config.WithLogger(zap.NewLogger(logger.DebugLevel)),
    )

    // Parse workflow definition
    workflow, err := parser.NewParser(engine.GetRegistry()).ParseFromFile("workflow.json")
    if err != nil {
        panic(err)
    }

    // Setup execution context
    ctx := logger.WithFields(context.Background(),
        logger.String("requestID", "123"),
        logger.String("userID", "456"),
    )

    // Define input data and globals
    input := map[string]interface{}{
        "user": map[string]interface{}{
            "type": "premium",
            "email": "user@example.com",
        },
    }

    globals := map[string]interface{}{
        "apiKey": "secret-key",
        "baseURL": "https://api.example.com",
    }

    // Execute workflow
    result, err := engine.Execute(ctx, workflow, input, globals)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Workflow result: %v\n", result.Data)
    if result.Debug != nil {
        fmt.Printf("Execution details: %v\n", result.Debug)
    }
}
```

## 🛠️ Creating Custom Activities

Extend JSONJuggler's capabilities by creating your own activities:

```go
type CustomActivity struct {
    activities.BaseActivity
}

func (a *CustomActivity) Execute(ctx context.Context, args map[string]any) (interface{}, error) {
    a.GetLogger().InfoContext(ctx, "Starting custom operation")
    
    // Your implementation here
    result := "hello world"
    
    a.GetLogger().DebugContextf(ctx, "Operation completed: %v", result)
    return result, nil
}

// Register your custom activity
engine := config.Initialize(
    config.WithActivity("CustomOp", &CustomActivity{}),
    config.WithDebug(true),
)
```

## 📝 Workflow Definition Examples

### Basic Data Processing
```json
{
    "id": "data-processing",
    "version": "1.0",
    "specVersion": "0.9",
    "name": "Data Processing Pipeline",
    "start": "TransformData",
    "states": [
        {
            "name": "TransformData",
            "type": "operation",
            "actions": [
                {
                    "functionRef": {
                        "refName": "JQ",
                        "arguments": {
                            "query": ".data | map(select(.value > 100))",
                            "data": "${ .current }"
                        }
                    }
                }
            ],
            "end": true
        }
    ]
}
```

### Complex Workflow with Conditions
```json
{
    "id": "loan-application",
    "version": "1.0",
    "specVersion": "0.9",
    "name": "Loan Application Process",
    "start": "ExtractData",
    "states": [
        {
            "name": "ExtractData",
            "type": "operation",
            "actions": [
                {
                    "functionRef": {
                        "refName": "JQ",
                        "arguments": {
                            "query": ".application.user",
                            "data": "${ .current }"
                        }
                    }
                }
            ],
            "transition": {
                "nextState": "EvaluateRisk"
            }
        },
        {
            "name": "EvaluateRisk",
            "type": "switch",
            "dataConditions": [
                {
                    "name": "HighRisk",
                    "condition": ".current.risk_score < 600",
                    "transition": {
                        "nextState": "RejectApplication"
                    }
                },
                {
                    "name": "LowRisk",
                    "condition": ".current.risk_score >= 600",
                    "transition": {
                        "nextState": "ApproveApplication"
                    }
                }
            ],
            "defaultCondition": {
                "transition": {
                    "nextState": "StandardProcess"
                }
            }
        }
    ]
}
```

## 🎨 Workflow Visualization

JSONJuggler provides a web-based visualization tool to help you design and understand your workflows better. Visit [JSONJuggler Workflow Editor](https://kshitiz1403.github.io/swf-editor/) to:

- Visualize your workflow definitions as interactive diagrams
- Generate workflow diagrams with different quality settings (Low/Medium/High)
- Toggle between light and dark themes
- Export workflow diagrams as images
- View your workflow in full-screen mode

The editor provides a real-time preview of your workflow structure, making it easier to understand and debug complex state transitions and conditions.

## 🔧 Built-in Activities Detail

### JQ Transform
Transform data using powerful JQ expressions:
```json
{
    "functionRef": {
        "refName": "JQ",
        "arguments": {
            "query": ".user | {name: .name, email: .email}",
            "data": "${ .current }"
        }
    }
}
```

### HTTP Request
Make HTTP requests with rich configuration:
```json
{
    "functionRef": {
        "refName": "HTTPRequest",
        "arguments": {
            "url": "http://api.example.com/users",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "Authorization": "Bearer ${.globals.apiToken}"
            },
            "body": "${ .current }",
            "timeoutSec": 30,
            "failOnError": true
        }
    }
}
```

### JWE Encrypt
Encrypt data using JSON Web Encryption:
```json
{
    "functionRef": {
        "refName": "JWEEncrypt",
        "arguments": {
            "payload": "${ .current.sensitive_data }",
            "publicKey": "${ .globals.encryption_key }",
            "contentEncryptionAlgorithm": "A256GCM",
            "keyManagementAlgorithm": "RSA-OAEP-256"
        }
    }
}
```

## 🛡️ Error Handling

JSONJuggler provides comprehensive error handling capabilities:
- Activity-level error handling with structured errors
- State-level error transitions with condition matching
- Default error handlers for unmatched errors
- Detailed error information in debug mode

Example error handling configuration:
```json
{
    "name": "ProcessOrder",
    "type": "operation",
    "actions": [...],
    "onErrors": [
        {
            "errorRef": "connection refused",
            "transition": "HandleConnectionError"
        },
        {
            "errorRef": "DefaultErrorRef",
            "transition": "HandleGenericError"
        }
    ]
}
```

## 🔍 Debug Mode

When debug mode is enabled, JSONJuggler provides detailed execution information:
- Complete state transition history
- Action execution details with timing
- Input and output data for each step
- Matched conditions in switch states
- Error details with full context

Example debug output:
```json
{
    "states": [
        {
            "name": "ExtractData",
            "type": "operation",
            "startTime": "2024-02-20T10:00:00Z",
            "endTime": "2024-02-20T10:00:01Z",
            "input": {...},
            "output": {...},
            "actions": [
                {
                    "activityName": "JQ",
                    "arguments": {...},
                    "startTime": "2024-02-20T10:00:00Z",
                    "endTime": "2024-02-20T10:00:01Z",
                    "output": {...}
                }
            ]
        }
    ]
}
```

## ⚠️ Known Limitations

- Limited support for ISO 8601 duration formats (fractional durations like "PT0.5S" are not properly parsed)

## 🗺️ Roadmap

- **Documentation**: 
  - Automated documentation generation from code comments
  - Validation for necessary documentation comments when registering new activities
- **Architecture**: 
  - Cluster-based approach for related activities
  - Consolidation of similar activities (e.g., JWE encryption, AES encryption) into unified activities with type parameters
  - Prevention of activity proliferation through smart consolidation
- **Examples**: 
  - Creation of example applications in a new module
  - Comprehensive usage examples and best practices

## 📜 License

MIT License

## 📚 Detailed Documentation

For more detailed documentation, examples, and advanced usage, please check the [v2 documentation](./v2/README.md).
