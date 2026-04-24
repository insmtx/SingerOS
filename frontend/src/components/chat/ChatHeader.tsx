import {
  IconDotsVertical,
  IconFileText,
  IconPlus,
  IconSearch,
  IconSettings2,
  IconShare,
} from '@tabler/icons-react';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { useChatStore, useLayoutStore } from '@/store/appStore';

export function ChatHeader() {
  const { selectedModel, modelOptions, tokenUsage } = useChatStore((s) => s);
  const { conversations, activeConversationId } = useLayoutStore((s) => s);

  const currentModel = modelOptions.find((m) => m.id === selectedModel);
  const activeConversation = conversations.find(
    (c) => c.id === activeConversationId,
  );

  const formatTokenCount = (count: number) => {
    if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
    return String(count);
  };

  return (
    <div
      data-slot="chat-header"
      className="flex h-12 items-center justify-between border-b border-slate-200 bg-white px-4"
    >
      <div className="flex items-center gap-3">
        <Avatar size="sm">
          <AvatarFallback className="bg-blue-500 text-white text-xs">
            AI
          </AvatarFallback>
        </Avatar>
        <div className="flex flex-col">
          <span className="text-sm font-medium text-slate-700">
            {activeConversation?.title ?? '选择一个会话'}
          </span>
          <span className="text-xs text-slate-400">
            {currentModel?.label ?? 'GPT-4'}
          </span>
        </div>
      </div>

      <div className="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon-sm"
          className="text-slate-500 hover:text-slate-700"
        >
          <IconPlus className="size-4" />
        </Button>

        <div className="flex items-center gap-1 rounded-md bg-slate-50 border border-slate-200 px-2.5 py-1 text-xs text-slate-500">
          <IconFileText className="size-3.5" />
          <span>{formatTokenCount(tokenUsage.total)}</span>
        </div>

        <Button
          variant="ghost"
          size="icon-sm"
          className="text-slate-500 hover:text-slate-700"
        >
          <IconSearch className="size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon-sm"
          className="text-slate-500 hover:text-slate-700"
        >
          <IconSettings2 className="size-4" />
        </Button>

        <Button
          variant="ghost"
          size="sm"
          className="text-slate-500 hover:text-slate-700"
        >
          <IconShare className="size-3.5" />
          <span className="ml-1">分享</span>
        </Button>

        <Button
          variant="ghost"
          size="icon-sm"
          className="text-slate-500 hover:text-slate-700"
        >
          <IconDotsVertical className="size-4" />
        </Button>
      </div>
    </div>
  );
}
