import { useState, useCallback } from "react";
import type { AgentEvent } from "../../types/events";
import { MessageBubble } from "../conversation/MessageBubble";
import { ChatInput } from "../conversation/ChatInput";
import { apiClient } from "../../services/api";
import { useSWRConfig } from "swr";
import { ErrorBubble } from "./ErrorBubble";

interface ConversationProps {
  events: AgentEvent[];
  runId: string;
}

export function Conversation({ events, runId }: ConversationProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { mutate } = useSWRConfig();

  const messageEvents = events.filter((e) => e.type === "message");
  const errorEvents = events.filter((e) => e.type === "error");

  const handleSendMessage = useCallback(async (messageContent: string) => {
    setIsSubmitting(true);

    try {
      await apiClient.resumeRun(runId, messageContent);
      mutate("/api/runs");
      mutate(`/api/runs/${runId}`);
    } catch (error) {
      console.error("Failed to send message:", error);
    } finally {
      setIsSubmitting(false);
    }
  }, [runId, mutate]);

  // Combine and sort all events by timestamp
  const allEvents = [...messageEvents, ...errorEvents].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  return (
    <div className="flex flex-col h-full min-h-0">
      {/* Messages */}
      <div className="flex-1 space-y-2 overflow-y-auto min-h-0 p-3">
        {allEvents.map((event) => {
          if (event.type === "error") {
            return (
              <ErrorBubble
                key={event.id}
                event={event}
              />
            );
          }
          return (
            <MessageBubble
              key={event.id}
              data={event.data as any}
              timestamp={event.timestamp}
            />
          );
        })}

        {/* Show "No messages yet" when no messages */}
        {allEvents.length === 0 && (
          <div className="flex items-center justify-center h-full text-xs text-muted-foreground">
            No messages yet
          </div>
        )}
      </div>

      {/* Input */}
      <ChatInput onSend={handleSendMessage} isSubmitting={isSubmitting} />
    </div>
  );
}
