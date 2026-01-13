import React, { useState } from "react";
import { ExaminationFilter } from "./ExaminationFilter";
import { ExaminationGroup } from "./ExaminationGroup";

export function MedicalRecordExamination({ isNewRecord = false }: { isNewRecord?: boolean }) {
  const [searchTerm, setSearchTerm] = useState("");
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");

  // Mock data for Examination results
  const examGroups = isNewRecord ? [] : [
    {
      id: 1,
      date: "2025/10/10 10:15",
      machine: "血液検査機器 A",
      items: [
        { id: 1, name: "WBC", result: "15.2", unit: "10^3/µL", ref: "6.0 - 17.0", status: "normal" },
        { id: 2, name: "RBC", result: "6.80", unit: "10^6/µL", ref: "5.50 - 8.50", status: "normal" },
        { id: 3, name: "HGB", result: "14.5", unit: "g/dL", ref: "12.0 - 18.0", status: "normal" },
        { id: 4, name: "PLT", result: "150", unit: "10^3/µL", ref: "200 - 500", status: "low" },
        { id: 5, name: "CRP", result: "3.5", unit: "mg/dL", ref: "< 1.0", status: "high" },
      ],
    },
    {
      id: 2,
      date: "2025/10/10 10:30",
      machine: "生化学検査機器 B",
      items: [
        { id: 6, name: "BUN", result: "25", unit: "mg/dL", ref: "7 - 27", status: "normal" },
        { id: 7, name: "CRE", result: "1.8", unit: "mg/dL", ref: "0.5 - 1.8", status: "normal" },
        { id: 8, name: "ALT", result: "120", unit: "U/L", ref: "10 - 100", status: "high" },
        { id: 9, name: "ALP", result: "45", unit: "U/L", ref: "20 - 150", status: "normal" },
      ],
    },
  ];

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-20 pr-1">
      {/* Search & Actions Header */}
      <ExaminationFilter
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        dateStart={dateStart}
        onDateStartChange={setDateStart}
        dateEnd={dateEnd}
        onDateEndChange={setDateEnd}
      />

      {/* Results Title */}
      <div>
         <h2 className="text-sm font-bold text-[#37352F] pl-1">検査結果一覧</h2>
      </div>

      {/* Exam Groups */}
      <div className="flex flex-col gap-4 pl-1">
        {examGroups.map((group) => (
          <ExaminationGroup key={group.id} group={group} />
        ))}
      </div>
    </div>
  );
}
