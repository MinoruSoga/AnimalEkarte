// FE-RC-045: 503行あった panels ファイルを StickyHeader / ClinicalTabs / ServiceTabs へ分割。
// 既存の import 経路 (`./MedicalRecordFormPanels`, vi.mock 対象) を維持するための re-export barrel。
import { LAYOUT } from "@/lib/design-tokens";
import { isMedicalRecordFinalizedStatus } from "../lib/medical-record-lock";
import { MedicalRecordClinicalTabs } from "./MedicalRecordClinicalTabs";
import { MedicalRecordServiceTabs } from "./MedicalRecordServiceTabs";
import type { MedicalRecordTabsAreaProps } from "../lib/medical-record-tabs-types";

export { MedicalRecordStickyHeader } from "./MedicalRecordStickyHeader";

export function MedicalRecordTabsArea(props: MedicalRecordTabsAreaProps) {
  const isFinalized = isMedicalRecordFinalizedStatus(props.recordStatus);
  return (
    <div className={`mt-4 ${LAYOUT.fullHeight}`}>
      <MedicalRecordClinicalTabs {...props} isFinalized={isFinalized} />
      <MedicalRecordServiceTabs {...props} isFinalized={isFinalized} />
    </div>
  );
}
