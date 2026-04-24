import type { SliceCreator } from '../types';
import { flattenActions } from '../utils';

export type WorkspaceMode = 'remote' | 'local';

export type Conversation = {
  id: string;
  title: string;
  workspaceId: string;
  createdAt: number;
  updatedAt: number;
};

export type Workspace = {
  id: string;
  name: string;
  mode: WorkspaceMode;
  collapsed: boolean;
};

export type NavGroup = {
  id: string;
  label: string;
  items: NavItem[];
};

export type NavItem = {
  id: string;
  label: string;
  icon: string;
  badge?: number;
};

export type LayoutState = {
  leftRailCollapsed: boolean;
  rightRailCollapsed: boolean;
  activeConversationId: string | null;
  activeWorkspaceId: string | null;
  workspaces: Workspace[];
  conversations: Conversation[];
  inputFocused: boolean;
  activeRightTab: 'shortcuts' | 'inbox' | 'artifacts';
  navGroups: NavGroup[];
  collapsedNavGroups: Set<string>;
  conversationSearchQuery: string;
};

export type LayoutAction = Pick<LayoutActionImpl, keyof LayoutActionImpl>;
export type LayoutStore = LayoutState & LayoutAction;

const _initialState: LayoutState = {
  leftRailCollapsed: false,
  rightRailCollapsed: false,
  activeConversationId: 'conv-1',
  activeWorkspaceId: null,
  workspaces: [
    { id: 'remote-1', name: '远程工作区', mode: 'remote', collapsed: false },
    { id: 'local-1', name: '本地工作区', mode: 'local', collapsed: false },
  ],
  conversations: [
    {
      id: 'conv-1',
      title: '代码审查讨论',
      workspaceId: 'remote-1',
      createdAt: Date.now(),
      updatedAt: Date.now(),
    },
    {
      id: 'conv-2',
      title: '需求指派',
      workspaceId: 'remote-1',
      createdAt: Date.now() - 3600000,
      updatedAt: Date.now() - 3600000,
    },
    {
      id: 'conv-3',
      title: '技术问答',
      workspaceId: 'local-1',
      createdAt: Date.now() - 7200000,
      updatedAt: Date.now() - 7200000,
    },
  ],
  inputFocused: false,
  activeRightTab: 'shortcuts',
  navGroups: [
    {
      id: 'core',
      label: '',
      items: [
        { id: 'ai-assistant', label: 'AI 助手', icon: 'IconRobot' },
        { id: 'workbench', label: '工作台', icon: 'IconCommand' },
      ],
    },
    {
      id: 'ai-capability',
      label: 'AI 能力',
      items: [
        { id: 'ai-employee', label: 'AI 员工', icon: 'IconUsers' },
        { id: 'knowledge', label: '知识库', icon: 'IconBook' },
        { id: 'skills', label: '技能管理', icon: 'IconStar' },
      ],
    },
    {
      id: 'dev-collab',
      label: '研发协作',
      items: [
        { id: 'insflow', label: 'InsFlow', icon: 'IconGitBranch' },
        { id: 'insgit', label: 'InsGit', icon: 'IconCode' },
        { id: 'jenkins', label: 'Jenkins', icon: 'IconHammer' },
        { id: 'inssketch', label: 'InsSketch', icon: 'IconPaint' },
      ],
    },
    {
      id: 'team',
      label: '团队效率',
      items: [
        { id: 'org', label: '组织管理', icon: 'IconNetwork' },
        { id: 'reports', label: '汇报中心', icon: 'IconCalendar' },
        { id: 'plans', label: '计划任务', icon: 'IconCalendar' },
      ],
    },
    {
      id: 'system',
      label: '系统',
      items: [{ id: 'settings', label: '个人设置', icon: 'IconSettings' }],
    },
  ],
  collapsedNavGroups: new Set([
    'ai-capability',
    'dev-collab',
    'team',
    'system',
  ]),
  conversationSearchQuery: '',
};

type SetState = (
  partial:
    | LayoutStore
    | Partial<LayoutStore>
    | ((state: LayoutStore) => LayoutStore | Partial<LayoutStore>),
  replace?: boolean,
) => void;

export const createLayoutSlice = (
  set: SetState,
  get: () => LayoutStore,
  _api?: unknown,
) => new LayoutActionImpl(set, get, _api);

export class LayoutActionImpl {
  readonly #set: SetState;

  constructor(set: SetState, _get: () => LayoutStore, _api?: unknown) {
    void _api;
    void _get;
    this.#set = set;
  }

  toggleLeftRail = () => {
    this.#set((state) => ({ leftRailCollapsed: !state.leftRailCollapsed }));
  };

  toggleRightRail = () => {
    this.#set((state) => ({ rightRailCollapsed: !state.rightRailCollapsed }));
  };

  toggleWorkspaceCollapse = (workspaceId: string) => {
    this.#set((state) => ({
      workspaces: state.workspaces.map((w) =>
        w.id === workspaceId ? { ...w, collapsed: !w.collapsed } : w,
      ),
    }));
  };

  switchConversation = (conversationId: string) => {
    this.#set({ activeConversationId: conversationId });
  };

  createConversation = (workspaceId: string, title: string) => {
    const now = Date.now();
    const newConversation: Conversation = {
      id: `conv-${now}`,
      title,
      workspaceId,
      createdAt: now,
      updatedAt: now,
    };
    this.#set((state) => ({
      conversations: [...state.conversations, newConversation],
      activeConversationId: newConversation.id,
    }));
    return newConversation.id;
  };

  deleteConversation = (conversationId: string) => {
    this.#set((state) => ({
      conversations: state.conversations.filter((c) => c.id !== conversationId),
      activeConversationId:
        state.activeConversationId === conversationId
          ? null
          : state.activeConversationId,
    }));
  };

  setInputFocused = (focused: boolean) => {
    this.#set({ inputFocused: focused });
  };

  setActiveRightTab = (tab: 'shortcuts' | 'inbox' | 'artifacts') => {
    this.#set({ activeRightTab: tab });
  };

  toggleNavGroup = (groupId: string) => {
    this.#set((state) => {
      const collapsed = new Set(state.collapsedNavGroups);
      if (collapsed.has(groupId)) {
        collapsed.delete(groupId);
      } else {
        collapsed.add(groupId);
      }
      return { collapsedNavGroups: collapsed };
    });
  };

  setConversationSearchQuery = (query: string) => {
    this.#set({ conversationSearchQuery: query });
  };
}

export const layoutSlice: SliceCreator<LayoutStore> = (...params) => ({
  ..._initialState,
  ...flattenActions<LayoutAction>([
    createLayoutSlice(params[0] as SetState, params[1], params[2]),
  ]),
});
