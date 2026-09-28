import { defineConfig, devices } from '@playwright/test';

// The suite drives an already-running Forge application (the ecommerce
// example in CI) backed by a real PostgreSQL database. It never starts a
// server itself: the caller owns the application and its database fixture.
const baseURL = process.env.ADMIN_E2E_BASE_URL || 'http://localhost:8000';

// Optional local override for environments that ship a preinstalled
// Chromium build that differs from the one this Playwright version expects.
const executablePath = process.env.ADMIN_E2E_CHROMIUM_PATH || undefined;

export default defineConfig({
  testDir: '.',
  testMatch: /.*\.spec\.ts/,
  globalSetup: './global-setup.ts',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  reporter: process.env.CI
    ? [['list'], ['html', { open: 'never' }], ['github']]
    : [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'off',
    viewport: { width: 1280, height: 720 },
    launchOptions: executablePath ? { executablePath } : {},
  },
  projects: [
    {
      name: 'desktop',
      testIgnore: /mobile\.spec\.ts/,
      use: { ...devices['Desktop Chrome'], launchOptions: executablePath ? { executablePath } : {} },
    },
    {
      // docs/design/admin-ui-system.md declares the phone layout at <=640px.
      name: 'mobile',
      testMatch: /mobile\.spec\.ts/,
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 375, height: 812 },
        hasTouch: true,
        launchOptions: executablePath ? { executablePath } : {},
      },
    },
  ],
});
