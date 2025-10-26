import { useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Send } from "lucide-react";

interface ChatInputProps {
  onSend: (message: string) => Promise<void>;
  isSubmitting: boolean;
}

export function ChatInput({ onSend, isSubmitting }: ChatInputProps) {
  const [message, setMessage] = useState("");

  const handleSendMessage = useCallback(async () => {
    if (!message.trim() || isSubmitting) return;
    
    const messageContent = message.trim();
    setMessage("");
    await onSend(messageContent);
  }, [message, isSubmitting, onSend]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  }, [handleSendMessage]);

  return (
    <div className="flex-shrink-0 border-t p-2 bg-background">
      <div className="flex gap-2 items-end">
        <div className="flex-1">
          <Textarea
            placeholder="Type a message..."
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            rows={2}
            disabled={isSubmitting}
            className="resize-none text-sm"
          />
        </div>
        <Button
          onClick={handleSendMessage}
          disabled={!message.trim() || isSubmitting}
          size="icon"
          className="h-[52px] flex-shrink-0"
        >
          <Send className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  );
}
