/**
 * SupportPanel — サポートウィジェットの本体パネル
 *
 * 「使い方を聞く」（マニュアル検索）と「バグを報告」（スクショ付き報告）の2タブ。
 */
import { useState } from "react";
import { X } from "lucide-react";

import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { C, STYLE } from "@/lib/design-tokens";

import { BugReportTab } from "./BugReportTab";
import { ManualHelpTab } from "./ManualHelpTab";

interface SupportPanelProps {
  onClose: () => void;
}

const SUPPORT_TABS = [
  { value: "help", label: "使い方を聞く" },
  { value: "bug", label: "バグを報告" },
] as const;

export function SupportPanel({ onClose }: SupportPanelProps) {
  const [tab, setTab] = useState<string>("help");
  return (
    <section
      id="support-panel"
      aria-label="サポート"
      className={`flex flex-col overflow-hidden rounded-xl border ${C.borderLight} ${C.bgWhite} shadow-level2 w-[calc(100vw-2rem)] max-w-[420px] h-[min(600px,75vh)]`}
    >
      <div className={`flex items-center justify-between border-b ${C.borderLight} px-3 py-2`}>
        <h2 className={`text-base font-semibold ${C.text}`}>サポート</h2>
        <button
          type="button"
          onClick={onClose}
          aria-label="サポートを閉じる"
          className={`${STYLE.iconBtn32} ${C.text60} ${C.hoverBgMedium}`}
        >
          <X className="size-5" />
        </button>
      </div>
      <UnifiedTabs
        items={SUPPORT_TABS}
        value={tab}
        onValueChange={setTab}
        className="flex-1 min-h-0 gap-0 flex flex-col"
        listClassName="mt-1"
      >
        {/* flex カラムにして HelpChat（flex-1）が高さ一杯に広がり、入力欄を最下部へ固定する。
            チャット無効時の ManualSearchView は高さ自動で、この領域がそのままスクロールする */}
        <UnifiedTabsContent value="help" className="flex min-h-0 flex-col overflow-y-auto">
          <ManualHelpTab onClose={onClose} />
        </UnifiedTabsContent>
        <UnifiedTabsContent value="bug" className="min-h-0 overflow-y-auto">
          <BugReportTab onClose={onClose} />
        </UnifiedTabsContent>
      </UnifiedTabs>
    </section>
  );
}
