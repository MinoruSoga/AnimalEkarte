// React/Framework
import type { ReactNode } from "react";
import { useCallback } from "react";
import { useNavigate } from "react-router";

// External
import { Building2, ChevronRight, ClipboardList, FolderTree, Scissors, Settings, Stethoscope } from "lucide-react";
import { CATEGORY_CONFIG } from "@/features/master/constants/category-config";
import type { MasterSettingsCategory } from "@/features/master/constants/category-config";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { usePermission, useAuth } from "@/features/auth";
import type { ResourceAction } from "@/features/auth";
import type { Resource } from "@/types/generated/models";
import { ResourceHospitalSettings, ResourceMasterMedical, ResourceMasterTrimming } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

/** Group keys that are NOT individual MasterSettingsCategory */
type GroupKey = "clinic" | "treatmentItems" | "diagnosisGroup" | "trimmingGroup" | "inquiry_template";

/** All possible card keys in the settings index */
type MasterCardKey = MasterSettingsCategory | GroupKey;

interface SectionDef {
  title: string;
  keys: MasterCardKey[];
}

// ─────────────────────────────────────────────────
// Group card definitions (hardcoded)
// ─────────────────────────────────────────────────
interface GroupCardConfig {
  label: string;
  description: string;
  IconComponent: (props: { className?: string }) => ReactNode;
  path: string;
  /** 権限チェック用リソース定数 */
  resource: Resource;
  /** Which individual categories to sum for the count */
  countCategories: MasterSettingsCategory[];
}

const GROUP_CARD_CONFIG: Record<GroupKey, GroupCardConfig> = {
  clinic: {
    label: "医院マスタ",
    description: "院名、住所、電話番号などの医院基本情報を管理します",
    IconComponent: Building2,
    path: "/settings/clinic",
    resource: ResourceHospitalSettings,
    countCategories: [],
  },
  treatmentItems: {
    label: "診療項目マスタ",
    description: "診察・検査・処置・予防接種・定期健診の項目と単価(税込)を管理します",
    IconComponent: Stethoscope,
    path: "/settings/treatment-items",
    resource: ResourceMasterMedical,
    countCategories: ["consultation", "examination", "procedure", "vaccine", "checkup"],
  },
  diagnosisGroup: {
    label: "診断マスタ",
    description: "診断カテゴリと診断名を管理します",
    IconComponent: FolderTree,
    path: "/settings/diagnosis",
    resource: ResourceMasterMedical,
    countCategories: ["diagnosis_category", "diagnosis_name"],
  },
  trimmingGroup: {
    label: "トリミングマスタ",
    description: "トリミングコースとオプションを管理します",
    IconComponent: Scissors,
    path: "/settings/trimming",
    resource: ResourceMasterTrimming,
    countCategories: ["trimming_course", "trimming_option"],
  },
  inquiry_template: {
    label: "問診テンプレート",
    description: "問診票のテンプレートを管理します",
    IconComponent: ClipboardList,
    path: "/settings/inquiry-templates",
    resource: ResourceMasterMedical,
    countCategories: [],
  },
};

// ─────────────────────────────────────────────────
// Section definitions
// ─────────────────────────────────────────────────
const MASTER_SECTIONS: SectionDef[] = [
  { title: "基本設定", keys: ["clinic", "animal_species"] },
  {
    title: "カルテ",
    keys: ["treatmentItems", "diagnosisGroup", "inquiry_template", "medicine"],
  },
  { title: "予約管理マスタ", keys: ["reservationCategory"] },
  { title: "入院・ケージ管理", keys: ["hospitalization", "cage"] },
  { title: "トリミング関連", keys: ["trimmingGroup"] },
  { title: "会計・商品", keys: ["merchandise_item", "insurance"] },
  { title: "スタッフ・権限", keys: ["staff", "occupations", "permission_group"] },
];

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
        <span className={`text-base ${C.text40} tabular-nums shrink-0`}>
          {count}件
        </span>
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
function getResourceForKey(key: MasterCardKey): Resource {
  if (key in GROUP_CARD_CONFIG) {
    return GROUP_CARD_CONFIG[key as GroupKey].resource;
  }
  return CATEGORY_CONFIG[key as MasterSettingsCategory].resource;
}

function PermissionFilteredCard({
  cardKey,
  navigate,
}: {
  cardKey: MasterCardKey;
  navigate: (path: string) => void;
}) {
  const resource = getResourceForKey(cardKey);
  const { canView } = usePermission(resource);

  if (!canView) return null;

  if (cardKey in GROUP_CARD_CONFIG) {
    const cfg = GROUP_CARD_CONFIG[cardKey as GroupKey];
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

  const cat = cardKey as MasterSettingsCategory;
  const cfg = CATEGORY_CONFIG[cat];
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
    hasPermission(getResourceForKey(key), "view"),
  );
  if (!hasVisibleCards) return null;

  return (
    <div className="mb-5">
      <div className={`px-1 pb-1.5 text-base ${C.text40} uppercase tracking-wide select-none`}>
        {section.title}
      </div>
      <div className={`bg-white rounded-lg border ${C.borderLight} overflow-hidden divide-y ${C.divideDivider}`}>
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
            }}
            className={C.text}
          >
            マスタ設定
          </h2>
        </div>
        <p className={`text-base ${C.text50} mb-6`}>
          動物病院の各種マスタデータを管理します
        </p>

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
