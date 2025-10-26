import { formatJSON } from "../../utils/formatters";

interface JsonViewerProps {
  data: unknown;
  collapsed?: boolean;
}

export function JsonViewer({ data, collapsed = false }: JsonViewerProps) {
  const formatted = formatJSON(data);

  return (
    <details open={!collapsed} className="group">
      <summary className="cursor-pointer text-sm font-medium text-gray-700 hover:text-gray-900">
        View JSON
      </summary>
      <pre className="mt-2 bg-gray-50 p-3 rounded border border-gray-200 overflow-x-auto text-xs">
        {formatted}
      </pre>
    </details>
  );
}
