import { ICON, LAYOUT } from "@/lib/design-tokens";
import { memo, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { Plus } from "lucide-react";
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { usePermission } from "@/hooks/use-permission";
import type { Resource } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface MasterPageShellProps {
  /** Page title (e.g. "ケージマスタ") */
  title: string;
  /** Page icon element */
  icon: ReactNode;
  /** 権限バッジ表示用リソース */
  resource?: Resource;
  /** Handler for "新規登録" button */
  onNew: () => void;
  /** SidePanel rendered next to main content */
  sidePanel: ReactNode;
  /** PageLayout body — MasterListPage wraps with PropertyFilter, MasterTabPage passes UnifiedTabs directly */
  children: ReactNode;
}

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

/**
 * MasterPageShell — common chrome shared by MasterListPage (single-entity
 * flat list) and MasterTabPage (multi-entity tabbed page): the
 * flex-h-full/PageLayout/canCreate-gated "新規登録" button/sidePanel slot.
 * Extracted to remove duplication between the two — see MasterListPage.tsx
 * and MasterTabPage.tsx for how each composes its own body/footer around it.
 */
export const MasterPageShell = memo(function MasterPageShell({
  title,
  icon,
  resource,
  onNew,
  sidePanel,
  children,
}: MasterPageShellProps) {
  const navigate = useNavigate();
  // BUG-124: resource が指定されている場合、create 権限がないなら「新規登録」ボタンを非表示
  // FE6-2: resource は任意。フック呼び出し順序維持のための sentinel（"" は未定義扱い）。
  const { canCreate } = usePermission((resource ?? "") as Resource);

  return (
    <div className="flex h-full">
      <div className="flex-1 min-w-0">
        <PageLayout
          title={title}
          icon={icon}
          resource={resource}
          onBack={() => navigate(paths.settings.getHref())}
          maxWidth={LAYOUT.pageContentMaxWidth.full}
          headerAction={
            canCreate ? (
              <PrimaryButton onClick={onNew}>
                <Plus className={`mr-1.5 ${ICON.action}`} />
                新規登録
              </PrimaryButton>
            ) : null
          }
        >
          {children}
        </PageLayout>
      </div>
      {sidePanel}
    </div>
  );
});
