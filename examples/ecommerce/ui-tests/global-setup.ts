import { existsSync, unlinkSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const adminApiBase = 'http://localhost:8000/admin/api';

async function waitForAdminApi(timeoutMs = 30000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const resp = await fetch(`${adminApiBase}/meta`);
      if (resp.ok) {
        return;
      }
    } catch {
      // ignore and retry
    }
    await new Promise((resolve) => setTimeout(resolve, 500));
  }
  throw new Error('Admin API did not become ready in time');
}

async function cleanupModel(modelName: string) {
  const resp = await fetch(`${adminApiBase}/${modelName}?page_size=200`);
  if (!resp.ok) {
    return;
  }
  const payload = await resp.json();
  const results = Array.isArray(payload?.results) ? payload.results : [];
  for (const row of results) {
    if (row?.id === undefined || row?.id === null) {
      continue;
    }
    await fetch(`${adminApiBase}/${modelName}/${row.id}`, { method: 'DELETE' });
  }
}

export default async function globalSetup() {
  const __filename = fileURLToPath(import.meta.url);
  const __dirname = path.dirname(__filename);
  const rootDir = path.resolve(__dirname, '..');
  const sqlitePath = path.join(rootDir, 'ecommerce.sqlite');

  // Ensure a clean sqlite database for UI tests.
  if (existsSync(sqlitePath)) {
    unlinkSync(sqlitePath);
  }

  // Clean stale admin data in Postgres-backed runs.
  await waitForAdminApi();
  await cleanupModel('categories');
  await cleanupModel('brands');
}
