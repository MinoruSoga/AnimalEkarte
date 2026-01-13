import React from "react";
import { DiagnosisHeaderChiefComplaint } from "./DiagnosisHeaderChiefComplaint";
import { DiagnosisHeaderPhysicalExam } from "./DiagnosisHeaderPhysicalExam";
import { DiagnosisHeaderDiagnosis } from "./DiagnosisHeaderDiagnosis";

interface DiagnosisHeaderProps {
  policy: string;
  setPolicy: (v: string) => void;
  diagnosisDetails: string;
  setDiagnosisDetails: (v: string) => void;
}

export const DiagnosisHeader = React.memo(function DiagnosisHeader({
  policy,
  setPolicy,
  diagnosisDetails,
  setDiagnosisDetails,
}: DiagnosisHeaderProps) {
  return (
    <div className="grid grid-cols-12 gap-3 shrink-0 h-[240px]">
      <DiagnosisHeaderChiefComplaint />
      <DiagnosisHeaderPhysicalExam policy={policy} setPolicy={setPolicy} />
      <DiagnosisHeaderDiagnosis
        diagnosisDetails={diagnosisDetails}
        setDiagnosisDetails={setDiagnosisDetails}
      />
    </div>
  );
});
