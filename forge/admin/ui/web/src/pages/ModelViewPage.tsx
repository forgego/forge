import { useNavigate, useParams } from "@tanstack/react-router";
import { Loader2, ArrowLeft, Edit } from "lucide-react";
import AdminLayout from "../components/layout/AdminLayout";
import { Card, CardContent } from "../components/ui/card";
import { Button } from "../components/ui/button";
import { useModelDetail, useModelMetadata } from "../api/hooks/adminHooks";

function formatValue(value: any): string {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  if (typeof value === "boolean") {
    return value ? "Yes" : "No";
  }
  if (Array.isArray(value)) {
    if (value.length === 0) {
      return "-";
    }
    return value
      .map((item) => {
        if (item && typeof item === "object") {
          return item.name || item.title || item.label || item.id || "[object]";
        }
        return String(item);
      })
      .join(", ");
  }
  if (typeof value === "object") {
    try {
      return JSON.stringify(value, null, 2);
    } catch {
      return String(value);
    }
  }
  return String(value);
}

export default function ModelViewPage() {
  const params = useParams({ strict: false }) as any;
  const { model, id } = params;
  const navigate = useNavigate();

  const modelName = model as string;
  const objectID = id as string;

  const { data: metadata, isLoading: metaLoading } = useModelMetadata(modelName);
  const {
    data: objectData,
    isLoading: objectLoading,
    error,
  } = useModelDetail(modelName, objectID, {
    enabled: !!modelName && !!objectID,
  } as any);

  if (metaLoading || objectLoading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      </AdminLayout>
    );
  }

  if (!metadata || error || !objectData) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-destructive">
              Unable to load record
            </h2>
            <p className="text-muted-foreground mt-2">
              Failed to load {metadata?.verbose_name || modelName} #{objectID}
            </p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div className="space-y-1">
            <h1 className="text-3xl font-semibold tracking-tight text-foreground">
              {metadata.verbose_name} details
            </h1>
            <p className="text-sm text-muted-foreground">
              Viewing {metadata.verbose_name} #{objectID}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              onClick={() => navigate({ to: "/$model", params: { model: modelName } })}
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to list
            </Button>
            {metadata.permissions.change && (
              <Button
                onClick={() =>
                  navigate({
                    to: "/$model/$id",
                    params: { model: modelName, id: objectID },
                  })
                }
              >
                <Edit className="h-4 w-4 mr-2" />
                Edit
              </Button>
            )}
          </div>
        </div>

        <Card className="glass-lite border-border/50 shadow-xl shadow-black/5 overflow-hidden max-w-4xl">
          <CardContent className="p-8">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
              {metadata.fields
                .filter((field) => field.name !== "id")
                .map((field) => {
                  const value = objectData[field.name];
                  const isJSON = typeof value === "object" && value !== null;
                  return (
                    <div key={field.name} className="space-y-2">
                      <p className="text-xs font-bold uppercase tracking-widest text-muted-foreground/80">
                        {field.label}
                      </p>
                      {isJSON ? (
                        <pre className="rounded-md border border-border/50 bg-muted/20 p-3 text-xs whitespace-pre-wrap break-words">
                          {formatValue(value)}
                        </pre>
                      ) : (
                        <p className="text-sm text-foreground/90 break-words">
                          {formatValue(value)}
                        </p>
                      )}
                    </div>
                  );
                })}
            </div>
          </CardContent>
        </Card>
      </div>
    </AdminLayout>
  );
}
