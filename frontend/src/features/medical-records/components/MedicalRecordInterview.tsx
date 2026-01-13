import React, { useState, useCallback } from "react";
import { InterviewChiefComplaint } from "./InterviewChiefComplaint";
import { InterviewTreatmentPolicy } from "./InterviewTreatmentPolicy";
import { InterviewHistory } from "./InterviewHistory";

export function MedicalRecordInterview() {
  const [chiefComplaint, setChiefComplaint] = useState(
    "# どんな症状\n\n# どこが\n\n# いつから\n\n# その他・備考\n\n# フリースペース"
  );
  const [treatmentPolicy, setTreatmentPolicy] = useState("# 治療方針");

  const templates = [
    { label: "定期検診", text: "# 定期検診\n特に異常なし。食欲・元気あり。" },
    { label: "ワクチン", text: "# 混合ワクチン接種\n体調良好。" },
    { label: "下痢・嘔吐", text: "# 消化器症状\n・嘔吐：あり（回数：　）\n・下痢：あり（性状：　）\n・食欲：なし" },
    { label: "皮膚", text: "# 皮膚症状\n・痒み：あり\n・発赤：あり\n・部位：" },
  ];

  const handleInsertTemplate = useCallback((text: string) => {
    setChiefComplaint(text);
  }, []);

  const historyItems = [
    {
      id: 1,
      date: "2022/10/10 (月)",
      author: "医者A",
      type: "再診",
      title: "消化器症状",
      content: "嘔吐2回、下痢なし。食欲低下。",
    },
    {
      id: 2,
      date: "2022/10/09 (日)",
      author: "医者B",
      type: "初診",
      title: "定期検診",
      content: "異常なし。体重3.5kg",
    },
    {
      id: 3,
      date: "2022/09/15 (木)",
      author: "医者A",
      type: "ワクチン",
      title: "混合ワクチン",
      content: "5種混合ワクチン接種。副反応なし。",
    },
  ];

  return (
    <div className="grid grid-cols-12 gap-3 h-[calc(100vh-220px)] min-h-[500px]">
      {/* Left Column: 主訴情報 (Chief Complaint) */}
      <div className="col-span-3 flex flex-col gap-3 min-h-0">
        <InterviewChiefComplaint
          chiefComplaint={chiefComplaint}
          setChiefComplaint={setChiefComplaint}
          templates={templates}
          onInsertTemplate={handleInsertTemplate}
        />
      </div>

      {/* Middle Column: 治療方針 (Treatment Policy) */}
      <div className="col-span-4 flex flex-col gap-3 min-h-0">
        <InterviewTreatmentPolicy
          treatmentPolicy={treatmentPolicy}
          setTreatmentPolicy={setTreatmentPolicy}
        />
      </div>

      {/* Right Column: カルテ履歴 (Medical History) */}
      <div className="col-span-5 flex flex-col gap-3 min-h-0">
        <InterviewHistory historyItems={historyItems} />
      </div>
    </div>
  );
}
