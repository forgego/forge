import { test as base, expect, type Page } from '@playwright/test';
import { AdminApi } from './api';
import { adminURL, uniqueTag } from './env';

type Fixtures = {
  /** Admin API client logged in as the configured superuser. */
  api: AdminApi;
  /** Browser page already signed in with the API session token. */
  adminPage: Page;
  /** Unique lowercase tag for this test's fixture rows. */
  tag: string;
};

export const test = base.extend<Fixtures>({
  api: async ({ request }, use) => {
    await use(await AdminApi.login(request));
  },
  tag: async ({}, use) => {
    await use(uniqueTag());
  },
  adminPage: async ({ page, api }, use) => {
    await signInWithToken(page, api.token);
    await use(page);
  },
});

export { expect };

/**
 * Stores a session token the way the login page does, without re-running
 * the login form in every test (the login journey has its own spec).
 */
export async function signInWithToken(page: Page, token: string) {
  await page.goto(adminURL('/login'));
  await page.evaluate((t) => localStorage.setItem('admin_token', t), token);
}

/** Opens a model's list page and waits for its rows to render. */
export async function openList(page: Page, model: string) {
  await page.goto(adminURL(`/${model}`));
  await expect(page.getByTestId('pagination-status')).toBeVisible();
}

/** Types into the list search box and waits for the debounced result. */
export async function searchList(page: Page, text: string, expectedCount: number) {
  await page.getByTestId('search-input').fill(text);
  await expect(page.getByTestId('pagination-status')).toContainText(`of ${expectedCount}`);
}

/** Picks a value in a Radix select rendered by the admin UI. */
export async function chooseOption(page: Page, triggerTestId: string, optionName: string) {
  await page.getByTestId(triggerTestId).click();
  await page.getByRole('option', { name: optionName, exact: true }).click();
}
