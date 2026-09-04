import type { ReactNode } from "react";
import { UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { C, LAYOUT } from "@/lib/design-tokens";

export function MedicalRecordMountedTab({
  tab,
  activeTab,
  mountedTabs,
  contentClassName,
  children,
}: {
  tab: string;
  activeTab: string;
  mountedTabs: Set<string>;
  contentClassName?: string;
  children: ReactNode;
}) {
  if (!mountedTabs.has(tab)) return null;
  return (
    <UnifiedTabsContent value={tab} className={contentClassName}>
      <div className={`${LAYOUT.fullHeight} ${activeTab === tab ? "" : "hidden"}`}>{children}</div>
    </UnifiedTabsContent>
  );
}

export function MedicalRecordSaveRequired({
  show,
  children,
}: {
  show: boolean;
  children: ReactNode;
}) {
  if (show) {
    return (
      <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
        カルテを保存してから使用できます
      </div>
    );
  }
  return children;
}
