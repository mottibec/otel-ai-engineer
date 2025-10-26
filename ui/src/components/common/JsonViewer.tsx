import { formatJSON } from "../../utils/formatters";
import { useState } from "react";
import { ChevronDown, ChevronUp, Copy, Check } from "lucide-react";

interface JsonViewerProps {
  data: unknown;
  collapsed?: boolean;
  title?: string;
}

export function JsonViewer({ data, collapsed = false, title }: JsonViewerProps) {
  const [isOpen, setIsOpen] = useState(!collapsed);
  const [copied, setCopied] = useState(false);
  const formatted = formatJSON(data);
  const isLarge = formatted.length > 500; // Consider large if > 500 chars
  
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(formatted);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <div className="border border-gray-200 rounded">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-3 py-2 text-left flex items-center justify-between hover:bg-gray-50 transition-colors"
      >
        <div className="flex items-center gap-2">
          {isOpen ? (
            <ChevronUp className="h-3 w-3 text-gray-500" />
          ) : (
            <ChevronDown className="h-3 w-3 text-gray-500" />
          )}
          <span className="text-xs font-medium text-gray-700">
            {title || "View JSON"}
          </span>
          {isLarge && (
            <span className="text-[10px] text-gray-500">
              ({formatted.length.toLocaleString()} chars)
            </span>
          )}
        </div>
        {isOpen && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              handleCopy();
            }}
            className="flex items-center gap-1 text-[10px] text-gray-600 hover:text-gray-900 transition-colors"
            title="Copy JSON"
          >
            {copied ? (
              <>
                <Check className="h-3 w-3" />
                Copied!
              </>
            ) : (
              <>
                <Copy className="h-3 w-3" />
                Copy
              </>
            )}
          </button>
        )}
      </button>
      
      {isOpen && (
        <div className="border-t border-gray-200">
          <pre className="bg-gray-50 p-3 text-xs overflow-x-auto max-h-96 overflow-y-auto">
            {formatted}
          </pre>
        </div>
      )}
    </div>
  );
}
