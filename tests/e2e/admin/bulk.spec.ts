import { test, expect, openList, searchList } from './support/fixtures';

test.describe('@core bulk operations', () => {
  test('reports each skipped record when an action partially succeeds', async ({ adminPage: page, api, tag }) => {
    // Both inactive; the second becomes primary, which the Warehouse admin
    // protects from changes, so "Activate" can only apply to the first.
    const regular = await api.createWarehouse(tag, { code: `${tag}-a`, name: `${tag} Regular` });
    const primary = await api.createWarehouse(tag, { code: `${tag}-b`, name: `${tag} Primary` });
    await api.patch('warehouses', primary.id, { is_primary: true });
    expect((await api.get('warehouses', regular.id)).is_active).toBe(false);
    expect((await api.get('warehouses', primary.id)).is_active).toBe(false);

    await openList(page, 'warehouses');
    await searchList(page, tag, 2);
    await page.getByTestId('select-all').check();
    await expect(page.getByTestId('bulk-toolbar')).toContainText('2 selected');
    await page.getByTestId('bulk-action-activate').click();

    const result = page.getByTestId('bulk-result');
    await expect(result).toBeVisible();
    await expect(page.getByTestId('bulk-result-summary')).toHaveText(
      /Activate Warehouses: applied to 1 of 2 selected; 1 record was not changed\./,
    );
    await expect(page.getByTestId(`bulk-result-item-${primary.id}`)).toContainText(
      "You don't have permission to change this record",
    );
    await expect(page.getByTestId(`bulk-result-item-${regular.id}`)).toHaveCount(0);
    // Only the skipped record stays selected for follow-up.
    await expect(page.getByTestId('bulk-toolbar')).toContainText('1 selected');
    await expect(page.getByTestId(`select-${primary.id}`)).toBeChecked();

    expect((await api.get('warehouses', regular.id)).is_active).toBe(true);
    expect((await api.get('warehouses', primary.id)).is_active).toBe(false);

    await page.getByTestId('bulk-result-dismiss').click();
    await expect(result).toHaveCount(0);
  });

  test('reports a fully rejected action without claiming success', async ({ adminPage: page, api, tag }) => {
    const primary = await api.createWarehouse(tag, { code: `${tag}-p`, name: `${tag} Primary` });
    await api.patch('warehouses', primary.id, { is_primary: true });

    await openList(page, 'warehouses');
    await searchList(page, tag, 1);
    await page.getByTestId(`select-${primary.id}`).check();
    await page.getByTestId('bulk-action-activate').click();

    await expect(page.getByTestId(`bulk-result-item-${primary.id}`)).toBeVisible();
    await expect(page.getByText('No permitted objects found for action')).toBeVisible();
    await expect(page.getByText('Success', { exact: true })).toHaveCount(0);
    expect((await api.get('warehouses', primary.id)).is_active).toBe(false);
  });

  test('deletes what it can and names each record it could not delete', async ({ adminPage: page, api, tag }) => {
    const parent = await api.createCategory(tag, { n: '-parent', name: `${tag} Parent` });
    await api.createCategory(tag, { n: '-child', name: `${tag} Child`, parent_id: parent.id });
    const loose = await api.createCategory(tag, { n: '-loose', name: `${tag} Loose` });

    await openList(page, 'categories');
    await searchList(page, tag, 3);
    await page.getByTestId(`select-${parent.id}`).check();
    await page.getByTestId(`select-${loose.id}`).check();
    await page.getByTestId('bulk-delete-button').click();
    await page.getByRole('alertdialog').getByRole('button', { name: 'Delete Selected' }).click();

    await expect(page.getByTestId('bulk-result-summary')).toHaveText(
      /Delete: applied to 1 of 2 selected; 1 record was not changed\./,
    );
    await expect(page.getByTestId(`bulk-result-item-${parent.id}`)).toContainText(
      'Other records still refer to this record',
    );
    await expect(page.locator('body')).not.toContainText('categories_parent_id_fkey');

    expect(await api.exists('categories', loose.id)).toBe(false);
    expect(await api.exists('categories', parent.id)).toBe(true);
    await expect(page.getByTestId('pagination-status')).toContainText('of 2');
  });
});
