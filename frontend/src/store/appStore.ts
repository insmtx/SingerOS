import { devtools, subscribeWithSelector } from 'zustand/middleware';
import { createWithEqualityFn } from 'zustand/traditional';
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

export type AppStore = LayoutStore & TopicStore;
export type AppAction = LayoutAction & TopicAction;

const createStore: SliceCreator<AppStore> = (...params) => ({
  ...layoutSlice(...params),
  ...topicSlice(...params),
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
