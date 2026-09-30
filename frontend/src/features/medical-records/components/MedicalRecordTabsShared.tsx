import type { ReactNode } from "react";
import { UnifiedTabsContent } from "@/components/shared/UnifiedTabs";
import { C, LAYOUT } from "@/lib/design-tokens";

export function MedicalRecordMountedTab({
  tab,
  activeTab,
  mountedTabs,
  contentClassName,
  isLocked,
  internalLock = false,
  children,
}: {
  tab: string;
  activeTab: string;
  mountedTabs: Set<string>;
  contentClassName?: string;
  /** 確定済み/送信権限なしのとき子コンテンツを disabled fieldset で包む。 */
  isLocked: boolean;
  /** true: ロック境界を子コンポーネント内で管理するタブ（読み取り専用パネルを fieldset 外に出す）。 */
  internalLock?: boolean;
  children: ReactNode;
}) {
  if (!mountedTabs.has(tab)) return null;
  return (
    <UnifiedTabsContent value={tab} className={contentClassName}>
      {internalLock ? (
        <div className={`${LAYOUT.fullHeight} ${activeTab === tab ? "" : "hidden"}`}>
          {children}
        </div>
      ) : (
        <fieldset
          disabled={isLocked}
          className={`${LAYOUT.fullHeight} ${activeTab === tab ? "" : "hidden"} border-0 p-0 m-0 min-w-0`}
        >
          {children}
        </fieldset>
      )}
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
