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
  actions: [],
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

export async function mockAdminApi(page: Page): Promise<void> {
  await page.addInitScript(() => {
    localStorage.setItem('admin_token', 'e2e-token');
  });

  // 1. Catch-all route registered FIRST
  await page.route('**/admin/api/**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({}),
    });
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

  await page.route('**/admin/api/categories?**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(CATEGORIES_LIST),
    });
  });

  await page.route('**/admin/api/categories', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(CATEGORIES_LIST),
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
