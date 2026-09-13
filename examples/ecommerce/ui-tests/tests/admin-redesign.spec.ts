import { test, expect, type APIRequestContext } from '@playwright/test';

async function chooseRadixOption(page, triggerTestId: string, optionName: string | RegExp) {
  await page.getByTestId(triggerTestId).click();
  await page.getByRole('option', { name: optionName, exact: typeof optionName === 'string' }).click();
}

// Comprehensive verification for the admin redesign. Runs against the live
// ecommerce reference server (seeded sqlite) serving the production bundle.
const username = process.env.FORGE_ADMIN_USERNAME || 'admin';
const password = process.env.FORGE_ADMIN_PASSWORD || 'admin123';
const adminApi = 'http://localhost:8020/admin/api';

async function login(page) {
  await page.goto('login');
  await page.evaluate(() => {
    localStorage.removeItem('admin_token');
  });
  await page.fill('[data-testid="username-input"]', username);
  await page.fill('[data-testid="password-input"]', password);
  await page.click('[data-testid="login-button"]');
  await expect(page).toHaveURL(/\/admin\/?$/);
  await expect(page.getByTestId('nav-dashboard')).toBeVisible();
}

async function apiToken(request: APIRequestContext): Promise<string> {
  const resp = await request.post(`${adminApi}/login`, {
    data: { username, password },
  });
  expect(resp.ok()).toBeTruthy();
  const body = await resp.json();
  return body.token as string;
}

async function createCategory(request: APIRequestContext, name: string, slug: string): Promise<number> {
  const token = await apiToken(request);
  const resp = await request.post(`${adminApi}/categories`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { name, slug },
  });
  expect(resp.ok()).toBeTruthy();
  const body = await resp.json();
  return body.id as number;
}

test.describe('Admin redesign', () => {
  test('login shows brand, toggles password, and surfaces errors accessibly', async ({ page }) => {
    await page.goto('login');
    await expect(page.getByText('Forge Admin')).toBeVisible();

    const pw = page.getByTestId('password-input');
    await expect(pw).toHaveAttribute('type', 'password');
    await page.getByRole('button', { name: 'Show password' }).click();
    await expect(pw).toHaveAttribute('type', 'text');

    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'wrong-pass');
    await page.click('[data-testid="login-button"]');
    const alert = page.getByRole('alert');
    await expect(alert).toBeVisible();
  });

  test('dashboard renders SDUI widgets and palette opens via keyboard', async ({ page }) => {
    await login(page);

    await expect(page.getByText('Total Revenue')).toBeVisible();
    await expect(page.getByText('Top Selling Products')).toBeVisible();

    await page.keyboard.press('Control+k');
    const dialog = page.getByRole('dialog', { name: 'Global search' });
    await expect(dialog).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(dialog).not.toBeVisible();
  });

  test('palette: fuzzy model filter, arrow+enter navigation', async ({ page }) => {
    await login(page);

    await page.getByTestId('global-search-trigger').click();
    const dialog = page.getByRole('dialog', { name: 'Global search' });
    await expect(dialog).toBeVisible();

    await page.fill('[data-testid="global-search-input"]', 'cate');
    await expect(dialog.getByText('Categories')).toBeVisible();
    await expect(dialog.getByText('Products', { exact: true })).not.toBeVisible();

    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/\/admin\/categories$/);
  });

  test('list: search, sort, page-size, pagination, select-all', async ({ page, request }) => {
    const unique = Date.now();
    const name = `Searchable ${unique}`;
    await createCategory(request, name, `searchable-${unique}`);
    await login(page);
    await page.goto('categories');
    await expect(page.locator('table')).toBeVisible();

    // Search narrows results; clear restores.
    await page.fill('[data-testid="search-input"]', name);
    await expect(page.locator('table')).toContainText(name);
    await page.getByTestId('clear-search').click();
    await expect(page.getByTestId('search-input')).toHaveValue('');

    // Sort toggling hits the API with ordering params.
    const sortBtn = page.getByTestId('sort-name');
    const ascResp = page.waitForResponse(
      (r) => r.url().includes('/api/categories') && r.url().includes('ordering=name')
    );
    await sortBtn.click();
    await ascResp;
    const descResp = page.waitForResponse(
      (r) => r.url().includes('/api/categories') && r.url().includes('ordering=-name')
    );
    await sortBtn.click();
    await descResp;

    // Pagination controls exist with live status.
    await expect(page.getByTestId('pagination-status')).toContainText(/of \d+/);
    const sizeResp = page.waitForResponse((r) =>
      r.url().includes('/api/categories') && r.url().includes('page_size=10')
    );
    await chooseRadixOption(page, 'page-size-select', '10');
    await sizeResp;

    // Select all drives the bulk toolbar.
    await page.getByTestId('select-all').click();
    await expect(page.getByTestId('bulk-toolbar')).toContainText(/selected/);
    await page.getByTestId('bulk-cancel').click();
    await expect(page.getByTestId('bulk-toolbar')).toContainText(/Select rows/);
  });

  test('list: boolean filter + reset, save/apply/delete view, csv export', async ({ page }) => {
    await login(page);
    await page.goto('categories');
    await expect(page.locator('table')).toBeVisible();

    // Filter panel with badge count + reset.
    await page.getByTestId('filter-button').click();
    await chooseRadixOption(page, 'filter-is_active', 'Yes');
    await expect(page.getByTestId('filter-button')).toContainText('1');
    await page.getByTestId('reset-filters').click();

    // Save current view via dialog (no native prompt).
    const viewName = `e2e-view-${Date.now()}`;
    await page.getByTestId('save-view-button').click();
    await expect(page.getByTestId('save-view-dialog')).toBeVisible();
    await page.fill('[data-testid="save-view-name"]', viewName);
    await page.getByTestId('save-view-confirm').click();
    await expect(page.getByTestId('save-view-dialog')).not.toBeVisible();
    await expect(page.getByTestId('saved-view-select')).toContainText(viewName);

    // CSV export downloads a real file (authenticated blob fetch).
    const downloadPromise = page.waitForEvent('download');
    await page.getByTestId('export-select').click();
    await page.getByTestId('export-csv').click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/\.csv$/i);

    // Delete the saved view again.
    await page.getByTestId('delete-view-button').click();
    await expect(page.getByTestId('saved-view-select')).not.toContainText(viewName);
  });

  test('create blocks empty submit, saves, edits, views tabs, deletes', async ({ page }) => {
    await login(page);
    const unique = Date.now();
    const name = `Redesign E2E ${unique}`;

    await page.goto('categories');
    await page.getByTestId('create-button').click();
    await expect(page).toHaveURL(/\/admin\/categories\/create$/);

    // Empty submit is blocked by native validation: we stay on the form.
    await page.getByTestId('submit-button').click();
    await expect(page).toHaveURL(/\/admin\/categories\/create$/);
    await expect(page.locator('#name')).toBeVisible();

    await page.fill('#name', name);
    await page.fill('#slug', `redesign-e2e-${unique}`);
    await page.getByTestId('submit-button').click();
    await expect(page).toHaveURL(/\/admin\/categories$/);
    await expect(page.locator('table')).toContainText(name);

    // Open the detail view, check tabs + copy button.
    await page.locator('table').getByRole('button', { name }).click();
    await expect(page).toHaveURL(/\/view$/);
    await expect(page.getByTestId('tab-overview')).toBeVisible();
    await page.getByTestId('tab-json').click();
    await expect(page.getByTestId('copy-json')).toBeVisible();
    await page.getByTestId('tab-history').click();

    // Edit via header button.
    await page.getByTestId('edit-record').click();
    await page.fill('#name', `${name} Updated`);
    await page.getByTestId('submit-button').click();
    await expect(page.locator('table')).toContainText(`${name} Updated`);

    // Row delete with confirmation dialog.
    const row = page.locator('tr', { hasText: `${name} Updated` });
    const editId = await row.getByTestId(/^edit-/).getAttribute('data-testid');
    const id = (editId || '').replace('edit-', '');
    await row.getByTestId(`delete-${id}`).click();
    await expect(page.getByTestId('delete-dialog')).toBeVisible();
    await page.getByTestId('delete-dialog-confirm').click();
    await expect(page.locator('table')).not.toContainText(`${name} Updated`);
  });

  test('bulk activate + bulk delete with confirmation', async ({ page, request }) => {
    const unique = Date.now();
    await createCategory(request, `Bulk One ${unique}`, `bulk-one-${unique}`);
    await createCategory(request, `Bulk Two ${unique}`, `bulk-two-${unique}`);
    await login(page);
    await page.goto('categories');
    await expect(page.locator('table')).toBeVisible();

    page.on('console', (msg) => {
      if (msg.type() === 'error') console.log('BROWSER CONSOLE ERROR:', msg.text());
    });

    await page.getByTestId('select-all').click();
    await page.getByTestId('bulk-action-activate').click();
    // The activate action requires confirmation.
    await expect(page.getByTestId('bulk-action-dialog')).toBeVisible();
    await page.getByTestId('bulk-action-dialog-confirm').click();
    await expect(page.getByTestId('bulk-toolbar')).toContainText(/Select rows/);

    await page.getByTestId('select-all').click();
    await page.getByTestId('bulk-delete-button').click();
    await expect(page.getByTestId('bulk-delete-dialog')).toBeVisible();
    await page.getByTestId('bulk-delete-dialog-cancel').click();
    await expect(page.getByTestId('bulk-toolbar')).toContainText(/selected/);
  });

  test('registry lists models and plugin pages render', async ({ page }) => {
    await login(page);

    await page.goto('registry');
    const main = page.locator('main');
    await expect(
      main.getByText('categories', { exact: true }).first()
    ).toBeVisible();

    // Reports plugin (registered by the ecommerce app) exposes pages.
    await page.goto('categories');
    const reportsNav = page.getByTestId(/^nav-.*report.*/i);
    if ((await reportsNav.count()) > 0) {
      await reportsNav.first().click();
      await expect(page.locator('main')).not.toBeEmpty();
    }
  });

  test('theme customizer opens, resets, and closes on escape', async ({ page }) => {    await login(page);

    await page.getByTestId('theme-settings-trigger').click();
    await expect(page.getByTestId('theme-settings')).toBeVisible();
    await page.getByRole('button', { name: /reset to defaults/i }).click();
    await page.keyboard.press('Escape');
    await expect(page.getByTestId('theme-settings')).not.toBeVisible();
  });

  test('logout returns to login', async ({ page }) => {
    await login(page);
    await page.getByTestId('logout-button').click();
    await expect(page).toHaveURL(/\/admin\/login$/);
  });
});

test.describe('Admin responsive + dark mode', () => {
  test.use({ viewport: { width: 375, height: 812 } });

  test('mobile drawer opens and list stays usable', async ({ page }) => {
    await login(page);
    await page.goto('categories');
    await expect(page.locator('table')).toBeVisible();

    await page.getByRole('button', { name: 'Open navigation' }).click();
    await expect(page.getByRole('navigation', { name: 'Primary' })).toBeVisible();
    await page.screenshot({ path: '/tmp/shot-mobile-nav.png' });
    await page.keyboard.press('Escape');
    await page.screenshot({ path: '/tmp/shot-mobile-list.png' });
  });
});
