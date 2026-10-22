import { test, expect } from '@playwright/test';
import { mockAdminApi } from './fixtures/admin-api-mock';

const ROUTES = [
  { name: 'dashboard', path: '/admin/', heading: 'Dashboard' },
  { name: 'registry', path: '/admin/registry', heading: 'Model Registry' },
  { name: 'list', path: '/admin/categories', heading: 'Categories' },
  { name: 'create', path: '/admin/categories/create', heading: 'Add Category' },
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
        await page.addInitScript((t) => {
          localStorage.setItem('admin-theme-storage', JSON.stringify({
            state: { theme: t, resolvedTheme: t, primary: 'iris', radius: 0.625 },
            version: 0,
          }));
        }, theme);
        await page.goto(route.path);

        await expect(page.locator('main')).toBeVisible();
        await expect(page.getByRole('heading', { name: route.heading, level: 1 })).toBeVisible();
        await expect(page.locator('html')).toHaveClass(new RegExp(`(^|\\s)${theme}(\\s|$)`));

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
