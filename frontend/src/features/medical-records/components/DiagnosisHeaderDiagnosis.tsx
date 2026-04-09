// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo } from "react";

// External
import { ChevronRight } from "lucide-react";

// Internal
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

// Relative
import { useGetDiagnosisCategories, useGetDiagnosisNames } from "../api/get-diagnosis-options";

interface DiagnosisHeaderDiagnosisProps {
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
  canEdit: boolean;
}

export const DiagnosisHeaderDiagnosis = memo(function DiagnosisHeaderDiagnosis({
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
  canEdit,
}: DiagnosisHeaderDiagnosisProps) {
  const { data: categories = [], isLoading: isCategoriesLoading } = useGetDiagnosisCategories();
  const { data: names1 = [], isLoading: isNames1Loading } = useGetDiagnosisNames(diagnosis1CategoryId);
  const { data: names2 = [], isLoading: isNames2Loading } = useGetDiagnosisNames(diagnosis2CategoryId);

  return (
    <div className="col-span-5 flex flex-col min-h-0">
      <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
        <CardHeader className="p-0 pb-2">
          <CardTitle className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
            <ChevronRight className={ICON.action} />
            診断
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0 flex-1 flex flex-col gap-2 min-h-0">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label className={`w-10 shrink-0 text-sm font-medium ${C.text60}`}>
                診断1
              </Label>
              <Select
                value={diagnosis1CategoryId ? String(diagnosis1CategoryId) : ""}
                onValueChange={(value) => {
                  setDiagnosis1CategoryId?.(value ? Number(value) : null);
                  setDiagnosis1NameId?.(null);
                }}
                disabled={isCategoriesLoading || !canEdit}
              >
                <SelectTrigger className={`flex-1 bg-white ${C.borderMedium} h-10 text-sm`}>
                  <SelectValue placeholder={isCategoriesLoading ? "読み込み中..." : "カテゴリを選択"} />
                </SelectTrigger>
                <SelectContent className="z-[9999]">
                  {categories.map((cat) => (
                    <SelectItem key={cat.id} value={String(cat.id)}>
                      {cat.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select
                value={diagnosis1NameId ? String(diagnosis1NameId) : ""}
                onValueChange={(value) => setDiagnosis1NameId?.(value ? Number(value) : null)}
                disabled={isNames1Loading || !diagnosis1CategoryId || !canEdit}
              >
                <SelectTrigger className={`flex-1 bg-white ${C.borderMedium} h-10 text-sm`}>
                  <SelectValue placeholder={isNames1Loading ? "読み込み中..." : "病名を選択"} />
                </SelectTrigger>
                <SelectContent className="z-[9999]">
                  {names1.map((name) => (
                    <SelectItem key={name.id} value={String(name.id)}>
                      {name.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center gap-2">
              <Label className={`w-10 shrink-0 text-sm font-medium ${C.text60}`}>
                診断2
              </Label>
              <Select
                value={diagnosis2CategoryId ? String(diagnosis2CategoryId) : ""}
                onValueChange={(value) => {
                  setDiagnosis2CategoryId?.(value ? Number(value) : null);
                  setDiagnosis2NameId?.(null);
                }}
                disabled={isCategoriesLoading || !canEdit}
              >
                <SelectTrigger className={`flex-1 bg-white ${C.borderMedium} h-10 text-sm`}>
                  <SelectValue placeholder={isCategoriesLoading ? "読み込み中..." : "カテゴリを選択"} />
                </SelectTrigger>
                <SelectContent className="z-[9999]">
                  {categories.map((cat) => (
                    <SelectItem key={cat.id} value={String(cat.id)}>
                      {cat.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select
                value={diagnosis2NameId ? String(diagnosis2NameId) : ""}
                onValueChange={(value) => setDiagnosis2NameId?.(value ? Number(value) : null)}
                disabled={isNames2Loading || !diagnosis2CategoryId || !canEdit}
              >
                <SelectTrigger className={`flex-1 bg-white ${C.borderMedium} h-10 text-sm`}>
                  <SelectValue placeholder={isNames2Loading ? "読み込み中..." : "病名を選択"} />
                </SelectTrigger>
                <SelectContent className="z-[9999]">
                  {names2.map((name) => (
                    <SelectItem key={name.id} value={String(name.id)}>
                      {name.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex-1 flex flex-col min-h-0">
            <Label className={`text-sm ${C.text60} mb-1.5`}>
              診断詳細
            </Label>
            <Textarea
              value={diagnosisDetails}
              onChange={(e) => setDiagnosisDetails(e.target.value)}
              className={`flex-1 resize-none bg-white ${C.borderMedium} text-sm p-3 font-mono ${C.focusRingMedicalBlue}`}
              disabled={!canEdit}
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
});
