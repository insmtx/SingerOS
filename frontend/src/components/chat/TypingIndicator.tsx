import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export function TypingIndicator() {
  return (
    <div data-slot="typing-indicator" className="flex items-start gap-3">
      <Avatar size="sm">
        <AvatarFallback className="bg-blue-500 text-white text-xs">
          AI
        </AvatarFallback>
      </Avatar>
      <div className="rounded-lg bg-white border border-slate-200 px-4 py-3">
        <div className="flex items-center gap-1.5">
          <span className="size-1.5 rounded-full bg-slate-400 animate-pulse" />
          <span className="size-1.5 rounded-full bg-slate-400 animate-pulse [animation-delay:200ms]" />
          <span className="size-1.5 rounded-full bg-slate-400 animate-pulse [animation-delay:400ms]" />
        </div>
      </div>
    </div>
  );
}
