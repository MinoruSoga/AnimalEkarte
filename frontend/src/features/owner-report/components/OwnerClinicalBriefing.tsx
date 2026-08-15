import type { Owner, Pet } from "@/types";

import { useOwnerClinicalBriefingData } from "../hooks/use-owner-clinical-briefing-data";
import { BasicInformationPanel } from "./OwnerClinicalBasicPanel";
import { ClinicalHistoryPanel } from "./OwnerClinicalHistoryPanel";
import {
  NextActionPanel,
  PreviousVisitPanel,
  PreVisitPanel,
  TodayVisitPanel,
} from "./OwnerClinicalVisitPanels";

import "../owner-report.css";

interface OwnerClinicalBriefingProps {
  owner: Owner;
  pet: Pet;
  firstVisitDate?: string | null;
  firstVisitLoading: boolean;
  firstVisitError: boolean;
}

/** 1画面で「今見る情報→次にすること→種類別履歴」を辿れる診療ブリーフィング。 */
export function OwnerClinicalBriefing(props: OwnerClinicalBriefingProps) {
  const data = useOwnerClinicalBriefingData(props.pet.id);

  return (
    <main
      className="min-h-0 flex-1 overflow-hidden p-1"
      aria-label="飼主ペット診療ブリーフィング"
    >
      <div className="owner-report-dashboard">
        <PreVisitPanel data={data} pet={props.pet} />
        <TodayVisitPanel data={data} />
        <NextActionPanel data={data} />
        <PreviousVisitPanel data={data} />
        <BasicInformationPanel
          owner={props.owner}
          pet={props.pet}
          firstVisitDate={props.firstVisitDate}
          firstVisitLoading={props.firstVisitLoading}
          firstVisitError={props.firstVisitError}
        />
        <ClinicalHistoryPanel data={data} />
      </div>
    </main>
  );
}
