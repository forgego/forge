import { test, expect } from './support/fixtures';
import { adminURL } from './support/env';

// Journeys that pin the current, wrong outcome of a defect outside the admin
// UI. Every step runs and must pass; only the final assertion states the
// defective result, so any other failure (a broken form, a failed save)
// still turns the suite red. When the defect is fixed the pinned assertion
// fails: flip it to the correct value then.
test.describe('known defects', () => {
  test('creating a record with a boolean unchecked stores the column default (#291)', async ({
    adminPage: page,
    api,
    tag,
  }) => {
    await page.goto(adminURL('/categories/create'));
    await page.locator('#name').fill(`${tag} Hidden`);
    await page.locator('#slug').fill(`${tag}-hidden`);
    await expect(page.locator('#is_active')).not.toBeChecked();
    await page.getByTestId('submit-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));

    const { results } = await api.list('categories', { search: `${tag}-hidden` });
    expect(results).toHaveLength(1);
    // Defect #291: orm Manager.Create omits zero-valued columns, so the
    // explicit false is replaced by the column default (Category.is_active
    // defaults to true). When #291 is fixed, change this to toBe(false) and
    // move the test out of this file.
    expect(results[0].is_active).toBe(true);
  });
});
