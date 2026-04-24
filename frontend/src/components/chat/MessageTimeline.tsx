import { useEffect, useRef } from 'react';
import { useChatStore } from '@/store/appStore';
import type { Message } from '@/types/chat';
import { AIMessageBubble } from './AIMessageBubble';
import { TypingIndicator } from './TypingIndicator';
import { UserMessageBubble } from './UserMessageBubble';
import { WelcomeScreen } from './WelcomeScreen';

function formatTime(timestamp: number): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function MessageTimeline() {
  const { messagesMap, messageIds, isGenerating, streamingMessageId } =
    useChatStore((s) => s);

  const scrollRef = useRef<HTMLDivElement>(null);
  const prevMessageCountRef = useRef(0);
  const prevStreamContentRef = useRef('');

  const messages = messageIds
    .map((id) => messagesMap[id])
    .filter(Boolean);

  useEffect(() => {
    const container = scrollRef.current;
    if (!container) return;

    const nearBottom =
      container.scrollHeight - container.scrollTop - container.clientHeight <
      120;

    const messageCountIncreased =
      messages.length > prevMessageCountRef.current;
    prevMessageCountRef.current = messages.length;

    const streamingMsg = streamingMessageId
      ? messagesMap[streamingMessageId]
      : null;
    const contentChanged =
      streamingMsg && streamingMsg.content !== prevStreamContentRef.current;
    prevStreamContentRef.current = streamingMsg?.content ?? '';

    if (nearBottom || messageCountIncreased || contentChanged) {
      container.scrollTop = container.scrollHeight;
    }
  }, [messages.length, streamingMessageId, messagesMap]);

  const isEmpty = messages.length === 0 && !isGenerating;

  return (
    <div
      ref={scrollRef}
      data-slot="message-timeline"
      className="min-h-0 flex-1 overflow-y-auto"
    >
      {isEmpty ? (
        <WelcomeScreen />
      ) : (
        <div className="mx-auto max-w-[720px] py-4 px-4 space-y-4">
          {messages.length > 0 && (
            <div className="flex items-center justify-center py-2">
              <span className="text-xs text-slate-400 bg-slate-100 rounded-full px-3 py-1">
                {formatTime(messages[0].timestamp)}
              </span>
            </div>
          )}
          {messages.map((msg: Message) => (
            <div key={msg.id}>
              {msg.role === 'user' ? (
                <UserMessageBubble message={msg} />
              ) : msg.role === 'assistant' ? (
                <AIMessageBubble
                  message={msg}
                  isStreaming={msg.id === streamingMessageId}
                />
              ) : null}
            </div>
          ))}
          {isGenerating && !streamingMessageId && <TypingIndicator />}
        </div>
      )}
    </div>
  );
}