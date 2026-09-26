/**
 * SupportPanel — サポートウィジェットの本体パネル
 *
 * 「使い方を聞く」（マニュアル検索）と「バグを報告」（スクショ付き報告）の2タブ。
 */
import { X } from "lucide-react";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { C, STYLE } from "@/lib/design-tokens";

import { BugReportTab } from "./BugReportTab";
import { ManualHelpTab } from "./ManualHelpTab";

interface SupportPanelProps {
  onClose: () => void;
}

export function SupportPanel({ onClose }: SupportPanelProps) {
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
          className={`${STYLE.iconBtn32} ${C.text45} ${C.hoverBgMedium}`}
        >
          <X className="size-5" />
        </button>
      </div>
      <Tabs defaultValue="help" className="flex-1 min-h-0 gap-0">
        <TabsList className="mx-3 mt-2">
          <TabsTrigger value="help">使い方を聞く</TabsTrigger>
          <TabsTrigger value="bug">バグを報告</TabsTrigger>
        </TabsList>
        <TabsContent value="help" className="min-h-0 overflow-y-auto">
          <ManualHelpTab onClose={onClose} />
        </TabsContent>
        <TabsContent value="bug" className="min-h-0 overflow-y-auto">
          <BugReportTab onClose={onClose} />
        </TabsContent>
      </Tabs>
    </section>
  );
}
