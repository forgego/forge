import { defineConfig } from '@playwright/test';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.resolve(__dirname, '..');

export default defineConfig({
  testDir: './tests',
  timeout: 120_000,
  expect: {
    timeout: 20_000,
  },
  use: {
    baseURL: 'http://localhost:8000/admin/',
    channel: 'chrome',
    trace: 'off',
    screenshot: 'only-on-failure',
    video: 'off',
  },
  webServer: {
    command: 'go run -tags embed .',
    cwd: rootDir,
    url: 'http://localhost:8000/admin/',
    timeout: 180_000,
    reuseExistingServer: true,
  },
  globalSetup: path.resolve(__dirname, './global-setup.ts'),
});
