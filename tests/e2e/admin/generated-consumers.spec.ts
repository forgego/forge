import { test, expect } from '@playwright/test';
import { readdirSync, readFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { spawn, type ChildProcess } from 'node:child_process';

// Opt-in: the Go CLI journey creates real apps/databases and consumer.json.
const artifacts = process.env.FORGE_CONSUMER_ARTIFACTS;
const consumers = artifacts ? readdirSync(artifacts).flatMap(dir => {
  const file = join(artifacts, dir, 'consumer.json');
  return existsSync(file) ? [JSON.parse(readFileSync(file, 'utf8'))] : [];
}) : [];
if (!artifacts) test('generated consumers require retained CLI journey artifacts', () => test.skip());
for (const app of consumers) {
  test.describe(app.name, () => {
    let server: ChildProcess;
    const base = `http://127.0.0.1:${app.port}`;
    test.beforeAll(async () => {
      server = spawn(join(app.project, 'bin/server'), [], { cwd: app.project, stdio: 'ignore' });
      await expect.poll(async () => {
        try { return (await fetch(`${base}/admin/login`)).status; } catch { return 0; }
      }).toBe(200);
    });
    test.afterAll(async () => {
      if (server && server.exitCode === null) {
        const stopped = new Promise(resolve => server.once('exit', resolve));
        server.kill('SIGTERM'); await stopped;
      }
    });
    for (const mobile of [false, true]) {
      test(`${mobile ? 'phone' : 'desktop'} login, create, edit, delete and revoke`, async ({ page, request }, testInfo) => {
        await page.setViewportSize(mobile ? { width: 375, height: 812 } : { width: 1280, height: 900 });
        await page.goto(`${base}/admin/login`);
        await page.getByTestId('username-input').fill('operator');
        await page.getByTestId('password-input').fill('local-acceptance-only');
        await page.getByTestId('login-button').click();
        await expect(page.getByTestId('logout-button')).toBeAttached();
        const token = await page.evaluate(() => localStorage.getItem('admin_token'));
        expect(token).toBeTruthy();
        const headers = { Authorization: `Bearer ${token}` };
        const tag = `browser-${Date.now()}-${mobile ? 'phone' : 'desktop'}`;
        await page.goto(`${base}/admin/${app.route}/create`);
        await page.locator('#name').fill(tag);
        for (const [field, value] of Object.entries(app.input)) {
          if (field === 'name' || field === 'active') continue;
          await page.locator(`#${field}`).fill(typeof value === 'number' ? String(value) : `${tag}-${value}`);
        }
        if (await page.locator('#active').count()) await page.locator('#active').uncheck();
        await expect(page.locator('#internal_token')).toHaveCount(0);
        await page.getByTestId('submit-button').click();
        await expect(page).toHaveURL(`${base}/admin/${app.route}`);
        await page.getByTestId('search-input').fill(tag);
        await expect(page.getByTestId('pagination-status')).toContainText('of 1');
        const listed = await request.get(`${base}/admin/api/${app.route}?search=${tag}`, { headers });
        expect(listed.status()).toBe(200);
        const row = (await listed.json()).results[0];
        expect(row.name).toBe(tag);
        expect(row).not.toHaveProperty('internal_token');
        const detail = await request.get(`${base}/admin/api/${app.route}/${row.id}`, { headers });
        expect(detail.status()).toBe(200);
        expect(await detail.json()).not.toHaveProperty('internal_token');
        if ('active' in row) expect(row.active).toBe(false);
        expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
        await page.screenshot({ path: testInfo.outputPath("created-record.png"), fullPage: true });
        await page.goto(`${base}/admin/${app.route}/${row.id}`);
        await expect(page.locator('#name')).toHaveValue(tag);
        await page.locator('#name').fill(`${tag}-edited`);
        await page.getByTestId('submit-button').click();
        await expect(page).toHaveURL(`${base}/admin/${app.route}`);
        const edited = await request.get(`${base}/admin/api/${app.route}/${row.id}`, { headers });
        expect((await edited.json()).name).toBe(`${tag}-edited`);
        await page.getByTestId('search-input').fill(tag);
        await expect(page.getByTestId('pagination-status')).toContainText('of 1');
        await page.getByTestId(`delete-${row.id}`).click();
        await page.getByRole('alertdialog').getByRole('button', { name: 'Delete', exact: true }).click();
        await expect(page.getByTestId('pagination-status')).toContainText('of 0');
        expect((await request.get(`${base}/admin/api/${app.route}/${row.id}`, { headers })).status()).toBe(404);
        if (mobile) {
          await page.getByRole('button', { name: 'Open navigation' }).click();
        }
        await page.getByTestId('logout-button').click();
        await expect(page).toHaveURL(`${base}/admin/login`);
        expect((await request.get(`${base}/admin/api/meta`, { headers })).status()).toBe(401);
      });
    }
  });
}
