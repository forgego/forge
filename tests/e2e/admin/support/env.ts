// Connection settings for the application under test. The suite never
// starts a server: CI (or a developer) starts the built application against
// an isolated PostgreSQL database and points these variables at it.

export const baseURL = process.env.ADMIN_E2E_BASE_URL || 'http://localhost:8000';
export const adminPath = normalizePath(process.env.ADMIN_E2E_PATH || '/admin');
export const apiPath = normalizePath(process.env.ADMIN_E2E_API_PATH || `${adminPath}/api`);
export const adminUser = process.env.ADMIN_E2E_USERNAME || 'admin';
export const adminPass = process.env.ADMIN_E2E_PASSWORD || 'admin123';

export function normalizePath(value: string) {
  let path = value.trim();
  if (!path.startsWith('/')) path = `/${path}`;
  if (path.length > 1 && path.endsWith('/')) path = path.slice(0, -1);
  return path;
}

/** URL of an admin UI route, e.g. adminURL('/categories'). */
export function adminURL(route = '/') {
  return `${adminPath}${route === '/' ? '/' : route}`;
}

/** Absolute URL of an admin API route, e.g. apiURL('/categories'). */
export function apiURL(route: string) {
  return new URL(`${apiPath}${route}`, baseURL).toString();
}

/** Short lowercase tag that isolates one test's fixture rows. */
export function uniqueTag(prefix = 'e2e') {
  const rand = Math.floor(Math.random() * 36 ** 4).toString(36).padStart(4, '0');
  return `${prefix}${Date.now().toString(36)}${rand}`;
}
