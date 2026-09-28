import { test, expect, openList, searchList } from './support/fixtures';
import { adminURL } from './support/env';

// Runs in the "mobile" project (375x812). docs/design/admin-ui-system.md
// declares a 1rem page gutter below 640px, an off-canvas navigation drawer
// below the lg breakpoint, and a horizontally scrolling table with a sticky
// identity column (VB-08) instead of a page that overflows.
test.describe('@core mobile layout', () => {
  test('opens navigation as a drawer and closes it after navigating', async ({ adminPage: page }) => {
    await page.goto(adminURL('/'));
    const drawer = page.getByRole('complementary', { name: 'Admin navigation' });
    await expect(page.getByRole('button', { name: 'Open navigation' })).toBeVisible();
    await expect(drawer).not.toBeInViewport();

    await page.getByRole('button', { name: 'Open navigation' }).click();
    await expect(drawer).toBeInViewport();
    await drawer.getByTestId('nav-categories').click();

    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));
    await expect(drawer).not.toBeInViewport();
  });

  test('keeps the page within the viewport and scrolls the table instead', async ({ adminPage: page, api, tag }) => {
    await api.createCategory(tag);
    await openList(page, 'categories');
    await searchList(page, tag, 1);

    const layout = await page.evaluate(() => {
      const table = document.querySelector('table')!;
      const scroller = table.parentElement!.closest('[class*="overflow-x-auto"]') as HTMLElement;
      const content = document.querySelector('main > div.flex-1') as HTMLElement;
      const stickyCell = document.querySelector('tbody tr td') as HTMLElement;
      return {
        viewport: window.innerWidth,
        pageWidth: document.documentElement.scrollWidth,
        tableWidth: table.scrollWidth,
        scrollerWidth: scroller.clientWidth,
        scrollerOverflow: getComputedStyle(scroller).overflowX,
        gutter: parseFloat(getComputedStyle(content).paddingLeft),
        stickyPosition: getComputedStyle(stickyCell).position,
      };
    });

    expect(layout.viewport).toBe(375);
    expect(layout.pageWidth).toBeLessThanOrEqual(layout.viewport);
    expect(layout.gutter).toBe(16);
    expect(layout.scrollerOverflow).toBe('auto');
    expect(layout.tableWidth).toBeGreaterThan(layout.scrollerWidth);
    expect(layout.stickyPosition).toBe('sticky');

    // The row stays usable: its checkbox is reachable after scrolling.
    const scroller = page.locator('table').locator('xpath=ancestor::div[contains(@class,"overflow-x-auto")][1]');
    await scroller.evaluate((el) => el.scrollTo({ left: el.scrollWidth }));
    await expect(page.locator('tbody tr').first().locator('input[type="checkbox"]')).toBeInViewport();
  });

  test('signs in on a phone-sized screen', async ({ page }) => {
    await page.goto(adminURL('/login'));
    const form = page.getByTestId('login-button');
    await expect(form).toBeInViewport();
    const pageWidth = await page.evaluate(() => document.documentElement.scrollWidth);
    expect(pageWidth).toBeLessThanOrEqual(375);
  });
});
