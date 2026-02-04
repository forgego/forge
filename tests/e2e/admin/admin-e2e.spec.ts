import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

type ModelListEntry = {
  name: string;
  permissions: { add: boolean; change: boolean; delete: boolean; view: boolean };
};

type FieldMetadata = {
  name: string;
  type: string;
  label: string;
  required: boolean;
  read_only: boolean;
  choices?: { value: any; label: string }[];
};

type RelationMetadata = {
  name: string;
  type: string;
  related_model: string;
};

type ModelMetadata = {
  name: string;
  verbose_name: string;
  list_display: string[];
  permissions: { add: boolean; change: boolean; delete: boolean; view: boolean };
  fields: FieldMetadata[];
  relations?: RelationMetadata[];
  filters: { name: string; type: string; choices?: { value: any }[] }[];
  search_fields?: string[];
  ordering?: string[];
};

const baseURL = process.env.ADMIN_E2E_BASE_URL || 'http://localhost:8000';
const adminPath = normalizePath(process.env.ADMIN_E2E_PATH || '/admin');
const apiPath = normalizePath(process.env.ADMIN_E2E_API_PATH || '/admin/api');
const adminUser = process.env.ADMIN_E2E_USERNAME || 'admin';
const adminPass = process.env.ADMIN_E2E_PASSWORD || 'password';

const displayFieldCandidates = [
  'name',
  'title',
  'label',
  'email',
  'username',
  'display_name',
  'code',
  'slug',
];

const relationTypeAliases = {
  foreign_key: ['foreignkey', 'foreign_key', 'foreignkeyfield', 'foreignkeyid'],
  one_to_one: ['onetoone', 'one_to_one'],
  many_to_many: ['manytomany', 'many_to_many'],
};

test.describe('Admin E2E', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.removeItem('admin_token');
      localStorage.removeItem('admin-tabs');
    });
  });

  test('login and list pages load', async ({ page, request }) => {
    const token = await loginAPI(request);
    await loginUI(page);

    const models = await fetchModels(request, token);
    const listTargets = models
      .filter((model) => model.permissions.view)
      .slice(0, 3);

    for (const model of listTargets) {
      await page.getByTestId(`nav-${model.name}`).click();
      await expect(page).toHaveURL(new RegExp(`${adminPath}/${model.name}`));
      await expect(page.locator('table')).toBeVisible();
    }
  });

  test('create, edit, filter/search, order, and delete records', async ({ page, request }) => {
    const token = await loginAPI(request);
    await loginUI(page);

    const models = await fetchModels(request, token);
    const modelMeta = await findModel(models, request, token, (meta) => {
      return (
        meta.permissions.add &&
        meta.permissions.change &&
        meta.permissions.delete &&
        (meta.relations ?? []).length === 0 &&
        meta.fields.some((field) => field.required && !field.read_only && field.name !== 'id')
      );
    });

    const displayField = pickDisplayField(modelMeta);
    const unique = `e2e-${Date.now()}`;
    const seedPayload = await buildPayload(modelMeta, request, token, {
      overrides: { [displayField]: `${modelMeta.name}-${unique}` },
    });

    await page.goto(`${adminPath}/${modelMeta.name}`);
    await page.click('[data-testid="create-button"]');
    await fillForm(page, modelMeta, seedPayload);
    await page.click('[data-testid="submit-button"]');

    await expect(page.locator('table')).toContainText(String(seedPayload[displayField]));

    const createdRow = page.locator('tr', { hasText: String(seedPayload[displayField]) });
    const editButton = createdRow.locator('[data-testid^="edit-"]');
    const editTestId = (await editButton.getAttribute('data-testid')) || '';
    const createdId = editTestId.replace('edit-', '');
    expect(createdId).not.toBe('');

    await editButton.click();
    await expect(page).toHaveURL(new RegExp(`${adminPath}/${modelMeta.name}/${createdId}$`));

    const updatedValue = `${seedPayload[displayField]}-updated`;
    await fillForm(page, modelMeta, { [displayField]: updatedValue }, { replace: true });
    await page.click('[data-testid="submit-button"]');

    await expect(page.locator('table')).toContainText(updatedValue);

    await exerciseSearchAndFilters(page, modelMeta, updatedValue);
    await exerciseOrdering(request, token, modelMeta, displayField, unique);

    const updatedRow = page.locator('tr', { hasText: updatedValue });
    await updatedRow.locator(`[data-testid="delete-${createdId}"]`).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await expect(page.locator('table')).not.toContainText(updatedValue);
  });

  test('relation fields (foreign key + many-to-many)', async ({ page, request }) => {
    const token = await loginAPI(request);
    await loginUI(page);

    const models = await fetchModels(request, token);

    const fkModel = await findModel(models, request, token, (meta) => {
      return (
        meta.permissions.add &&
        meta.permissions.change &&
        meta.relations?.some((rel) => isRelationType(rel.type, 'foreign_key') || isRelationType(rel.type, 'one_to_one'))
      );
    });

    const m2mModel = await findModel(models, request, token, (meta) => {
      return (
        meta.permissions.add &&
        meta.permissions.change &&
        meta.relations?.some((rel) => isRelationType(rel.type, 'many_to_many'))
      );
    });

    const fkRecord = await seedWithRelations(request, token, fkModel, { includeManyToMany: false });
    const m2mRecord = await seedWithRelations(request, token, m2mModel, { includeForeignKey: false });

    await page.goto(`${adminPath}/${fkModel.name}/${fkRecord.id}`);
    await expectRelationChips(page, fkRecord.displayLabels);

    await page.goto(`${adminPath}/${m2mModel.name}/${m2mRecord.id}`);
    await expectRelationChips(page, m2mRecord.displayLabels);
  });
});

function normalizePath(value: string) {
  let path = value.trim();
  if (!path.startsWith('/')) {
    path = `/${path}`;
  }
  if (path.length > 1 && path.endsWith('/')) {
    path = path.slice(0, -1);
  }
  return path;
}

function apiURL(path: string) {
  return new URL(`${apiPath}${path}`, baseURL).toString();
}

async function loginAPI(request: APIRequestContext) {
  const response = await request.post(apiURL('/login'), {
    data: { username: adminUser, password: adminPass },
  });
  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  return data.token as string;
}

async function loginUI(page: Page) {
  await page.goto(`${adminPath}/login`);
  await page.fill('[data-testid="username-input"]', adminUser);
  await page.fill('[data-testid="password-input"]', adminPass);
  await page.click('[data-testid="login-button"]');
  await expect(page.getByTestId('nav-dashboard')).toBeVisible();
}

async function fetchModels(request: APIRequestContext, token: string) {
  const response = await request.get(apiURL('/meta'), {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  return data.models as ModelListEntry[];
}

async function fetchModelMeta(request: APIRequestContext, token: string, modelName: string) {
  const response = await request.get(apiURL(`/meta/${modelName}`), {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(response.ok()).toBeTruthy();
  return (await response.json()) as ModelMetadata;
}

async function findModel(
  models: ModelListEntry[],
  request: APIRequestContext,
  token: string,
  predicate: (meta: ModelMetadata) => boolean
) {
  for (const model of models) {
    const meta = await fetchModelMeta(request, token, model.name);
    if (predicate(meta)) {
      return meta;
    }
  }
  throw new Error('No model matched the required criteria. Provide ADMIN_E2E_* overrides.');
}

function pickDisplayField(meta: ModelMetadata) {
  for (const candidate of meta.list_display ?? []) {
    if (candidate !== 'id') {
      return candidate;
    }
  }
  for (const candidate of displayFieldCandidates) {
    if (meta.fields.some((field) => field.name === candidate)) {
      return candidate;
    }
  }
  const fallback = meta.fields.find((field) => field.name !== 'id');
  return fallback?.name || 'id';
}

function normalizeFieldType(fieldType: string) {
  return fieldType.toLowerCase().replace(/[^a-z0-9]/g, '');
}

function isRelationType(value: string, relation: keyof typeof relationTypeAliases) {
  const normalized = normalizeFieldType(value);
  return relationTypeAliases[relation].some((alias) => alias === normalized);
}

function buildScalarValue(field: FieldMetadata, unique: string) {
  const normalized = normalizeFieldType(field.type);
  if (field.choices && field.choices.length > 0) {
    return field.choices[0].value;
  }

  if (['string', 'text', 'email', 'url', 'uuid'].includes(normalized)) {
    if (field.name.includes('slug')) {
      return `slug-${unique}`;
    }
    if (field.name.includes('email')) {
      return `e2e-${unique}@example.com`;
    }
    return `${field.label || field.name}-${unique}`.slice(0, 255);
  }

  if (['int64', 'int32', 'integer', 'number'].includes(normalized)) {
    return 10;
  }

  if (['float64', 'float32', 'decimal', 'float'].includes(normalized)) {
    return 10.5;
  }

  if (['bool', 'boolean'].includes(normalized)) {
    return true;
  }

  if (['date', 'datetime', 'time'].includes(normalized)) {
    return '2024-01-01';
  }

  if (normalized === 'json') {
    return { note: 'e2e' };
  }

  return `${field.label || field.name}-${unique}`.slice(0, 255);
}

async function buildPayload(
  meta: ModelMetadata,
  request: APIRequestContext,
  token: string,
  options: { overrides?: Record<string, any>; depth?: number } = {}
) {
  const overrides = options.overrides ?? {};
  const relationMap = new Map((meta.relations ?? []).map((rel) => [rel.name, rel]));
  const unique = `${Date.now()}-${Math.floor(Math.random() * 1000)}`;
  const payload: Record<string, any> = { ...overrides };

  for (const field of meta.fields) {
    if (payload[field.name] !== undefined) {
      continue;
    }

    if (field.read_only || field.name === 'id') {
      continue;
    }

    if (!field.required) {
      continue;
    }

    const relation = relationMap.get(field.name);
    if (relation) {
      const relatedMeta = await fetchModelMeta(request, token, relation.related_model);
      const related = await createObject(request, token, relatedMeta, { depth: (options.depth ?? 0) + 1 });
      payload[field.name] = related.id;
      continue;
    }

    payload[field.name] = buildScalarValue(field, unique);
  }

  return payload;
}

async function createObject(
  request: APIRequestContext,
  token: string,
  meta: ModelMetadata,
  options: { overrides?: Record<string, any>; depth?: number } = {}
) {
  if ((options.depth ?? 0) > 2) {
    throw new Error(`Relation depth too deep for ${meta.name}`);
  }

  const payload = await buildPayload(meta, request, token, options);
  const response = await request.post(apiURL(`/${meta.name}`), {
    headers: { Authorization: `Bearer ${token}` },
    data: payload,
  });

  expect(response.ok()).toBeTruthy();
  return await response.json();
}

async function fillForm(
  page: Page,
  meta: ModelMetadata,
  payload: Record<string, any>,
  options: { replace?: boolean } = {}
) {
  const fields = new Map(meta.fields.map((field) => [field.name, field]));

  for (const [name, value] of Object.entries(payload)) {
    const field = fields.get(name);
    if (!field || field.read_only) {
      continue;
    }

    const input = page.locator(`#${name}`);
    if (await input.count()) {
      const tagName = await input.evaluate((el) => el.tagName.toLowerCase());
      const normalizedType = normalizeFieldType(field.type);

      if (['bool', 'boolean'].includes(normalizedType)) {
        if (value) {
          await input.check();
        } else {
          await input.uncheck();
        }
        continue;
      }

      if (tagName === 'select') {
        await input.selectOption(String(value));
        continue;
      }

      if (options.replace) {
        await input.fill(String(value));
      } else {
        await input.fill(String(value));
      }
      continue;
    }
  }
}

async function exerciseSearchAndFilters(page: Page, meta: ModelMetadata, searchValue: string) {
  if (meta.search_fields && meta.search_fields.length > 0) {
    await page.fill('[data-testid="search-input"]', searchValue);
    await expect(page.locator('table')).toContainText(searchValue);
    await page.fill('[data-testid="search-input"]', '');
  }

  const filter = meta.filters.find((item) => ['boolean', 'choice'].includes(item.type));
  if (!filter) {
    return;
  }

  await page.click('[data-testid="filter-button"]');
  const filterSelector = `[data-testid="filter-${filter.name}"]`;

  if (filter.type === 'boolean') {
    await page.selectOption(filterSelector, 'true');
  } else if (filter.choices && filter.choices.length > 0) {
    await page.selectOption(filterSelector, String(filter.choices[0].value));
  }
}

async function exerciseOrdering(
  request: APIRequestContext,
  token: string,
  meta: ModelMetadata,
  displayField: string,
  unique: string
) {
  const orderingField = meta.ordering?.find((field) => !field.startsWith('-')) || meta.ordering?.[0];
  if (!orderingField) {
    return;
  }

  const normalizedField = orderingField.replace('-', '');
  const fieldMeta = meta.fields.find((field) => field.name === normalizedField);
  if (!fieldMeta || fieldMeta.read_only) {
    return;
  }

  const firstPayload = await buildPayload(meta, request, token, {
    overrides: { [displayField]: `${displayField}-${unique}-1`, [normalizedField]: 1 },
  });
  const secondPayload = await buildPayload(meta, request, token, {
    overrides: { [displayField]: `${displayField}-${unique}-2`, [normalizedField]: 2 },
  });

  await request.post(apiURL(`/${meta.name}`), {
    headers: { Authorization: `Bearer ${token}` },
    data: firstPayload,
  });
  await request.post(apiURL(`/${meta.name}`), {
    headers: { Authorization: `Bearer ${token}` },
    data: secondPayload,
  });

  const response = await request.get(apiURL(`/${meta.name}`), {
    headers: { Authorization: `Bearer ${token}` },
    params: { ordering: orderingField },
  });

  if (!response.ok()) {
    return;
  }

  const data = await response.json();
  const results = data.results as Record<string, any>[];
  if (results.length < 2) {
    return;
  }

  const firstValue = results[0][normalizedField];
  const secondValue = results[1][normalizedField];
  if (orderingField.startsWith('-')) {
    expect(firstValue).toBeGreaterThanOrEqual(secondValue);
  } else {
    expect(firstValue).toBeLessThanOrEqual(secondValue);
  }
}

async function seedWithRelations(
  request: APIRequestContext,
  token: string,
  meta: ModelMetadata,
  options: { includeForeignKey?: boolean; includeManyToMany?: boolean }
) {
  const relationMap = meta.relations ?? [];
  const payload = await buildPayload(meta, request, token, {});
  const displayLabels: string[] = [];

  for (const relation of relationMap) {
    if (isRelationType(relation.type, 'foreign_key') || isRelationType(relation.type, 'one_to_one')) {
      if (options.includeForeignKey === false) {
        continue;
      }
      const relatedMeta = await fetchModelMeta(request, token, relation.related_model);
      const relatedDisplayField = pickDisplayField(relatedMeta);
      const relatedValue = `fk-${relation.related_model}-${Date.now()}`;
      const related = await createObject(request, token, relatedMeta, {
        overrides: { [relatedDisplayField]: relatedValue },
      });
      payload[relation.name] = related.id;
      displayLabels.push(relatedValue);
    }

    if (isRelationType(relation.type, 'many_to_many')) {
      if (options.includeManyToMany === false) {
        continue;
      }
      const relatedMeta = await fetchModelMeta(request, token, relation.related_model);
      const relatedDisplayField = pickDisplayField(relatedMeta);
      const relatedValue = `m2m-${relation.related_model}-${Date.now()}`;
      const related = await createObject(request, token, relatedMeta, {
        overrides: { [relatedDisplayField]: relatedValue },
      });
      payload[relation.name] = [related.id];
      displayLabels.push(relatedValue);
    }
  }

  const response = await request.post(apiURL(`/${meta.name}`), {
    headers: { Authorization: `Bearer ${token}` },
    data: payload,
  });

  expect(response.ok()).toBeTruthy();
  const record = await response.json();

  return { id: record.id, displayLabels };
}

async function expectRelationChips(page: Page, labels: string[]) {
  for (const label of labels) {
    await expect(page.locator(`text=${label}`)).toBeVisible();
  }
}
