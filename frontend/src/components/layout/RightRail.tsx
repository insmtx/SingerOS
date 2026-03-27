import {
  IconBolt,
  IconFile,
  IconFileText,
  IconInbox,
  IconPhoto,
  IconUpload,
} from '@tabler/icons-react';
import type { ReactNode } from 'react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { useLayoutStore } from '@/store/appStore';

type QuickAction = {
  id: string;
  label: string;
  icon: ReactNode;
};

const QUICK_ACTIONS: QuickAction[] = [
  {
    id: 'review',
    label: '代码审查',
    icon: <IconFileText className="size-4" />,
  },
  {
    id: 'summarize',
    label: '总结文档',
    icon: <IconFileText className="size-4" />,
  },
  { id: 'explain', label: '解释代码', icon: <IconBolt className="size-4" /> },
  { id: 'test', label: '生成测试', icon: <IconFileText className="size-4" /> },
];

type ArtifactFile = {
  id: string;
  name: string;
  type: 'markdown' | 'image' | 'code';
  updatedAt: number;
};

const MOCK_ARTIFACTS: ArtifactFile[] = [
  {
    id: '1',
    name: 'review-summary.md',
    type: 'markdown',
    updatedAt: Date.now() - 60000,
  },
  {
    id: '2',
    name: 'architecture.png',
    type: 'image',
    updatedAt: Date.now() - 3600000,
  },
];

const TABS = [
  {
    id: 'shortcuts' as const,
    label: '快捷',
    icon: <IconBolt className="size-4" />,
  },
  {
    id: 'inbox' as const,
    label: '收件箱',
    icon: <IconInbox className="size-4" />,
  },
  {
    id: 'artifacts' as const,
    label: '工件',
    icon: <IconFile className="size-4" />,
  },
];

export function RightRail() {
  const { activeRightTab, setActiveRightTab } = useLayoutStore(
    (state) => state,
  );

  return (
    <div className="flex h-full w-[280px] flex-col border-l border-slate-200 bg-white">
      <div className="flex border-b border-slate-200">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            type="button"
            onClick={() => setActiveRightTab(tab.id)}
            className={cn(
              'flex flex-1 items-center justify-center gap-1.5 px-3 py-2.5 text-xs font-medium transition-colors',
              activeRightTab === tab.id
                ? 'text-blue-600 border-b-2 border-blue-500 bg-blue-50/50'
                : 'text-slate-500 hover:text-slate-700 hover:bg-slate-50',
            )}
          >
            {tab.icon}
            <span className="tracking-wide uppercase">{tab.label}</span>
          </button>
        ))}
      </div>

      <ScrollArea className="flex-1">
        <div className="p-3">
          {activeRightTab === 'shortcuts' && (
            <div className="space-y-2">
              <h3 className="text-xs font-medium text-slate-500 uppercase tracking-wide mb-3">
                快捷操作
              </h3>
              {QUICK_ACTIONS.map((action) => (
                <Button
                  key={action.id}
                  variant="outline"
                  size="sm"
                  className="w-full justify-start text-slate-600 hover:text-slate-800"
                >
                  {action.icon}
                  <span className="ml-2">{action.label}</span>
                </Button>
              ))}
            </div>
          )}

          {activeRightTab === 'inbox' && (
            <div className="space-y-3">
              <h3 className="text-xs font-medium text-slate-500 uppercase tracking-wide mb-3">
                文件收件箱
              </h3>
              <div className="border-2 border-dashed border-slate-200 rounded-lg p-6 text-center hover:border-slate-300 transition-colors cursor-pointer">
                <IconUpload className="size-8 mx-auto text-slate-300 mb-2" />
                <p className="text-sm text-slate-500">拖放文件到此处</p>
                <p className="text-xs text-slate-400 mt-1">或点击选择文件</p>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2 p-2 rounded-md hover:bg-slate-50 transition-colors cursor-pointer">
                  <IconFile className="size-4 text-slate-400" />
                  <span className="text-sm text-slate-600 truncate">
                    example.py
                  </span>
                </div>
              </div>
            </div>
          )}

          {activeRightTab === 'artifacts' && (
            <div className="space-y-3">
              <h3 className="text-xs font-medium text-slate-500 uppercase tracking-wide mb-3">
                工件文件
              </h3>
              {MOCK_ARTIFACTS.map((artifact) => (
                <div
                  key={artifact.id}
                  className="flex items-center gap-3 p-2 rounded-md hover:bg-slate-50 transition-colors cursor-pointer"
                >
                  {artifact.type === 'image' ? (
                    <IconPhoto className="size-4 text-green-500" />
                  ) : (
                    <IconFileText className="size-4 text-blue-500" />
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-slate-700 truncate">
                      {artifact.name}
                    </p>
                    <p className="text-xs text-slate-400">
                      {new Date(artifact.updatedAt).toLocaleString('zh-CN', {
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </p>
                  </div>
                </div>
              ))}
              {MOCK_ARTIFACTS.length === 0 && (
                <p className="text-sm text-slate-400 text-center py-4">
                  暂无工件
                </p>
              )}
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
