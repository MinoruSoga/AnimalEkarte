import { memo } from "react";
import { C, ICON } from "@/lib/design-tokens";
import { Settings } from "lucide-react";
import { Link } from "react-router";
import { paths } from "@/config/paths";

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
  | "diagnosis_name"
  | "chief_complaint"
  | "interview_template";

const CATEGORY_PATH_MAP: Record<MasterLinkCategory, string> = {
  serviceType: paths.settings.serviceType.getHref(),
  vaccine: paths.settings.vaccine.getHref(),
  examination: paths.settings.examination.getHref(),
  cage: paths.settings.cage.getHref(),
  trimming_course: paths.settings.trimmingCourse.getHref(),
  trimming_option: paths.settings.trimmingOption.getHref(),
  medicine: paths.settings.medicine.getHref(),
  consultation: paths.settings.consultation.getHref(),
  procedure: paths.settings.procedure.getHref(),
  hospitalization: paths.settings.hospitalization.getHref(),
  insurance: paths.settings.insurance.getHref(),
  staff: paths.settings.staff.getHref(),
  diagnosis_category: paths.settings.diagnosisCategory.getHref(),
  diagnosis_name: paths.settings.diagnosisName.getHref(),
  chief_complaint: paths.settings.interview.chiefComplaint.getHref(),
  interview_template: paths.settings.interview.interviewTemplate.getHref(),
};

interface MasterLinkProps {
  /** Master category key (e.g. "serviceType", "vaccine") */
  category: MasterLinkCategory;
  /** Override the display label. Defaults to "マスタ管理" */
  label?: string;
  className?: string;
}

export const MasterLink = memo(function MasterLink({ category, label = "マスタ管理", className }: MasterLinkProps) {
  const path = CATEGORY_PATH_MAP[category] || "/settings";

  return (
    <Link
      to={path}
      className={`inline-flex items-center gap-1 text-xs ${C.text40} ${C.hoverText}/70 transition-colors ${className || ""}`}
    >
      <Settings className={ICON.xs} />
      <span>{label}</span>
    </Link>
  );
});
