import { useState } from 'react';
import { Download, FileJson, FileSpreadsheet, Loader2 } from 'lucide-react';
import { Button } from '../components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../components/ui/dropdown-menu';
import { adminAPI } from '../api/client';

interface ExportOptions {
  model: string;
  fields?: string[];
  format?: 'json' | 'csv' | 'xlsx';
  filename?: string;
}

interface UseExportOptions {
  onSuccess?: (data: any) => void;
  onError?: (error: Error) => void;
}

export function useExport(options: UseExportOptions = {}) {
  const [exporting, setExporting] = useState(false);
  const [progress, setProgress] = useState(0);

  const exportData = async (exportOptions: ExportOptions) => {
    const { model, fields, format = 'json', filename } = exportOptions;

    setExporting(true);
    setProgress(0);

    try {
      // For large datasets, we might want to use the API directly
      const response = await adminAPI.listObjects(model, {
        page: 1,
        page_size: 10000, // Max export size
      });

      setProgress(50);

      const data = response.results;
      let content: string;
      let mimeType: string;
      let extension: string;

      if (format === 'csv') {
        content = convertToCSV(data, fields);
        mimeType = 'text/csv';
        extension = 'csv';
      } else {
        content = JSON.stringify(data, null, 2);
        mimeType = 'application/json';
        extension = 'json';
      }

      setProgress(75);

      downloadFile(content, filename || `${model}_export.${extension}`, mimeType);

      setProgress(100);
      options.onSuccess?.(data);
    } catch (error) {
      console.error('Export failed:', error);
      options.onError?.(error as Error);
    } finally {
      setExporting(false);
      setProgress(0);
    }
  };

  const exportViaUrl = (exportOptions: ExportOptions) => {
    const { model, format = 'json', filename } = exportOptions;
    const url = adminAPI.getExportURL(model, format as 'json' | 'csv');
    
    // Create a link to download
    const link = document.createElement('a');
    link.href = url;
    link.download = filename || `${model}_export.${format}`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return {
    exporting,
    progress,
    exportData,
    exportViaUrl,
  };
}

function convertToCSV(data: any[], fields?: string[]): string {
  if (!data || data.length === 0) return '';

  const headers = fields || Object.keys(data[0]);
  const csvRows: string[] = [];

  // Add header row
  csvRows.push(headers.join(','));

  // Add data rows
  for (const row of data) {
    const values = headers.map((header) => {
      const value = row[header];
      if (value === null || value === undefined) return '';
      if (typeof value === 'string') {
        // Escape quotes and wrap in quotes if contains comma or quote
        const escaped = value.replace(/"/g, '""');
        if (escaped.includes(',') || escaped.includes('"') || escaped.includes('\n')) {
          return `"${escaped}"`;
        }
        return escaped;
      }
      return String(value);
    });
    csvRows.push(values.join(','));
  }

  return csvRows.join('\n');
}

function downloadFile(content: string, filename: string, mimeType: string) {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

interface ExportButtonProps {
  model: string;
  metadata?: {
    fields: { name: string; label?: string }[];
  };
  onExport?: (format: string) => void;
  variant?: 'default' | 'outline' | 'ghost';
  size?: 'default' | 'sm' | 'icon';
}

export function ExportButton({
  model,
  metadata,
  onExport,
  variant = 'outline',
  size = 'sm',
}: ExportButtonProps) {
  const { exporting, exportData } = useExport();

  const handleExport = (format: string) => {
    const fields = metadata?.fields.map((f) => f.name);
    exportData({
      model,
      fields,
      format: format as 'json' | 'csv',
    });
    onExport?.(format);
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant={variant} size={size} disabled={exporting}>
          {exporting ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Download className="h-4 w-4" />
          )}
          <span className="ml-2">Export</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>Export Format</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => handleExport('json')}
          className="cursor-pointer"
        >
          <FileJson className="h-4 w-4 mr-2" />
          JSON
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => handleExport('csv')}
          className="cursor-pointer"
        >
          <FileSpreadsheet className="h-4 w-4 mr-2" />
          CSV
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
