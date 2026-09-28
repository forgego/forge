import { test, expect } from '@playwright/test';

// Admin CRUD smoke flow. Requires a running backend for the dev proxy
// (vite `server.proxy` -> http://localhost:8080) with a `categories`
// model (name/slug required, is_active boolean, activate action).
// NOTE: the ecommerce reference app is covered more thoroughly by
// examples/ecommerce/ui-tests/tests/admin-redesign.spec.ts against the
// production bundle; this spec keeps the vite-dev path honest.
const MODEL = process.env.FORGE_E2E_MODEL || 'categories';

test.describe('Admin CRUD Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.removeItem('admin_token');
    });
  });

  test('login → navigate → create → edit → bulk action → filter → delete', async ({ page }) => {
    // 1) Login
    await page.goto('/login');
    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'admin123');
    await page.click('[data-testid="login-button"]');

    await expect(page).toHaveURL('/');
    await expect(page.getByTestId('nav-dashboard')).toBeVisible();

    // 2) Navigate to the model via sidebar
    await page.click(`[data-testid="nav-${MODEL}"]`);
    await expect(page).toHaveURL(new RegExp(`/${MODEL}$`));

    // List loads
    await expect(page.locator('table')).toBeVisible();

    // 3) Create new
    await page.click('[data-testid="create-button"]');
    await expect(page).toHaveURL(new RegExp(`/${MODEL}/create$`));

    const unique = Date.now();
    const createdName = `E2E Category ${unique}`;

    await page.fill('#name', createdName);
    await page.fill('#slug', `e2e-${unique}`);
    await page.click('[data-testid="submit-button"]');

    await expect(page).toHaveURL(new RegExp(`/${MODEL}$`));
    await expect(page.locator('table')).toContainText(createdName);

    // Capture new id from the row's edit button (actions are always
    // present in the DOM; no hover needed).
    const row = page.locator('tr', { hasText: createdName });
    const editBtn = row.locator('[data-testid^="edit-"]');
    const editTestId = (await editBtn.getAttribute('data-testid')) || '';
    const id = editTestId.replace('edit-', '');
    expect(id).not.toBe('');

    // 4) Edit
    await editBtn.click();
    await expect(page).toHaveURL(new RegExp(`/${MODEL}/${id}$`));

    const updatedName = `${createdName} Updated`;
    await page.fill('#name', updatedName);
    await page.click('[data-testid="submit-button"]');

    await expect(page).toHaveURL(new RegExp(`/${MODEL}$`));
    await expect(page.locator('table')).toContainText(updatedName);

    // 5) Bulk action (activate requires confirmation in the reference app)
    await page.click(`[data-testid="select-${id}"]`);
    await expect(page.getByTestId('bulk-toolbar')).toContainText('1 selected');
    await page.click('[data-testid="bulk-action-activate"]');
    const actionDialog = page.getByTestId('bulk-action-dialog');
    if (await actionDialog.isVisible()) {
      await page.getByTestId('bulk-action-dialog-confirm').click();
    }
    await expect(page.getByTestId('bulk-toolbar')).toContainText(/Select rows/);

    // 6) Filter: inactive only hides the active row, then reset shows it.
    await page.click('[data-testid="filter-button"]');
    await page.selectOption('[data-testid="filter-is_active"]', 'false');
    await expect(page.locator('table')).not.toContainText(updatedName);
    await page.getByTestId('reset-filters').click();
    await expect(page.locator('table')).toContainText(updatedName);

    // 7) Delete via the Radix confirmation dialog
    const updatedRow = page.locator('tr', { hasText: updatedName });
    await updatedRow.locator(`[data-testid="delete-${id}"]`).click();
    await expect(page.getByTestId('delete-dialog')).toBeVisible();
    await page.getByTestId('delete-dialog-confirm').click();

    await expect(page.locator('table')).not.toContainText(updatedName);
  });
});
