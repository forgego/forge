export type FieldKind =
  | "boolean"
  | "number"
  | "date"
  | "datetime"
  | "time"
  | "text"
  | "textarea"
  | "choice"
  | "password"
  | "email"
  | "url"
  | "tel"
  | "color"
  | "json"
  | "fk"
  | "m2m";

export interface FieldRendererProps {
  field: any;
  value: any;
  onChange: (val: any) => void;
  relation?: any;
  mode: "create" | "edit";
  overrideKey?: string;
  metadata?: any;
}

export interface InlineFieldRendererProps {
  field: any;
  value: any;
  onChange: (val: any) => void;
  relation?: any;
}

export interface InlineRelationDetail {
  relation: any;
  inlineConfig: any;
  relatedModel: string;
  relatedField: string | undefined;
}
