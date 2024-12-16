# Logger Package

## Overview

This package provides a structured JSON logging solution that wraps `log/slog` and optionally integrates with Sentry for error tracking and monitoring. The package is designed for convenience and flexibility, allowing developers to easily configure and use the logger with minimal boilerplate.

## Features

- Structured JSON logging using `log/slog`.
- Optional integration with Sentry for error tracking.
- Simple configuration using a `Config` struct.
- Method chaining for convenient logging.
- Instance-Specific Configuration: Each logger instance is initialized with its own Config struct, which includes the Sentry configuration. This ensures that each instance operates independently.
- Concurrency Safety: Since there is no shared mutable state, the logger instances can be used concurrently without risking race conditions.

## Installation

To install the package, use:

```sh
go get github.com/stratastor/logger
```

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "github.com/stratastor/logger"
)

func main() {
    config := logger.Config{
        LogLevel:    "info",
        SentryDSN:   "",
        EnableSentry: false,
    }

    log, err := logger.New(config)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }

    log.Info("This is an info message")
}
```

### Using Tags

```go
package main

import (
    "log"
    "github.com/stratastor/logger"
)

func main() {
    config := logger.Config{
        LogLevel:    "info",
        SentryDSN:   "",
        EnableSentry: false,
    }

    log, err := logger.NewTag(config, "example")
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }

    log.Info("This is an info message with a tag")
}
```

### Enabling Sentry

```go
package main

import (
    "log"
    "github.com/stratastor/logger"
)

func main() {
    config := logger.Config{
        LogLevel:    "error",
        SentryDSN:   "your-sentry-dsn",
        EnableSentry: true,
    }

    log, err := logger.New(config)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }

    log.Error("This is an error message")
}
```

### Concurency safe usage

```go
package main

import (
    "log"
    "sync"

    "github.com/stratastor/logger"
)

func main() {
    config1 := logger.Config{
        LogLevel:    "info",
        SentryDSN:   "dsn1",
        EnableSentry: true,
    }

    config2 := logger.Config{
        LogLevel:    "debug",
        SentryDSN:   "",
        EnableSentry: false,
    }

    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        log1, err := logger.NewTag(config1, "tag1")
        if err != nil {
            log.Fatalf("Failed to initialize logger1: %v", err)
        }
        log1.Info("This is an info message from logger1")
    }()

    go func() {
        defer wg.Done()
        log2, err := logger.NewTag(config2, "tag2")
        if err != nil {
            log.Fatalf("Failed to initialize logger2: %v", err)
        }
        log2.Debug("This is a debug message from logger2")
    }()

    wg.Wait()
}
```

## Design Decisions

### Direct Alias for `*slog.Logger`

The package uses `type Logger = *slog.Logger` to directly alias `*slog.Logger`. This decision was made to:

- Simplify the codebase by avoiding the need to wrap all methods of `slog.Logger`.
- Allow method chaining directly on the Logger instance.
- Provide convenience and flexibility for developers.

### Configuration Struct

A `Config` struct is used to encapsulate configuration options such as log level, Sentry DSN, and whether to enable Sentry. This approach makes the logger easy to configure and extend in the future.

### Sentry Integration

The package provides optional integration with Sentry for error tracking. If Sentry is enabled and configured, log messages with a level of `warn` or higher are sent to Sentry. This integration helps in monitoring and tracking errors in production environments.


### Concurency Safetey

By managing the Sentry state within the Logger instance or configuration and avoiding global state, the logger package is concurrency-safe. Each logger instance is configured independently, ensuring that concurrent usage does not lead to race conditions or inconsistent behavior.

## Conclusion

This logger package provides a convenient and flexible solution for structured JSON logging with optional Sentry integration. By leveraging `log/slog` and providing a simple configuration interface, the package aims to make logging easy and effective for developers.
