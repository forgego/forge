import { test, expect } from '@playwright/test';

const username = 'admin';
const password = 'password';

async function login(page) {
  await page.addInitScript(() => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin-tabs');
  });

  await page.goto('login');
  await page.fill('[data-testid="username-input"]', username);
  await page.fill('[data-testid="password-input"]', password);
  await page.click('[data-testid="login-button"]');
  await expect(page).toHaveURL(/\/admin\/?$/);
  await expect(page.getByTestId('nav-dashboard')).toBeVisible();
}

test.describe('Ecommerce Admin UI', () => {
  test('login and load all model list pages', async ({ page }) => {
    await login(page);

    const metaResp = await page.request.get('api/meta');
    expect(metaResp.ok()).toBeTruthy();
    const meta = await metaResp.json();
    const models = meta.models || [];
    expect(models.length).toBeGreaterThan(0);

    for (const model of models) {
      await page.goto(model.name);
      await expect(page.locator('table')).toBeVisible();
    }
  });

  test('create category and brand via admin forms', async ({ page }) => {
    await login(page);

    const unique = Date.now();

    // Categories
    await page.goto('categories');
    await expect(page.locator('table')).toBeVisible();
    await page.click('[data-testid="create-button"]');
    await expect(page).toHaveURL(/\/admin\/categories\/(new|create)$/);

    const categoryName = `Category ${unique}`;
    await page.fill('#name', categoryName);
    await page.fill('#slug', `category-${unique}`);
    await page.click('[data-testid="submit-button"]');
    await expect(page).toHaveURL(/\/admin\/categories$/);
    await expect(page.locator('table')).toContainText(categoryName);

    // Brands
    await page.goto('brands');
    await expect(page.locator('table')).toBeVisible();
    await page.click('[data-testid="create-button"]');
    await expect(page).toHaveURL(/\/admin\/brands\/(new|create)$/);

    const brandName = `Brand ${unique}`;
    await page.fill('#name', brandName);
    await page.fill('#slug', `brand-${unique}`);
    await page.click('[data-testid="submit-button"]');
    await expect(page).toHaveURL(/\/admin\/brands$/);
    await expect(page.locator('table')).toContainText(brandName);
  });
});
