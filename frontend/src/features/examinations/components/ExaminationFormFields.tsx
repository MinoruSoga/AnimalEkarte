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
import { FormFieldError } from "@/components/shared/FormFieldError";
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
  /** Server-persisted confirmed — full lock (S02). */
  isConfirmed: boolean;
  /** BUG-033: first-pass completed seal — results/delete lock; status transition save allowed. */
  isCompletedLocked?: boolean;
  canEdit: boolean;
  canCreate: boolean;
  canDelete: boolean;
  fieldErrors?: Record<string, string>;
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
  isCompletedLocked = false,
  canEdit,
  canCreate,
  canDelete,
  fieldErrors,
  onSetFormData,
  onBack,
  onDeleteClick,
}: ExaminationFormFieldsProps) {
  const canSubmit = isEdit ? canEdit : canCreate && canEdit;
  const fieldsLocked = isConfirmed || isCompletedLocked;
  // completed seal: hide save/delete while status remains 完了; allow save after status change (confirm/unlock).
  const showActions =
    !isConfirmed && !(isCompletedLocked && formData.status === "完了");
  const testTypeError = fieldErrors?.testTypeId;
  const doctorError = fieldErrors?.doctorId;

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
      {isCompletedLocked && !isConfirmed ? (
        <p className={`text-sm font-medium ${C.text60}`}>
          完了済みのため結果の編集・削除はできません。確定する場合はステータスを「確定」に変更して保存してください。
        </p>
      ) : null}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label htmlFor="testTypeId" className={`text-sm ${C.text60}`}>検査種別</Label>
            <MasterLink category="examination" label="編集" className="text-2xs" />
          </div>
          {masterLoading ? (
            <div className={`h-10 ${C.bgGray100} rounded-md animate-pulse`} />
          ) : (
            <SearchableSelect
              value={formData.testTypeId ?? ""}
              disabled={fieldsLocked}
              onValueChange={(value) => {
                const item = examTypes.find((examType) => examType.id === value);
                onSetFormData({ testTypeId: value, testType: item?.name ?? value });
              }}
              options={examTypeOptions}
              placeholder="選択してください"
              searchPlaceholder="検査種別を検索..."
              id="testTypeId"
              ariaInvalid={Boolean(testTypeError)}
              ariaDescribedBy={testTypeError ? "testTypeId-error" : undefined}
            />
          )}
          <FormFieldError id="testTypeId-error" message={testTypeError} />
        </div>
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label htmlFor="doctorId" className={`text-sm ${C.text60}`}>担当医</Label>
            <MasterLink category="staff" label="編集" className="text-2xs" />
          </div>
          {masterLoading ? (
            <div className={`h-10 ${C.bgGray100} rounded-md animate-pulse`} />
          ) : (
            <SearchableSelect
              value={formData.doctorId ?? ""}
              disabled={fieldsLocked}
              onValueChange={(value) => {
                const staff = staffList.find((item) => String(item.id) === value);
                onSetFormData({ doctorId: value, doctor: staff?.name ?? value });
              }}
              options={staffOptions}
              placeholder="選択してください"
              searchPlaceholder="担当医を検索..."
              id="doctorId"
              ariaInvalid={Boolean(doctorError)}
              ariaDescribedBy={doctorError ? "doctorId-error" : undefined}
            />
          )}
          <FormFieldError id="doctorId-error" message={doctorError} />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="examination-date" className={`text-sm ${C.text60}`}>検査日</Label>
        <DatePicker
          id="examination-date"
          value={formData.date ? formData.date.split("T")[0] : ""}
          onChange={(value) => {
            if (fieldsLocked) return;
            onSetFormData({ date: value ? jstDateStartISOString(value) : jstDateStartISOString(todayJSTISO()) });
          }}
          disabledDays={fieldsLocked ? () => true : { after: toJSTWallDate(new Date()) }}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="examination-status" className={`text-sm ${C.text60}`}>ステータス</Label>
        <Select
          value={formData.status}
          disabled={isConfirmed}
          onValueChange={(value: "依頼中" | "検査中" | "結果入力済み" | "完了" | "確定") => onSetFormData({ status: value })}
        >
          <SelectTrigger id="examination-status" className={`h-11 min-w-11 text-sm ${C.text} ${C.bgWhite} ${C.borderMedium}`}>
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>
            {EXAM_STATUS_ITEMS}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="examination-notes" className={`text-sm ${C.text60}`}>備考・所見</Label>
        <Textarea
          id="examination-notes"
          className={`h-24 text-sm ${C.text} ${C.bgWhite} ${C.borderMedium} resize-none`}
          placeholder="検査結果や備考を入力"
          value={formData.resultSummary || ""}
          disabled={fieldsLocked}
          onChange={(event) => onSetFormData({ resultSummary: event.target.value })}
        />
      </div>

      {showActions ? (
        <div className="flex justify-end gap-2 pt-2">
          {canDelete && isEdit && !isCompletedLocked ? (
            <Button
              variant="ghost"
              type="button"
              className={`text-sm ${STYLE.btnDangerGhost} mr-auto`}
              onClick={onDeleteClick}
              disabled={isDeleting}
            >
              <Trash2 className={`mr-1.5 ${ICON.action}`} />
              {isDeleting ? "削除中..." : "削除"}
            </Button>
          ) : null}
          <Button variant="outline" type="button" onClick={onBack} className="text-sm">キャンセル</Button>
          {canSubmit ? (
            <SubmitButton className="text-sm">
              保存
            </SubmitButton>
          ) : null}
        </div>
      ) : isCompletedLocked ? (
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="outline" type="button" onClick={onBack} className="text-sm">戻る</Button>
        </div>
      ) : null}
    </div>
  );
}

export const ExaminationFormFields = memo(ExaminationFormFieldsBase);
