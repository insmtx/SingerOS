import {
  IconBolt,
  IconBook,
  IconCode,
  IconFileText,
  IconMessage,
} from '@tabler/icons-react';
import { Button } from '@/components/ui/button';

export function WelcomeScreen() {
  return (
    <div
      data-slot="welcome-screen"
      className="flex flex-col items-center justify-center py-16 px-4"
    >
      <div className="size-12 rounded-full bg-blue-50 flex items-center justify-center mb-4">
        <IconBolt className="size-6 text-blue-500" />
      </div>
      <h2 className="text-lg font-medium text-slate-700 mb-1">
        SingerOS AI 助手
      </h2>
      <p className="text-sm text-slate-500 mb-8 text-center max-w-[320px]">
        选择一个快捷操作开始，或直接输入你的问题
      </p>
      <div className="grid grid-cols-2 gap-2 max-w-[360px]">
        <Button
          variant="outline"
          size="sm"
          className="justify-start text-slate-600 hover:text-slate-800"
        >
          <IconCode className="size-4" />
          <span className="ml-2">代码审查</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="justify-start text-slate-600 hover:text-slate-800"
        >
          <IconFileText className="size-4" />
          <span className="ml-2">总结文档</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="justify-start text-slate-600 hover:text-slate-800"
        >
          <IconBook className="size-4" />
          <span className="ml-2">解释代码</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="justify-start text-slate-600 hover:text-slate-800"
        >
          <IconBolt className="size-4" />
          <span className="ml-2">生成测试</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="justify-start text-slate-600 hover:text-slate-800"
        >
          <IconMessage className="size-4" />
          <span className="ml-2">需求指派</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="justify-start text-slate-600 hover:text-slate-800"
        >
          <IconBook className="size-4" />
          <span className="ml-2">知识问答</span>
        </Button>
      </div>
    </div>
  );
}
