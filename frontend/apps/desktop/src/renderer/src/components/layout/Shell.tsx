"use client";

import { useLayoutStore } from "@leros/store";
import { AssistantDetailPanel } from "../digitalAssistant/AssistantDetailPanel";
import { AssistantListView } from "../digitalAssistant/AssistantListView";
import { CenterCanvas } from "./CenterCanvas";
import { ConversationListPanel } from "./ConversationListPanel";
import { LeftRail } from "./LeftRail";
import { TopBar } from "./TopBar";
import { WorkbenchPanel } from "./WorkbenchPanel";

export function Shell() {
	const currentView = useLayoutStore((s) => s.currentView);

	return (
		<div className="flex h-screen w-screen flex-col overflow-hidden bg-slate-50">
			<TopBar />
			<div className="flex flex-1 overflow-hidden">
				<LeftRail />
				{currentView === "chat" && (
					<>
						<ConversationListPanel />
						<CenterCanvas />
					</>
				)}
				{currentView === "workbench" && <WorkbenchPanel />}
				{currentView === "digitalAssistant" && (
					<>
						<AssistantListView />
						<AssistantDetailPanel />
					</>
				)}
			</div>
		</div>
	);
}
