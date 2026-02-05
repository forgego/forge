---
sidebar_position: 1
description: forge framework development roadmap. See what's coming next and help shape the future of the project.
keywords:
  - forge roadmap
  - development roadmap
  - forge future
  - django go roadmap
  - framework roadmap
image: /forge-social-card.svg
---

# Roadmap

Here's what we're planning for forge. This roadmap shows our priorities and timeline for future development.

## Current Status: MVP Complete ✅

forge is production-ready with all core features implemented. You can build real applications today.

## Q1 2025: Advanced ORM

### Priority: High
These features will make the ORM complete:

#### SelectRelated & PrefetchRelated
```go
// Instead of N+1 queries
users, err := User.Objects.SelectRelated("profile").All(ctx)

// Instead of separate queries
posts, err := Post.Objects.PrefetchRelated("comments").All(ctx)
```

#### Aggregates & Annotations
```go
// Count posts per user
users, err := User.Objects.Annotate("post_count", 
    Count("posts")).All(ctx)

// Average rating per category
categories, err := Category.Objects.Annotate("avg_rating",
    Avg("posts__rating")).All(ctx)
```

#### F() Expressions
```go
// Update based on other fields
err := Product.Objects.Filter(ID.Equals(1)).Update(map[string]interface{}{
    "price": F("cost") * 1.5,
})

// Complex expressions
err := Post.Objects.Update(map[string]interface{}{
    "view_count": F("view_count") + 1,
})
```

#### Subqueries
```go
// Posts with more than 10 comments
posts, err := Post.Objects.Filter(
    CommentCount.GT(
        Subquery(Comment.Objects.Filter(PostID.EqualsOuterRef("id")).Count())
    ),
).All(ctx)
```

### Timeline
- **January**: SelectRelated implementation
- **February**: PrefetchRelated implementation  
- **March**: Aggregates, Annotations, F() expressions

## Q2 2025: Performance & Scaling

### Priority: High
Make forge ready for high-traffic applications:

#### Query Optimization
- Automatic query optimization
- Query plan analysis
- N+1 detection and prevention
- Index suggestions

#### Caching Layer
```go
// Cache query results
posts, err := Post.Objects.Cache("popular_posts", time.Hour).
    Filter(Published.Equals(true)).
    OrderBy("-view_count").
    Limit(10).
    All(ctx)
```

#### Connection Pooling
- Advanced connection pool configuration
- Connection health monitoring
- Automatic failover
- Load balancing

### Timeline
- **April**: Query optimization
- **May**: Caching layer
- **June**: Connection pooling improvements

## Q3 2025: Developer Experience

### Priority: Medium
Make development even better:

#### Hot Reload
```bash
forge runserver --reload
# Automatically restarts when files change
```

#### Debug Toolbar
- Request/response inspection
- SQL query analysis
- Performance profiling
- Template context viewer

#### Better Error Messages
- Contextual error messages
- Suggestions for fixes
- Stack trace improvements
- Development vs production errors

#### Development Tools
```bash
forge shell          # Interactive shell
forge dbshell         # Database shell
forge test            # Run tests
forge benchmark       # Performance tests
```

### Timeline
- **July**: Hot reload and debug toolbar
- **August**: Error message improvements
- **September**: Development tools

## Q4 2025: Advanced Features

### Priority: Medium
Advanced features for complex applications:

#### GraphQL Support
```go
type Query struct {
    posts []Post `graphql:"posts"`
    users []User `graphql:"users"`
}

// Auto-generated GraphQL schema
// Type-safe resolvers
// Query optimization
```

#### WebSocket Support
```go
// Real-time features
ws.Handle("/ws/posts", PostWebSocketHandler)
// Automatic connection management
// Message broadcasting
// Room support
```

#### Background Tasks
```go
// Async task processing
tasks.Enqueue(&SendEmailTask{UserID: 123})
// Task scheduling
// Retry logic
// Task monitoring
```

### Timeline
- **October**: GraphQL support
- **November**: WebSocket support
- **December**: Background tasks

## 2026: Enterprise Features

### Priority: Low
Features for large organizations:

#### Multi-Tenancy
```go
// Tenant-aware models
type TenantModel struct {
    schema.BaseSchema
    TenantID int64 `db:"tenant_id"`
}

// Automatic tenant filtering
// Tenant isolation
// Tenant management
```

#### Audit Logging
```go
// Automatic audit trails
type AuditLog struct {
    Action     string
    Model      string
    ObjectID   int64
    Changes    JSON
    UserID     int64
    Timestamp  time.Time
}
```

#### Advanced Monitoring
- Metrics collection
- Distributed tracing
- Health checks
- Performance monitoring

### Timeline
- **Q1 2026**: Multi-tenancy
- **Q2 2026**: Audit logging
- **Q3 2026**: Advanced monitoring

## Ongoing Improvements

These happen continuously:

### Documentation
- More examples and tutorials
- API reference improvements
- Video tutorials
- Community contributions

### Performance
- Benchmark improvements
- Memory usage optimization
- Startup time reduction
- Database query optimization

### Ecosystem
- Third-party integrations
- Plugin marketplace
- Community packages
- Tooling support

## How We Prioritize

### Factors We Consider
1. **User Feedback** - What do users actually need?
2. **Django Parity** - What Django features are missing?
3. **Performance Impact** - Will it make forge faster?
4. **Developer Experience** - Will it make development easier?
5. **Maintenance Cost** - Can we maintain this long-term?

### Priority Levels
- **High** - Core functionality that many users need
- **Medium** - Important features for specific use cases
- **Low** - Nice-to-have features for edge cases

## Get Involved

### Contribute Code
- Pick an issue from the roadmap
- Submit a pull request
- Write tests
- Update documentation

### Provide Feedback
- Open issues for missing features
- Vote on existing issues
- Share your use cases
- Suggest improvements

### Spread the Word
- Blog about your forge projects
- Share on social media
- Present at meetups
- Help others learn

## Timeline Summary

```
2024: ✅ MVP Complete (Production Ready)
2025:
  Q1: Advanced ORM (SelectRelated, Aggregates, F() expressions)
  Q2: Performance & Scaling (Caching, Query Optimization)
  Q3: Developer Experience (Hot Reload, Debug Toolbar)
  Q4: Advanced Features (GraphQL, WebSockets)
2026:
  Q1-Q3: Enterprise Features (Multi-tenancy, Audit Logging)
```

## Reality Check

This roadmap is ambitious but realistic. We:

- **Focus on quality** over features
- **Listen to users** for priorities
- **Maintain stability** while adding features
- **Keep it simple** and avoid over-engineering

Some features might move around based on:
- User demand
- Technical challenges
- Resource availability
- Community contributions

## What Won't We Do

### Out of Scope
- **NoSQL databases** - Focus on PostgreSQL
- **Microservices framework** - Stay focused on monoliths
- **Frontend framework** - Let users choose their frontend
- **Cloud platform** - Focus on the framework itself

### Why Not?
- **Scope creep** - We want to stay focused
- **Maintenance burden** - More features = more maintenance
- **User choice** - Let users pick their own tools
- **Expertise** - Focus on what we do best

## Stay Updated

- **GitHub** - Watch releases and issues
- **Blog** - Follow development updates
- **Twitter/X** - Quick updates and announcements
- **Discord** - Chat with the community

## Bottom Line

forge is already production-ready. This roadmap is about making it even better - more features, better performance, and an even greater developer experience.

The core framework is solid. Everything from here is about making forge the best Go web framework it can be.
