import {
  IconAt,
  IconChevronDown,
  IconPaperclip,
  IconPlayerStop,
  IconSend,
} from '@tabler/icons-react';
import { useCallback, useState } from 'react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { useLayoutStore } from '@/store/appStore';

type Message = {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: number;
};

const mockMessages: Message[] = [
  {
    id: '1',
    role: 'user',
    content: '帮我审查这个 Pull Request，检查代码质量和潜在问题。',
    timestamp: Date.now() - 60000,
  },
  {
    id: '2',
    role: 'assistant',
    content: `好的，我来帮你审查这个 Pull Request。

## 代码审查总结

### 发现的问题

1. **潜在空指针异常** - 在 \`src/utils/parser.ts\` 第 45 行，未对输入参数进行空值检查。

2. **性能优化建议** - \`src/components/List.tsx\` 中的 renderItem 回调建议使用 useMemo 包裹以避免不必要的重渲染。

### 改进建议

- 添加类型注释以提高代码可维护性
- 考虑提取公共逻辑到独立的工具函数中

整体代码质量良好，建议处理上述问题后合并。`,
    timestamp: Date.now() - 30000,
  },
];

const MODEL_OPTIONS = ['GPT-4', 'Claude-3', 'DeepSeek'] as const;

function formatTime(timestamp: number) {
  return new Date(timestamp).toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function CenterCanvas() {
  const { activeConversationId, setInputFocused } = useLayoutStore(
    (state) => state,
  );
  const [inputValue, setInputValue] = useState('');
  const [isGenerating] = useState(false);
  const [selectedModel] = useState(MODEL_OPTIONS[0]);

  const handleSend = useCallback(() => {
    if (!inputValue.trim()) return;
    setInputValue('');
  }, [inputValue]);

  return (
    <div className="flex h-full flex-1 flex-col bg-slate-50">
      <div className="flex h-12 items-center justify-between border-b border-slate-200 bg-white px-6">
        <h1 className="text-sm font-medium text-slate-700">
          {activeConversationId ? '代码审查讨论' : '选择一个会话'}
        </h1>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" className="text-slate-500">
            分享
          </Button>
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="mx-auto max-w-[650px] py-6 px-4">
          {!activeConversationId ? (
            <div className="flex flex-col items-center justify-center py-20 text-slate-400">
              <p>选择或创建一个会话开始对话</p>
            </div>
          ) : (
            <div className="space-y-6">
              {mockMessages.map((message) => (
                <div
                  key={message.id}
                  className={cn(
                    'rounded-lg px-4 py-3',
                    message.role === 'user'
                      ? 'bg-white border border-slate-200 ml-8'
                      : 'bg-transparent mr-8',
                  )}
                >
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-xs font-medium text-slate-500 uppercase tracking-wide">
                      {message.role === 'user' ? '你' : 'AI 助手'}
                    </span>
                    <span className="text-xs text-slate-400">
                      {formatTime(message.timestamp)}
                    </span>
                  </div>
                  <div
                    className={cn(
                      'text-sm leading-relaxed',
                      message.role === 'assistant'
                        ? 'font-serif text-slate-700'
                        : 'text-slate-600',
                    )}
                  >
                    {message.content.split('\n').map((line) => (
                      <p
                        key={`${message.id}-${line.slice(0, 20)}`}
                        className={line ? '' : 'mt-2'}
                      >
                        {line || '\u00A0'}
                      </p>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </ScrollArea>

      <div className="border-t border-slate-200 bg-white">
        <div className="mx-auto max-w-[800px] p-4">
          <div className="relative rounded-lg border border-slate-200 bg-white shadow-sm">
            <textarea
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onFocus={() => setInputFocused(true)}
              onBlur={() => setInputFocused(false)}
              placeholder="输入消息... 使用 @ 提及，/ 命令"
              className="w-full resize-none rounded-lg px-4 py-3 text-sm min-h-[80px] max-h-[200px] focus:outline-none placeholder:text-slate-400"
              rows={1}
            />
            <div className="flex items-center justify-between border-t border-slate-100 px-3 py-2">
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  className="text-slate-400 hover:text-slate-600"
                >
                  <IconAt className="size-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  className="text-slate-400 hover:text-slate-600"
                >
                  <IconPaperclip className="size-4" />
                </Button>
                <button
                  type="button"
                  className="flex items-center gap-1 rounded-md px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 transition-colors"
                >
                  {selectedModel}
                  <IconChevronDown className="size-3" />
                </button>
              </div>
              <div className="flex items-center gap-2">
                {isGenerating ? (
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-red-500 border-red-200 hover:bg-red-50"
                  >
                    <IconPlayerStop className="size-4 mr-1" />
                    停止
                  </Button>
                ) : (
                  <Button
                    size="sm"
                    className="bg-blue-500 hover:bg-blue-600 text-white"
                    onClick={handleSend}
                    disabled={!inputValue.trim()}
                  >
                    <IconSend className="size-4 mr-1" />
                    发送
                  </Button>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
