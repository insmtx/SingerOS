"use client";

import type { NavItem, ViewMode } from "@leros/store";
import { useLayoutStore } from "@leros/store";
import { Button } from "@leros/ui/components/ui/button";
import { ScrollArea } from "@leros/ui/components/ui/scroll-area";
import { cn } from "@leros/ui/lib/utils";
import {
	BookOpen,
	Bot,
	Calendar,
	ChevronDown,
	ChevronLeft,
	ChevronRight,
	Code2,
	GitBranch,
	Hammer,
	MessageSquare,
	Network,
	Paintbrush,
	Settings,
	Star,
	Terminal,
	Users,
} from "lucide-react";

const iconMap: Record<string, React.ReactNode> = {
	IconRobot: <Bot className="size-4" />,
	IconCommand: <Terminal className="size-4" />,
	IconUsers: <Users className="size-4" />,
	IconBook: <BookOpen className="size-4" />,
	IconStar: <Star className="size-4" />,
	IconGitBranch: <GitBranch className="size-4" />,
	IconCode: <Code2 className="size-4" />,
	IconHammer: <Hammer className="size-4" />,
	IconPaint: <Paintbrush className="size-4" />,
	IconNetwork: <Network className="size-4" />,
	IconReport: <Calendar className="size-4" />,
	IconCalendar: <Calendar className="size-4" />,
	IconSettings2: <Settings className="size-4" />,
	IconMessage: <MessageSquare className="size-4" />,
};

const navIdToView: Record<string, ViewMode> = {
	"ai-assistant": "chat",
	workbench: "workbench",
	"ai-employee": "digitalAssistant",
	knowledge: "knowledge",
	skills: "skills",
	settings: "settings",
};

export function LeftRail() {
	const {
		leftRailCollapsed,
		navGroups,
		collapsedNavGroups,
		currentView,
		toggleLeftRail,
		toggleNavGroup,
		switchView,
	} = useLayoutStore((s) => s);

	const handleNavClick = (item: NavItem) => {
		const view = navIdToView[item.id] ?? "chat";
		switchView(view);
	};

	const isItemActive = (item: NavItem) => {
		const view = navIdToView[item.id] ?? "chat";
		return currentView === view;
	};

	return (
		<div
			className={cn(
				"flex h-full flex-col border-r border-slate-200 bg-white transition-all duration-300",
				leftRailCollapsed ? "w-[52px]" : "w-[260px]",
			)}
		>
			<ScrollArea className="flex-1">
				<div className="p-1.5">
					{navGroups.map((group) => {
						const isCollapsed = collapsedNavGroups.has(group.id);

						if (leftRailCollapsed) {
							return (
								<div key={group.id} className="mb-1">
									{group.items.map((item: NavItem) => (
										<CollapsedNavItemButton
											key={item.id}
											item={item}
											active={isItemActive(item)}
											onClick={() => handleNavClick(item)}
										/>
									))}
								</div>
							);
						}

						return (
							<div key={group.id} className="mb-0.5">
								{group.label && (
									<button
										type="button"
										onClick={() => toggleNavGroup(group.id)}
										className="flex w-full items-center gap-1 rounded-md px-2 py-1.5 text-xs font-medium text-slate-500 hover:bg-slate-50 transition-colors"
									>
										{isCollapsed ? (
											<ChevronRight className="size-3.5" />
										) : (
											<ChevronDown className="size-3.5" />
										)}
										<span className="tracking-wide uppercase">{group.label}</span>
									</button>
								)}

								{!isCollapsed && (
									<div className={cn("mt-0.5", group.label && "ml-2")}>
										{group.items.map((item: NavItem) => (
											<NavItemButton
												key={item.id}
												item={item}
												active={isItemActive(item)}
												onClick={() => handleNavClick(item)}
											/>
										))}
									</div>
								)}
							</div>
						);
					})}
				</div>
			</ScrollArea>

			<div className="border-t border-slate-200 p-2">
				<Button
					variant="ghost"
					size="sm"
					className={cn(
						"w-full justify-start text-slate-500 hover:text-slate-700",
						leftRailCollapsed && "justify-center",
					)}
					onClick={toggleLeftRail}
				>
					{leftRailCollapsed ? (
						<ChevronRight className="size-4" />
					) : (
						<>
							<ChevronLeft className="size-4 mr-1.5" />
							收起侧栏
						</>
					)}
				</Button>
			</div>
		</div>
	);
}

function NavItemButton({
	item,
	active,
	onClick,
}: {
	item: NavItem;
	active: boolean;
	onClick: () => void;
}) {
	const icon = iconMap[item.icon];
	return (
		<button
			type="button"
			onClick={onClick}
			className={cn(
				"group flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm cursor-pointer transition-colors w-full text-left",
				active
					? "bg-blue-50 text-blue-700"
					: "text-slate-600 hover:bg-slate-50 hover:text-slate-800",
			)}
		>
			{icon}
			<span className="truncate">{item.label}</span>
			{item.badge && (
				<span className="ml-auto rounded-full bg-red-100 text-red-600 px-1.5 py-0.5 text-xs">
					{item.badge}
				</span>
			)}
		</button>
	);
}

function CollapsedNavItemButton({
	item,
	active,
	onClick,
}: {
	item: NavItem;
	active: boolean;
	onClick: () => void;
}) {
	const icon = iconMap[item.icon];
	return (
		<button
			type="button"
			onClick={onClick}
			className={cn(
				"flex items-center justify-center rounded-md p-2 transition-colors w-full cursor-pointer",
				active
					? "bg-blue-50 text-blue-700"
					: "text-slate-500 hover:bg-slate-50 hover:text-slate-700",
			)}
			title={item.label}
		>
			{icon}
		</button>
	);
}
