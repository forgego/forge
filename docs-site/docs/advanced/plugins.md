---
sidebar_position: 2
---

# Plugins

forge has a plugin system that allows you to extend and customize framework behavior.

## Plugin Interface

Plugins implement the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Install() error
}
```

## Creating a Plugin

### Basic Plugin

```go
package plugins

import (
    "github.com/forgego/forge/registry"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "0.1.0"
}

func (p *MyPlugin) Install() error {
    // Plugin initialization logic
    return nil
}
```

### Registering a Plugin

```go
import (
    "github.com/forgego/forge/registry"
    "myapp/plugins"
)

func main() {
    if err := registry.RegisterPlugin(&plugins.MyPlugin{}); err != nil {
        log.Fatal(err)
    }
}
```

## Plugin Hooks

Plugins can hook into various framework events:

### Model Registration

```go
type AuditModelPlugin struct{}

func (p *AuditModelPlugin) Name() string {
    return "audit"
}

func (p *AuditModelPlugin) Version() string {
    return "0.1.0"
}

func (p *AuditModelPlugin) Install() error {
    return nil
}

func (p *AuditModelPlugin) ExtendModel(model interface{}) error {
    log.Printf("Model registered: %T", model)
    return nil
}

func (p *AuditModelPlugin) GetModelFields(modelName string) []interface{} {
    return nil
}

func (p *AuditModelPlugin) GetModelRelations(modelName string) []interface{} {
    return nil
}

func (p *AuditModelPlugin) GetModelHooks(modelName string) interface{} {
    return nil
}
```

### Before Request

```go
func (p *MyPlugin) Initialize(app *registry.Application) error {
    app.OnBeforeRequest(func(w http.ResponseWriter, r *http.Request) {
        // Called before each request
    })
    return nil
}
```

### After Request

```go
func (p *MyPlugin) Initialize(app *registry.Application) error {
    app.OnAfterRequest(func(w http.ResponseWriter, r *http.Request) {
        // Called after each request
    })
    return nil
}
```

## Example Plugins

### Logging Plugin

```go
type LoggingPlugin struct{}

func (p *LoggingPlugin) Name() string {
    return "logging"
}

func (p *LoggingPlugin) Initialize(app *registry.Application) error {
    app.OnBeforeRequest(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
    })
    return nil
}
```

### Caching Plugin

```go
type CachingPlugin struct {
    cache cache.Cache
}

func (p *CachingPlugin) Name() string {
    return "caching"
}

func (p *CachingPlugin) Initialize(app *registry.Application) error {
    p.cache = cache.NewRedisCache(redisClient)
    
    app.OnModelRegister(func(model schema.Schema) {
        // Add caching to models
    })
    
    return nil
}
```

## Plugin Configuration

Plugins can have configuration:

```go
type MyPlugin struct {
    config Config
}

type Config struct {
    Enabled bool
    Option  string
}

func (p *MyPlugin) Initialize(app *registry.Application) error {
    // Use configuration
    if !p.config.Enabled {
        return nil
    }
    
    // Plugin logic
    return nil
}
```

## Best Practices

1. **Keep Plugins Focused** - Each plugin should do one thing well
2. **Use Hooks** - Use framework hooks instead of modifying core code
3. **Handle Errors** - Always return errors from Initialize
4. **Document Plugins** - Document what your plugin does
5. **Test Plugins** - Write tests for your plugins

## See Also

- [Code Generation](/docs/advanced/code-generation) - Code generation system
- [Custom Fields](/docs/advanced/custom-fields) - Create custom field types

