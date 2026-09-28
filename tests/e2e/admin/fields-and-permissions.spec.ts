import { test, expect, openList } from './support/fixtures';
import { adminURL } from './support/env';

test.describe('@core read-only and auto-managed fields', () => {
  test('hides them on create and disables them on edit', async ({ adminPage: page, api, tag }) => {
    // created_at/updated_at are auto-managed by the database; level is
    // listed in the Category admin's ReadOnlyFields.
    const meta = await (await api.raw('GET', '/meta/categories')).json();
    const readOnly = Object.fromEntries(meta.fields.map((f: any) => [f.name, f.read_only]));
    expect(readOnly).toMatchObject({ created_at: true, updated_at: true, level: true, name: false });

    await page.goto(adminURL('/categories/create'));
    await expect(page.locator('#name')).toBeEditable();
    for (const name of ['created_at', 'updated_at', 'level']) {
      await expect(page.locator(`#${name}`)).toHaveCount(0);
    }

    const row = await api.createCategory(tag);
    await page.goto(adminURL(`/categories/${row.id}`));
    await expect(page.locator('#name')).toBeEditable();
    for (const name of ['created_at', 'updated_at', 'level']) {
      await expect(page.locator(`#${name}`)).toBeDisabled();
    }
  });

  test('ignores writes to them and keeps the stored values', async ({ api, tag }) => {
    const row = await api.createCategory(tag);
    const before = await api.get('categories', row.id);

    const response = await api.raw('PATCH', `/categories/${row.id}`, {
      level: 9,
      created_at: '2001-01-01T00:00:00Z',
      name: `${tag} renamed`,
    });
    expect(response.status()).toBe(200);

    const after = await api.get('categories', row.id);
    expect(after.name).toBe(`${tag} renamed`);
    expect(after.level).toBe(before.level);
    // The ecommerce models type timestamps as strings, which read back
    // empty through both APIs, so created_at cannot be compared here.
  });

  test('rejects an unknown field with a validation error instead of a server error', async ({ api, tag }) => {
    const response = await api.raw('POST', '/categories', {
      name: `${tag} Unknown`,
      slug: `${tag}-unknown`,
      not_a_field: 1,
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.error.code).toBe('validation_error');
    expect(body.error.details.not_a_field).toEqual(['unknown field']);
    expect((await api.list('categories', { search: tag })).count).toBe(0);
  });
});

test.describe('@core object permissions', () => {
  test('denies editing a protected record and leaves it unchanged', async ({ adminPage: page, api, tag }) => {
    // Warehouse admin: a primary warehouse may not be changed in the admin.
    const warehouse = await api.createWarehouse(tag, { is_active: true });
    await api.patch('warehouses', warehouse.id, { is_primary: true });

    await page.goto(adminURL(`/warehouses/${warehouse.id}`));
    await expect(page.locator('#name')).toHaveValue(`${tag} Warehouse`);
    await page.locator('#name').fill(`${tag} Hijacked`);
    await page.getByTestId('submit-button').click();

    await expect(page.getByText("You don't have permission to change this object")).toBeVisible();
    await expect(page).toHaveURL(new RegExp(`${adminURL(`/warehouses/${warehouse.id}`)}$`));

    const stored = await api.get('warehouses', warehouse.id);
    expect(stored).toMatchObject({ name: `${tag} Warehouse`, is_primary: true });
  });

  test('hides delete controls where deletion is not permitted and refuses direct deletes', async ({ adminPage: page, api }) => {
    // Order admin: orders are never deleted, only cancelled.
    const meta = await (await api.raw('GET', '/meta/orders')).json();
    expect(meta.permissions).toMatchObject({ view: true, delete: false });

    await openList(page, 'orders');
    await expect(page.getByTestId('bulk-toolbar')).toBeVisible();
    await expect(page.getByTestId('bulk-delete-button')).toHaveCount(0);
    await expect(page.locator('[data-testid^="delete-"]')).toHaveCount(0);

    const response = await api.raw('DELETE', '/orders/1');
    expect(response.status()).toBe(403);
    const bulk = await api.raw('DELETE', '/orders/bulk-delete', { ids: [1] });
    expect(bulk.status()).toBe(403);
  });
});
