import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: '.', testMatch: 'generated-consumers.spec.ts', workers: 1,
  timeout: 60_000, expect: { timeout: 10_000 }, reporter: 'list',
  outputDir: '/tmp/forge-consumer-browser-results',
  use: { trace: 'retain-on-failure', screenshot: 'only-on-failure',
    launchOptions: process.env.ADMIN_E2E_CHROMIUM_PATH ? { executablePath: process.env.ADMIN_E2E_CHROMIUM_PATH } : {} },
});
