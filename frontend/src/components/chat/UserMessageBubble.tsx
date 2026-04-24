import { useState } from 'react';
import { IconCheck, IconCopy } from '@tabler/icons-react';
import { Button } from '@/components/ui/button';
import { formatTime } from '@/utils/format';
import type { Message } from '@/types/chat';

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };
  return (
    <Button
      variant="ghost"
      size="icon-xs"
      className={
        copied
          ? 'text-green-400'
          : 'text-slate-300 opacity-0 group-hover:opacity-100 transition-opacity hover:text-slate-400'
      }
      onClick={handleCopy}
    >
      {copied ? <IconCheck className="size-3.5" /> : <IconCopy className="size-3.5" />}
    </Button>
  );
}

export function UserMessageBubble({ message }: { message: Message }) {
  return (
    <div data-slot="user-message" className="flex justify-end group">
      <div className="mr-2 max-w-[80%] flex flex-col items-end">
        <div className="flex items-center gap-2 mb-1">
          <CopyButton text={message.content} />
          <span className="text-xs text-slate-400">
            {formatTime(message.timestamp)}
          </span>
          <span className="text-xs font-medium text-slate-500 uppercase tracking-wide">
            你
          </span>
        </div>
        <div className="rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 px-4 py-3 text-white text-sm leading-relaxed shadow-sm">
          {message.content}
        </div>
      </div>
    </div>
  );
}