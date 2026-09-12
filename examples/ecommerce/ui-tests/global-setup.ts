const adminPort = process.env.FORGE_PORT || '8020';
const adminApiBase = `http://localhost:${adminPort}/admin/api`;

async function waitForAdminApi(timeoutMs = 2000): Promise<boolean> {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const resp = await fetch(`http://localhost:${adminPort}/health`);
      if (resp.ok) {
        return true;
      }
    } catch {
      // ignore and retry
    }
    await new Promise((resolve) => setTimeout(resolve, 300));
  }
  return false;
}

async function cleanupModel(modelName: string, token: string) {
  if (!token) return;
  const resp = await fetch(`${adminApiBase}/${modelName}?page_size=200`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!resp.ok) {
    return;
  }
  const payload = await resp.json();
  const results = Array.isArray(payload?.results) ? payload.results : [];
  for (const row of results) {
    if (row?.id === undefined || row?.id === null) {
      continue;
    }
    await fetch(`${adminApiBase}/${modelName}/${row.id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    });
  }
}

export default async function globalSetup() {
  // Clean stale admin data in already-running or Postgres-backed runs.
  const isReady = await waitForAdminApi(2000);
  if (isReady) {
    try {
      const loginResp = await fetch(`${adminApiBase}/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: process.env.FORGE_ADMIN_USERNAME || 'admin',
          password: process.env.FORGE_ADMIN_PASSWORD || 'admin123',
        }),
      });
      if (loginResp.ok) {
        const { token } = await loginResp.json();
        await cleanupModel('categories', token);
        await cleanupModel('brands', token);
      }
    } catch {
      // Ignore cleanup error if already clean
    }
  }
}
