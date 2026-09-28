import { test, expect } from './support/fixtures';
import { AdminApi } from './support/api';
import { adminPass, adminURL, adminUser, apiURL } from './support/env';

test.describe('@core authentication', () => {
  test('rejects a wrong password with a message and keeps the operator signed out', async ({ page }) => {
    await page.goto(adminURL('/login'));
    await page.getByTestId('username-input').fill(adminUser);
    await page.getByTestId('password-input').fill(`${adminPass}-wrong`);
    await page.getByTestId('login-button').click();

    await expect(page.getByText(/invalid/i)).toBeVisible();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/login')}$`));
    expect(await page.evaluate(() => localStorage.getItem('admin_token'))).toBeNull();
  });

  test('redirects a signed-out deep link to the login page', async ({ page }) => {
    await page.goto(adminURL('/categories'));
    await expect(page).toHaveURL(new RegExp(`${adminURL('/login')}$`));
    await expect(page.getByTestId('login-button')).toBeVisible();
  });

  test('signs in, then signs out and revokes the session token', async ({ page, request }) => {
    await page.goto(adminURL('/login'));
    await page.getByTestId('username-input').fill(adminUser);
    await page.getByTestId('password-input').fill(adminPass);
    await page.getByTestId('login-button').click();

    await expect(page.getByTestId('nav-dashboard')).toBeVisible();
    const token = await page.evaluate(() => localStorage.getItem('admin_token'));
    expect(token).toBeTruthy();

    // The session works for API calls made on the operator's behalf.
    const signedIn = new AdminApi(request, token!);
    expect((await signedIn.raw('GET', '/meta')).status()).toBe(200);

    await page.getByTestId('logout-button').click();
    await expect(page).toHaveURL(new RegExp(`${adminURL('/login')}$`));

    // The server revoked the token, not only the browser copy of it.
    const afterLogout = await request.get(apiURL('/meta'), {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(afterLogout.status()).toBe(401);

    // Protected pages are no longer reachable.
    await page.goto(adminURL('/categories'));
    await expect(page).toHaveURL(new RegExp(`${adminURL('/login')}$`));
  });
});
