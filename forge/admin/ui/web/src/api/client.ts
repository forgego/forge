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
  SearchRequest,
  SearchResponse,
  AutocompleteResponse,
  UploadResponse,
  ErrorResponse,
  MetadataResponse,
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
          // Handle unauthorized
          localStorage.removeItem("admin_token");
          window.location.href = "/login";
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
  async get(url: string): Promise<any> {
    return this.client.get(url);
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
  // Bulk operations
  async bulkCreate<T = any>(
    model: string,
    data: ModelFormData[]
  ): Promise<{ created: number; objects: T[] }> {
    const response = await this.client.post(`/${model}/bulk-create`, data);
    return response.data;
  }

  async bulkUpdate(
    model: string,
    ids: (string | number)[],
    data: Partial<ModelFormData>
  ): Promise<{ updated: number }> {
    const response = await this.client.post(`/${model}/bulk-update`, {
      ids,
      data,
    });
    return response.data;
  }

  async bulkDelete(model: string, ids: (string | number)[]): Promise<void> {
    await this.client.delete(`/${model}/bulk-delete`, { data: { ids } });
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

  // File upload
  async uploadFile(model: string, file: File): Promise<UploadResponse> {
    const formData = new FormData();
    formData.append("file", file);

    const response = await this.client.post(`/${model}/upload`, formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
    return response.data;
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
        searchParams.set(key, String(value));
      });
    }

    const query = searchParams.toString();
    return `${this.baseURL}/${model}/export${query ? `?${query}` : ""}`;
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
