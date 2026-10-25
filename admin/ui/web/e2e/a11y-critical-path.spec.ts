import { test, expect } from '@playwright/test';
import type { Page } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { mockAdminApi } from './fixtures/admin-api-mock';

const MODEL = 'categories';

async function checkA11y(page: Page, contextName: string) {
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa'])
    // Exclude transient sonner toaster region from the accessibility audit
    .exclude('[data-sonner-toaster]')
    .analyze();

  const violations = results.violations.filter(
    (v) => v.impact === 'serious' || v.impact === 'critical'
  );

  if (violations.length > 0) {
    console.error(
      `A11y violations in [${contextName}]:`,
      JSON.stringify(
        violations.map((v) => ({
          id: v.id,
          impact: v.impact,
          description: v.description,
          nodes: v.nodes.map((n) => ({ html: n.html, failureSummary: n.failureSummary })),
        })),
        null,
        2
      )
    );
  }

  expect(violations, `Expected zero serious/critical a11y violations in ${contextName}`).toEqual([]);
}

test.describe('Admin Accessibility Critical Path', () => {
  test.beforeEach(async ({ page }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await mockAdminApi(page, false);

    // Provide 25 items so total_pages = 2 with default page_size = 20
    const allCategories = Array.from({ length: 25 }, (_, i) => ({
      id: i + 1,
      name: i === 0 ? 'Electronics' : i === 1 ? 'Books' : i === 2 ? 'Clothing' : `Category ${i + 1}`,
      slug: i === 1 ? null : `cat-${i + 1}`,
      is_active: i !== 1, // Books is inactive (false), others active (true)
      created_at: `2025-01-${String((i % 28) + 1).padStart(2, '0')}T12:00:00Z`,
    }));
    let nextId = 26;

    await page.route('**/admin/api/logout', async (route) => {
      await route.fulfill({ json: { message: 'Logged out' } });
    });

    await page.route(/\/admin\/api\/categories(?:\?.*)?$/, async (route) => {
      const method = route.request().method();
      if (method === 'POST') {
        const body = route.request().postDataJSON();
        if (body.name === 'TRIGGER_VALIDATION_ERROR') {
          await route.fulfill({
            status: 400,
            json: {
              error: {
                message: 'Validation failed',
                details: {
                  name: ['Name cannot be TRIGGER_VALIDATION_ERROR'],
                },
              },
            },
          });
          return;
        }
        const created = {
          id: nextId++,
          name: '',
          slug: null,
          is_active: false,
          created_at: new Date().toISOString(),
          ...body,
        };
        allCategories.unshift(created);
        await route.fulfill({ status: 201, json: created });
        return;
      }

      const url = new URL(route.request().url());
      const active = url.searchParams.get('is_active');
      const search = url.searchParams.get('search');
      const pageNum = Number(url.searchParams.get('page') ?? '1');
      const pageSize = Number(url.searchParams.get('page_size') ?? '20');

      const filtered = allCategories.filter((c) => {
        if (active !== null && active !== '' && c.is_active !== (active === 'true')) return false;
        if (search && !c.name.toLowerCase().includes(search.toLowerCase())) return false;
        return true;
      });

      const totalCount = filtered.length;
      const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));
      const start = (pageNum - 1) * pageSize;
      const results = filtered.slice(start, start + pageSize);

      await route.fulfill({
        json: {
          count: totalCount,
          page: pageNum,
          page_size: pageSize,
          total_pages: totalPages,
          results,
        },
      });
    });

    await page.route(/\/admin\/api\/categories\/\d+(?:\?.*)?$/, async (route) => {
      const id = Number(new URL(route.request().url()).pathname.split('/').pop());
      const idx = allCategories.findIndex((c) => c.id === id);
      const method = route.request().method();

      if (idx < 0) {
        await route.fulfill({ status: 404, json: { error: 'Category not found' } });
        return;
      }

      if (method === 'DELETE') {
        allCategories.splice(idx, 1);
        await route.fulfill({ status: 204, body: '' });
      } else if (method === 'PATCH' || method === 'PUT') {
        const body = route.request().postDataJSON();
        if (body.name === 'TRIGGER_VALIDATION_ERROR') {
          await route.fulfill({
            status: 400,
            json: {
              error: {
                message: 'Validation failed',
                details: {
                  name: ['Name cannot be TRIGGER_VALIDATION_ERROR'],
                },
              },
            },
          });
          return;
        }
        allCategories[idx] = { ...allCategories[idx], ...body };
        await route.fulfill({ json: allCategories[idx] });
      } else {
        await route.fulfill({ json: allCategories[idx] });
      }
    });
  });

  test('drives critical path via keyboard only and verifies axe compliance', async ({ page }) => {
    // ----------------------------------------------------
    // 1) Sign in screen
    // ----------------------------------------------------
    await page.goto('/admin/login');
    await expect(page.getByTestId('username-input')).toBeVisible();
    await checkA11y(page, 'Sign-in Screen');

    // Drive sign in purely via keyboard
    await page.getByTestId('username-input').focus();
    await page.keyboard.type('admin');
    await page.keyboard.press('Tab');
    await expect(page.getByTestId('password-input')).toBeFocused();
    await page.keyboard.type('admin123');
    await page.keyboard.press('Enter');

    // ----------------------------------------------------
    // 2) Dashboard
    // ----------------------------------------------------
    await expect(page).toHaveURL(/\/admin\/?$/);
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    await checkA11y(page, 'Dashboard');

    // Navigate to Categories via sidebar with keyboard
    const navCategories = page.getByTestId(`nav-${MODEL}`);
    await navCategories.focus();
    await page.keyboard.press('Enter');

    // ----------------------------------------------------
    // 3) Model List Screen
    // ----------------------------------------------------
    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/?$`));
    await expect(page.locator('table')).toBeVisible();
    await checkA11y(page, 'Model List Screen');

    // List Search (Keyboard only)
    const searchInput = page.getByTestId('search-input');
    await searchInput.focus();
    await page.keyboard.type('Books');
    await expect(page.locator('table')).toContainText('Books');
    await expect(page.locator('table')).not.toContainText('Electronics');
    await checkA11y(page, 'Model List - Search Active');

    // Clear search with keyboard
    const clearSearch = page.getByTestId('clear-search');
    await clearSearch.focus();
    await page.keyboard.press('Enter');
    await expect(page.locator('table')).toContainText('Electronics');

    // Filter (Keyboard only)
    const filterBtn = page.getByTestId('filter-button');
    await filterBtn.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByTestId('filter-is_active')).toBeVisible();
    await checkA11y(page, 'Model List - Filter Panel Open');

    const filterActive = page.getByTestId('filter-is_active');
    await filterActive.focus();
    await page.keyboard.press('Space'); // Open select dropdown
    const noOption = page.getByRole('option', { name: 'No' });
    await expect(noOption).toBeVisible();
    await noOption.focus();
    await page.keyboard.press('Enter'); // Select 'No'
    await expect(page.locator('table')).toContainText('Books');
    await expect(page.locator('table')).not.toContainText('Electronics');

    // Reset filters with keyboard
    const resetFiltersBtn = page.getByTestId('reset-filters');
    await resetFiltersBtn.focus();
    await page.keyboard.press('Enter');
    await expect(page.locator('table')).toContainText('Electronics');

    // Pagination (Next page) with keyboard
    const nextBtn = page.getByTestId('pagination-next');
    await expect(nextBtn).toBeEnabled();
    await nextBtn.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByTestId('pagination-status')).toContainText('Showing 21–25');
    await checkA11y(page, 'Model List - Next Page');

    // Previous page with keyboard
    const prevBtn = page.getByTestId('pagination-prev');
    await prevBtn.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByTestId('pagination-status')).toContainText('Showing 1–20');

    // ----------------------------------------------------
    // 4) Action Confirmation Dialog (Confirm + Escape)
    // ----------------------------------------------------
    // Select first row with Space
    const firstRowSelect = page.getByTestId('select-1');
    await firstRowSelect.focus();
    await page.keyboard.press('Space');
    await expect(page.getByTestId('bulk-toolbar')).toContainText('1 selected');

    // Open bulk action confirmation dialog with Enter
    const activateBtn = page.getByTestId('bulk-action-activate');
    await activateBtn.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByTestId('bulk-action-dialog')).toBeVisible();
    await checkA11y(page, 'Action Confirmation Dialog Open');

    // Test Escape closes it
    await page.keyboard.press('Escape');
    await expect(page.getByTestId('bulk-action-dialog')).not.toBeVisible();

    // Re-open and confirm with keyboard
    await activateBtn.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByTestId('bulk-action-dialog')).toBeVisible();

    const dialogConfirmBtn = page.getByTestId('bulk-action-dialog-confirm');
    await dialogConfirmBtn.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByTestId('bulk-action-dialog')).not.toBeVisible();

    // ----------------------------------------------------
    // 5) Create Screen
    // ----------------------------------------------------
    const createBtn = page.getByTestId('create-button');
    await createBtn.focus();
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/create$`));
    await expect(page.getByRole('heading', { name: 'Add Category' })).toBeVisible();
    await checkA11y(page, 'Create Category Screen');

    const createdName = `A11y Category ${Date.now()}`;
    const nameInput = page.locator('#name');
    await nameInput.focus();
    await page.keyboard.type(createdName);

    await page.keyboard.press('Tab');
    await expect(page.locator('#slug')).toBeFocused();
    await page.keyboard.type('a11y-cat');

    const submitCreateBtn = page.getByTestId('submit-button');
    await submitCreateBtn.focus();
    await page.keyboard.press('Enter');

    // Expect return to list screen
    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/?$`));
    await expect(page.locator('table')).toContainText(createdName);

    // ----------------------------------------------------
    // 6) Edit Screen & Error Recovery
    // ----------------------------------------------------
    const row = page.locator('tr', { hasText: createdName });
    const editBtn = row.locator('[data-testid^="edit-"]');
    await editBtn.focus();
    await page.keyboard.press('Enter');

    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/\\d+$`));
    await expect(page.getByRole('heading', { name: 'Edit Category' })).toBeVisible();
    await checkA11y(page, 'Edit Category Screen');

    // Trigger validation error
    const editNameInput = page.locator('#name');
    await editNameInput.focus();
    await page.keyboard.press('ControlOrMeta+A');
    await page.keyboard.type('TRIGGER_VALIDATION_ERROR');

    const submitEditBtn = page.getByTestId('submit-button');
    await submitEditBtn.focus();
    await page.keyboard.press('Enter');

    // Validation error announced & alert role reachable
    const alert = page.getByRole('alert');
    await expect(alert.first()).toBeVisible();
    await checkA11y(page, 'Edit Category Screen - Validation Error');

    // Fix validation error and save
    const updatedName = `${createdName} Fixed`;
    await editNameInput.focus();
    await page.keyboard.press('ControlOrMeta+A');
    await page.keyboard.type(updatedName);

    await submitEditBtn.focus();
    await page.keyboard.press('Enter');

    await expect(page).toHaveURL(new RegExp(`/admin/${MODEL}/?$`));
    await expect(page.locator('table')).toContainText(updatedName);

    // ----------------------------------------------------
    // 7) Sign Out
    // ----------------------------------------------------
    const logoutBtn = page.getByTestId('logout-button');
    await logoutBtn.focus();
    await page.keyboard.press('Enter');

    await expect(page).toHaveURL(/\/admin\/login\/?$/);
    await expect(page.getByTestId('username-input')).toBeVisible();
    await checkA11y(page, 'Signed Out Screen');
  });
});
