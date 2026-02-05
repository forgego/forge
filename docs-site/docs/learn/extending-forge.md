---
sidebar_position: 5
---

# Extending forge

forge is designed to be fully extensible. You can extend almost every part of the framework to fit your needs.

## Why extend forge?

You might want to extend forge to:

- Add custom field types for your domain
- Create reusable components
- Integrate with third-party services
- Customize admin interface
- Add custom validators
- Build plugins

## Extension points

forge provides several extension points:

### 1. Custom field types

Create custom field types for specialized data:

```go
type ColorFieldBuilder struct {
    *schema.BaseFieldBuilder
}

func Color(name string) *ColorFieldBuilder {
    return &ColorFieldBuilder{
        BaseFieldBuilder: &schema.BaseFieldBuilder{
            field: schema.Field{
                Name: name,
                Type: schema.TypeString,
            },
        },
    }
}

func (b *ColorFieldBuilder) Build() schema.Field {
    // Add custom validation
    b.field.Validators = append(b.field.Validators, &ColorValidator{})
    return b.field
}
```

### 2. Custom validators

Add custom validation logic:

```go
type PhoneValidator struct{}

func (v *PhoneValidator) Validate(value interface{}) error {
    str, ok := value.(string)
    if !ok {
        return fmt.Errorf("expected string")
    }
    
    // Validate phone format
    matched, _ := regexp.MatchString(`^\+?[1-9]\d{1,14}$`, str)
    if !matched {
        return fmt.Errorf("invalid phone number format")
    }
    
    return nil
}

// Usage
schema.String("phone").
    Required().
    Validators(&PhoneValidator{}).
    Build()
```

### 3. Custom admin widgets

Create custom form widgets:

```go
type RichTextWidget struct {
    *admin.BaseWidget
}

func (w *RichTextWidget) Render(field *admin.Field, value interface{}) string {
    // Render rich text editor
    return fmt.Sprintf(`<textarea class="rich-text">%v</textarea>`, value)
}

// Register widget
admin.RegisterWidget("richtext", &RichTextWidget{})
```

### 4. Custom QuerySet methods

Add custom query methods:

```go
func (qs *PostQuerySet) Published(ctx context.Context) ([]*Post, error) {
    return qs.Filter(Post.Fields.Published.Equals(true)).All(ctx)
}

func (qs *PostQuerySet) ByAuthor(ctx context.Context, authorID int64) ([]*Post, error) {
    return qs.Filter(Post.Fields.AuthorID.Equals(authorID)).All(ctx)
}
```

### 5. Custom middleware

Create custom middleware:

```go
func CustomMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your middleware logic
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

// Use it
router.Use(CustomMiddleware)
```

### 6. Plugin system

Build reusable plugins:

```go
package myplugin

import "github.com/forgego/forge/pkg/registry"

func init() {
    registry.RegisterPlugin(&MyPlugin{})
}

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "myplugin"
}

func (p *MyPlugin) Initialize(app *App) error {
    // Plugin initialization
    return nil
}
```

## Extension patterns

### Builder pattern

Use builder pattern for field types:

```go
type CustomFieldBuilder struct {
    *schema.BaseFieldBuilder
    customOption string
}

func (b *CustomFieldBuilder) CustomOption(value string) *CustomFieldBuilder {
    b.customOption = value
    return b
}
```

### Strategy pattern

Use strategy pattern for interchangeable components:

```go
type AuthenticationStrategy interface {
    Authenticate(r *http.Request) (*User, error)
}

type TokenAuth struct{}
type JWTAuth struct{}
type SessionAuth struct{}
```

### Factory pattern

Use factory pattern for creating instances:

```go
type WidgetFactory interface {
    CreateWidget(widgetType string) Widget
}

func NewWidgetFactory() WidgetFactory {
    return &DefaultWidgetFactory{}
}
```

## Best practices

### 1. Keep extensions focused

Each extension should do one thing well:

```go
// ✅ Good: Focused validator
type EmailValidator struct{}

// ❌ Bad: Does too much
type EmailAndPhoneAndAddressValidator struct{}
```

### 2. Document your extensions

Document how to use your extensions:

```go
// EmailValidator validates email addresses.
// Usage:
//   schema.String("email").Validators(&EmailValidator{}).Build()
type EmailValidator struct{}
```

### 3. Test your extensions

Write tests for your extensions:

```go
func TestEmailValidator(t *testing.T) {
    v := &EmailValidator{}
    
    err := v.Validate("test@example.com")
    assert.NoError(t, err)
    
    err = v.Validate("invalid")
    assert.Error(t, err)
}
```

### 4. Follow forge patterns

Match forge's existing patterns:

- Use builder pattern for field types
- Use interfaces for extensibility
- Follow naming conventions
- Match code style

## Example: Complete extension

Here's a complete example of extending forge:

```go
package myfields

import (
    "github.com/forgego/forge/pkg/schema"
    "github.com/forgego/forge/pkg/schema/validators"
)

// PhoneField creates a phone number field with validation
func PhoneField(name string) *PhoneFieldBuilder {
    return &PhoneFieldBuilder{
        BaseFieldBuilder: &schema.BaseFieldBuilder{
            field: schema.Field{
                Name: name,
                Type: schema.TypeString,
            },
        },
    }
}

type PhoneFieldBuilder struct {
    *schema.BaseFieldBuilder
}

func (b *PhoneFieldBuilder) Required() *PhoneFieldBuilder {
    b.field.Required = true
    return b
}

func (b *PhoneFieldBuilder) Build() schema.Field {
    // Add phone validation
    b.field.Validators = append(b.field.Validators, &PhoneValidator{})
    return b.field
}

type PhoneValidator struct{}

func (v *PhoneValidator) Validate(value interface{}) error {
    str, ok := value.(string)
    if !ok {
        return fmt.Errorf("expected string")
    }
    
    matched, _ := regexp.MatchString(`^\+?[1-9]\d{1,14}$`, str)
    if !matched {
        return fmt.Errorf("invalid phone number format")
    }
    
    return nil
}

// Usage
func (User) Fields() []schema.Field {
    return []schema.Field{
        PhoneField("phone").Required().Build(),
    }
}
```

## Next steps

- **[Custom fields](/docs/advanced/custom-fields)** - Create custom field types
- **[Plugins](/docs/advanced/plugins)** - Build plugins
- **[API Reference](/docs/api-reference/schema)** - Complete API documentation
