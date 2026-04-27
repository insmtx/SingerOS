import {
  IconPlus,
  IconSearch,
  IconTrash,
} from '@tabler/icons-react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { useChatStore, useLayoutStore } from '@/store/appStore';

export function ConversationListPanel() {
  const {
    conversations,
    activeConversationId,
    conversationSearchQuery,
    conversationListOpen,
    switchConversation,
    createConversation,
    deleteConversation,
    setConversationSearchQuery,
  } = useLayoutStore((s) => s);

  const { loadConversationMessages } = useChatStore((s) => s);

  const filteredConversations = conversationSearchQuery
    ? conversations.filter((c) =>
        c.title.toLowerCase().includes(conversationSearchQuery.toLowerCase()),
      )
    : conversations;

  const handleConversationClick = (id: string) => {
    switchConversation(id);
    loadConversationMessages(id);
  };

  const handleCreateConversation = () => {
    const id = createConversation('remote-1', '新会话');
    handleConversationClick(id);
  };

  if (!conversationListOpen) return null;

  return (
    <div
      data-slot="conversation-list-panel"
      className="flex h-full w-[260px] flex-col border-r border-slate-200 bg-white transition-all duration-300"
    >
      <div className="flex items-center gap-2 border-b border-slate-200 px-3 py-2.5">
        <div className="relative flex-1">
          <IconSearch className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-slate-400" />
          <input
            type="text"
            value={conversationSearchQuery}
            onChange={(e) => setConversationSearchQuery(e.target.value)}
            placeholder="搜索会话"
            className="w-full rounded-md border border-slate-200 bg-slate-50 py-1.5 pl-7 pr-2 text-xs text-slate-600 placeholder:text-slate-400 focus:border-blue-300 focus:bg-white focus:outline-none transition-colors"
          />
        </div>
        <Button
          variant="ghost"
          size="icon-sm"
          className="text-slate-500 hover:text-slate-700 hover:bg-slate-50 shrink-0"
          onClick={handleCreateConversation}
        >
          <IconPlus className="size-4" />
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="px-3 pb-2">
          {filteredConversations.map((conv) => (
            <button
              key={conv.id}
              type="button"
              className={cn(
                'group relative flex items-center rounded-md px-2 py-1.5 text-sm cursor-pointer transition-colors w-full text-left',
                activeConversationId === conv.id
                  ? 'bg-blue-50 text-blue-700'
                  : 'text-slate-600 hover:bg-slate-50',
              )}
              onClick={() => handleConversationClick(conv.id)}
            >
              <span className="truncate flex-1">{conv.title}</span>
              <Button
                variant="ghost"
                size="icon-xs"
                className="opacity-0 group-hover:opacity-100 transition-opacity text-slate-400 hover:text-red-500"
                onClick={(e) => {
                  e.stopPropagation();
                  deleteConversation(conv.id);
                }}
              >
                <IconTrash className="size-3" />
              </Button>
            </button>
          ))}
        </div>
      </ScrollArea>

      </div>
  );
}