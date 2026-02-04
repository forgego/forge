---
sidebar_position: 3
---

# E-commerce Website Sample

This page walks through running the ecommerce sample locally, the main entry points, and a quick UI walkthrough.

## Run the sample locally

Use the ecommerce example under `examples/ecommerce`.

1. **Install dependencies**

   ```bash
   cd examples/ecommerce
   make install
   # or: go mod download
   ```

2. **Create the database**

   ```bash
   make db-create
   # or: createdb ecommerce_db
   ```

3. **Generate code + run migrations**

   ```bash
   make generate
   make migrate
   # or: forge generate && forge migrate
   ```

4. **Create a superuser (admin login)**

   ```bash
   make superuser
   # or: forge createsuperuser
   ```

5. **Start the server**

   ```bash
   make run
   # or: forge runserver
   ```

> Tip: If you want the quick all-in-one flow, `make setup` covers install, generate, migrate, and superuser creation. Use `make run` afterward.

## Entry points

Once the server is running, use these URLs:

- **Homepage**: `http://localhost:8000/`
- **Admin UI**: `http://localhost:8000/admin/`
- **REST API**: `http://localhost:8000/api/v1/`

### Default credentials and seed data

- There are **no default credentials** checked into the repo. You create the admin user with `forge createsuperuser` or `make superuser`.
- There is **no preloaded seed data** by default. You can:
  - Add records manually in the admin UI.
  - Create your own seed script (the setup guide suggests `scripts/seed.go` and `make seed`).

## UI walkthrough (text)

If you do not have screenshots available locally, use the following walkthrough to confirm the main pages:

1. **Homepage** (`/`)
   - The storefront landing page. Start the server and verify the page loads without error.
2. **Admin UI** (`/admin/`)
   - Log in with the superuser account you created.
   - Explore catalog, customers, orders, inventory, and marketing sections.
3. **REST API** (`/api/v1/`)
   - Confirm the browsable API (or JSON response) for endpoints like `/api/v1/products/` and `/api/v1/orders/`.

## Related docs

- [E-commerce Example Overview](/docs/examples/ecommerce)
