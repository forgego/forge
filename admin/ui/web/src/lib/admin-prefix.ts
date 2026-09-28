export const DEFAULT_PREFIX = "/admin";

declare global {
  interface Window {
    __forgeAssetUrl?: (filename: string) => string;
  }
}

/**
 * Normalizes mount prefix:
 * null/undefined -> DEFAULT_PREFIX; "" or "/" -> ""; ensure leading "/", strip trailing "/"
 */
export function normalizePrefix(p: string | null | undefined): string {
  if (p === null || p === undefined) {
    return DEFAULT_PREFIX;
  }
  const trimmed = p.trim();
  if (trimmed === "" || trimmed === "/") {
    return "";
  }
  let normalized = trimmed;
  if (!normalized.startsWith("/")) {
    normalized = "/" + normalized;
  }
  normalized = normalized.replace(/\/+$/, "");
  if (normalized === "") {
    return "";
  }
  return normalized;
}

/**
 * Reads meta[name="forge-admin-prefix"]; if the tag exists use its content (even ""),
 * otherwise DEFAULT_PREFIX; always normalizePrefix.
 */
export function getAdminPrefix(
  doc: Document = typeof document !== "undefined" ? document : (undefined as unknown as Document)
): string {
  if (!doc) {
    return DEFAULT_PREFIX;
  }
  const meta = doc.querySelector('meta[name="forge-admin-prefix"]');
  if (meta) {
    const content = meta.getAttribute("content");
    return normalizePrefix(content ?? "");
  }
  return DEFAULT_PREFIX;
}

/**
 * If prefix === "" return path; if path === prefix return "/";
 * if path starts with prefix + "/" return path.slice(prefix.length); else return path
 */
export function stripAdminPrefix(path: string, prefix = getAdminPrefix()): string {
  const normPrefix = normalizePrefix(prefix);
  if (normPrefix === "") return path;
  if (path === normPrefix) return "/";
  if (path.startsWith(normPrefix + "/")) return path.slice(normPrefix.length);
  return path;
}

/**
 * Joins prefix + path with exactly one "/"
 */
export function withAdminPrefix(path: string, prefix = getAdminPrefix()): string {
  const p = normalizePrefix(prefix);
  const cleanPath = path.startsWith("/") ? path : `/${path}`;
  if (!p) {
    return cleanPath;
  }
  return `${p}${cleanPath}`;
}

/**
 * Returns the asset URL for a given filename relative to the admin mount prefix.
 * Used by renderBuiltUrl at runtime.
 */
export function assetUrl(filename: string): string {
  const cleanFilename = filename.startsWith("/") ? filename.slice(1) : filename;
  return withAdminPrefix("/" + cleanFilename);
}

if (typeof window !== "undefined") {
  (window as unknown as { __forgeAssetUrl?: (f: string) => string }).__forgeAssetUrl = assetUrl;
}
