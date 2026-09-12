import { test, expect, type APIRequestContext, type Page } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const adminUser = process.env.FORGE_ADMIN_USERNAME || 'admin';
const adminPass = process.env.FORGE_ADMIN_PASSWORD || 'admin123';
const rootDir = path.resolve(process.cwd(), '..');
const serverBaseUrl = process.env.FORGE_SERVER_BASE_URL || 'http://localhost:8020';

test.describe('Ecommerce Framework Features', () => {
  test('REST API resources are reachable', async ({ request }) => {
    const resources = discoverAPIResources();
    expect(resources.length).toBeGreaterThan(0);

    const health = await request.get(`${serverBaseUrl}/health`);
    expect(health.ok()).toBeTruthy();

    for (const resource of resources) {
      const resp = await request.get(`${serverBaseUrl}/api/v1/${resource}/`);
      expect(resp.ok(), `list endpoint failed for ${resource}`).toBeTruthy();
    }
  });

  test('admin config/meta/search/saved-views/autocomplete/export/crud', async ({ page, request }) => {
    await loginUI(page);
    const token = await loginAPI(request);

    const configResp = await request.get(`${serverBaseUrl}/admin/api/config`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(configResp.ok()).toBeTruthy();

    const metaResp = await request.get(`${serverBaseUrl}/admin/api/meta`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(metaResp.ok()).toBeTruthy();
    const metaPayload = await metaResp.json();
    const models = Array.isArray(metaPayload.models) ? metaPayload.models : [];
    expect(models.length).toBeGreaterThan(0);

    const target = models.find((m: any) => m?.name === 'categories') ?? models[0];
    const modelName = target.name as string;

    const modelMetaResp = await request.get(`${serverBaseUrl}/admin/api/meta/${modelName}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(modelMetaResp.ok()).toBeTruthy();
    const modelMeta = await modelMetaResp.json();
    const displayField = pickDisplayField(modelMeta);

    const unique = `e2e-${Date.now()}`;
    const payload = await buildCreatePayload(modelMeta, displayField, unique);

    const createResp = await request.post(`${serverBaseUrl}/admin/api/${modelName}`, {
      headers: { Authorization: `Bearer ${token}` },
      data: payload,
    });
    if (!createResp.ok()) {
      const b = await createResp.text();
      throw new Error(`POST /admin/api/${modelName} failed: status=${createResp.status()} body=${b} payload=${JSON.stringify(payload)}`);
    }
    expect(createResp.ok()).toBeTruthy();
    const created = await createResp.json();
    const createdID = created.id;
    expect(createdID).toBeTruthy();

    const detailResp = await request.get(`${serverBaseUrl}/admin/api/${modelName}/${createdID}/`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(detailResp.ok()).toBeTruthy();

    const updatedValue = `${payload[displayField]}-updated`;
    const patchResp = await request.patch(`${serverBaseUrl}/admin/api/${modelName}/${createdID}/`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { [displayField]: updatedValue },
    });
    if (!patchResp.ok()) {
      const patchBody = await patchResp.text();
      throw new Error(`PATCH failed: status=${patchResp.status()} body=${patchBody}`);
    }

    const searchResp = await request.get(`${serverBaseUrl}/admin/api/search`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { q: updatedValue },
    });
    expect(searchResp.ok()).toBeTruthy();

    const savedViewResp = await request.post(`${serverBaseUrl}/admin/api/saved-views/${modelName}`, {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        name: `view-${unique}`,
        filters: {},
        ordering: [],
        display: [],
      },
    });
    expect(savedViewResp.ok()).toBeTruthy();

    const savedViewListResp = await request.get(`${serverBaseUrl}/admin/api/saved-views/${modelName}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(savedViewListResp.ok()).toBeTruthy();

    const exportResp = await request.get(`${serverBaseUrl}/admin/api/${modelName}/export`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { format: 'json' },
    });
    expect(exportResp.ok()).toBeTruthy();

    const autocompleteResp = await request.get(`${serverBaseUrl}/admin/api/${modelName}/autocomplete`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { q: updatedValue, limit: '5' },
    });
    if (!autocompleteResp.ok()) {
      const autocompleteBody = await autocompleteResp.text();
      throw new Error(
        `AUTOCOMPLETE failed: status=${autocompleteResp.status()} body=${autocompleteBody}`,
      );
    }

    const deleteResp = await request.delete(`${serverBaseUrl}/admin/api/${modelName}/${createdID}/`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(deleteResp.status()).toBe(204);
  });

  test('storefront homepage, openapi spec, and showcase endpoints', async ({ page, request }) => {
    // 1. OpenAPI Specification endpoint
    const openapiResp = await request.get(`${serverBaseUrl}/api/openapi.json`);
    expect(openapiResp.ok()).toBeTruthy();
    const openapi = await openapiResp.json();
    expect(openapi.openapi).toBe('3.0.3');
    expect(openapi.info.title).toContain('Forge Ecommerce Reference API');
    expect(openapi.paths['/api/v1/catalog/search']).toBeDefined();
    expect(openapi.paths['/api/v1/orders/checkout']).toBeDefined();

    // 2. Catalog Stats endpoint (ORM Aggregations)
    const statsResp = await request.get(`${serverBaseUrl}/api/v1/catalog/stats`);
    expect(statsResp.ok()).toBeTruthy();
    const stats = await statsResp.json();
    expect(stats.total_products).toBeGreaterThanOrEqual(0);

    // 3. Faceted Search endpoint
    const searchResp = await request.get(`${serverBaseUrl}/api/v1/catalog/search`, {
      params: { q: 'pro', in_stock: 'true' },
    });
    expect(searchResp.ok()).toBeTruthy();
    const searchData = await searchResp.json();
    expect(Array.isArray(searchData.results)).toBeTruthy();

    // 4. Order Checkout Flow with Lifecycle Hooks
    const checkoutResp = await request.post(`${serverBaseUrl}/api/v1/orders/checkout`, {
      data: {
        customer_id: 1,
        shipping_address: '100 Tech Blvd, Silicon Valley, CA',
        billing_address: '100 Tech Blvd, Silicon Valley, CA',
        payment_method: 'credit_card',
        items: [
          {
            product_id: 1,
            quantity: 1,
            unit_price: 199.99,
          },
        ],
      },
    });
    expect(checkoutResp.status()).toBe(201);
    const orderData = await checkoutResp.json();
    expect(orderData.order_number).toMatch(/^ORD-\d{8}-[A-F0-9]{4}$/);
    expect(orderData.status).toBe('success');
    expect(orderData.order_status).toBe('pending');
    expect(orderData.total_amount).toBeGreaterThan(0);

    // 5. Storefront UI in browser
    await page.goto(`${serverBaseUrl}/`);
    await expect(page.locator('h1')).toContainText('Full-Stack Go Framework');
    await expect(page.locator('.brand')).toContainText('Forge');
    await expect(page.locator('#productGrid')).toBeVisible();
    await expect(page.locator('#api-console')).toBeVisible();

    // Test API console button click
    await page.click('button:has-text("Execute Query")');
    await expect(page.locator('#apiStatus')).toBeVisible({ timeout: 10000 });
  });
});

async function loginUI(page: Page) {
  await page.goto('login');
  await page.evaluate(() => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin-tabs');
  });

  await page.fill('[data-testid="username-input"]', adminUser);
  await page.fill('[data-testid="password-input"]', adminPass);
  await page.click('[data-testid="login-button"]');
  await expect(page.getByTestId('nav-dashboard')).toBeVisible();
}

async function loginAPI(request: APIRequestContext) {
  const response = await request.post(`${serverBaseUrl}/admin/api/login`, {
    data: { username: adminUser, password: adminPass },
  });
  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  return data.token as string;
}

function pickDisplayField(modelMeta: any) {
  const preferred = ['name', 'title', 'label', 'slug', 'code'];
  const fields = Array.isArray(modelMeta?.fields) ? modelMeta.fields : [];
  for (const candidate of preferred) {
    const field = fields.find((f: any) => f?.name === candidate && !f.read_only);
    if (field) {
      return candidate;
    }
  }
  for (const field of fields) {
    if (!field?.read_only && field?.name !== 'id') {
      return field.name;
    }
  }
  return 'id';
}

async function buildCreatePayload(modelMeta: any, displayField: string, unique: string) {
  const fields = Array.isArray(modelMeta?.fields) ? modelMeta.fields : [];
  const payload: Record<string, any> = { [displayField]: `${displayField}-${unique}` };

  for (const field of fields) {
    if (!field || field.read_only || !field.required || field.name === 'id') {
      continue;
    }
    if (payload[field.name] !== undefined) {
      continue;
    }
    payload[field.name] = defaultFieldValue(field, unique);
  }

  return payload;
}

function defaultFieldValue(field: any, unique: string) {
  const fieldType = String(field?.type ?? '').toLowerCase();
  const choices = Array.isArray(field?.choices) ? field.choices : [];
  if (choices.length > 0 && choices[0]?.value !== undefined) {
    return choices[0].value;
  }
  if (fieldType.includes('bool')) {
    return true;
  }
  if (fieldType.includes('int') || fieldType.includes('number')) {
    return 10;
  }
  if (fieldType.includes('float') || fieldType.includes('decimal')) {
    return 10.5;
  }
  if (fieldType.includes('date') || fieldType.includes('time')) {
    return '2024-01-01';
  }
  if (fieldType.includes('json')) {
    return { note: `e2e-${unique}` };
  }
  if (String(field?.name).includes('email')) {
    return `e2e-${unique}@example.com`;
  }
  if (String(field?.name).includes('slug')) {
    return `slug-${unique}`;
  }
  return `${field?.name ?? 'field'}-${unique}`;
}

function discoverAPIResources() {
  const apps = ['catalog', 'customers', 'inventory', 'marketing', 'orders'];
  const resources = new Set<string>();

  for (const app of apps) {
    const filePath = path.join(rootDir, 'app', app, 'api.go');
    if (!fs.existsSync(filePath)) {
      continue;
    }
    const source = fs.readFileSync(filePath, 'utf8');
    const regex = /router\.Register\("([^"]+)"/g;
    let match: RegExpExecArray | null;
    while ((match = regex.exec(source)) !== null) {
      resources.add(match[1]);
    }
  }

  return Array.from(resources).sort();
}
