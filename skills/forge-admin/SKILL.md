---
name: forge-admin
description: Set up and customize Forge's auto-generated admin interface. Use when registering models with admin, configuring list displays, search, filters, fieldsets, or wiring admin routes.
---

# Forge Admin

## Overview
Register models with the admin system to get a full CRUD web UI. This skill focuses on admin registration, configuration, and routing.

## When to Use
- The task mentions admin registration, list display, search, filters, or admin routes.
- You need CRUD admin screens for Forge models.
- The user asks about `/admin` setup or admin UI behavior.

## Quick Start
1. Register your model with admin using the admin core registry.
2. Configure list display, search, filters, and ordering.
3. Register admin routes under an `/admin` path and start the server.

## Common Tasks
- Set `ListDisplay`, `SearchFields`, `ListFilter`, and `Ordering`.
- Configure read-only fields and form layout with fieldsets.
- Register multiple models and expose them under a shared admin router.

## Gotchas
- Admin requires schema definitions and managers for models.
- Admin routes must be mounted on your main router to be accessible.
- Changes to models often require re-running `forge generate` and migrations.

## References
- [Admin configuration options](references/admin.md)
- [Admin tutorial](references/admin-tutorial.md)
