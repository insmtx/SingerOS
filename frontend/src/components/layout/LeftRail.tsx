import {
  IconBook,
  IconCalendar,
  IconChevronDown,
  IconChevronRight,
  IconCode,
  IconCommand,
  IconGitBranch,
  IconHammer,
  IconMessage,
  IconNetwork,
  IconPaint,
  IconPlus,
  IconRobot,
  IconSearch,
  IconSettings2,
  IconStar,
  IconTrash,
  IconUsers,
} from '@tabler/icons-react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { useChatStore, useLayoutStore } from '@/store/appStore';
import type { NavItem } from '@/store/slices/layoutSlice';

const iconMap: Record<string, React.ReactNode> = {
  IconRobot: <IconRobot className="size-4" />,
  IconCommand: <IconCommand className="size-4" />,
  IconUsers: <IconUsers className="size-4" />,
  IconBook: <IconBook className="size-4" />,
  IconStar: <IconStar className="size-4" />,
  IconGitBranch: <IconGitBranch className="size-4" />,
  IconCode: <IconCode className="size-4" />,
  IconHammer: <IconHammer className="size-4" />,
  IconPaint: <IconPaint className="size-4" />,
  IconNetwork: <IconNetwork className="size-4" />,
  IconReport: <IconCalendar className="size-4" />,
  IconCalendar: <IconCalendar className="size-4" />,
  IconSettings2: <IconSettings2 className="size-4" />,
  IconMessage: <IconMessage className="size-4" />,
};

export function LeftRail() {
  const {
    navGroups,
    collapsedNavGroups,
    conversations,
    activeConversationId,
    conversationSearchQuery,
    toggleNavGroup,
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

  return (
    <div className="flex h-full w-[260px] flex-col border-r border-slate-200 bg-white">
      <div className="flex h-12 items-center justify-between border-b border-slate-200 px-4">
        <h2 className="text-sm font-medium tracking-wide uppercase text-slate-600">
          导航
        </h2>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-1.5">
          {navGroups.map((group) => {
            const isCollapsed = collapsedNavGroups.has(group.id);

            return (
              <div key={group.id} className="mb-0.5">
                {group.label && (
                  <button
                    type="button"
                    onClick={() => toggleNavGroup(group.id)}
                    className="flex w-full items-center gap-1 rounded-md px-2 py-1.5 text-xs font-medium text-slate-500 hover:bg-slate-50 transition-colors"
                  >
                    {isCollapsed ? (
                      <IconChevronRight className="size-3.5" />
                    ) : (
                      <IconChevronDown className="size-3.5" />
                    )}
                    <span className="tracking-wide uppercase">
                      {group.label}
                    </span>
                  </button>
                )}

                {(isCollapsed && group.label) || !group.label ? null : (
                  <div className={cn('mt-0.5', group.label && 'ml-2')}>
                    {group.items.map((item: NavItem) =>
                      item.id === 'ai-assistant' ? (
                        <AiAssistantSection
                          key={item.id}
                          activeConversationId={activeConversationId}
                          filteredConversations={filteredConversations}
                          onConversationClick={handleConversationClick}
                          onDeleteConversation={deleteConversation}
                          onCreateConversation={createConversation}
                          conversationSearchQuery={conversationSearchQuery}
                          onSearchChange={setConversationSearchQuery}
                        />
                      ) : (
                        <NavItemButton
                          key={item.id}
                          item={item}
                          active={false}
                        />
                      ),
                    )}
                  </div>
                )}

                {isCollapsed && group.label && (
                  <div className="ml-2">
                    {group.items.map((item: NavItem) => (
                      <NavItemButton key={item.id} item={item} active={false} />
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </ScrollArea>

      <div className="border-t border-slate-200 p-2">
        <Button
          variant="ghost"
          size="sm"
          className="w-full justify-start text-slate-500"
        >
          <IconPlus className="size-4 mr-1.5" />
          新建会话
        </Button>
      </div>
    </div>
  );
}

function NavItemButton({ item, active }: { item: NavItem; active: boolean }) {
  const icon = iconMap[item.icon];
  return (
    <button
      type="button"
      className={cn(
        'group flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm cursor-pointer transition-colors w-full text-left',
        active
          ? 'bg-blue-50 text-blue-700'
          : 'text-slate-600 hover:bg-slate-50 hover:text-slate-800',
      )}
    >
      {icon}
      <span className="truncate">{item.label}</span>
      {item.badge && (
        <span className="ml-auto rounded-full bg-red-100 text-red-600 px-1.5 py-0.5 text-xs">
          {item.badge}
        </span>
      )}
    </button>
  );
}

function AiAssistantSection({
  activeConversationId,
  filteredConversations,
  onConversationClick,
  onDeleteConversation,
  onCreateConversation,
  conversationSearchQuery,
  onSearchChange,
}: {
  activeConversationId: string | null;
  filteredConversations: { id: string; title: string; updatedAt: number }[];
  onConversationClick: (id: string) => void;
  onDeleteConversation: (id: string) => void;
  onCreateConversation: (workspaceId: string, title: string) => string;
  conversationSearchQuery: string;
  onSearchChange: (query: string) => void;
}) {
  return (
    <div data-slot="ai-assistant-section" className="py-1">
      <NavItemButton
        item={{ id: 'ai-assistant', label: 'AI 助手', icon: 'IconRobot' }}
        active={true}
      />
      <div className="mt-1 ml-2">
        <div className="relative mb-1">
          <IconSearch className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3 text-slate-400" />
          <input
            type="text"
            value={conversationSearchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="搜索会话"
            className="w-full rounded-md border border-slate-200 bg-slate-50 py-1.5 pl-7 pr-2 text-xs text-slate-600 placeholder:text-slate-400 focus:border-slate-300 focus:bg-white focus:outline-none transition-colors"
          />
        </div>
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
            onClick={() => onConversationClick(conv.id)}
          >
            <span className="truncate flex-1">{conv.title}</span>
            <Button
              variant="ghost"
              size="icon-xs"
              className="opacity-0 group-hover:opacity-100 transition-opacity text-slate-400 hover:text-red-500"
              onClick={(e) => {
                e.stopPropagation();
                onDeleteConversation(conv.id);
              }}
            >
              <IconTrash className="size-3" />
            </Button>
          </button>
        ))}
        <button
          type="button"
          onClick={() => {
            const id = onCreateConversation('remote-1', '新会话');
            onConversationClick(id);
          }}
          className="flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-sm text-slate-400 hover:text-slate-600 hover:bg-slate-50 transition-colors"
        >
          <IconPlus className="size-3.5" />
          <span>新建会话</span>
        </button>
      </div>
    </div>
  );
}
