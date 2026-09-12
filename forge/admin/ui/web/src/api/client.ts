import axios from "axios";
import type { AxiosInstance, AxiosError } from "axios";
import type {
  Metadata,
  ModelListMetadata,
  PaginatedResponse,
  ListParams,
  ModelFormData,
  BulkActionRequest,
  BulkActionResponse,
  BulkCreateResponse,
  BulkUpdateResponse,
  BulkDeleteResponse,
  SearchRequest,
  SearchResponse,
  AutocompleteResponse,
  UploadResponse,
  ErrorResponse,
  MetadataResponse,
  HistoryResponse,
  SavedView,
  SavedViewRequest,
} from "./types";

export class AdminAPIClient {
  private client: AxiosInstance;
  private baseURL: string;

  constructor(baseURL: string = "/admin/api") {
    this.baseURL = baseURL;
    this.client = axios.create({
      baseURL,
      headers: {
        "Content-Type": "application/json",
      },
      withCredentials: true,
    });

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        // Add auth token if available
        const token = localStorage.getItem("admin_token");
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ErrorResponse>) => {
        if (error.response?.status === 401) {
          // Do not redirect if this is the login request itself
          const isLoginRequest = error.config?.url?.endsWith("/login");
          if (!isLoginRequest) {
            localStorage.removeItem("admin_token");
            const currentPath = window.location.pathname;
            if (!currentPath.endsWith("/login")) {
              const adminPrefix = currentPath.startsWith("/admin") ? "/admin" : "";
              window.location.href = `${adminPrefix}/login`;
            }
          }
        }
        return Promise.reject(error);
      }
    );
  }

  // Metadata endpoints
  async getConfig(): Promise<any> {
    const response = await this.client.get("/config");
    return response.data;
  }

  async getMetadata(): Promise<MetadataResponse> {
    const response = await this.client.get("/meta");
    return response.data;
  }

  // Generic GET request for dynamic routes (e.g. plugins)
  async get<T = any>(url: string): Promise<T> {
    const response = await this.client.get(url);
    return response.data as T;
  }

  async getModels(): Promise<{ models: ModelListMetadata[] }> {
    return this.getMetadata();
  }

  async getModelMetadata(model: string): Promise<Metadata> {
    const response = await this.client.get(`/meta/${model}`);
    return response.data;
  }

  // CRUD endpoints
  async listObjects<T = any>(
    model: string,
    params?: ListParams
  ): Promise<PaginatedResponse<T>> {
    const response = await this.client.get(`/${model}`, { params });
    return response.data;
  }

  async getObject<T = any>(model: string, id: string | number): Promise<T> {
    const response = await this.client.get(`/${model}/${id}`);
    return response.data;
  }

  async getHistory(
    model: string,
    id: string | number
  ): Promise<HistoryResponse> {
    const response = await this.client.get(`/${model}/${id}/history`);
    return response.data;
  }

  async createObject<T = any>(model: string, data: ModelFormData): Promise<T> {
    const response = await this.client.post(`/${model}`, data);
    return response.data;
  }

  async updateObject<T = any>(
    model: string,
    id: string | number,
    data: Partial<ModelFormData>
  ): Promise<T> {
    const response = await this.client.patch(`/${model}/${id}`, data);
    return response.data;
  }

  async replaceObject<T = any>(
    model: string,
    id: string | number,
    data: ModelFormData
  ): Promise<T> {
    const response = await this.client.put(`/${model}/${id}`, data);
    return response.data;
  }

  async deleteObject(model: string, id: string | number): Promise<void> {
    await this.client.delete(`/${model}/${id}`);
  }

  // Auth
  async login(credentials: any): Promise<{ token: string }> {
    // Try standard auth endpoint
    const response = await this.client.post("/login", credentials);
    return response.data;
  }

  async logout(): Promise<{ message?: string }> {
    try {
      const response = await this.client.post("/logout");
      return response.data ?? {};
    } finally {
      localStorage.removeItem("admin_token");
    }
  }
  // Bulk operations
  async bulkCreate<T = any>(
    model: string,
    data: ModelFormData[]
  ): Promise<BulkCreateResponse<T>> {
    const response = await this.client.post(`/${model}/bulk-create`, data);
    return response.data;
  }

  async bulkUpdate<T = any>(
    model: string,
    ids: (string | number)[],
    data: Partial<ModelFormData>
  ): Promise<BulkUpdateResponse<T>> {
    const response = await this.client.post(`/${model}/bulk-update`, {
      ids,
      data,
    });
    return response.data;
  }

  // Bulk delete returns 204 (full success, empty body) or 207 with a
  // {deleted, errors} payload on partial success. Always normalized to
  // a BulkDeleteResponse so callers can surface partial failures.
  async bulkDelete(
    model: string,
    ids: (string | number)[]
  ): Promise<BulkDeleteResponse> {
    const response = await this.client.delete(`/${model}/bulk-delete`, {
      data: { ids },
    });
    if (response.status === 204 || response.data == null || response.data === "") {
      return { deleted: ids.length };
    }
    return response.data as BulkDeleteResponse;
  }

  // Actions
  async executeAction(
    model: string,
    action: string,
    request: BulkActionRequest
  ): Promise<BulkActionResponse> {
    const response = await this.client.post(
      `/${model}/action/${action}`,
      request
    );
    return response.data;
  }

  // Search and autocomplete
  async globalSearch(request: SearchRequest): Promise<SearchResponse> {
    const response = await this.client.get("/search", {
      params: { q: request.query, models: request.models?.join(",") },
    });
    return response.data;
  }

  async autocomplete(
    model: string,
    field: string,
    query: string,
    limit: number = 10
  ): Promise<AutocompleteResponse> {
    const response = await this.client.get(`/${model}/autocomplete`, {
      params: { field, q: query, limit },
    });
    return response.data;
  }

  // File upload is not implemented by the admin REST API (no /upload
  // route exists). Kept as an explicit stub so misuse fails loudly
  // instead of 404ing silently.
  async uploadFile(_model: string, _file: File): Promise<UploadResponse> {
    throw new Error(
      "File upload is not supported by the admin API. Implement an upload endpoint or a custom widget first."
    );
  }

  // Export list results
  getExportURL(model: string, format: "csv" | "json", params?: ListParams) {
    const searchParams = new URLSearchParams();

    if (format) {
      searchParams.set("format", format);
    }
    if (params?.search) {
      searchParams.set("search", params.search);
    }
    if (params?.ordering) {
      searchParams.set("ordering", params.ordering);
    }

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (
          value === undefined ||
          value === null ||
          value === "" ||
          key === "search" ||
          key === "ordering" ||
          key === "page" ||
          key === "page_size"
        ) {
          return;
        }
        if (Array.isArray(value)) {
          if (value.length > 0) {
            searchParams.set(key, value.join(","));
          }
          return;
        }
        if (typeof value === "object") {
          searchParams.set(key, JSON.stringify(value));
          return;
        }
        searchParams.set(key, String(value));
      });
    }

    const query = searchParams.toString();
    return `${this.baseURL}/${model}/export${query ? `?${query}` : ""}`;
  }

  // Download an export as a file. Uses an authenticated fetch (blob)
  // instead of window.open so the Bearer token is sent and popup
  // blockers cannot swallow the download.
  async downloadExport(
    model: string,
    format: "csv" | "json",
    params?: ListParams
  ): Promise<{ blob: Blob; filename: string }> {
    const url = this.getExportURL(model, format, params);
    const token = localStorage.getItem("admin_token");
    const response = await fetch(url, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      credentials: "include",
    });
    if (!response.ok) {
      let message = `Export failed (${response.status})`;
      try {
        const data = await response.json();
        if (data?.error?.message) message = data.error.message;
      } catch {
        /* keep default message */
      }
      throw new Error(message);
    }
    const blob = await response.blob();
    const disposition = response.headers.get("Content-Disposition") || "";
    const match = disposition.match(/filename="?([^";]+)"?/);
    const filename =
      match?.[1] || `${model}-export.${format === "csv" ? "csv" : "json"}`;
    return { blob, filename };
  }

  // Saved views
  async listSavedViews(model: string): Promise<{ views: SavedView[] }> {
    const response = await this.client.get(`/saved-views/${model}`);
    return response.data;
  }

  async saveSavedView(
    model: string,
    request: SavedViewRequest
  ): Promise<SavedView> {
    const response = await this.client.post(`/saved-views/${model}`, request);
    return response.data;
  }

  async deleteSavedView(model: string, id: string): Promise<void> {
    await this.client.delete(`/saved-views/${model}/${id}`);
  }

  // Helper to construct URLs
  getObjectURL(model: string, id: string | number): string {
    return `${this.baseURL}/${model}/${id}`;
  }
}

// Singleton instance
export const adminAPI = new AdminAPIClient();

// Export for testing or custom instances
export default adminAPI;
