export type StreamSimulatorOptions = {
  content: string;
  chunkSize?: number;
  chunkDelay?: number;
  onChunk: (chunk: string) => void;
  onComplete: () => void;
};

export function mockStreamResponse(options: StreamSimulatorOptions): {
  cancel: () => void;
} {
  const {
    content,
    chunkSize = 2,
    chunkDelay = 25,
    onChunk,
    onComplete,
  } = options;

  let cancelled = false;
  let index = 0;

  function tick() {
    if (cancelled) return;

    if (index >= content.length) {
      onComplete();
      return;
    }

    const end = Math.min(index + chunkSize, content.length);
    const chunk = content.slice(index, end);
    index = end;
    onChunk(chunk);
    setTimeout(tick, chunkDelay);
  }

  setTimeout(tick, 300);

  return {
    cancel: () => {
      cancelled = true;
    },
  };
}

export type ToolCallSimulatorOptions = {
  toolCallId: string;
  toolName: string;
  executionDelay?: number;
  onStatusChange: (status: 'running' | 'success') => void;
  onResult: (result: Record<string, unknown>) => void;
};

export function mockToolCallExecution(options: ToolCallSimulatorOptions): {
  cancel: () => void;
} {
  const {
    toolCallId: _toolCallId,
    toolName: _toolName,
    executionDelay = 800,
    onStatusChange,
    onResult,
  } = options;

  void _toolCallId;
  void _toolName;

  let cancelled = false;

  setTimeout(() => {
    if (cancelled) return;
    onStatusChange('running');
  }, 100);

  setTimeout(() => {
    if (cancelled) return;
    onStatusChange('success');
    onResult({ message: '操作完成', assigned_to: '张三' });
  }, executionDelay);

  return {
    cancel: () => {
      cancelled = true;
    },
  };
}
