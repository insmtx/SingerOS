import { devtools, subscribeWithSelector } from 'zustand/middleware';
import { createWithEqualityFn } from 'zustand/traditional';
import { type ChatAction, type ChatStore, chatSlice } from './slices/chatSlice';
import {
  type LayoutAction,
  type LayoutStore,
  layoutSlice,
} from './slices/layoutSlice';
import {
  type TopicAction,
  type TopicStore,
  topicSlice,
} from './slices/topicSlice';
import type { SliceCreator } from './types';

export type AppStore = LayoutStore & TopicStore & ChatStore;
export type AppAction = LayoutAction & TopicAction & ChatAction;

const createStore: SliceCreator<AppStore> = (...params) => ({
  ...layoutSlice(...params),
  ...topicSlice(...params),
  ...chatSlice(...params),
});

export const useAppStore = createWithEqualityFn<AppStore>()(
  subscribeWithSelector(devtools(createStore)),
  Object.is,
);

export const useLayoutStore = <T>(
  selector: (state: LayoutStore & LayoutAction) => T,
): T => useAppStore(selector);

export const useTopicStore = <T>(
  selector: (state: TopicStore & TopicAction) => T,
): T => useAppStore(selector);

export const useChatStore = <T>(
  selector: (state: ChatStore & ChatAction) => T,
): T => useAppStore(selector);
