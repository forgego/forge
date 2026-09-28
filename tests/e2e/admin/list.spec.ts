import { test, expect, openList, searchList, chooseOption } from './support/fixtures';
import { adminURL } from './support/env';

test.describe('@core list pages', () => {
  test('navigates from the sidebar to a model list', async ({ adminPage: page }) => {
    await page.goto(adminURL('/'));
    await page.getByTestId('nav-categories').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));
    await expect(page.getByRole('heading', { name: 'Categories' })).toBeVisible();
    await expect(page.locator('table')).toBeVisible();
  });

  test('searches, filters and paginates PostgreSQL rows', async ({ adminPage: page, api, tag }) => {
    // 30 tagged rows: more than the configured page size (25). Every third
    // row is inactive so the boolean filter has something to remove.
    for (let n = 1; n <= 30; n++) {
      await api.createCategory(tag, {
        n: String(n).padStart(2, '0'),
        is_active: n % 3 !== 0,
        sort_order: n,
      });
    }

    await openList(page, 'categories');
    await searchList(page, tag, 30);

    const status = page.getByTestId('pagination-status');
    await expect(status).toHaveText(/Showing 1–25 of 30/);
    await expect(page.locator('tbody tr')).toHaveCount(25);

    await page.getByTestId('pagination-next').click();
    await expect(status).toHaveText(/Showing 26–30 of 30/);
    await expect(page.locator('tbody tr')).toHaveCount(5);
    await expect(page.getByTestId('pagination-next')).toBeDisabled();

    await page.getByTestId('pagination-prev').click();
    await expect(status).toHaveText(/Showing 1–25 of 30/);

    // Filter composes with search and resets to the first page.
    await page.getByTestId('filter-button').click();
    await chooseOption(page, 'filter-is_active', 'No');
    await expect(status).toHaveText(/Showing 1–10 of 10/);
    const expected = await api.list('categories', { search: tag, is_active: 'false', page_size: 100 });
    expect(expected.count).toBe(10);
    for (const row of expected.results) {
      await expect(page.locator('tbody')).toContainText(row.name);
    }

    // A search with no matches offers a way back.
    await page.getByTestId('search-input').fill(`${tag}-nothing-matches`);
    await expect(page.getByText('No results found')).toBeVisible();
    await page.getByTestId('clear-filters').click();
    await expect(page.getByTestId('search-input')).toHaveValue('');
  });

  test('changes the page size and sorts by a column', async ({ adminPage: page, api, tag }) => {
    for (const [n, name] of [['1', 'Charlie'], ['2', 'Alpha'], ['3', 'Bravo']]) {
      await api.createCategory(tag, { n, name: `${tag} ${name}` });
    }

    await openList(page, 'categories');
    await searchList(page, tag, 3);

    await chooseOption(page, 'page-size-select', '10');
    await expect(page.getByTestId('pagination-status')).toHaveText(/Showing 1–3 of 3/);

    await page.getByTestId('sort-name').click();
    const names = page.locator('tbody tr').filter({ hasText: tag });
    await expect(names.nth(0)).toContainText(`${tag} Alpha`);
    await expect(names.nth(2)).toContainText(`${tag} Charlie`);

    await page.getByTestId('sort-name').click();
    await expect(names.nth(0)).toContainText(`${tag} Charlie`);
    await expect(names.nth(2)).toContainText(`${tag} Alpha`);
  });
});
