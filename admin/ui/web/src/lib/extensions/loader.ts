import { ForgeAdmin, FORGE_ADMIN_API_VERSION, setCurrentExtensionContext } from "./api";
import type { MetadataResponse, ExtensionMetadata } from "../../api/types";

export interface ExtensionFailure {
  name: string;
  version?: number;
  reason: "unsupported_version" | "import_failed" | "execution_failed";
  error: Error;
}

let extensionFailures: ExtensionFailure[] = [];
const loadedExtensions = new Set<string>();

export function getExtensionFailures(): ExtensionFailure[] {
  return [...extensionFailures];
}

export function clearExtensionFailures(): void {
  extensionFailures = [];
  loadedExtensions.clear();
}

/**
 * Loads registered admin extensions from metadata.
 * Executes module.default(ForgeAdmin) for each valid extension.
 * Fails safely: records and logs failures, never throws.
 */
export async function loadExtensions(
  meta?: (MetadataResponse & { extensions?: ExtensionMetadata[] }) | { extensions?: ExtensionMetadata[] } | null
): Promise<ExtensionFailure[]> {
  if (!meta || !meta.extensions || !Array.isArray(meta.extensions)) {
    return getExtensionFailures();
  }

  for (const ext of meta.extensions) {
    if (!ext || !ext.name) continue;

    // Skip if already loaded
    if (loadedExtensions.has(ext.name)) continue;

    // Check version compatibility
    if (ext.version !== FORGE_ADMIN_API_VERSION) {
      const err = new Error(
        `Extension "${ext.name}" specifies unsupported API version ${ext.version} (supported version: ${FORGE_ADMIN_API_VERSION})`
      );
      console.warn(`[ExtensionLoader] ${err.message}`);
      extensionFailures.push({
        name: ext.name,
        version: ext.version,
        reason: "unsupported_version",
        error: err,
      });
      continue;
    }

    // Attempt dynamic import and execution
    setCurrentExtensionContext(ext.name);
    try {
      const mod = await import(/* @vite-ignore */ ext.entry);
      if (typeof mod?.default === "function") {
        await mod.default(ForgeAdmin);
        loadedExtensions.add(ext.name);
      } else if (mod && typeof mod === "object" && typeof (mod as any).init === "function") {
        await (mod as any).init(ForgeAdmin);
        loadedExtensions.add(ext.name);
      } else {
        throw new Error(
          `Extension "${ext.name}" loaded from "${ext.entry}" does not provide a default export function`
        );
      }
    } catch (err: any) {
      const errorObj = err instanceof Error ? err : new Error(String(err));
      console.error(
        `[ExtensionLoader] Failed to load extension "${ext.name}":`,
        errorObj
      );
      extensionFailures.push({
        name: ext.name,
        version: ext.version,
        reason: "import_failed",
        error: errorObj,
      });
    } finally {
      setCurrentExtensionContext(null);
    }
  }

  return getExtensionFailures();
}
