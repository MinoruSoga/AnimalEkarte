// React/Framework
import type { ReactNode } from "react";
import { useNavigate } from "react-router";

// External
import {
  ChevronRight,
  Settings,
} from "lucide-react";
import { CATEGORY_CONFIG } from "@/features/master/constants/category-config";
import type { MasterSettingsCategory } from "@/features/master/constants/category-config";

// Internal
import { PageLayout } from "@/components/shared/PageLayout";
import { useMasterItems } from "@/hooks/use-master-items";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";

// ─────────────────────────────────────────────────
// Category → backend label map (for item counting)
// ─────────────────────────────────────────────────
const CATEGORY_LABEL_MAP: Record<MasterSettingsCategory, string> = {
  serviceType: "診療内容",
  medicine: "薬剤",
  hospitalization: "入院",
  cage: "ケージ",
  staff: "スタッフ",
  job_title: "職種",
  insurance: "保険",
  consultation: "診察",
  examination: "検査",
  procedure: "処置",
  vaccine: "予防",
  checkup: "定期健診",
  diagnosis_category: "診断カテゴリ",
  diagnosis_name: "診断名",
  trimming_course: "トリミングコース",
  trimming_option: "トリミングオプション",
};

interface SectionDef {
  title: string;
  keys: MasterSettingsCategory[];
}

// ─── "clinic" is special (not a MasterSettingsCategory) ───
interface ClinicCard {
  type: "clinic";
  label: string;
  description: string;
  IconComponent: (props: { className?: string }) => ReactNode;
  path: string;
}

// Suppress unused variable warning — ClinicCard is referenced via inline object literals below
type _ClinicCardRef = ClinicCard;

const SECTIONS: SectionDef[] = [
  {
    title: "診療関連マスタ",
    keys: ["serviceType", "consultation", "examination", "procedure", "vaccine", "checkup", "medicine", "diagnosis_category", "diagnosis_name"],
  },
  {
    title: "入院・ケージ管理",
    keys: ["hospitalization", "cage"],
  },
  {
    title: "トリミング関連",
    keys: ["trimming_course", "trimming_option"],
  },
  {
    title: "スタッフ・保険",
    keys: ["staff", "job_title", "insurance"],
  },
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
        <div className={`text-sm font-medium ${C.text} leading-tight`}>{label}</div>
        <div className={`text-xs ${C.text45} mt-0.5 truncate`}>{description}</div>
      </div>
      {count !== undefined && (
        <span className={`text-xs ${C.text40} tabular-nums shrink-0`}>
          {count}件
        </span>
      )}
      <ChevronRight className={`size-4 ${C.text35} shrink-0`} />
    </button>
  );
}

// ─────────────────────────────────────────────────
// MasterSettingsIndex
// ─────────────────────────────────────────────────
export function MasterSettingsIndex() {
  const navigate = useNavigate();
  const { data: allItems } = useMasterItems();

  // Build count map: backend category label → count
  const countByLabel: Record<string, number> = {};
  for (const item of allItems) {
    if (item.category) {
      countByLabel[item.category] = (countByLabel[item.category] ?? 0) + 1;
    }
  }

  function getCount(cat: MasterSettingsCategory): number | undefined {
    const label = CATEGORY_LABEL_MAP[cat];
    if (!label) return undefined;
    return countByLabel[label] ?? 0;
  }

  return (
    <PageLayout
      title="マスタ設定"
      icon={<Settings className="size-5 text-[#37352F]" />}
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
        <p className={`text-sm ${C.text50} mb-6`}>
          動物病院の各種マスタデータを管理します
        </p>

        {/* Thin divider */}
        <div className={`${STYLE.sectionDivider} mb-6`} />

        {/* ── 基本設定 ── */}
        <div className="mb-5">
          <div className={`px-1 pb-1.5 text-xs ${C.text40} uppercase tracking-wide select-none`}>
            基本設定
          </div>
          <div className={`bg-white rounded-lg border ${C.borderLight} overflow-hidden divide-y ${C.divideDivider}`}>
            <CardRow
              label="病院情報"
              description="病院名、住所、電話番号などの基本情報を管理します"
              icon={<Settings className="size-[16px]" />}
              count={undefined}
              onClick={() => navigate("/settings/clinic")}
            />
          </div>
        </div>

        {/* ── Category sections ── */}
        {SECTIONS.map((section) => (
          <div key={section.title} className="mb-5">
            <div className={`px-1 pb-1.5 text-xs ${C.text40} uppercase tracking-wide select-none`}>
              {section.title}
            </div>
            <div className={`bg-white rounded-lg border ${C.borderLight} overflow-hidden divide-y ${C.divideDivider}`}>
              {section.keys.map((cat) => {
                const cfg = CATEGORY_CONFIG[cat];
                const Icon = cfg.IconComponent;
                return (
                  <CardRow
                    key={cat}
                    label={cfg.label}
                    description={cfg.description}
                    icon={<Icon className="size-[16px]" />}
                    count={getCount(cat)}
                    onClick={() => navigate(cfg.settingsPath)}
                  />
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </PageLayout>
  );
}
