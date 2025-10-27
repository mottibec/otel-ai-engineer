import { useMemo } from "react";
import type { AgentEvent, ErrorData } from "../../types/events";
import { AlertCircle, AlertTriangle, Info } from "lucide-react";
import { formatTimestamp } from "../../utils/formatters";
import { Badge } from "@/components/ui/badge";

interface ErrorBubbleProps {
  event: AgentEvent;
}

interface ParsedError {
  type: "rate_limit" | "api_error" | "general";
  title: string;
  message: string;
  details?: string;
  suggestion?: string;
}

export function ErrorBubble({ event }: ErrorBubbleProps) {
  const data = event.data as ErrorData;
  
  const parsedError = useMemo<ParsedError>(() => {
    const message = data.message || "";
    
    // Check for rate limit errors
    if (message.includes("rate limit") || message.includes("rate_limit")) {
      // Extract relevant information from the error
      let title = "Rate Limit Exceeded";
      let suggestion = "Please wait a moment and try again. Consider reducing the prompt length or max tokens.";
      
      // Try to extract limit information if present
      if (message.includes("30,000 input tokens per minute")) {
        title = "Rate Limit: 30,000 input tokens/minute";
        suggestion = "You've exceeded the rate limit. Please wait before trying again, or reduce your prompt size.";
      } else if (message.includes("per minute")) {
        // Extract the limit if present
        const limitMatch = message.match(/(\d+,\d+|\d+)\s+(?:input\s+)?tokens per minute/);
        if (limitMatch) {
          title = `Rate Limit: ${limitMatch[1]} tokens/minute`;
        }
      }
      
      return {
        type: "rate_limit",
        title,
        message: "This request would exceed your organization's rate limit.",
        suggestion,
      };
    }
    
    // Check for API errors
    if (message.includes("API call failed")) {
      const apiMatch = message.match(/API call failed: (.+)/);
      const apiError = apiMatch ? apiMatch[1] : message;
      
      return {
        type: "api_error",
        title: "API Error",
        message: apiError,
        suggestion: "Please check your API key and try again.",
      };
    }
    
    // Generic error
    return {
      type: "general",
      title: "Error",
      message: data.message || "An error occurred",
    };
  }, [data.message]);
  
  const getIcon = () => {
    switch (parsedError.type) {
      case "rate_limit":
        return <AlertTriangle className="h-4 w-4 text-yellow-600 dark:text-yellow-500" />;
      case "api_error":
        return <AlertCircle className="h-4 w-4 text-red-600 dark:text-red-500" />;
      default:
        return <Info className="h-4 w-4 text-blue-600 dark:text-blue-500" />;
    }
  };
  
  const getBadgeVariant = () => {
    switch (parsedError.type) {
      case "rate_limit":
        return "secondary" as const;
      case "api_error":
        return "destructive" as const;
      default:
        return "secondary" as const;
    }
  };
  
  return (
    <div className="flex gap-3 max-w-full animate-in slide-in-from-bottom-1 duration-300">
      <div className="flex flex-col items-end flex-shrink-0 gap-1 w-12">
        <div className="flex items-center justify-center w-6 h-6 rounded-full bg-destructive/10 dark:bg-destructive/20">
          {getIcon()}
        </div>
        <span className="text-[10px] text-muted-foreground leading-tight">
          {formatTimestamp(event.timestamp)}
        </span>
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="bg-destructive/5 border border-destructive/20 dark:bg-destructive/10 dark:border-destructive/30 rounded-lg p-3 space-y-2">
          {/* Header */}
          <div className="flex items-center gap-2 flex-wrap">
            <Badge variant={getBadgeVariant()} className="text-xs h-5 px-2">
              {parsedError.title}
            </Badge>
          </div>
          
          {/* Message */}
          <div className="text-sm text-destructive dark:text-destructive/90">
            {parsedError.message}
          </div>
          
          {/* Suggestion */}
          {parsedError.suggestion && (
            <div className="text-xs text-muted-foreground bg-background/50 rounded px-2 py-1.5 border border-border/50">
              <div className="font-medium mb-0.5">💡 Suggestion:</div>
              {parsedError.suggestion}
            </div>
          )}
          
          {/* Full error details (collapsible) */}
          {data.message && (
            <details className="text-xs">
              <summary className="cursor-pointer text-muted-foreground hover:text-foreground transition-colors">
                Show technical details
              </summary>
              <pre className="mt-2 p-2 bg-muted rounded text-[10px] overflow-x-auto">
                {JSON.stringify(data, null, 2)}
              </pre>
            </details>
          )}
        </div>
      </div>
    </div>
  );
}
