import { expect, type APIRequestContext, type APIResponse } from '@playwright/test';
import { adminPass, adminUser, apiURL } from './env';

type Row = Record<string, any>;

/**
 * Thin client for the admin REST API. Tests use it to seed fixtures and to
 * read back what the database actually stored after a browser operation.
 */
export class AdminApi {
  constructor(
    private readonly request: APIRequestContext,
    readonly token: string,
  ) {}

  static async login(request: APIRequestContext, username = adminUser, password = adminPass) {
    const response = await request.post(apiURL('/login'), { data: { username, password } });
    expect(response.status(), await response.text()).toBe(200);
    const body = await response.json();
    expect(typeof body.token).toBe('string');
    return new AdminApi(request, body.token as string);
  }

  private get headers() {
    return { Authorization: `Bearer ${this.token}` };
  }

  raw(method: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE', route: string, data?: unknown): Promise<APIResponse> {
    return this.request.fetch(apiURL(route), { method, headers: this.headers, data });
  }

  async create(model: string, data: Row): Promise<Row> {
    const response = await this.raw('POST', `/${model}`, data);
    expect(response.status(), `create ${model}: ${await response.text()}`).toBe(201);
    return response.json();
  }

  async get(model: string, id: number | string): Promise<Row> {
    const response = await this.raw('GET', `/${model}/${id}`);
    expect(response.status(), `get ${model}/${id}: ${await response.text()}`).toBe(200);
    return response.json();
  }

  async exists(model: string, id: number | string): Promise<boolean> {
    const response = await this.raw('GET', `/${model}/${id}`);
    if (response.status() === 404) return false;
    expect(response.status(), await response.text()).toBe(200);
    return true;
  }

  async list(model: string, params: Record<string, string | number> = {}) {
    const query = new URLSearchParams(Object.entries(params).map(([k, v]) => [k, String(v)]));
    const response = await this.raw('GET', `/${model}?${query}`);
    expect(response.status(), await response.text()).toBe(200);
    const body = await response.json();
    return { count: body.count as number, results: (body.results ?? []) as Row[] };
  }

  async patch(model: string, id: number | string, data: Row): Promise<Row> {
    const response = await this.raw('PATCH', `/${model}/${id}`, data);
    expect(response.status(), `patch ${model}/${id}: ${await response.text()}`).toBe(200);
    return response.json();
  }

  /**
   * Creates a row, then applies falsy values with an update.
   *
   * Create drops zero values such as `false` so a column default wins
   * (see the known-defect spec), so fixtures that need `false` on a column
   * that defaults to true set it in a second step.
   */
  private async createThenApplyFalsy(model: string, data: Row) {
    const falsy: Row = {};
    const initial: Row = {};
    for (const [key, value] of Object.entries(data)) {
      if (value === false || value === 0) falsy[key] = value;
      else initial[key] = value;
    }
    const created = await this.create(model, initial);
    if (Object.keys(falsy).length === 0) return created;
    return this.patch(model, created.id, falsy);
  }

  async createCategory(tag: string, overrides: Row = {}) {
    const { n = '', ...rest } = overrides;
    return this.createThenApplyFalsy('categories', {
      name: `${tag} Category${n}`,
      slug: `${tag}-category${n}`.toLowerCase(),
      is_active: true,
      ...rest,
    });
  }

  async createWarehouse(tag: string, overrides: Row = {}) {
    const code = (overrides.code as string) ?? `${tag}-wh`;
    return this.createThenApplyFalsy('warehouses', {
      name: `${tag} Warehouse`,
      code: code.slice(0, 50),
      address_line1: '1 Dock Road',
      city: 'Rotterdam',
      postal_code: '3011',
      country_code: 'NL',
      country_name: 'Netherlands',
      is_active: false,
      ...overrides,
    });
  }
}
