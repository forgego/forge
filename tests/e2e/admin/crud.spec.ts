import { test, expect, openList, searchList } from './support/fixtures';
import { adminURL } from './support/env';

test.describe('@core create, update and delete', () => {
  test('creates a record with a related parent, edits it and deletes it', async ({ adminPage: page, api, tag }) => {
    const parent = await api.createCategory(tag, { n: '-parent', name: `${tag} Parent` });

    // Create through the form, choosing the parent with the related-object picker.
    await openList(page, 'categories');
    await page.getByTestId('create-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories/create')}$`));

    await page.locator('#name').fill(`${tag} Child`);
    await page.locator('#slug').fill(`${tag}-child`);
    await page.locator('#sort_order').fill('4');
    await page.getByLabel(/^Parent/).click();
    await page.getByPlaceholder('Search...').fill(`${tag} Parent`);
    await page.getByRole('button', { name: `${tag} Parent`, exact: true }).click();
    await expect(page.getByLabel(/^Parent/)).toHaveText(`${tag} Parent`);
    await page.getByTestId('submit-button').click();

    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));
    await searchList(page, `${tag}-child`, 1);
    const row = page.locator('tbody tr', { hasText: `${tag} Child` });
    await expect(row).toContainText(`${tag} Parent`);

    // What PostgreSQL stored, read back through the API.
    const created = (await api.list('categories', { search: `${tag}-child` })).results[0];
    expect(created).toMatchObject({ name: `${tag} Child`, slug: `${tag}-child`, parent_id: parent.id, sort_order: 4 });

    // Update through the edit form.
    await row.getByTestId(`edit-${created.id}`).click();
    await expect(page).toHaveURL(new RegExp(`${adminURL(`/categories/${created.id}`)}$`));
    await expect(page.locator('#name')).toHaveValue(`${tag} Child`);
    await page.locator('#name').fill(`${tag} Child renamed`);
    await page.getByTestId('submit-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));
    expect(await api.get('categories', created.id)).toMatchObject({
      name: `${tag} Child renamed`,
      parent_id: parent.id,
      sort_order: 4,
    });

    // Delete from the list with confirmation.
    await searchList(page, `${tag}-child`, 1);
    await page.getByTestId(`delete-${created.id}`).click();
    await page.getByRole('alertdialog').getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByTestId('pagination-status')).toContainText('of 0');
    expect(await api.exists('categories', created.id)).toBe(false);
    expect(await api.exists('categories', parent.id)).toBe(true);
  });

  test('saves an edit to a root record without re-sending untouched fields', async ({ adminPage: page, api, tag }) => {
    // A root category has no parent; the API reports parent_id as 0, which
    // PostgreSQL's foreign key rejects if the form sends it back.
    const root = await api.createCategory(tag, { name: `${tag} Root` });
    expect(root.parent_id).toBe(0);

    await page.goto(adminURL(`/categories/${root.id}`));
    await expect(page.locator('#name')).toHaveValue(`${tag} Root`);
    await page.locator('#description').fill('Edited in the browser');
    await page.getByTestId('submit-button').click();

    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));
    expect(await api.get('categories', root.id)).toMatchObject({
      name: `${tag} Root`,
      description: 'Edited in the browser',
    });
  });

  test('prefills the create form with schema defaults and stores them untouched', async ({ adminPage: page, api, tag }) => {
    // Category.is_active has Default(true) and sort_order Default(0).
    await page.goto(adminURL('/categories/create'));
    await expect(page.locator('#is_active')).toBeChecked();
    await expect(page.locator('#sort_order')).toHaveValue('0');
    await page.locator('#name').fill(`${tag} Shown`);
    await page.locator('#slug').fill(`${tag}-shown`);
    await page.getByTestId('submit-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));

    const { results } = await api.list('categories', { search: `${tag}-shown` });
    expect(results).toHaveLength(1);
    expect(results[0].is_active).toBe(true);
  });

  test('stores an unchecked boolean as false, not the column default (#291)', async ({ adminPage: page, api, tag }) => {
    // Category.is_active defaults to true, so the form starts checked; the
    // operator unchecks it and false must be stored.
    await page.goto(adminURL('/categories/create'));
    await page.locator('#name').fill(`${tag} Hidden`);
    await page.locator('#slug').fill(`${tag}-hidden`);
    await expect(page.locator('#is_active')).toBeChecked();
    await page.locator('#is_active').uncheck();
    await expect(page.locator('#is_active')).not.toBeChecked();
    await page.getByTestId('submit-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories')}$`));

    const { results } = await api.list('categories', { search: `${tag}-hidden` });
    expect(results).toHaveLength(1);
    expect(results[0].is_active).toBe(false);
  });

  test('explains a duplicate value next to the field and stores nothing', async ({ adminPage: page, api, tag }) => {
    await api.createCategory(tag, { name: `${tag} Original`, slug: `${tag}-taken` });

    await page.goto(adminURL('/categories/create'));
    await page.locator('#name').fill(`${tag} Duplicate`);
    await page.locator('#slug').fill(`${tag}-taken`);
    await page.getByTestId('submit-button').click();

    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories/create')}$`));
    const slugField = page.locator('#slug').locator('xpath=ancestor::div[contains(@class,"space-y-2")][1]');
    await expect(slugField).toContainText('A record with this slug already exists.');
    // No driver text (constraint or table names) reaches the operator.
    await expect(page.locator('body')).not.toContainText('categories_slug_key');

    const stored = await api.list('categories', { search: tag });
    expect(stored.count).toBe(1);
    expect(stored.results[0].name).toBe(`${tag} Original`);
  });

  test('blocks submitting a form with a required field left empty', async ({ adminPage: page, api, tag }) => {
    await page.goto(adminURL('/categories/create'));
    await page.locator('#slug').fill(`${tag}-noname`);
    await page.getByTestId('submit-button').click();

    await expect(page).toHaveURL(new RegExp(`${adminURL('/categories/create')}$`));
    const missing = await page.locator('#name').evaluate((el: HTMLInputElement) => ({
      valueMissing: el.validity.valueMissing,
      message: el.validationMessage,
    }));
    expect(missing.valueMissing).toBe(true);
    expect(missing.message).not.toBe('');
    expect((await api.list('categories', { search: tag })).count).toBe(0);
  });
});
