// Auto-generated types from backend admin API

export interface FieldMetadata {
  name: string;
  type: string;
  label: string;
  help_text?: string;
  required: boolean;
  read_only: boolean;
  choices?: Choice[];
  widget: string;
  validators?: ValidatorMetadata[];
  default_value?: any;
  max_length?: number;
  min_length?: number;
  max_value?: any;
  min_value?: any;
  accept?: string;
  max_size?: number;
  multiple?: boolean;
  allow_create?: boolean;
}

export interface Choice {
  value: any;
  label: string;
}

export interface ValidatorMetadata {
  name: string;
  message?: string;
  params?: Record<string, any>;
}

export interface RelationMetadata {
  name: string;
  type: string; // 'foreign_key' | 'one_to_one' | 'many_to_many'
  related_model: string;
  related_field?: string;
  label: string;
}

export interface PermissionMetadata {
  add: boolean;
  change: boolean;
  delete: boolean;
  view: boolean;
}

export interface ActionMetadata {
  name: string;
  label: string;
  description?: string;
  confirmation?: string;
  permissions?: string[];
  dangerous?: boolean;
  icon?: string;
  ui_component?: string;
}

export interface FilterMetadata {
  name: string;
  type: string; // 'text' | 'number' | 'boolean' | 'choice' | 'date_range' | 'relation'
  label: string;
  choices?: Choice[];
  related_model?: string;
  multiple?: boolean;
  params?: Record<string, any>;
  ui_component?: string;
}

export interface PaginationConfig {
  page_size: number;
  max_page_size: number;
}

export interface Metadata {
  name: string;
  verbose_name: string;
  verbose_name_plural: string;
  description?: string;
  icon?: string;
  fields: FieldMetadata[];
  relations?: RelationMetadata[];
  permissions: PermissionMetadata;
  actions: ActionMetadata[];
  filters: FilterMetadata[];
  list_display: string[];
  list_filter?: string[];
  search_fields?: string[];
  ordering?: string[];
  pagination: PaginationConfig;
  page_type?: string;
  ui_overrides?: Record<string, string>;
}

export interface ModelListMetadata {
  name: string;
  verbose_name: string;
  verbose_name_plural: string;
  icon?: string;
  count: number;
  permissions: PermissionMetadata;
}

export interface PaginatedResponse<T = any> {
  count: number;
  next?: string;
  previous?: string;
  page_size: number;
  page: number;
  total_pages: number;
  results: T[];
}

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
    details?: Record<string, any>;
  };
}

export interface BulkActionRequest {
  ids: number[];
  params?: Record<string, any>;
}

export interface BulkActionResponse {
  success: boolean;
  affected: number;
  message: string;
  errors?: BulkActionError[];
}

export interface BulkActionError {
  id: number;
  message: string;
}

export interface SearchRequest {
  query: string;
  models?: string[];
}

export interface SearchResponse {
  results: SearchResultGroup[];
}

export interface SearchResultGroup {
  model: string;
  count: number;
  items: SearchResultItem[];
}

export interface SearchResultItem {
  id: number;
  title: string;
  highlight?: string;
  url: string;
}

export interface AutocompleteResponse {
  results: AutocompleteItem[];
}

export interface AutocompleteItem {
  value: any;
  label: string;
}

export interface UploadResponse {
  url: string;
  filename: string;
  size: number;
  mime_type: string;
  width?: number;
  height?: number;
  uploaded_at: string;
}

export interface HistoryEntry {
  id: number;
  timestamp: string;
  user_id?: string;
  user_name?: string;
  model_name?: string;
  object_id?: string;
  object_repr?: string;
  action?: string;
  change_stats?: string;
}

export interface HistoryResponse {
  entries: HistoryEntry[];
}

// Frontend-specific types

export interface ListParams {
  page?: number;
  page_size?: number;
  search?: string;
  ordering?: string;
  [key: string]: any; // filters
}

export interface ModelFormData {
  [key: string]: any;
}

export interface PluginMetadata {
  name: string;
  label: string;
  icon?: string;
  pages?: CustomPageMetadata[];
  widgets?: WidgetMetadata[];
  menuEntries?: MenuEntryMetadata[];
  ui_overrides?: Record<string, string>;
}

export interface CustomPageMetadata {
  name: string;
  label: string;
  path: string;
  component: string;
}

export interface WidgetMetadata {
  type: string;
  title: string;
  description?: string;
  defaultSize?: string;
  config?: Record<string, any>;
}

export interface MenuEntryMetadata {
  label: string;
  icon?: string;
  path?: string;
  children?: MenuEntryMetadata[];
  order?: number;
}

export interface MetadataResponse {
  models: ModelListMetadata[];
  plugins: PluginMetadata[];
}
