import { useState, useMemo, useCallback } from "react";
import type { AgentEvent } from "../../types/events";
import { MessageBubble } from "../conversation/MessageBubble";
import { ChatInput } from "../conversation/ChatInput";
import { apiClient } from "../../services/api";
import { useSWRConfig } from "swr";

interface ConversationProps {
  events: AgentEvent[];
  runId: string;
  markMessageAsSent: (messageContent: string) => void;
}

export function Conversation({ events, runId, markMessageAsSent }: ConversationProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { mutate } = useSWRConfig();

  const messageEvents = useMemo(
    () => events.filter((e) => e.type === "message"),
    [events]
  );

  const handleSendMessage = useCallback(async (messageContent: string) => {
    setIsSubmitting(true);
    
    try {
      markMessageAsSent(messageContent);
      await apiClient.resumeRun(runId, messageContent);
      mutate("/api/runs");
      mutate(`/api/runs/${runId}`);
    } catch (error) {
      console.error("Failed to send message:", error);
    } finally {
      setIsSubmitting(false);
    }
  }, [runId, markMessageAsSent, mutate]);

  return (
    <div className="flex flex-col h-full min-h-0">
      {/* Messages */}
      <div className="flex-1 space-y-2 overflow-y-auto min-h-0 p-3">
        {messageEvents.map((event) => (
          <MessageBubble
            key={event.id}
            data={event.data as any}
            timestamp={event.timestamp}
          />
        ))}

        {messageEvents.length === 0 && (
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
