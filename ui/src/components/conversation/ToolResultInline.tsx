import { Badge } from "@/components/ui/badge";
import { CheckCircle, XCircle } from "lucide-react";
import { CodeBlock } from "../common/CodeBlock";

interface ToolResultInlineProps {
  toolResult: {
    content: string;
    is_error: boolean;
  };
}

export function ToolResultInline({ toolResult }: ToolResultInlineProps) {
  return (
    <div className={
      toolResult.is_error
        ? "border-l-2 border-destructive/20 bg-destructive/5 pl-2 py-1 my-1"
        : "border-l-2 border-green-200/40 bg-green-50/10 dark:bg-green-950/5 pl-2 py-1 my-1"
    }>
      <div className="flex items-center gap-1 mb-0.5">
        {toolResult.is_error ? (
          <>
            <XCircle className="h-2.5 w-2.5 text-destructive/60 flex-shrink-0" />
            <Badge variant="destructive" className="text-[8px] h-3 px-1 leading-tight opacity-70">
              Error
            </Badge>
          </>
        ) : (
          <>
            <CheckCircle className="h-2.5 w-2.5 text-green-600/60 dark:text-green-400/50 flex-shrink-0" />
            <Badge variant="secondary" className="text-[8px] h-3 px-1 leading-tight text-green-700/70 dark:text-green-400/60 border-green-200/40">
              Success
            </Badge>
          </>
        )}
      </div>
      <div className="ml-3.5">
        <CodeBlock code={toolResult.content} />
      </div>
    </div>
  );
}
