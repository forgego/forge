import { request } from '@playwright/test';
import { adminPass, adminUser, apiURL, baseURL } from './support/env';

/**
 * Fails fast, with a clear message, when the application under test is not
 * reachable or the configured superuser cannot sign in. Every test seeds
 * its own uniquely tagged rows, so no shared data is created here.
 */
export default async function globalSetup() {
  const ctx = await request.newContext({ baseURL });
  try {
    const deadline = Date.now() + 60_000;
    let lastError = '';
    while (Date.now() < deadline) {
      try {
        const response = await ctx.post(apiURL('/login'), {
          data: { username: adminUser, password: adminPass },
        });
        if (response.ok()) return;
        lastError = `${response.status()} ${await response.text()}`;
        if (response.status() === 401 || response.status() === 503) break;
      } catch (err) {
        lastError = String(err);
      }
      await new Promise((resolve) => setTimeout(resolve, 1000));
    }
    throw new Error(`Admin login at ${apiURL('/login')} failed: ${lastError}`);
  } finally {
    await ctx.dispose();
  }
}
