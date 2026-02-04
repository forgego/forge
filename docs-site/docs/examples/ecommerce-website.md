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
