import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { test, expect, type Page } from '@playwright/test';
import { mockAdminApi } from './fixtures/admin-api-mock';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const EXTENSIONS = [
  { name: 'reference', entry: '/admin/ext/reference/index.js', version: 1 },
];

const ASSET_PATH = path.resolve(
  __dirname,
  '../../../../tests/testdata/reference_app/admin/assets/index.js',
);

async function login(page: Page) {
  await page.goto('/admin/login');
  await page.fill('[data-testid="username-input"]', 'admin');
  await page.fill('[data-testid="password-input"]', 'admin123');
  await page.click('[data-testid="login-button"]');

  await expect(page).toHaveURL(/\/admin\/?$/);
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
}

test.describe('Reference Admin Extension', () => {
  test('loads reference extension, shows nav, renders page, and freezes ForgeAdmin', async ({
    page,
  }) => {
    expect(fs.existsSync(ASSET_PATH)).toBe(true);
    const assetContent = fs.readFileSync(ASSET_PATH, 'utf-8');

    await mockAdminApi(page, false, { extensions: EXTENSIONS });

    await page.route(/\/admin\/ext\/reference\/index\.js(?:\?.*)?$/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'text/javascript',
        body: assetContent,
      });
    });

    // 1) Login
    await login(page);

    // 2) Verify ForgeAdmin is frozen and has version 1
    const isFrozen = await page.evaluate(() => {
      const win = window as Window & { ForgeAdmin?: unknown };
      return win.ForgeAdmin ? Object.isFrozen(win.ForgeAdmin) : false;
    });
    expect(isFrozen).toBe(true);

    const version = await page.evaluate(() => {
      const win = window as Window & { ForgeAdmin?: { version?: number } };
      return win.ForgeAdmin?.version;
    });
    expect(version).toBe(1);

    // 3) No extension failure banner is shown
    await expect(page.getByTestId('extension-error-banner')).not.toBeVisible();

    // 4) Check nav entry for "Approvals"
    if (test.info().project.name === 'mobile') {
      await page.getByRole('button', { name: 'Open navigation' }).click();
    }
    const approvalsNav = page.getByRole('button', { name: 'Approvals' });
    await expect(approvalsNav).toBeVisible();

    // 5) Clicking it navigates to /admin/ext/reference/approvals and shows ref-approvals-page
    await approvalsNav.click();
    await expect(page).toHaveURL(/\/admin\/ext\/reference\/approvals\/?$/);
    await expect(page.getByTestId('ref-approvals-page')).toBeVisible();

    // 6) Verify no failure banner is shown on extension page
    await expect(page.getByTestId('extension-error-banner')).not.toBeVisible();
  });

  test('when index.js returns 500, admin still loads and failure banner is shown', async ({
    page,
  }) => {
    await mockAdminApi(page, false, { extensions: EXTENSIONS });

    await page.route(/\/admin\/ext\/reference\/index\.js(?:\?.*)?$/, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'text/plain',
        body: 'Internal Server Error',
      });
    });

    // Login
    await login(page);

    // Failure banner is shown
    const banner = page.getByTestId('extension-error-banner');
    await expect(banner).toBeVisible();
    await expect(banner).toContainText('Some extensions failed to load');
  });
});
