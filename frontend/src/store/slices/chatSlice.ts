import {
  mockMessages,
  mockModelOptions,
  mockStreamContent,
} from '@/mocks/chatMocks';
import { mockStreamResponse } from '@/mocks/streamSimulator';
import type {
  Attachment,
  Message,
  ModelOption,
  ToolCallStatus,
} from '@/types/chat';
import type { SliceCreator } from '../types';
import { flattenActions } from '../utils';

export type ChatState = {
  messagesMap: Record<string, Message>;
  messageIds: string[];
  streamingMessageId: string | null;
  isGenerating: boolean;
  streamCancelRef: (() => void) | null;

  inputText: string;
  inputAttachments: Attachment[];
  inputFocused: boolean;
  selectedModel: string;
  modelOptions: ModelOption[];

  tokenUsage: { total: number; currentSession: number };
};

export type ChatAction = Pick<ChatActionImpl, keyof ChatActionImpl>;
export type ChatStore = ChatState & ChatAction;

const _initialState: ChatState = {
  messagesMap: {},
  messageIds: [],
  streamingMessageId: null,
  isGenerating: false,
  streamCancelRef: null,

  inputText: '',
  inputAttachments: [],
  inputFocused: false,
  selectedModel: 'gpt-4',
  modelOptions: mockModelOptions,

  tokenUsage: { total: 239500, currentSession: 12500 },
};

type SetState = (
  partial:
    | ChatStore
    | Partial<ChatStore>
    | ((state: ChatStore) => ChatStore | Partial<ChatStore>),
  replace?: boolean,
) => void;

export const createChatSlice = (
  set: SetState,
  get: () => ChatStore,
  _api?: unknown,
) => new ChatActionImpl(set, get, _api);

export class ChatActionImpl {
  readonly #set: SetState;
  readonly #get: () => ChatStore;

  constructor(set: SetState, get: () => ChatStore, _api?: unknown) {
    void _api;
    this.#set = set;
    this.#get = get;
  }

  #dispatchChat = (action: ChatActionType) => {
    this.#set((state) => chatReducer(state, action));
  };

  sendMessage = (content: string, attachments?: Attachment[]) => {
    if (!content.trim() && !attachments?.length) return;

    const now = Date.now();
    const conversationId = 'conv-1';

    const userMsg: Message = {
      id: `msg-user-${now}`,
      conversationId,
      role: 'user',
      content,
      status: 'complete',
      timestamp: now,
    };

    const assistantMsg: Message = {
      id: `msg-assistant-${now}`,
      conversationId,
      role: 'assistant',
      content: '',
      status: 'streaming',
      timestamp: now + 100,
    };

    this.#dispatchChat({ type: 'addMessage', value: userMsg });
    this.#dispatchChat({ type: 'addMessage', value: assistantMsg });
    this.#set({
      streamingMessageId: assistantMsg.id,
      isGenerating: true,
      inputText: '',
      inputAttachments: [],
      tokenUsage: {
        total: this.#get().tokenUsage.total + Math.floor(Math.random() * 200 + 100),
        currentSession: this.#get().tokenUsage.currentSession + Math.floor(Math.random() * 200 + 100),
      },
    });

    this.internal_startStream(assistantMsg.id);
  };

  internal_startStream = (messageId: string) => {
    const { cancel } = mockStreamResponse({
      content: mockStreamContent,
      chunkSize: 3,
      chunkDelay: 30,
      onChunk: (chunk) => {
        const state = this.#get();
        const msg = state.messagesMap[messageId];
        if (!msg) return;
        this.#dispatchChat({
          type: 'updateMessage',
          id: messageId,
          value: { ...msg, content: msg.content + chunk },
        });
      },
      onComplete: () => {
        this.#set({
          streamingMessageId: null,
          isGenerating: false,
          streamCancelRef: null,
        });
        const state = this.#get();
        const msg = state.messagesMap[messageId];
        if (msg) {
          this.#dispatchChat({
            type: 'updateMessage',
            id: messageId,
            value: { ...msg, status: 'complete' },
          });
        }
      },
    });

    this.#set({ streamCancelRef: cancel });
  };

  cancelGeneration = () => {
    const state = this.#get();
    state.streamCancelRef?.();
    const streamingId = state.streamingMessageId;
    if (streamingId) {
      const msg = state.messagesMap[streamingId];
      if (msg) {
        this.#dispatchChat({
          type: 'updateMessage',
          id: streamingId,
          value: { ...msg, status: 'complete' },
        });
      }
    }
    this.#set({
      streamingMessageId: null,
      isGenerating: false,
      streamCancelRef: null,
    });
  };

  loadConversationMessages = (conversationId: string) => {
    const msgs = mockMessages[conversationId];
    if (!msgs) {
      this.#set({ messagesMap: {}, messageIds: [] });
      return;
    }

    const maps: Record<string, Message> = {};
    const ids: string[] = [];
    for (const m of msgs) {
      if (m.status === 'streaming') {
        continue;
      }
      maps[m.id] = m;
      ids.push(m.id);
    }

    this.#set({ messagesMap: maps, messageIds: ids });
  };

  setInputText = (text: string) => {
    this.#set({ inputText: text });
  };

  addAttachment = (file: File) => {
    const id = `att-${Date.now()}`;
    const url = URL.createObjectURL(file);
    const attachment: Attachment = {
      id,
      type: file.type.startsWith('image/') ? 'image' : 'file',
      name: file.name,
      size: file.size,
      url,
      file,
    };
    this.#set((state) => ({
      inputAttachments: [...state.inputAttachments, attachment],
    }));
  };

  removeAttachment = (id: string) => {
    const state = this.#get();
    const att = state.inputAttachments.find((a) => a.id === id);
    if (att?.url) URL.revokeObjectURL(att.url);
    this.#set((state) => ({
      inputAttachments: state.inputAttachments.filter((a) => a.id !== id),
    }));
  };

  setInputFocused = (focused: boolean) => {
    this.#set({ inputFocused: focused });
  };

  setSelectedModel = (modelId: string) => {
    this.#set({ selectedModel: modelId });
  };

  resendMessage = (messageId: string) => {
    const state = this.#get();
    const oldMsg = state.messagesMap[messageId];
    if (!oldMsg || oldMsg.role !== 'assistant') return;

    const now = Date.now();
    const newMsg: Message = {
      id: `msg-assistant-${now}`,
      conversationId: oldMsg.conversationId,
      role: 'assistant',
      content: '',
      status: 'streaming',
      timestamp: now,
    };

    this.#dispatchChat({ type: 'addMessage', value: newMsg });
    this.#set({
      streamingMessageId: newMsg.id,
      isGenerating: true,
      tokenUsage: {
        total: state.tokenUsage.total + Math.floor(Math.random() * 150 + 80),
        currentSession: state.tokenUsage.currentSession + Math.floor(Math.random() * 150 + 80),
      },
    });

    this.internal_startStream(newMsg.id);
  };
}

type ChatActionType =
  | { type: 'addMessage'; value: Message }
  | { type: 'updateMessage'; id: string; value: Message }
  | { type: 'removeMessage'; id: string }
  | {
      type: 'updateToolCallStatus';
      toolCallId: string;
      status: ToolCallStatus;
      result?: Record<string, unknown>;
    };

function chatReducer(state: ChatState, action: ChatActionType): ChatState {
  switch (action.type) {
    case 'addMessage': {
      const msg = action.value;
      return {
        ...state,
        messagesMap: { ...state.messagesMap, [msg.id]: msg },
        messageIds: [...state.messageIds, msg.id],
      };
    }

    case 'updateMessage': {
      const { id, value } = action;
      if (!state.messagesMap[id]) return state;
      return {
        ...state,
        messagesMap: { ...state.messagesMap, [id]: value },
      };
    }

    case 'removeMessage': {
      const { id } = action;
      const { [id]: _, ...remainingMaps } = state.messagesMap;
      return {
        ...state,
        messagesMap: remainingMaps,
        messageIds: state.messageIds.filter((mid) => mid !== id),
      };
    }

    case 'updateToolCallStatus': {
      const { toolCallId, status, result } = action;
      const msgId =
        state.streamingMessageId ??
        state.messageIds.find((id) => {
          const msg = state.messagesMap[id];
          return msg?.toolCalls?.some((tc) => tc.id === toolCallId);
        });

      if (!msgId) return state;
      const msg = state.messagesMap[msgId];
      if (!msg?.toolCalls) return state;

      const updatedToolCalls = msg.toolCalls.map((tc) =>
        tc.id === toolCallId
          ? { ...tc, status, ...(result ? { result } : {}) }
          : tc,
      );

      return {
        ...state,
        messagesMap: {
          ...state.messagesMap,
          [msgId]: { ...msg, toolCalls: updatedToolCalls },
        },
      };
    }

    default:
      return state;
  }
}

export const chatSlice: SliceCreator<ChatStore> = (...params) => ({
  ..._initialState,
  ...flattenActions<ChatAction>([
    createChatSlice(params[0] as SetState, params[1], params[2]),
  ]),
});
