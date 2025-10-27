export function TypingIndicator() {
  return (
    <div className="flex items-start gap-3 p-3 animate-in fade-in duration-300">
      {/* Bot Avatar */}
      <div className="flex-shrink-0">
        <div className="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center">
          <span className="text-xs">🤖</span>
        </div>
      </div>

      {/* Typing dots */}
      <div className="flex items-center gap-1 pt-1">
        <div className="w-2 h-2 rounded-full bg-muted-foreground/40 animate-bounce [animation-delay:-0.3s]" />
        <div className="w-2 h-2 rounded-full bg-muted-foreground/40 animate-bounce [animation-delay:-0.15s]" />
        <div className="w-2 h-2 rounded-full bg-muted-foreground/40 animate-bounce" />
      </div>
    </div>
  );
}
