// React/Framework
import { memo, useDeferredValue, useMemo, useState } from "react";

// Relative
import { useGetRecordExaminations } from "../api/get-record-examinations";
import { ExaminationFilter } from "./ExaminationFilter";
import { ExaminationGroup } from "./ExaminationGroup";
import { C } from "@/lib/design-tokens";

interface MedicalRecordExaminationProps {
  isNewRecord?: boolean;
  petId?: string;
}

export const MedicalRecordExamination = memo(function MedicalRecordExamination({
  isNewRecord = false,
  petId,
}: MedicalRecordExaminationProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");

  const { data: apiExamGroups = [], isLoading } = useGetRecordExaminations(
    isNewRecord ? undefined : petId,
  );

  const examGroups = useMemo(
    () =>
      apiExamGroups.filter((g) =>
        deferredSearch
          ? g.items.some((item) => item.name.includes(deferredSearch))
          : true,
      ),
    [apiExamGroups, deferredSearch],
  );

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-20 pr-1">
      {/* Search & Actions Header */}
      <ExaminationFilter
        searchTerm={deferredSearch}
        onSearchChange={setSearchTerm}
        dateStart={dateStart}
        onDateStartChange={setDateStart}
        dateEnd={dateEnd}
        onDateEndChange={setDateEnd}
      />

      {/* Results Title */}
      <div>
        <h2 className={`text-sm font-bold ${C.text} pl-1`}>検査結果一覧</h2>
      </div>

      {/* Exam Groups */}
      {isLoading ? (
        <div className={`flex items-center justify-center h-24 text-sm ${C.text40} pl-1`}>
          読み込み中...
        </div>
      ) : examGroups.length === 0 ? (
        <div className={`flex items-center justify-center h-24 text-sm ${C.text40} pl-1`}>
          検査記録がありません
        </div>
      ) : null}
      <div className="flex flex-col gap-4 pl-1">
        {!isLoading
          ? examGroups.map((group) => (
              <ExaminationGroup key={group.id} group={group} />
            ))
          : null}
      </div>
    </div>
  );
});
