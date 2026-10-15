import { memo, type ReactNode } from "react";
import { MasterPageShell } from "./MasterPageShell";
import type { Resource } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface MasterTabPageProps {
  /** Page title (e.g. "診断病名マスタ") */
  title: string;
  /** 権限バッジ表示用リソース */
  resource?: Resource;

  /** Handler for "新規登録" button. Route decides which tab's entity to create. */
  onNew: () => void;

  /** SidePanel(s) rendered next to main content — route composes its own multi-entity panels here */
  sidePanel: ReactNode;
  /** Delete confirmation dialog(s) — route composes its own multi-entity dialogs here */
  deleteDialogs: ReactNode;

  /** PageLayout maxWidth への転送（EMR-333: 少列マスタ画面の横幅制限等に適用） */
  maxWidth?: string;

  /** Tab content — route renders <UnifiedTabs>...</UnifiedTabs> here */
  children: ReactNode;
}

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

/**
 * MasterTabPage — shared chrome for multi-entity tabbed master pages
 * (DiagnosisSettings, TrimmingSettings, TreatmentPlanMaster).
 *
 * Counterpart to MasterListPage: MasterListPage owns a single entity's
 * search/filter/ConfirmDialog; MasterTabPage instead leaves entity count,
 * per-tab filtering, and delete confirmation entirely to the route, since
 * a tabbed master page manages 2+ independent entities with their own
 * PropertyFilter state and dialogs (see DiagnosisSettingsSidePanels /
 * DiagnosisDeleteDialogs for the multi-entity slot pattern).
 *
 * Shares the flex-h-full/PageLayout/canCreate-gated button/sidePanel chrome
 * with MasterListPage via MasterPageShell — only the body content and the
 * footer slot (ConfirmDialog vs. route-owned deleteDialogs) differ.
 */
export const MasterTabPage = memo(function MasterTabPage({
  title,
  resource,
  onNew,
  sidePanel,
  deleteDialogs,
  maxWidth,
  children,
}: MasterTabPageProps) {
  return (
    <>
      <MasterPageShell
        title={title}
        resource={resource}
        onNew={onNew}
        sidePanel={sidePanel}
        maxWidth={maxWidth}
        // タブバーはページ内ナビとして sticky ヘッダに接続するため、コンテンツ上余白（py-6 の pt 分）を
        // 除去する。先頭要素が PropertyFilter 等のコンテンツである MasterListPage には適用しない。
        contentClassName="pt-0"
      >
        {children}
      </MasterPageShell>

      {deleteDialogs}
    </>
  );
});
