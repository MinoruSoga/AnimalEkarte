import { Settings } from "lucide-react";
import { Link } from "react-router";

/** 設定ページへのリンクに対応する全マスタカテゴリ */
export type MasterLinkCategory =
  | "serviceType"
  | "vaccine"
  | "examination"
  | "cage"
  | "trimming_course"
  | "trimming_option"
  | "medicine"
  | "consultation"
  | "procedure"
  | "hospitalization"
  | "insurance"
  | "staff"
  | "diagnosis_category"
  | "diagnosis_name";

const CATEGORY_PATH_MAP: Record<MasterLinkCategory, string> = {
  serviceType: "/settings/service-type",
  vaccine: "/settings/vaccine",
  examination: "/settings/examination",
  cage: "/settings/cage",
  trimming_course: "/settings/trimming-course",
  trimming_option: "/settings/trimming-option",
  medicine: "/settings/medicine",
  consultation: "/settings/consultation",
  procedure: "/settings/procedure",
  hospitalization: "/settings/hospitalization",
  insurance: "/settings/insurance",
  staff: "/settings/staff",
  diagnosis_category: "/settings/diagnosis-category",
  diagnosis_name: "/settings/diagnosis-name",
};

interface MasterLinkProps {
  /** Master category key (e.g. "serviceType", "vaccine") */
  category: MasterLinkCategory;
  /** Override the display label. Defaults to "マスタ管理" */
  label?: string;
  className?: string;
}

export function MasterLink({ category, label = "マスタ管理", className }: MasterLinkProps) {
  const path = CATEGORY_PATH_MAP[category] || "/settings";

  return (
    <Link
      to={path}
      className={`inline-flex items-center gap-1 text-xs text-[#37352F]/40 hover:text-[#37352F]/70 transition-colors ${className || ""}`}
    >
      <Settings className="size-3" />
      <span>{label}</span>
    </Link>
  );
}
