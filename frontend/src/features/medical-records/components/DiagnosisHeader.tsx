// React/Framework
import React from "react";

// Internal
import { usePermission } from "@/features/auth";

// Relative
import { DiagnosisHeaderChiefComplaint } from "./DiagnosisHeaderChiefComplaint";
import { DiagnosisHeaderPhysicalExam } from "./DiagnosisHeaderPhysicalExam";
import { DiagnosisHeaderDiagnosis } from "./DiagnosisHeaderDiagnosis";

interface DiagnosisHeaderProps {
  chiefComplaint?: string;
  policy: string;
  setPolicy: (v: string) => void;
  diagnosisDetails: string;
  setDiagnosisDetails: (v: string) => void;
  diagnosis1CategoryId?: number | null;
  setDiagnosis1CategoryId?: (id: number | null) => void;
  diagnosis1NameId?: number | null;
  setDiagnosis1NameId?: (id: number | null) => void;
  diagnosis2CategoryId?: number | null;
  setDiagnosis2CategoryId?: (id: number | null) => void;
  diagnosis2NameId?: number | null;
  setDiagnosis2NameId?: (id: number | null) => void;
}

export const DiagnosisHeader = React.memo(function DiagnosisHeader({
  chiefComplaint,
  policy,
  setPolicy,
  diagnosisDetails,
  setDiagnosisDetails,
  diagnosis1CategoryId,
  setDiagnosis1CategoryId,
  diagnosis1NameId,
  setDiagnosis1NameId,
  diagnosis2CategoryId,
  setDiagnosis2CategoryId,
  diagnosis2NameId,
  setDiagnosis2NameId,
}: DiagnosisHeaderProps) {
  const { canEdit } = usePermission("medical-records");
  return (
    <div className="grid grid-cols-12 gap-3 shrink-0 h-[240px]">
      <DiagnosisHeaderChiefComplaint content={chiefComplaint} />
      <DiagnosisHeaderPhysicalExam policy={policy} setPolicy={setPolicy} canEdit={canEdit} />
      <DiagnosisHeaderDiagnosis
        diagnosisDetails={diagnosisDetails}
        setDiagnosisDetails={setDiagnosisDetails}
        diagnosis1CategoryId={diagnosis1CategoryId}
        setDiagnosis1CategoryId={setDiagnosis1CategoryId}
        diagnosis1NameId={diagnosis1NameId}
        setDiagnosis1NameId={setDiagnosis1NameId}
        diagnosis2CategoryId={diagnosis2CategoryId}
        setDiagnosis2CategoryId={setDiagnosis2CategoryId}
        diagnosis2NameId={diagnosis2NameId}
        setDiagnosis2NameId={setDiagnosis2NameId}
        canEdit={canEdit}
      />
    </div>
  );
});
