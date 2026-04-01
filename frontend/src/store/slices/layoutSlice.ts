import type { SliceCreator } from '../types';
import { flattenActions } from '../utils';

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
  collapsed: boolean;
};

export type LayoutState = {
  leftRailCollapsed: boolean;
  rightRailCollapsed: boolean;
  activeConversationId: string | null;
  activeWorkspaceId: string | null;
  workspaces: Workspace[];
  conversations: Conversation[];
  inputFocused: boolean;
  activeRightTab: 'shortcuts' | 'uploads' | 'generated';
};

export type LayoutAction = Pick<LayoutActionImpl, keyof LayoutActionImpl>;
export type LayoutStore = LayoutState & LayoutAction;

const _initialState: LayoutState = {
  leftRailCollapsed: false,
  rightRailCollapsed: false,
  activeConversationId: null,
  activeWorkspaceId: null,
  workspaces: [{ id: 'ws-1', name: '工作区', collapsed: false }],
  conversations: [
    {
      id: 'conv-1',
      title: '代码审查讨论',
      workspaceId: 'ws-1',
      createdAt: Date.now(),
      updatedAt: Date.now(),
    },
    {
      id: 'conv-2',
      title: '项目规划',
      workspaceId: 'ws-1',
      createdAt: Date.now() - 3600000,
      updatedAt: Date.now() - 3600000,
    },
  ],
  inputFocused: false,
  activeRightTab: 'shortcuts',
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

  setActiveRightTab = (tab: 'shortcuts' | 'uploads' | 'generated') => {
    this.#set({ activeRightTab: tab });
  };
}

export const layoutSlice: SliceCreator<LayoutStore> = (...params) => ({
  ..._initialState,
  ...flattenActions<LayoutAction>([
    createLayoutSlice(params[0] as SetState, params[1], params[2]),
  ]),
});
