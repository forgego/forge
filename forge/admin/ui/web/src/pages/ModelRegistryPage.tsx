import { useModels } from "../api/hooks/adminHooks";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { Database, Loader2 } from "lucide-react";
import { Link } from "@tanstack/react-router";
import AdminLayout from "../components/layout/AdminLayout";

export default function ModelsListPage() {
  const { data, isLoading, error } = useModels();

  if (isLoading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      </AdminLayout>
    );
  }

  if (error) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-destructive">Error</h2>
            <p className="text-muted-foreground mt-2">{error.message}</p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  const models = data?.models || [];

  return (
    <AdminLayout>
      <div className="container mx-auto p-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold">Admin Dashboard</h1>
          <p className="text-muted-foreground mt-2">
            Manage your application models
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {models.map((model) => (
            <Link
              key={model.name}
              to="/$model"
              params={{ model: model.name }}
              className="transition-transform hover:scale-105"
            >
              <Card className="cursor-pointer hover:border-primary">
                <CardHeader>
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10 rounded-lg">
                      <Database className="h-6 w-6 text-primary" />
                    </div>
                    <div>
                      <CardTitle>{model.verbose_name_plural}</CardTitle>
                      <CardDescription>{model.name}</CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">
                      {model.count} records
                    </span>
                    <div className="flex gap-1">
                      {model.permissions.view && (
                        <span className="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded">
                          View
                        </span>
                      )}
                      {model.permissions.add && (
                        <span className="text-xs bg-green-100 text-green-700 px-2 py-1 rounded">
                          Add
                        </span>
                      )}
                      {model.permissions.change && (
                        <span className="text-xs bg-yellow-100 text-yellow-700 px-2 py-1 rounded">
                          Edit
                        </span>
                      )}
                      {model.permissions.delete && (
                        <span className="text-xs bg-red-100 text-red-700 px-2 py-1 rounded">
                          Delete
                        </span>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      </div>
    </AdminLayout>
  );
}
