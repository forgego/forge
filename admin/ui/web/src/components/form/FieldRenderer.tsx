import { Input } from "../ui/input";
import { Checkbox } from "../ui/checkbox";
import { SearchableSelect } from "../ui/searchable-select";
import { Switch } from "../ui/switch";
import { X } from "lucide-react";
import { useUIComponent } from "../../hooks/useUIComponent";
import { resolveFieldKind, isIntegerField } from "./resolve-field-kind";
import type { FieldRendererProps } from "./types";

// Override slot for a single field. Extracted as a component so the
// useUIComponent hook call has a stable call order (calling it inside
// renderField would violate the Rules of Hooks as field counts change).
export function FieldOverride({
  overrideKey,
  field,
  value,
  onChange,
  metadata,
}: {
  overrideKey: string;
  field: any;
  value: any;
  onChange: (val: any) => void;
  metadata: any;
}) {
  const CustomField = useUIComponent(overrideKey, null as any);
  if (!CustomField) return null;
  return (
    // eslint-disable-next-line react-hooks/static-components -- useUIComponent returns a stable registry ref
    <CustomField
      field={field}
      value={value}
      onChange={onChange}
      metadata={metadata}
    />
  );
}

export function FieldRenderer({
  field,
  value,
  onChange,
  relation,
  mode,
  overrideKey,
  metadata,
}: FieldRendererProps) {
  if (field.read_only && mode === "create") return null;

  const fieldOverride = overrideKey;
  const isReadOnly = field.read_only;
  const isSwitchWidget =
    field.widget === "switch" || field.widget === "toggle";

  if (fieldOverride) {
    return (
      <FieldOverride
        overrideKey={fieldOverride}
        field={field}
        value={value}
        onChange={(val: any) => onChange(val)}
        metadata={metadata}
      />
    );
  }

  const kind = resolveFieldKind(field, relation);

  switch (kind) {
    case "boolean":
      if (isSwitchWidget) {
        return (
          <div className="flex items-center gap-3 p-3 rounded-lg border border-border/50 bg-muted/20 hover:bg-muted/30 transition-colors">
            <Switch
              checked={!!value}
              onCheckedChange={(checked) => onChange(checked)}
              disabled={isReadOnly}
            />
            <label className="text-sm font-semibold leading-none cursor-pointer select-none">
              {field.label}
            </label>
          </div>
        );
      }
      return (
        <div className="flex items-center space-x-3 p-3 rounded-lg border border-border/50 bg-muted/20 hover:bg-muted/30 transition-colors">
          <Checkbox
            id={field.name}
            checked={!!value}
            onChange={(e) => onChange(e.target.checked)}
            disabled={isReadOnly}
          />
          <label
            htmlFor={field.name}
            className="text-sm font-semibold leading-none cursor-pointer select-none"
          >
            {field.label}
          </label>
        </div>
      );

    case "textarea":
      return (
        <textarea
          id={field.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="flex min-h-[120px] w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all"
          required={field.required}
          disabled={isReadOnly}
        />
      );

    case "text":
    case "email":
    case "url":
    case "tel":
    case "color":
      return (
        <Input
          type={kind}
          id={field.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
          required={field.required}
          maxLength={field.max_length}
          minLength={field.min_length}
          disabled={isReadOnly}
        />
      );

    case "number":
      return (
        <Input
          type="number"
          id={field.name}
          value={value}
          onChange={(e) =>
            onChange(
              e.target.value === "" ? "" : Number(e.target.value)
            )
          }
          className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
          required={field.required}
          min={field.min_value}
          max={field.max_value}
          step={isIntegerField(field) ? 1 : "any"}
          disabled={isReadOnly}
        />
      );

    case "date":
      return (
        <Input
          type="date"
          id={field.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
          required={field.required}
          disabled={isReadOnly}
        />
      );

    case "datetime":
      return (
        <Input
          type="datetime-local"
          id={field.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
          required={field.required}
          disabled={isReadOnly}
        />
      );

    case "time":
      return (
        <Input
          type="time"
          id={field.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
          required={field.required}
          disabled={isReadOnly}
        />
      );

    case "choice":
      if (!field.choices?.length) {
        return (
          <Input
            id={field.name}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
            disabled={isReadOnly}
          />
        );
      }
      return (
        <select
          id={field.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="flex h-10 w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all appearance-none"
          required={field.required}
          disabled={isReadOnly}
        >
          <option value="">Select {field.label}</option>
          {field.choices?.map((choice: any) => (
            <option key={choice.value} value={choice.value}>
              {choice.label}
            </option>
          ))}
        </select>
      );

    case "password":
      return (
        <Input
          type="password"
          id={field.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
          required={field.required}
          autoComplete="new-password"
          disabled={isReadOnly}
        />
      );

    case "json":
      return (
        <textarea
          id={field.name}
          value={
            typeof value === "object" ? JSON.stringify(value, null, 2) : value
          }
          onChange={(e) => {
            const val = e.target.value;
            onChange(val);
            // Basic validation hint could go here
          }}
          className="flex min-h-[150px] w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-xs font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all shadow-inner"
          placeholder="{}"
          required={field.required}
          disabled={isReadOnly}
        />
      );

    case "fk":
      return (
        <SearchableSelect
          model={relation?.related_model || field.related_model}
          value={value}
          onChange={(val) => onChange(val)}
          placeholder={`Select ${field.label}...`}
          required={field.required}
          disabled={isReadOnly}
        />
      );

    case "m2m":
      const m2mValue = Array.isArray(value) ? value : [];
      return (
        <div className="space-y-4">
          <div className="flex flex-wrap gap-2 mb-2">
            {m2mValue.map((item: any, idx: number) => (
              <div
                key={item.id || idx}
                className="flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary border border-primary/20 text-xs font-medium animate-in fade-in zoom-in-95"
              >
                <span>
                  {item.name ||
                    item.title ||
                    item.label ||
                    `ID: ${item.id || item}`}
                </span>
                <button
                  type="button"
                  aria-label={`Remove ${field.label} item ${idx + 1}`}
                  onClick={() => {
                    const newValue = m2mValue.filter(
                      (_: any, i: number) => i !== idx
                    );
                    onChange(newValue);
                  }}
                  className="hover:text-destructive transition-colors"
                  disabled={isReadOnly}
                >
                  <X className="h-3 w-3" />
                </button>
              </div>
            ))}
          </div>
          <SearchableSelect
            model={field.related_model}
            value={null}
            onChange={(val) => {
              if (!val) return;
              // Avoid duplicates
              if (
                m2mValue.some(
                  (v: any) => (typeof v === "object" ? v.id : v) === val
                )
              )
                return;
              onChange([...m2mValue, val]);
            }}
            placeholder={`Add ${field.label}...`}
            disabled={isReadOnly}
          />
        </div>
      );

    default:
      return (
        <Input
          id={field.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="rounded-lg border-border/50 bg-background/50 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
          required={field.required}
          disabled={isReadOnly}
        />
      );
  }
}
