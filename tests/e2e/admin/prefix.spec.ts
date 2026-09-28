import { test, expect } from './support/fixtures';
import { adminPath, adminURL, apiPath } from './support/env';

// CI mounts the admin under a non-default prefix (ADMIN_E2E_PATH), so every
// spec already runs through it; these tests pin the mount itself.
test.describe('@core admin mount prefix', () => {
  test('serves a deep link and every asset under the configured prefix', async ({ adminPage: page, api, tag }) => {
    const row = await api.createCategory(tag);

    const requests: { url: string; status: number; type: string }[] = [];
    page.on('response', (response) => {
      const request = response.request();
      if (['script', 'stylesheet', 'font', 'image', 'fetch', 'xhr'].includes(request.resourceType())) {
        requests.push({ url: response.url(), status: response.status(), type: request.resourceType() });
      }
    });

    // A cold load of a nested route must return the app shell, not a 404.
    const response = await page.goto(adminURL(`/categories/${row.id}`));
    expect(response?.status()).toBe(200);
    await expect(page.locator('#name')).toHaveValue(row.name);
    await page.reload();
    await expect(page.locator('#name')).toHaveValue(row.name);

    const prefix = adminPath === '/' ? '' : adminPath;
    const assets = requests.filter((r) => ['script', 'stylesheet', 'font', 'image'].includes(r.type));
    expect(assets.length).toBeGreaterThan(0);
    for (const asset of assets) {
      expect(new URL(asset.url).pathname.startsWith(`${prefix}/`), asset.url).toBe(true);
      expect(asset.status, asset.url).toBeLessThan(400);
    }
    const apiCalls = requests.filter((r) => r.type === 'fetch' || r.type === 'xhr');
    expect(apiCalls.length).toBeGreaterThan(0);
    for (const call of apiCalls) {
      expect(new URL(call.url).pathname.startsWith(`${apiPath}/`), call.url).toBe(true);
    }
  });

  test('keeps in-app links inside the prefix', async ({ adminPage: page }) => {
    await page.goto(adminURL('/'));
    await page.getByTestId('nav-categories').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));
    await page.getByTestId('create-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories/create')}$`));
    await page.getByTestId('cancel-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));
  });

  test('redirects the bare prefix to the trailing-slash entry point', async ({ request }) => {
    test.skip(adminPath === '/', 'no bare prefix to redirect when mounted at the root');
    const response = await request.get(adminPath, { maxRedirects: 0 });
    expect([301, 302, 307, 308]).toContain(response.status());
    expect(response.headers()['location']).toBe(`${adminPath}/`);
  });
});
