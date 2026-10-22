import { test, expect } from '@playwright/test';
import { mockAdminApi } from './fixtures/admin-api-mock';

const MODEL = 'categories';

test.describe('Admin CRUD Flow', () => {
  test.beforeEach(async ({ page }) => {
    await mockAdminApi(page, false);
  });

  test('login → navigate → create → edit → bulk action → filter → delete', async ({ page }) => {
    // 1) Login
    await page.goto('/admin/login');
    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'admin123');
    await page.click('[data-testid="login-button"]');

    await expect(page).toHaveURL(/\/admin\/?$/);
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

    // 2) Navigate to the model via sidebar
    if (test.info().project.name === 'mobile') {
      await page.getByRole('button', { name: 'Open navigation' }).click();
    }
    await page.click(`[data-testid="nav-${MODEL}"]`);
    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/?$`));

    // List loads
    await expect(page.locator('table')).toBeVisible();

    // 3) Create new
    await page.click('[data-testid="create-button"]');
    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/create$`));

    const unique = Date.now();
    const createdName = `E2E Category ${unique}`;

    await page.fill('#name', createdName);
    await page.fill('#slug', `e2e-${unique}`);
    await page.click('[data-testid="submit-button"]');

    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/?$`));
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
    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/${id}$`));

    const updatedName = `${createdName} Updated`;
    await page.fill('#name', updatedName);
    await page.click('[data-testid="submit-button"]');

    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/?$`));
    await expect(page.locator('table')).toContainText(updatedName);

    // 5) Bulk action (activate requires confirmation in the reference app)
    await page.click(`[data-testid="select-${id}"]`);
    await expect(page.getByTestId('bulk-toolbar')).toContainText('1 selected');
    await page.click('[data-testid="bulk-action-activate"]');
    await expect(page.getByTestId('bulk-action-dialog')).toBeVisible();
    await page.getByTestId('bulk-action-dialog-confirm').click();
    await expect(page.getByTestId('bulk-toolbar')).toContainText(/Select rows/);

    // 6) Filter: inactive only hides the active row, then reset shows it.
    await page.click('[data-testid="filter-button"]');
    await page.getByTestId('filter-is_active').click();
    await page.getByRole('option', { name: 'No' }).click();
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
