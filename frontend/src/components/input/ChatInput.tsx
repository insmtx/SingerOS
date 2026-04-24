import {
  IconAt,
  IconChevronDown,
  IconPaperclip,
  IconPlayerStop,
  IconSend,
  IconX,
} from '@tabler/icons-react';
import { useCallback, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { useChatStore } from '@/store/appStore';
import type { Attachment } from '@/types/chat';

export function ChatInput() {
  const {
    inputText,
    inputAttachments,
    isGenerating,
    selectedModel,
    modelOptions,
    setInputText,
    sendMessage,
    cancelGeneration,
    addAttachment,
    removeAttachment,
    setInputFocused,
    setSelectedModel,
  } = useChatStore((s) => s);

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [showModelDropdown, setShowModelDropdown] = useState(false);

  const currentModel = modelOptions.find((m) => m.id === selectedModel);

  const adjustHeight = useCallback(() => {
    const textarea = textareaRef.current;
    if (!textarea) return;
    textarea.style.height = 'auto';
    const maxHeight = 200;
    textarea.style.height = `${Math.min(textarea.scrollHeight, maxHeight)}px`;
  }, []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        if (inputText.trim() || inputAttachments.length > 0) {
          sendMessage(inputText, inputAttachments);
          if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
          }
        }
      }
    },
    [inputText, inputAttachments, sendMessage],
  );

  const handleTextareaChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      setInputText(e.target.value);
      adjustHeight();
    },
    [setInputText, adjustHeight],
  );

  const handlePaste = useCallback(
    (e: React.ClipboardEvent) => {
      const files = Array.from(e.clipboardData.files);
      for (const file of files) {
        if (file.type.startsWith('image/') || file.type.startsWith('text/')) {
          addAttachment(file);
        }
      }
    },
    [addAttachment],
  );

  const handleFileSelect = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = Array.from(e.target.files ?? []);
      for (const file of files) {
        addAttachment(file);
      }
      e.target.value = '';
    },
    [addAttachment],
  );

  const handleSend = useCallback(() => {
    if (inputText.trim() || inputAttachments.length > 0) {
      sendMessage(inputText, inputAttachments);
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto';
      }
    }
  }, [inputText, inputAttachments, sendMessage]);

  return (
    <div data-slot="chat-input" className="border-t border-slate-200 bg-white">
      <div className="mx-auto max-w-[800px] p-4">
        {inputAttachments.length > 0 && (
          <AttachmentPreview
            attachments={inputAttachments}
            onRemove={removeAttachment}
          />
        )}
        <div className="relative rounded-lg border border-slate-200 bg-white shadow-sm focus-within:border-blue-300 focus-within:shadow-blue-100 transition-all">
          <textarea
            ref={textareaRef}
            value={inputText}
            onChange={handleTextareaChange}
            onKeyDown={handleKeyDown}
            onPaste={handlePaste}
            onFocus={() => setInputFocused(true)}
            onBlur={() => setInputFocused(false)}
            placeholder="请描述您的问题，支持 Ctrl+V 粘贴图片。输入 @ 提及成员，/ 使用命令，# 引用工作项。"
            className="w-full resize-none rounded-lg px-4 py-3 text-sm min-h-[52px] max-h-[200px] focus:outline-none placeholder:text-slate-400"
            rows={1}
          />
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            accept="image/*,.pdf,.txt,.md,.json,.csv"
            multiple
            onChange={handleFileSelect}
          />
          <div className="flex items-center justify-between border-t border-slate-100 px-3 py-2">
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="icon-sm"
                className="text-slate-400 hover:text-slate-600"
                onClick={() => fileInputRef.current?.click()}
              >
                <IconPaperclip className="size-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon-sm"
                className="text-slate-400 hover:text-slate-600"
              >
                <IconAt className="size-4" />
              </Button>
              <div className="relative">
                <button
                  type="button"
                  onClick={() => setShowModelDropdown(!showModelDropdown)}
                  className="flex items-center gap-1 rounded-md px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 transition-colors"
                >
                  {currentModel?.label ?? 'GPT-4'}
                  <IconChevronDown className="size-3" />
                </button>
                {showModelDropdown && (
                  <div className="absolute bottom-full left-0 mb-1 rounded-lg border border-slate-200 bg-white shadow-lg py-1 z-10 min-w-[140px]">
                    {modelOptions.map((model) => (
                      <button
                        key={model.id}
                        type="button"
                        onClick={() => {
                          setSelectedModel(model.id);
                          setShowModelDropdown(false);
                        }}
                        className={cn(
                          'flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-slate-50 transition-colors',
                          model.id === selectedModel
                            ? 'text-blue-600 bg-blue-50/50'
                            : 'text-slate-600',
                        )}
                      >
                        <span>{model.label}</span>
                        <span className="text-xs text-slate-400">
                          {model.provider}
                        </span>
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </div>
            <div className="flex items-center gap-2">
              {isGenerating ? (
                <Button
                  variant="outline"
                  size="sm"
                  className="text-red-500 border-red-200 hover:bg-red-50"
                  onClick={cancelGeneration}
                >
                  <IconPlayerStop className="size-4 mr-1" />
                  停止
                </Button>
              ) : (
                <Button
                  size="sm"
                  className="bg-blue-500 hover:bg-blue-600 text-white"
                  onClick={handleSend}
                  disabled={!inputText.trim() && inputAttachments.length === 0}
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
  );
}

function AttachmentPreview({
  attachments,
  onRemove,
}: {
  attachments: Attachment[];
  onRemove: (id: string) => void;
}) {
  return (
    <div data-slot="attachment-preview" className="flex gap-2 mb-3 flex-wrap">
      {attachments.map((att) => (
        <div
          key={att.id}
          className="flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm"
        >
          {att.type === 'image' && att.url ? (
            <img
              src={att.url}
              alt={att.name}
              className="size-8 rounded object-cover"
            />
          ) : (
            <IconPaperclip className="size-3.5 text-slate-400" />
          )}
          <span className="text-slate-600 truncate max-w-[120px]">
            {att.name}
          </span>
          <button
            type="button"
            onClick={() => onRemove(att.id)}
            className="text-slate-400 hover:text-slate-600 transition-colors"
          >
            <IconX className="size-3.5" />
          </button>
        </div>
      ))}
    </div>
  );
}
