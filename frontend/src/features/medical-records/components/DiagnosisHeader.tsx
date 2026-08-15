// React/Framework
import { memo } from "react";

// Internal
import { usePermission } from "@/hooks/use-permission";

// Relative
import { DiagnosisHeaderChiefComplaint } from "./DiagnosisHeaderChiefComplaint";
import { DiagnosisHeaderPhysicalExam } from "./DiagnosisHeaderPhysicalExam";
import { DiagnosisHeaderDiagnosis } from "./DiagnosisHeaderDiagnosis";

interface DiagnosisHeaderProps {
  chiefComplaint?: string;
  physicalExam: string;
  setPhysicalExam: (v: string) => void;
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
  diagnosis1NameIdError?: string | null;
}

export const DiagnosisHeader = memo(function DiagnosisHeader({
  chiefComplaint,
  physicalExam,
  setPhysicalExam,
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
  diagnosis1NameIdError,
}: DiagnosisHeaderProps) {
  const { canEdit } = usePermission("medical-records");
  return (
    <div className="grid grid-cols-12 grid-rows-[auto_auto_minmax(0,1fr)] gap-x-3 gap-y-2 shrink-0 h-[300px]">
      <DiagnosisHeaderChiefComplaint content={chiefComplaint} />
      <DiagnosisHeaderPhysicalExam
        physicalExam={physicalExam}
        setPhysicalExam={setPhysicalExam}
        canEdit={canEdit}
      />
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
        diagnosis1NameIdError={diagnosis1NameIdError}
      />
    </div>
  );
});
