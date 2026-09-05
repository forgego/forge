# Admin UI Verification Checklist (Ecommerce Example)

This checklist is for manual verification of the ecommerce admin UI using the local SQLite runtime.

## Setup (SQLite)
1. Open a terminal in `examples/ecommerce`.
2. Confirm `config/config.yaml` has `database.driver: sqlite` and a valid `sqlite_path`.
3. Generate code (if needed):
   - `./forge.exe generate`
4. Create migrations (if none exist):
   - `./forge.exe makemigrations init --auto`
5. Apply migrations:
   - `./forge.exe migrate`
6. Create a superuser:
   - `./forge.exe createsuperuser`
7. Start the server:
   - `./forge.exe runserver`
8. Open the admin UI:
   - `http://localhost:8000/admin/`

## Login
- Log in with the superuser account.
- Expected: dashboard loads, no 500 errors, admin sidebar visible.

## Catalog
- Categories list view loads.
  - Expected: filters for `is_active`, `parent_id`, `level`, search works.
- Create a Category with a parent.
  - Expected: hierarchy fields save and appear in list.
- Products list view loads.
  - Expected: filters for `is_active`, `is_featured`, `category_id`, `brand_id`.
- Create a Product with at least one Variant and Image inline.
  - Expected: inline rows save; product shows related variants/images.
- Use bulk actions:
  - Activate/Deactivate, Feature/Unfeature products.

## Customers
- Customers list view loads.
  - Expected: filters for active/verified/marketing/group.
- Create a Customer with Address and WishList inline.
  - Expected: inline address and wishlist items save.

## Orders
- Orders list view loads.
  - Expected: filters for status, payment, fulfillment.
- Create an Order with OrderItems inline.
  - Expected: totals and status fields persist.
- Use `Mark Delivered` action.
  - Expected: status updates and `delivered_at` timestamp set.

## Inventory
- Stock list view loads.
  - Expected: filters for low stock, warehouse.
- Create a Stock record and a StockMovement inline (if present).

## Marketing / Promotions
- Coupons list view loads.
  - Expected: filters for active, date ranges.
- Create a Coupon, ensure validation on required fields.

## Support / Engagement
- Support ticket list view loads (if configured).
- Create a ticket and update status via actions.

## Validation & Errors
- Attempt saving a form without required fields.
  - Expected: validation errors shown, no server crash.

## Notes
Record any failures with:
- model name
- action performed
- expected vs actual result
- console or server error output
