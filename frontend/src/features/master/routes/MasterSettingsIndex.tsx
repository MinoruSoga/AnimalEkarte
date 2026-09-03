// React/Framework
import type { ReactNode } from "react";
import { useCallback } from "react";
import { useNavigate } from "react-router";

// External
import { ChevronRight, Settings } from "lucide-react";
import { CATEGORY_CONFIG } from "../constants/category-config";
import {
  getResourceForCardKey,
  GROUP_CARD_CONFIG,
  isGroupCardKey,
  MASTER_SECTIONS,
} from "./master-settings-index-model";
import type { MasterCardKey, SectionDef } from "./master-settings-index-model";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";
import type { ResourceAction } from "@/hooks/use-permission";
import type { Resource } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// CardRow
// ─────────────────────────────────────────────────
interface CardRowProps {
  label: string;
  description: string;
  icon: ReactNode;
  count: number | undefined;
  onClick: () => void;
}

function CardRow({ label, description, icon, count, onClick }: CardRowProps) {
  return (
    <button type="button" className={STYLE.settingsRow} onClick={onClick}>
      <span className={STYLE.settingsRowIcon}>{icon}</span>
      <div className="flex-1 min-w-0 text-left">
        <div className={`text-base font-medium ${C.text} leading-tight`}>{label}</div>
        <div className={`text-base ${C.text45} mt-0.5 truncate`}>{description}</div>
      </div>
      {count !== undefined ? (
        <span className={`text-base ${C.text40} tabular-nums shrink-0`}>{count}件</span>
      ) : null}
      <ChevronRight className={`${ICON.action} ${C.text35} shrink-0`} />
    </button>
  );
}

// ─────────────────────────────────────────────────
// Permission-filtered card wrapper (BUG-123)
// Hook のルール: usePermission はコンポーネントトップレベルでのみ呼べるため、
// カード単位のラッパーコンポーネントで権限チェックを行う。
// ─────────────────────────────────────────────────
function PermissionFilteredCard({
  cardKey,
  navigate,
}: {
  cardKey: MasterCardKey;
  navigate: (path: string) => void;
}) {
  const resource = getResourceForCardKey(cardKey);
  const { canView } = usePermission(resource);

  if (!canView) return null;

  if (isGroupCardKey(cardKey)) {
    const cfg = GROUP_CARD_CONFIG[cardKey];
    const Icon = cfg.IconComponent;
    return (
      <CardRow
        label={cfg.label}
        description={cfg.description}
        icon={<Icon className={ICON.action} />}
        count={undefined}
        onClick={() => navigate(cfg.path)}
      />
    );
  }

  const cfg = CATEGORY_CONFIG[cardKey];
  const Icon = cfg.IconComponent;
  return (
    <CardRow
      label={cfg.label}
      description={cfg.description}
      icon={<Icon className={ICON.action} />}
      count={undefined}
      onClick={() => navigate(cfg.settingsPath)}
    />
  );
}

/** セクション内の全カードが非表示の場合、セクションごと非表示にするラッパー。
 *  useAuth の hasPermission を使い、フックのルールに違反しない形で一括判定する。 */
function PermissionFilteredSection({
  section,
  navigate,
  hasPermission,
}: {
  section: SectionDef;
  navigate: (path: string) => void;
  hasPermission: (resource: Resource, action: ResourceAction) => boolean;
}) {
  const hasVisibleCards = section.keys.some((key) =>
    hasPermission(getResourceForCardKey(key), "view"),
  );
  if (!hasVisibleCards) return null;

  return (
    <div className="mb-6">
      <div className={`px-1 pb-1.5 ${STYLE.sectionLabel}`}>{section.title}</div>
      <div
        className={`${C.bgWhite} rounded-lg border ${C.borderLight} overflow-hidden divide-y ${C.divideDivider}`}
      >
        {section.keys.map((key) => (
          <PermissionFilteredCard key={key} cardKey={key} navigate={navigate} />
        ))}
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// MasterSettingsIndex
// ─────────────────────────────────────────────────
export function MasterSettingsIndex() {
  const navigate = useNavigate();
  const { hasPermission } = useAuth();
  const handleNavigate = useCallback((path: string) => navigate(path), [navigate]);

  return (
    <PageLayout
      title="マスタ設定"
      icon={<Settings className={`${ICON.page} ${C.text}`} />}
      maxWidth="max-w-3xl"
      align="left"
    >
      <div className="px-6 pb-12">
        {/* Notion-style page icon */}
        <div className="pt-6 pb-2">
          <div className={STYLE.pageIcon}>
            <Settings className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>

        {/* Large page title */}
        <div className="pb-1 mb-1">
          <h2
            style={{
              fontSize: LAYOUT.pageTitle.fontSize,
              fontWeight: LAYOUT.pageTitle.fontWeight,
              lineHeight: LAYOUT.pageTitle.lineHeight,
              letterSpacing: LAYOUT.pageTitle.letterSpacing,
            }}
            className={C.text}
          >
            マスタ設定
          </h2>
        </div>
        <p className={`text-base ${C.text50} mb-6`}>動物病院の各種マスタデータを管理します</p>

        {/* Thin divider */}
        <div className={`${STYLE.sectionDivider} mb-6`} />

        {/* Sections — 権限フィルタリング付き (BUG-123) */}
        {MASTER_SECTIONS.map((section) => (
          <PermissionFilteredSection
            key={section.title}
            section={section}
            navigate={handleNavigate}
            hasPermission={hasPermission}
          />
        ))}
      </div>
    </PageLayout>
  );
}
