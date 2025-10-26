import { useState } from "react";
import type { AgentEvent, MessageData } from "../../types/events";
import { CodeBlock } from "../common/CodeBlock";
import { JsonViewer } from "../common/JsonViewer";
import { formatTimestamp } from "../../utils/formatters";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { User, Bot, Send } from "lucide-react";
import ReactMarkdown from "react-markdown";
import { apiClient } from "../../services/api";
import { useSWRConfig } from "swr";


interface ConversationProps {
  events: AgentEvent[];
  runId: string;
  runStatus: string;
  markMessageAsSent: (messageContent: string) => void;
}

export function Conversation({ events, runId, runStatus, markMessageAsSent }: ConversationProps) {
  const messageEvents = events.filter((e) => e.type === "message");
  const [message, setMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { mutate } = useSWRConfig();

  const handleSendMessage = async () => {
    if (!message.trim() || isSubmitting) return;
    
    const messageContent = message.trim();
    setIsSubmitting(true);
    
    try {
      // Resume the conversation - this will update the existing run and continue it
      // Mark message as sent to prevent WebSocket echo
      markMessageAsSent(messageContent);

      // Resume the run with the new message
      await apiClient.resumeRun(runId, messageContent);

      setMessage("");

      // Refresh the current run data and run list
      mutate("/api/runs");
      mutate(`/api/runs/${runId}`);
    } catch (error) {
      console.error("Failed to send message:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div className="flex flex-col h-full min-h-0">
      <div className="flex-1 space-y-3 overflow-y-auto min-h-0 pb-3">
        {messageEvents.map((event) => {
        const data = event.data as MessageData;
        const isAssistant = data.role === "assistant";

        return (
          <div
            key={event.id}
            className={
              isAssistant
                ? "border-l-2 border-primary/50 pl-3 py-2"
                : "border-l-2 border-border/50 pl-3 py-2"
            }
          >
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-1.5">
                {isAssistant ? (
                  <Bot className="h-3 w-3 text-primary" />
                ) : (
                  <User className="h-3 w-3 text-muted-foreground" />
                )}
                <span className="text-xs font-medium">
                  {isAssistant ? "Assistant" : "User"}
                </span>
              </div>
              <span className="text-[10px] text-muted-foreground">
                {formatTimestamp(event.timestamp)}
              </span>
            </div>

            <div className="space-y-2">
              {data.content.map((block, index) => (
                <div key={index}>
                  {block.type === "text" && block.text && (
                    <div className="text-sm text-foreground/90 prose prose-sm prose-neutral dark:prose-invert max-w-none">
                      <ReactMarkdown
                        components={{
                          code({ className, children, ...props }: any) {
                            const isInline = !className?.includes("language-");
                            const match = /language-(\w+)/.exec(className || "");
                            const language = match ? match[1] : "text";
                            const codeString = String(children).replace(/\n$/, "");
                            
                            return isInline ? (
                              <code
                                className="bg-muted px-1 py-0.5 rounded text-xs font-mono"
                                {...props}
                              >
                                {children}
                              </code>
                            ) : (
                              <CodeBlock code={codeString} language={language} />
                            );
                          },
                          p({ children }) {
                            return <p className="mb-2 last:mb-0">{children}</p>;
                          },
                          ul({ children }) {
                            return <ul className="list-disc list-inside mb-2">{children}</ul>;
                          },
                          ol({ children }) {
                            return <ol className="list-decimal list-inside mb-2">{children}</ol>;
                          },
                          li({ children }) {
                            return <li className="mb-1">{children}</li>;
                          },
                          h1({ children }) {
                            return <h1 className="text-lg font-bold mb-2">{children}</h1>;
                          },
                          h2({ children }) {
                            return <h2 className="text-base font-semibold mb-2">{children}</h2>;
                          },
                          h3({ children }) {
                            return <h3 className="text-sm font-semibold mb-1">{children}</h3>;
                          },
                          blockquote({ children }) {
                            return (
                              <blockquote className="border-l-4 border-primary/20 pl-4 italic text-muted-foreground">
                                {children}
                              </blockquote>
                            );
                          },
                          a({ children, href }) {
                            return (
                              <a
                                href={href}
                                className="text-primary hover:underline"
                                target="_blank"
                                rel="noopener noreferrer"
                              >
                                {children}
                              </a>
                            );
                          },
                        }}
                      >
                        {block.text}
                      </ReactMarkdown>
                    </div>
                  )}

                  {block.type === "tool_use" && block.tool_use && (
                    <div className="border border-yellow-200 bg-yellow-50/50 rounded px-2 py-1.5">
                      <div className="text-xs font-medium text-yellow-900 mb-1">
                        🔧 {block.tool_use.name}
                      </div>
                      <JsonViewer data={block.tool_use.input} collapsed />
                    </div>
                  )}

                  {block.type === "tool_result" && block.tool_result && (
                    <div
                      className={
                        block.tool_result.is_error
                          ? "border border-destructive/20 bg-destructive/5 rounded px-2 py-1.5"
                          : "border border-green-200 bg-green-50/50 rounded px-2 py-1.5"
                      }
                    >
                      <div
                        className={`text-xs font-medium mb-1 ${
                          block.tool_result.is_error
                            ? "text-destructive"
                            : "text-green-900"
                        }`}
                      >
                        {block.tool_result.is_error ? "❌ Error" : "✅ Result"}
                      </div>
                      <CodeBlock code={block.tool_result.content} />
                    </div>
                  )}
                </div>
              ))}
            </div>

            {data.usage && (
              <div className="flex items-center gap-3 mt-2 text-[10px] text-muted-foreground">
                <span>↓ {data.usage.input_tokens}</span>
                <span>↑ {data.usage.output_tokens}</span>
              </div>
            )}
          </div>
        );
      })}

        {messageEvents.length === 0 && (
          <div className="text-center text-muted-foreground py-8 text-sm">
            No messages yet
          </div>
        )}
      </div>

      {/* Chat Input - Always visible at bottom */}
      <div className="flex-shrink-0 pt-3 border-t border-border/50">
        <div className="flex gap-2 items-end">
          <div className="flex-1">
            <Textarea
              placeholder="Type a message to continue the conversation..."
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={3}
              disabled={isSubmitting}
              className="resize-none"
            />
          </div>
          <Button
            onClick={handleSendMessage}
            disabled={!message.trim() || isSubmitting}
            size="icon"
            className="h-[72px] flex-shrink-0"
          >
            <Send className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
