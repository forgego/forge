import type { Page } from '@playwright/test';
import type {
  Metadata,
  MetadataResponse,
  PaginatedResponse,
  SavedView,
  SearchResponse,
} from '../../src/api/types';

const CONFIG_DATA = {
  plugins: [],
  user: {
    name: 'E2E Admin',
    role: 'Super Admin',
  },
};

const META_DATA: MetadataResponse = {
  models: [
    {
      name: 'categories',
      verbose_name: 'Category',
      verbose_name_plural: 'Categories',
      count: 3,
      permissions: {
        add: true,
        change: true,
        delete: true,
        view: true,
      },
    },
    {
      name: 'products',
      verbose_name: 'Product',
      verbose_name_plural: 'Products',
      count: 10,
      permissions: {
        add: true,
        change: true,
        delete: true,
        view: true,
      },
    },
  ],
  plugins: [],
};

const CATEGORIES_METADATA: Metadata = {
  name: 'categories',
  verbose_name: 'Category',
  verbose_name_plural: 'Categories',
  description: 'Category management',
  fields: [
    {
      name: 'id',
      type: 'integer',
      label: 'ID',
      required: false,
      read_only: true,
      widget: 'number',
    },
    {
      name: 'name',
      type: 'string',
      label: 'Name',
      required: true,
      read_only: false,
      widget: 'text',
    },
    {
      name: 'slug',
      type: 'string',
      label: 'Slug',
      required: false,
      read_only: false,
      widget: 'text',
    },
    {
      name: 'is_active',
      type: 'boolean',
      label: 'Active',
      required: false,
      read_only: false,
      widget: 'checkbox',
    },
    {
      name: 'created_at',
      type: 'datetime',
      label: 'Created At',
      required: false,
      read_only: true,
      widget: 'datetime',
    },
  ],
  relations: [],
  permissions: {
    add: true,
    change: true,
    delete: true,
    view: true,
  },
  actions: [{ name: 'activate', label: 'Activate', confirmation: 'Activate selected categories?' }],
  filters: [
    {
      name: 'is_active',
      type: 'boolean',
      label: 'Active',
    },
  ],
  list_display: ['id', 'name', 'slug', 'is_active', 'created_at'],
  pagination: {
    page_size: 20,
    max_page_size: 100,
  },
};

const CATEGORIES_LIST: PaginatedResponse = {
  count: 3,
  page: 1,
  page_size: 20,
  total_pages: 1,
  results: [
    {
      id: 1,
      name: 'Electronics',
      slug: 'electronics',
      is_active: true,
      created_at: '2025-01-01T12:00:00Z',
    },
    {
      id: 2,
      name: 'Books',
      slug: null,
      is_active: false,
      created_at: '2025-01-02T12:00:00Z',
    },
    {
      id: 3,
      name: 'Clothing',
      slug: 'clothing',
      is_active: true,
      created_at: '2025-01-03T12:00:00Z',
    },
  ],
};

const SAVED_VIEWS_DATA: { views: SavedView[] } = {
  views: [],
};

const SEARCH_DATA: SearchResponse = {
  results: [],
};

export async function mockAdminApi(page: Page, authenticated = true): Promise<void> {
  if (authenticated) {
    await page.addInitScript(() => {
      localStorage.setItem('admin_token', 'e2e-token');
    });
  }

  // A strict fallback makes missing fixture endpoints fail visibly.
  await page.route('**/admin/api/**', async (route) => {
    await route.fulfill({
      status: 404,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'Unmocked admin API request' }),
    });
  });

  const categories = CATEGORIES_LIST.results.map((category) => ({ ...category }));
  let nextId = 4;

  await page.route('**/admin/api/login', async (route) => {
    const credentials = route.request().postDataJSON();
    await route.fulfill({
      status: credentials.username === 'admin' && credentials.password === 'admin123' ? 200 : 401,
      json: credentials.username === 'admin' && credentials.password === 'admin123'
        ? { token: 'e2e-token' }
        : { error: { message: 'Invalid username or password' } },
    });
  });

  await page.route('**/admin/api/categories/action/activate', async (route) => {
    const { ids } = route.request().postDataJSON() as { ids: number[] };
    for (const category of categories) {
      if (ids.includes(category.id)) category.is_active = true;
    }
    await route.fulfill({ json: { affected: ids.length, message: 'Categories activated' } });
  });

  await page.route(/\/admin\/api\/categories\/\d+\/history$/, async (route) => {
    await route.fulfill({ json: { entries: [] } });
  });

  await page.route(/\/admin\/api\/categories\/\d+(?:\?.*)?$/, async (route) => {
    const id = Number(new URL(route.request().url()).pathname.split('/').pop());
    const index = categories.findIndex((category) => category.id === id);
    if (index < 0) {
      await route.fulfill({ status: 404, json: { error: 'Category not found' } });
      return;
    }
    const method = route.request().method();
    if (method === 'DELETE') {
      categories.splice(index, 1);
      await route.fulfill({ status: 204, body: '' });
    } else if (method === 'PATCH' || method === 'PUT') {
      categories[index] = { ...categories[index], ...route.request().postDataJSON() };
      await route.fulfill({ json: categories[index] });
    } else {
      await route.fulfill({ json: categories[index] });
    }
  });

  await page.route(/\/admin\/api\/categories(?:\?.*)?$/, async (route) => {
    if (route.request().method() === 'POST') {
      const category = {
        id: nextId++,
        name: '',
        slug: null,
        is_active: false,
        created_at: new Date().toISOString(),
        ...route.request().postDataJSON(),
      };
      categories.push(category);
      await route.fulfill({ status: 201, json: category });
      return;
    }
    const params = new URL(route.request().url()).searchParams;
    const active = params.get('is_active');
    const results = categories.filter((category) =>
      active === null || category.is_active === (active === 'true')
    );
    await route.fulfill({ json: {
      count: results.length,
      page: 1,
      page_size: 20,
      total_pages: 1,
      results,
    } });
  });

  // 2. Specific routes registered AFTER (Playwright matches most recently added first)
  await page.route('**/admin/api/search?**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(SEARCH_DATA),
    });
  });

  await page.route('**/admin/api/search', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(SEARCH_DATA),
    });
  });

  await page.route('**/admin/api/saved-views/**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(SAVED_VIEWS_DATA),
    });
  });

  await page.route('**/admin/api/meta', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(META_DATA),
    });
  });

  await page.route('**/admin/api/meta/categories', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(CATEGORIES_METADATA),
    });
  });

  await page.route('**/admin/api/config', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(CONFIG_DATA),
    });
  });
}
