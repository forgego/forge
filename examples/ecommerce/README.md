# Forge ecommerce example

This is a reference store for Forge's generated models, REST resources, admin UI, and a few custom catalog and order endpoints. It is a development example. The `/api/v1/` model resources allow writes without authentication; the admin API requires an admin login. Add API permissions before exposing this app to untrusted clients.

## Run locally

From the repository root, build the local CLI and put it on `PATH`:

```bash
GOTOOLCHAIN=go1.26.10 go build -o /tmp/forge ./cmd/forge
export PATH=/tmp:$PATH
cd examples/ecommerce
export GOTOOLCHAIN=go1.26.10
```

The default `config/config.yaml` uses SQLite and creates `./ecommerce.sqlite` on first use. To use PostgreSQL, create an empty database first and set the connection through environment variables. For example, with a PostgreSQL database named `ecom` listening on local port `55433`:

```bash
export FORGE_DATABASE_DRIVER=postgres
export FORGE_DATABASE_HOST=127.0.0.1
export FORGE_DATABASE_PORT=55433
export FORGE_DATABASE_USER=ecom
export FORGE_DATABASE_PASSWORD='<your database password>'
export FORGE_DATABASE_NAME=ecom
```

Then install dependencies, generate model code and a migration for the selected driver, apply it, and build the admin UI:

```bash
make install
make generate
make makemigrations
make migrate
make admin-ui
```

`make makemigrations` creates SQL under `migrations/` using the current database driver. Review the SQL before applying it. The server also runs `SetupSchema` at startup for the example's models. Do not reuse a PostgreSQL migration against SQLite, or vice versa.

Pick an unused local port (`ss -ltn` shows listeners), then run the server:

```bash
export FORGE_SERVER_PORT=8020
make run-go
```

The homepage is at `/`, the admin UI at `/admin/`, and the OpenAPI document at `/api/openapi.json`. `/api/v1/` is a route prefix, not an index or documentation page. The admin login uses the credentials configured under `admin` in `config/config.yaml`, or `FORGE_ADMIN_USERNAME` and `FORGE_ADMIN_PASSWORD`. `forge createsuperuser` manages Forge identity users in a separate `users` table; it does not create this example's admin login.

For a standalone binary with the built UI embedded, run `make build-embed` and then `./ecommerce` with the same database and port settings.

## Exercise the REST API

Set `BASE` to the port selected above. Collection URLs end in `/`; detail URLs use the ID without a trailing `/`.

```bash
BASE=http://localhost:8020
curl "$BASE/api/v1/categories/"
curl "$BASE/api/v1/products/"
curl "$BASE/api/v1/orders/"
curl "$BASE/api/openapi.json"
```

Create a category, then use its returned ID when creating a product. Product `slug` and `sku` are required.

```bash
curl -X POST "$BASE/api/v1/categories/" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Electronics","slug":"electronics"}'

curl -X POST "$BASE/api/v1/products/" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Laptop","slug":"laptop","sku":"LAP-001","description":"Laptop","price":999.99,"category_id":1}'

curl "$BASE/api/v1/products/1"
curl "$BASE/api/v1/products/?category_id=1&price__gte=100"
curl "$BASE/api/v1/products/?search=laptop&ordering=-created_at"
curl -X PATCH "$BASE/api/v1/products/1" \
  -H 'Content-Type: application/json' -d '{"price":899.99}'
curl -X DELETE "$BASE/api/v1/products/1"
```

Use the IDs returned by your own requests; they may differ from `1`. Direct field filters such as `category_id`, `price__gte`, `status`, and `status__in` work on the model resources. Deep relation filters such as `category__name` and `customer__email` are not implemented by these resources. The custom catalog endpoint supports `q`, `category_id`, `brand_id`, `min_price`, `max_price`, and `in_stock`:

```bash
curl "$BASE/api/v1/catalog/search?q=lap&min_price=100&in_stock=true"
curl "$BASE/api/v1/catalog/stats"
```

An order resource requires an existing customer and the order's required snapshot and total fields. `items` is not accepted by the ordinary `/orders/` create route. For an order with line items, use the checkout endpoint with an existing customer and product ID:

```bash
curl -X POST "$BASE/api/v1/orders/checkout" \
  -H 'Content-Type: application/json' \
  -d '{"customer_id":1,"items":[{"product_id":1,"quantity":2,"unit_price":99.99}]}'
curl "$BASE/api/v1/orders/?status=pending"
curl "$BASE/api/v1/orders/summary"
```

## Admin API and UI tests

The admin API returns 401 without authentication. POST to `/admin/api/login` with the configured admin username and password, then send its `token` as a Bearer token. For example:

```bash
curl -X POST "$BASE/admin/api/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"<admin username>","password":"<admin password>"}'
curl "$BASE/admin/api/categories" -H 'Authorization: Bearer <token>'
```

The admin UI has model list, create, edit, detail, search, export, and bulk action views. The example registers models across `catalog`, `customers`, `orders`, `inventory`, `marketing`, `commerce`, `engagement`, `promotions`, `support`, and `users`.

Run the Go and browser suites from this directory:

```bash
cd ui-tests && npm ci && npx playwright install chromium && cd ..
make test
make ui-test
```

The browser suite starts the server if the configured port is free. It can reuse a server already running on that port. Build the admin UI first with `make admin-ui`. Set `FORGE_SERVER_PORT` and `FORGE_SERVER_BASE_URL` together when using a port other than `8020`.
