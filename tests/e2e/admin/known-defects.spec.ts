import { test, expect } from './support/fixtures';
import { adminURL } from './support/env';

// Journeys that currently fail because of defects outside the admin UI.
// Each is marked test.fail(): the suite stays green while the defect
// exists and turns red once it is fixed, so the marker gets removed.
test.describe('known defects', () => {
  test('creating a record with a boolean unchecked stores false', async ({ adminPage: page, api, tag }) => {
    test.fail(
      true,
      'orm: Create omits zero-valued non-required columns, so an explicit false ' +
        'is replaced by the column default (Category.is_active defaults to true).',
    );

    await page.goto(adminURL('/categories/create'));
    await page.locator('#name').fill(`${tag} Hidden`);
    await page.locator('#slug').fill(`${tag}-hidden`);
    await expect(page.locator('#is_active')).not.toBeChecked();
    await page.getByTestId('submit-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));

    const stored = (await api.list('categories', { search: `${tag}-hidden` })).results[0];
    expect(stored.is_active).toBe(false);
  });
});
