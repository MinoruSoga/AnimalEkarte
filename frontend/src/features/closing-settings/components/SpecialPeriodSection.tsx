import { memo, useActionState, useCallback, useLayoutEffect, useRef, useState } from "react";
import { Calendar, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { EmptyState } from "@/components/shared/DataStates";
import { MasterSidePanel, PropertyRow } from "@/components/shared/SidePeek";
import { getFormString } from "@/lib/form-data";
import type { ClosingSpecialPeriod } from "@/types/generated/models";
import { useCreateSpecialPeriod, useDeleteSpecialPeriod } from "../api/special-periods";
import {
  computeClosingTimeRanges,
  DEFAULT_CLOSING_AM_START,
  formatRangeText,
} from "../lib/closing-time-ranges";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

// 導出時間帯の表示行（key は ClosingTimeRanges のキーと一致）。
const TIME_RANGE_ROWS = [
  { key: "am", label: "AM" },
  { key: "pm", label: "PM" },
  { key: "emg", label: "EMG" },
] as const;

interface SpecialPeriodSectionProps {
  periods: ClosingSpecialPeriod[];
  canCreate: boolean;
  canDelete: boolean;
  /** 標準締め設定の closing_am_start（特別期間は標準設定の am_start を継承する）。省略時は 09:00。 */
  amStart?: string;
}

export const SpecialPeriodSection = memo(function SpecialPeriodSection({
  periods,
  canCreate,
  canDelete,
  amStart = DEFAULT_CLOSING_AM_START,
}: SpecialPeriodSectionProps) {
  const [showForm, setShowForm] = useState(false);
  const [note, setNote] = useState("");
  // 時間帯プレビュー用に区切り・終了時刻を制御値として保持する。
  // 送信は form action が FormData(DOM 値) から読むため、このプレビュー state とは独立。
  const [amPmBoundary, setAmPmBoundary] = useState("");
  const [pmEnd, setPmEnd] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<ClosingSpecialPeriod | null>(null);
  const createMutation = useCreateSpecialPeriod();
  const deleteMutation = useDeleteSpecialPeriod();
  const { mutateAsync } = deleteMutation;
  // backend は POST→create / DELETE→delete を要求するため、各アクションのガードも対応する権限で判定する。
  const canCreateRef = useRef(canCreate);
  const canDeleteRef = useRef(canDelete);
  useLayoutEffect(() => {
    canCreateRef.current = canCreate;
    canDeleteRef.current = canDelete;
  }, [canCreate, canDelete]);

  const handleShowForm = useCallback(() => setShowForm(true), []);
  const handleHideForm = useCallback(() => {
    setShowForm(false);
    setNote("");
    setAmPmBoundary("");
    setPmEnd("");
  }, []);

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    if (canCreateRef.current !== true) {
      toast.error(PERMISSION_DENIED_MESSAGE);
      return null;
    }
    try {
      await createMutation.mutateAsync({
        start_date: getFormString(formData, "start_date"),
        end_date: getFormString(formData, "end_date"),
        am_pm_boundary: getFormString(formData, "am_pm_boundary"),
        pm_end: getFormString(formData, "pm_end"),
        note: getFormString(formData, "note") || undefined,
      });
      toast.success("特別期間を追加しました");
      handleHideForm();
    } catch {
      // FE-RC-005: useCreateSpecialPeriod.onError が既に handleApiError で通知済み。
    }
    return null;
  }, null);

  const handleDelete = useCallback(
    async (id: number) => {
      if (canDeleteRef.current !== true) {
        toast.error(PERMISSION_DENIED_MESSAGE);
        return;
      }
      try {
        await mutateAsync(id);
        toast.success("特別期間を削除しました");
      } catch {
        // FE-RC-005: useDeleteSpecialPeriod.onError が既に handleApiError で通知済み。
      }
    },
    [mutateAsync],
  );

  const handleConfirmDelete = useCallback(() => {
    if (deleteTarget != null) {
      void handleDelete(deleteTarget.id);
    }
    setDeleteTarget(null);
  }, [deleteTarget, handleDelete]);
  const handleCloseDeleteDialog = useCallback(() => setDeleteTarget(null), []);

  // 派生値は描画時に計算する（useEffect で同期しない）。特別期間は標準設定の am_start を継承する。
  const previewRanges = computeClosingTimeRanges(amPmBoundary, pmEnd, amStart);

  return (
    <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
      <div className="flex items-center justify-between mb-4">
        <h2 className={`text-base font-semibold ${C.text}`}>特別期間</h2>
        {canCreate ? (
          <button
            type="button"
            onClick={handleShowForm}
            className={`flex min-h-11 min-w-11 items-center gap-1.5 text-base ${C.textActionPrimary} ${C.hoverBgActionPrimary} ${C.hoverTextOnActionPrimary} rounded-xs px-3 transition-colors`}
          >
            <Plus className="size-4" />
            新規登録
          </button>
        ) : null}
      </div>

      {showForm ? (
        <MasterSidePanel
          isNew
          title={note}
          onTitleChange={setNote}
          titlePlaceholder="メモ（例: 年末年始）"
          titleDescription="この特別期間のメモ・名称です。年末年始や夏季休診など、期間の識別に使われます。"
          onClose={handleHideForm}
          action={formAction}
          icon={<Calendar className={LAYOUT.pageIcon.innerIcon} />}
        >
          <input type="hidden" name="note" value={note} />
          <PropertyRow label="開始日" description="特別期間が始まる日付です。">
            <input
              id="start_date"
              name="start_date"
              type="date"
              aria-label="開始日"
              className={`${STYLE.formInput} w-full rounded-xs border px-3`}
              required
            />
          </PropertyRow>
          <PropertyRow label="終了日" description="特別期間が終わる日付です。">
            <input
              id="end_date"
              name="end_date"
              type="date"
              aria-label="終了日"
              className={`${STYLE.formInput} w-full rounded-xs border px-3`}
              required
            />
          </PropertyRow>
          <PropertyRow
            label="午前・午後 区切り時間"
            description="午前（AM）と午後（PM）の区切りとなる時刻です。締め時間帯の計算に使われます。"
          >
            <input
              id="am_pm_boundary"
              name="am_pm_boundary"
              type="time"
              aria-label="午前・午後 区切り時間"
              value={amPmBoundary}
              onChange={(event) => setAmPmBoundary(event.target.value)}
              className={`${STYLE.formInput} w-full rounded-xs border px-3`}
              required
            />
          </PropertyRow>
          <PropertyRow
            label="午後 終了時間"
            description="午後（PM）時間帯の終了時刻です。これ以降はEMG（時間外）として扱われます。"
          >
            <input
              id="pm_end"
              name="pm_end"
              type="time"
              aria-label="午後 終了時間"
              value={pmEnd}
              onChange={(event) => setPmEnd(event.target.value)}
              className={`${STYLE.formInput} w-full rounded-xs border px-3`}
              required
            />
          </PropertyRow>
          <PropertyRow
            label="時間帯プレビュー"
            description="入力した区切り・終了時刻から自動計算されるAM/PM/EMGの時間帯です。確認用の表示で直接は編集できません。"
          >
            <div className="space-y-0.5">
              {TIME_RANGE_ROWS.map((row) => (
                <p key={row.key} className={`text-sm tabular-nums ${C.text60}`}>
                  {row.label} {formatRangeText(previewRanges[row.key])}
                </p>
              ))}
            </div>
          </PropertyRow>
        </MasterSidePanel>
      ) : null}

      {periods.length > 0 ? (
        <div className="space-y-2">
          {periods.map((period) => {
            const ranges = computeClosingTimeRanges(period.am_pm_boundary, period.pm_end, amStart);
            return (
              <div
                key={period.id}
                className={`flex items-center justify-between p-3 rounded-lg border ${C.borderLight} ${C.bgPage}`}
              >
                <div className="flex flex-col gap-0.5">
                  <span className={`text-base font-medium ${C.text}`}>
                    {period.start_date} 〜 {period.end_date}
                  </span>
                  <span className={`text-base ${C.text60}`}>
                    区切り: {period.am_pm_boundary} / 終了: {period.pm_end}
                    {period.note ? ` — ${period.note}` : ""}
                  </span>
                  <span className={`text-sm tabular-nums ${C.text60}`}>
                    AM {formatRangeText(ranges.am)} / PM {formatRangeText(ranges.pm)} / EMG{" "}
                    {formatRangeText(ranges.emg)}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() => setDeleteTarget(period)}
                  aria-label={`${period.start_date}から${period.end_date}の特別期間を削除`}
                  className={`flex size-11 min-h-11 min-w-11 shrink-0 items-center justify-center rounded-xxs ${C.text50} ${C.hoverTextDanger} ${C.hoverBgDanger5} transition-colors`}
                >
                  <Trash2 className="size-4" />
                </button>
              </div>
            );
          })}
        </div>
      ) : (
        <EmptyState message="特別期間は登録されていません" />
      )}

      <ConfirmDialog
        open={deleteTarget != null}
        onClose={handleCloseDeleteDialog}
        onConfirm={handleConfirmDelete}
        title="特別期間を削除しますか？"
        description={
          deleteTarget != null
            ? `${deleteTarget.start_date} 〜 ${deleteTarget.end_date} の特別期間を削除します。この操作は取り消せません。`
            : undefined
        }
        confirmLabel="削除"
        variant="destructive"
      />
    </section>
  );
});
