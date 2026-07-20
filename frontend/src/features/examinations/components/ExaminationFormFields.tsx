import { memo, useMemo } from "react";
import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { MasterLink } from "@/components/shared/MasterLink";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { jstDateStartISOString, todayJSTISO, toJSTWallDate } from "@/lib/jst-date";
import type { ExaminationRecord } from "../api/transforms";

const EXAM_STATUS_ITEMS = (
  <>
    <SelectItem value="依頼中">依頼中</SelectItem>
    <SelectItem value="検査中">検査中</SelectItem>
    <SelectItem value="結果入力済み">結果入力済み</SelectItem>
    <SelectItem value="完了">完了</SelectItem>
    <SelectItem value="確定">確定</SelectItem>
  </>
);

interface ExaminationFormFieldsProps {
  formData: Partial<ExaminationRecord>;
  examTypes: { id: string; name: string }[];
  staffList: { id: string; name: string }[];
  masterLoading: boolean;
  isEdit: boolean;
  isDeleting: boolean;
  isConfirmed: boolean;
  canEdit: boolean;
  canCreate: boolean;
  canDelete: boolean;
  onSetFormData: (next: Partial<ExaminationRecord>) => void;
  onBack: () => void;
  onDeleteClick: () => void;
}

function ExaminationFormFieldsBase({
  formData,
  examTypes,
  staffList,
  masterLoading,
  isEdit,
  isDeleting,
  isConfirmed,
  canEdit,
  canCreate,
  canDelete,
  onSetFormData,
  onBack,
  onDeleteClick,
}: ExaminationFormFieldsProps) {
  const canSubmit = isEdit ? canEdit : canCreate;

  const examTypeOptions = useMemo<SearchableSelectOption[]>(
    () => examTypes.map((item) => ({ value: String(item.id), label: item.name })),
    [examTypes],
  );

  const staffOptions = useMemo<SearchableSelectOption[]>(
    () => staffList.map((staff) => ({ value: String(staff.id), label: staff.name })),
    [staffList],
  );

  return (
    <div className={`${C.bgWhite} p-4 rounded-lg border ${C.borderMedium} space-y-4`}>
      {isConfirmed ? (
        <p className={`text-sm font-medium ${C.text60}`}>確定済みのため編集できません。</p>
      ) : null}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label className={`text-sm ${C.text60}`}>検査種別</Label>
            <MasterLink category="examination" label="編集" className="text-2xs" />
          </div>
          {masterLoading ? (
            <div className={`h-10 ${C.bgGray100} rounded-md animate-pulse`} />
          ) : (
            <SearchableSelect
              value={formData.testTypeId ?? ""}
              disabled={isConfirmed}
              onValueChange={(value) => {
                const item = examTypes.find((examType) => examType.id === value);
                onSetFormData({ testTypeId: value, testType: item?.name ?? value });
              }}
              options={examTypeOptions}
              placeholder="選択してください"
              searchPlaceholder="検査種別を検索..."
              id="testTypeId"
            />
          )}
        </div>
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label className={`text-sm ${C.text60}`}>担当医</Label>
            <MasterLink category="staff" label="編集" className="text-2xs" />
          </div>
          {masterLoading ? (
            <div className={`h-10 ${C.bgGray100} rounded-md animate-pulse`} />
          ) : (
            <SearchableSelect
              value={formData.doctorId ?? ""}
              disabled={isConfirmed}
              onValueChange={(value) => {
                const staff = staffList.find((item) => String(item.id) === value);
                onSetFormData({ doctorId: value, doctor: staff?.name ?? value });
              }}
              options={staffOptions}
              placeholder="選択してください"
              searchPlaceholder="担当医を検索..."
              id="doctorId"
            />
          )}
        </div>
      </div>

      <div className="space-y-1.5">
        <Label className={`text-sm ${C.text60}`}>検査日</Label>
        <DatePicker
          value={formData.date ? formData.date.split("T")[0] : ""}
          onChange={(value) => onSetFormData({ date: value ? jstDateStartISOString(value) : jstDateStartISOString(todayJSTISO()) })}
          disabledDays={{ after: toJSTWallDate(new Date()) }}
        />
      </div>

      <div className="space-y-1.5">
        <Label className={`text-sm ${C.text60}`}>ステータス</Label>
        <Select
          value={formData.status}
          disabled={isConfirmed}
          onValueChange={(value: "依頼中" | "検査中" | "結果入力済み" | "完了" | "確定") => onSetFormData({ status: value })}
        >
          <SelectTrigger className={`h-10 text-sm ${C.text} ${C.bgWhite} ${C.borderMedium}`}>
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>
            {EXAM_STATUS_ITEMS}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1.5">
        <Label className={`text-sm ${C.text60}`}>備考・所見</Label>
        <Textarea
          className={`h-24 text-sm ${C.text} ${C.bgWhite} ${C.borderMedium} resize-none`}
          placeholder="検査結果や備考を入力"
          value={formData.resultSummary || ""}
          disabled={isConfirmed}
          onChange={(event) => onSetFormData({ resultSummary: event.target.value })}
        />
      </div>

      {isConfirmed ? null : (
        <div className="flex justify-end gap-2 pt-2">
          {canDelete && isEdit ? (
            <Button
              variant="ghost"
              type="button"
              className={`h-10 text-sm ${STYLE.btnDangerGhost} mr-auto`}
              onClick={onDeleteClick}
              disabled={isDeleting}
            >
              <Trash2 className={`mr-1.5 ${ICON.action}`} />
              {isDeleting ? "削除中..." : "削除"}
            </Button>
          ) : null}
          <Button variant="outline" type="button" onClick={onBack} className="h-10 text-sm">キャンセル</Button>
          {canSubmit ? (
            <SubmitButton className="h-10 text-sm">
              保存
            </SubmitButton>
          ) : null}
        </div>
      )}
    </div>
  );
}

export const ExaminationFormFields = memo(ExaminationFormFieldsBase);
