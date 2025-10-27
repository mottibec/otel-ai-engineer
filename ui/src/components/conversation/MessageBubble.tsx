import { useMemo } from "react";
import type { MessageData } from "../../types/events";
import { formatTimestamp } from "../../utils/formatters";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { User, Bot } from "lucide-react";
import ReactMarkdown from "react-markdown";
import { CodeBlock } from "../common/CodeBlock";
import { ToolCallInline } from "./ToolCallInline";
import { ToolResultInline } from "./ToolResultInline";

interface MessageBubbleProps {
  data: MessageData;
  timestamp: string;
}

export function MessageBubble({ data, timestamp }: MessageBubbleProps) {
  const isAssistant = data.role === "assistant";

  const markdownComponents = useMemo(
    () => ({
      code({ className, children, ...props }: any) {
        const isInline = !className?.includes("language-");
        const match = /language-(\w+)/.exec(className || "");
        const language = match ? match[1] : "text";
        const codeString = String(children).replace(/\n$/, "");

        return isInline ? (
          <code
            className="bg-muted px-0.5 py-0.5 rounded text-xs font-mono"
            {...props}
          >
            {children}
          </code>
        ) : (
          <CodeBlock code={codeString} language={language} />
        );
      },
      p({ children }: any) {
        return (
          <p className="mb-1 last:mb-0 text-sm leading-snug">{children}</p>
        );
      },
      ul({ children }: any) {
        return (
          <ul className="list-disc list-inside mb-1 text-sm">{children}</ul>
        );
      },
      ol({ children }: any) {
        return (
          <ol className="list-decimal list-inside mb-1 text-sm">{children}</ol>
        );
      },
      li({ children }: any) {
        return <li className="text-sm">{children}</li>;
      },
      h1({ children }: any) {
        return <h1 className="text-sm font-semibold mb-1">{children}</h1>;
      },
      h2({ children }: any) {
        return <h2 className="text-sm font-semibold mb-1">{children}</h2>;
      },
      h3({ children }: any) {
        return <h3 className="text-xs font-semibold mb-0.5">{children}</h3>;
      },
      blockquote({ children }: any) {
        return (
          <blockquote className="border-l-2 border-primary/20 pl-2 py-0.5 italic text-muted-foreground text-sm my-1">
            {children}
          </blockquote>
        );
      },
      a({ children, href }: any) {
        return (
          <a
            href={href}
            className="text-primary hover:underline text-sm"
            target="_blank"
            rel="noopener noreferrer"
          >
            {children}
          </a>
        );
      },
    }),
    [],
  );

  return (
    <div className="flex gap-2 group">
      {/* Avatar */}
      <div className="flex-shrink-0">
        <div
          className={cn(
            "w-6 h-6 rounded-full flex items-center justify-center",
            isAssistant
              ? "bg-primary/10 text-primary"
              : "bg-muted text-muted-foreground",
          )}
        >
          {isAssistant ? (
            <Bot className="h-3 w-3" />
          ) : (
            <User className="h-3 w-3" />
          )}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0 space-y-1">
        {/* Header */}
        <div className="flex items-center gap-2">
          <span className="text-[10px] font-medium">
            {isAssistant ? "Assistant" : "User"}
          </span>
          <span className="text-[9px] text-muted-foreground font-mono">
            {formatTimestamp(timestamp)}
          </span>
          {data.usage && (
            <Badge
              variant="outline"
              className="text-[8px] h-3 px-1 font-mono ml-auto"
            >
              {data.usage.input_tokens + data.usage.output_tokens}t
            </Badge>
          )}
        </div>

        {/* Messages */}
        <div className="space-y-1.5">
          {data.content.map((block, index) => (
            <div key={index}>
              {block.type === "text" && block.text && (
                <div className="text-sm leading-snug prose prose-sm max-w-none dark:prose-invert prose-headings:my-1 prose-p:my-1">
                  <ReactMarkdown components={markdownComponents}>
                    {block.text}
                  </ReactMarkdown>
                </div>
              )}

              {block.type === "tool_use" && block.tool_use && (
                <ToolCallInline toolUse={block.tool_use} />
              )}

              {block.type === "tool_result" && block.tool_result && (
                <ToolResultInline toolResult={block.tool_result} />
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
