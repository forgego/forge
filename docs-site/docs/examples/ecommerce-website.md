---
sidebar_position: 4
---

# Ecommerce Website Example

Use the ecommerce sample in `examples/ecommerce` to explore a full storefront, admin, and API setup.

## Run the sample

1. Navigate to the example:
   
   ```bash
   cd examples/ecommerce
   ```
2. Install dependencies:
   
   ```bash
   make install
   ```
3. Create the database and configure `config/config.yaml` as needed.
   
   ```bash
   make db-create
   ```
4. Generate code and run migrations:
   
   ```bash
   make generate
   make migrate
   ```
5. Create a superuser for the admin:
   
   ```bash
   make superuser
   ```
6. Start the server:
   
   ```bash
   make run
   ```

The app is now available at http://localhost:8000/.【F:examples/ecommerce/SETUP.md†L9-L95】

## Admin access

- **Admin URL:** http://localhost:8000/admin/【F:examples/ecommerce/SETUP.md†L86-L95】
- **Credentials:** There are no default credentials. Create a superuser with `make superuser` (or `forge createsuperuser`) and log in with the account you created.【F:examples/ecommerce/SETUP.md†L66-L95】

## Main user flows to try

- Create categories and products (including variants) in the admin.
- Add a customer and place a test order.
- Add a review to exercise the marketing flow.

These flows are called out in the setup guide’s “Try” section after logging into the admin.【F:examples/ecommerce/SETUP.md†L111-L131】

## Related documentation

- [Admin guide](/docs/guides/admin)
- [Migrations guide](/docs/guides/migrations)
- [ORM system overview](/docs/features/orm-system)
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
