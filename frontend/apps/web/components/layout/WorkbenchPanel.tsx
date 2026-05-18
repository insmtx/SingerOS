"use client";

export function WorkbenchPanel() {
	return (
		<div
			data-slot="workbench-panel"
			className="flex h-full flex-1 flex-col items-center justify-center bg-slate-50"
		>
			<div className="flex flex-col items-center gap-3 text-slate-400">
				<span className="text-lg font-medium">工作台</span>
				<span className="text-sm">功能开发中，敬请期待</span>
			</div>
		</div>
	);
}