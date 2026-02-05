import { test, expect } from '@playwright/test';

test.describe('Admin CRUD Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Ensure we start clean (token affects routing via API client interceptor)
    await page.addInitScript(() => {
      localStorage.removeItem('admin_token');
      localStorage.removeItem('admin-tabs');
    });
  });

  test('login → navigate → create → edit → bulk action → filter → delete', async ({ page }) => {
    // 1) Login
    await page.goto('/login');
    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'password');
    await page.click('[data-testid="login-button"]');

    await expect(page).toHaveURL('/');
    await expect(page.getByTestId('nav-dashboard')).toBeVisible();

    // 2) Navigate to Example Models via sidebar
    await page.click('[data-testid="nav-examplemodels"]');
    await expect(page).toHaveURL('/admin/examplemodels');

    // List loads
    await expect(page.locator('table')).toBeVisible();

    // 3) Create new
    await page.click('[data-testid="create-button"]');
    await expect(page).toHaveURL('/admin/examplemodels/new');

    const unique = Date.now();
    const createdName = `E2E User ${unique}`;
    const createdEmail = `e2e_${unique}@example.com`;

    await page.fill('#name', createdName);
    await page.fill('#email', createdEmail);
    await page.check('#is_active');
    await page.click('[data-testid="submit-button"]');

    await expect(page).toHaveURL('/admin/examplemodels');
    await expect(page.locator('table')).toContainText(createdName);

    // Capture new id from row's edit button
    const row = page.locator('tr', { hasText: createdName });
    await row.hover();
    const editBtn = row.locator('[data-testid^="edit-"]');
    const editTestId = (await editBtn.getAttribute('data-testid')) || '';
    const id = editTestId.replace('edit-', '');
    expect(id).not.toBe('');

    // 4) Edit
    await editBtn.click();
    await expect(page).toHaveURL(new RegExp(`/admin/examplemodels/${id}$`));

    const updatedName = `${createdName} Updated`;
    await page.fill('#name', updatedName);
    await page.uncheck('#is_active');
    await page.click('[data-testid="submit-button"]');

    await expect(page).toHaveURL('/admin/examplemodels');
    await expect(page.locator('table')).toContainText(updatedName);

    // 5) Bulk action: activate selected
    await page.check(`[data-testid="select-${id}"]`);
    await expect(page.getByTestId('bulk-toolbar')).toBeVisible();
    await page.click('[data-testid="bulk-action-activate_selected"]');

    // Wait for action success handler to clear selection (toolbar disappears)
    await expect(page.getByTestId('bulk-toolbar')).toBeHidden();

    // 6) Filter: show inactive only, ensure our row disappears (it should be active after bulk activation)
    await page.click('[data-testid="filter-button"]');
    await page.selectOption('[data-testid="filter-is_active"]', 'false');
    await expect(page.locator('table')).not.toContainText(updatedName);

    // Filter: show active only, our row should be present after bulk activation
    await page.selectOption('[data-testid="filter-is_active"]', 'true');
    await expect(page.locator('table')).toContainText(updatedName);

    // 7) Delete
    page.once('dialog', (dialog) => dialog.accept());
    const updatedRow = page.locator('tr', { hasText: updatedName });
    await updatedRow.hover();
    await updatedRow.locator(`[data-testid="delete-${id}"]`).click();

    // After deletion, it should disappear
    await expect(page.locator('table')).not.toContainText(updatedName);
  });
});
