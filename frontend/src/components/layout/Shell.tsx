import { CenterCanvas } from './CenterCanvas';
import { ConversationListPanel } from './ConversationListPanel';
import { LeftRail } from './LeftRail';
import { TopBar } from './TopBar';

export function Shell() {
  return (
    <div className="flex h-screen w-screen flex-col overflow-hidden bg-slate-50">
      <TopBar />
      <div className="flex flex-1 overflow-hidden">
        <LeftRail />
        <ConversationListPanel />
        <CenterCanvas />
      </div>
    </div>
  );
}
