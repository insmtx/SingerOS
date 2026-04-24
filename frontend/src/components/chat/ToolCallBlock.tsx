import {
  IconCheck,
  IconChevronDown,
  IconChevronRight,
  IconLoader2,
  IconX,
} from '@tabler/icons-react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import type { ToolCall } from '@/types/chat';

export function ToolCallBlock({ toolCalls }: { toolCalls: ToolCall[] }) {
  const [expanded, setExpanded] = useState(false);

  const totalCalls = toolCalls.length;
  const successCount = toolCalls.filter((tc) => tc.status === 'success').length;
  const runningCount = toolCalls.filter((tc) => tc.status === 'running').length;

  return (
    <div
      data-slot="tool-call-block"
      className="rounded-lg border border-slate-200 bg-slate-50"
    >
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between px-3 py-2 text-sm cursor-pointer hover:bg-slate-100 transition-colors"
      >
        <div className="flex items-center gap-2">
          {expanded ? (
            <IconChevronDown className="size-3.5 text-slate-500" />
          ) : (
            <IconChevronRight className="size-3.5 text-slate-500" />
          )}
          <span className="text-slate-600 font-medium">
            工具调用 ({totalCalls})
          </span>
          {runningCount > 0 && (
            <span className="relative flex size-2">
              <span className="absolute inline-flex size-full rounded-full bg-yellow-400 opacity-75 animate-ping" />
              <span className="relative inline-flex size-2 rounded-full bg-yellow-500" />
            </span>
          )}
        </div>
        {!expanded && (
          <div className="flex items-center gap-1.5 text-xs">
            {successCount > 0 && (
              <span className="text-green-600">{successCount} 完成</span>
            )}
            {runningCount > 0 && (
              <span className="text-yellow-600">{runningCount} 执行中</span>
            )}
          </div>
        )}
      </button>

      {expanded && (
        <div className="border-t border-slate-200 px-3 py-2 space-y-2">
          {toolCalls.map((tc) => (
            <ToolCallItem key={tc.id} toolCall={tc} />
          ))}
        </div>
      )}
    </div>
  );
}

function ToolCallItem({ toolCall }: { toolCall: ToolCall }) {
  const [showArgs, setShowArgs] = useState(false);
  const [showResult, setShowResult] = useState(false);

  return (
    <div data-slot="tool-call-item" className="space-y-1">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {toolCall.status === 'running' && (
            <IconLoader2 className="size-3.5 text-yellow-500 animate-spin" />
          )}
          {toolCall.status === 'success' && (
            <IconCheck className="size-3.5 text-green-500" />
          )}
          {toolCall.status === 'error' && (
            <IconX className="size-3.5 text-red-500" />
          )}
          {toolCall.status === 'pending' && (
            <span className="size-3.5 rounded-full border-2 border-slate-300" />
          )}
          <span className="text-sm font-medium text-slate-700">
            {toolCall.name}
          </span>
          {toolCall.duration && (
            <span className="text-xs text-slate-400">
              {toolCall.duration}ms
            </span>
          )}
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-xs"
            className="text-slate-400 hover:text-slate-600"
            onClick={() => setShowArgs(!showArgs)}
          >
            <IconChevronDown
              className={cn(
                'size-3 transition-transform',
                showArgs && 'rotate-180',
              )}
            />
          </Button>
          {toolCall.result && (
            <Button
              variant="ghost"
              size="icon-xs"
              className="text-slate-400 hover:text-slate-600"
              onClick={() => setShowResult(!showResult)}
            >
              结果
            </Button>
          )}
        </div>
      </div>

      {showArgs && (
        <div className="rounded bg-slate-100 px-2 py-1.5 text-xs text-slate-600 overflow-x-auto">
          <pre className="whitespace-pre-wrap">
            {JSON.stringify(toolCall.arguments, null, 2)}
          </pre>
        </div>
      )}

      {showResult && toolCall.result && (
        <div className="rounded bg-green-50 px-2 py-1.5 text-xs text-green-700 overflow-x-auto">
          <pre className="whitespace-pre-wrap">
            {JSON.stringify(toolCall.result, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}
