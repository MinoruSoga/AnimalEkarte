import { memo, type ReactNode } from "react";
import { MasterPageShell } from "./MasterPageShell";
import type { Resource } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface MasterTabPageProps {
  /** Page title (e.g. "診断病名マスタ") */
  title: string;
  /** Page icon element */
  icon: ReactNode;
  /** 権限バッジ表示用リソース */
  resource?: Resource;

  /** Handler for "新規登録" button. Route decides which tab's entity to create. */
  onNew: () => void;

  /** SidePanel(s) rendered next to main content — route composes its own multi-entity panels here */
  sidePanel: ReactNode;
  /** Delete confirmation dialog(s) — route composes its own multi-entity dialogs here */
  deleteDialogs: ReactNode;

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
  icon,
  resource,
  onNew,
  sidePanel,
  deleteDialogs,
  children,
}: MasterTabPageProps) {
  return (
    <>
      <MasterPageShell
        title={title}
        icon={icon}
        resource={resource}
        onNew={onNew}
        sidePanel={sidePanel}
      >
        {children}
      </MasterPageShell>

      {deleteDialogs}
    </>
  );
});
