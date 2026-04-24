import { ChatHeader } from '@/components/chat/ChatHeader';
import { MessageTimeline } from '@/components/chat/MessageTimeline';
import { ChatInput } from '@/components/input/ChatInput';

export function CenterCanvas() {
  return (
    <div
      data-slot="center-canvas"
      className="flex h-full flex-1 flex-col bg-slate-50"
    >
      <ChatHeader />
      <MessageTimeline />
      <ChatInput />
    </div>
  );
}
