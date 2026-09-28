import { test, expect } from '@playwright/test';
import { mockAdminApi } from './fixtures/admin-api-mock';

const ROUTES = [
  { name: 'dashboard', path: '/admin/' },
  { name: 'registry', path: '/admin/registry' },
  { name: 'list', path: '/admin/categories' },
  { name: 'create', path: '/admin/categories/create' },
] as const;

const THEMES = ['light', 'dark'] as const;

test.describe('Design & Health Sweep', () => {
  for (const route of ROUTES) {
    for (const theme of THEMES) {
      test(`${route.name} (${theme})`, async ({ page }, testInfo) => {
        const errors: string[] = [];

        page.on('console', (msg) => {
          if (msg.type() === 'error') {
            const text = msg.text();
            if (
              !text.includes('Download the React DevTools') &&
              !text.includes('favicon')
            ) {
              errors.push(`Console error: ${text}`);
            }
          }
        });

        page.on('pageerror', (err) => {
          const text = err.message;
          if (
            !text.includes('Download the React DevTools') &&
            !text.includes('favicon')
          ) {
            errors.push(`Page error: ${text}`);
          }
        });

        await mockAdminApi(page);
        await page.goto(route.path);

        await page.evaluate(
          (t) => document.documentElement.classList.toggle('dark', t === 'dark'),
          theme
        );

        await expect(page.locator('main')).toBeVisible();
        await expect(page.locator('h1').first()).toBeVisible();

        const overflow = await page.evaluate(
          () => document.documentElement.scrollWidth - document.documentElement.clientWidth
        );
        expect(overflow).toBeLessThanOrEqual(1);

        await page.screenshot({
          path: testInfo.outputPath(`${route.name}-${theme}.png`),
          fullPage: true,
        });

        expect(errors).toEqual([]);
      });
    }
  }
});
