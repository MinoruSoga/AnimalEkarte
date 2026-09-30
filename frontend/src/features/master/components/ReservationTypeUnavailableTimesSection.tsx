import { useActionState, useCallback, useMemo, useState } from "react";
import { Plus, Clock } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { FieldHelp } from "@/components/shared/FieldHelp";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { DAY_OF_WEEK_LABELS } from "@/constants/day-of-week";
import {
  useGetUnavailableTimes,
  useCreateUnavailableTime,
  useDeleteUnavailableTime,
} from "../api/reservation-type-unavailable-times";
import { UnavailableTypeWeekly, UnavailableTypeSpecific } from "@/types/generated/models";
import { DAY_OF_WEEK_SELECT_ITEMS } from "./DayOfWeekSelectItems";
import type { CreateUnavailableTimeRequest } from "../api/reservation-type-unavailable-times";
import type { ReservationTypeUnavailableTime } from "@/hooks/use-reservation-type-unavailable-times";

// ─────────────────────────────────────────────────
// 静的定数（rendering-hoist-jsx）
// ─────────────────────────────────────────────────

/** 30分刻みの時刻選択肢 "HH:MM" */
const TIME_OPTIONS: string[] = [];
for (let h = 0; h < 24; h++) {
  for (const m of [0, 30]) {
    TIME_OPTIONS.push(`${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`);
  }
}

const TIME_SELECT_ITEMS = TIME_OPTIONS.map((t) => (
  <SelectItem key={t} value={t}>
    {t}
  </SelectItem>
));

// ─────────────────────────────────────────────────
// Default form state
// ─────────────────────────────────────────────────

interface FormState {
  unavailableType: string;
  dayOfWeek: string;
  specificDate: string;
  startTime: string;
  endTime: string;
}

const DEFAULT_FORM: FormState = {
  unavailableType: UnavailableTypeWeekly,
  dayOfWeek: "1",
  specificDate: "",
  startTime: "09:00",
  endTime: "18:00",
};

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

function unavailableTimeLabel(item: ReservationTypeUnavailableTime): string {
  return item.unavailableType === UnavailableTypeWeekly
    ? `毎週${DAY_OF_WEEK_LABELS[item.dayOfWeek ?? 0]}曜日`
    : (item.specificDate ?? "");
}

interface Props {
  clinicId: string;
  reservationTypeId: string;
  /** 参照権限のみのパネル表示。追加フォーム・削除ボタン等の mutation UI を描画しない */
  readOnly?: boolean;
}

export function ReservationTypeUnavailableTimesSection({
  clinicId,
  reservationTypeId,
  readOnly = false,
}: Props) {
  const { data: items = [], isLoading } = useGetUnavailableTimes(clinicId, reservationTypeId);
  const createMutation = useCreateUnavailableTime(clinicId, reservationTypeId);
  const deleteMutation = useDeleteUnavailableTime(clinicId, reservationTypeId);
  const { mutate } = deleteMutation;

  const [form, setForm] = useState<FormState>(DEFAULT_FORM);

  const handleFieldChange = useCallback(
    <K extends keyof FormState>(key: K, value: FormState[K]) => {
      setForm((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  // 特定日選択時に日付未入力だと空の specific_date を送るデッドな送信になるため抑止
  const isSpecificDateMissing =
    form.unavailableType === UnavailableTypeSpecific && form.specificDate === "";

  const [, formAction] = useActionState(async () => {
    try {
      const req: CreateUnavailableTimeRequest = {
        unavailable_type: form.unavailableType,
        start_time: form.startTime,
        end_time: form.endTime,
        ...(form.unavailableType === UnavailableTypeWeekly
          ? { day_of_week: Number(form.dayOfWeek) }
          : { specific_date: form.specificDate }),
      };
      await createMutation.mutateAsync(req);
      setForm(DEFAULT_FORM);
    } catch {
      // エラー通知は useCreateUnavailableTime の onError に一本化（二重トースト防止）
    }
  }, null);

  // 破壊的削除は ConfirmDialog 経由（直行削除禁止）
  const [pendingDelete, setPendingDelete] = useState<ReservationTypeUnavailableTime | null>(null);

  const handleDeleteRequest = useCallback((item: ReservationTypeUnavailableTime) => {
    setPendingDelete(item);
  }, []);

  const handleDeleteCancel = useCallback(() => {
    setPendingDelete(null);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    if (pendingDelete === null) return;
    mutate(pendingDelete.id);
    setPendingDelete(null);
  }, [mutate, pendingDelete]);

  const itemList = useMemo(
    () =>
      items.map((item) => (
        <div
          key={item.id}
          className={`flex items-center justify-between gap-2 py-1.5 px-2 rounded-xxs ${C.hoverBgLight} transition-colors group`}
        >
          <Clock className={`${ICON.smXs} ${C.text40} shrink-0`} />
          <span className={`flex-1 text-sm ${C.text}`}>{unavailableTimeLabel(item)}</span>
          <span className={`text-sm ${C.text50} tabular-nums`}>
            {item.startTime}〜{item.endTime}
          </span>
          {readOnly ? null : (
            <DeleteIconButton
              onClick={() => handleDeleteRequest(item)}
              className="opacity-0 group-hover:opacity-100 focus-visible:opacity-100 group-focus-within:opacity-100 transition-opacity"
            />
          )}
        </div>
      )),
    [items, handleDeleteRequest, readOnly],
  );

  return (
    <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-3">
        <Clock className={`${ICON.smXs} ${C.text50}`} />
        <p className={`text-xs font-medium ${C.text50}`}>予約不可時間</p>
        <FieldHelp
          label="予約不可時間"
          content="この予約区分で予約を受け付けない時間帯です。毎週の曜日指定または特定日で登録できます。"
        />
      </div>

      {/* 既存リスト */}
      {isLoading ? (
        <p className={`text-sm ${C.text60} py-2`}>読み込み中...</p>
      ) : items.length > 0 ? (
        <div className="mb-3 space-y-0.5">{itemList}</div>
      ) : null}

      {/* 追加フォーム */}
      {/* EMR-212: MasterSidePanel がコンテンツ全体を <form action> で包むため、ここに
          <form> を置くとネスト form となりブラウザが破棄して送信不能になる。
          form 要素は使わず、SubmitButton の formAction で送信する（EMR-208 と同型） */}
      {readOnly ? null : (
        <div className="space-y-2">
          {/* 種別 */}
          <div className="flex items-center gap-2">
            <FieldHelp
              label="不可時間の種別"
              content="「毎週」は曜日ごとの繰り返し、「特定日」は指定した日付のみの不可時間です。"
            />
            <Select
              value={form.unavailableType}
              onValueChange={(v) => handleFieldChange("unavailableType", v)}
            >
              <SelectTrigger className={STYLE.selectCompact} aria-label="不可時間の種別">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={UnavailableTypeWeekly}>毎週</SelectItem>
                <SelectItem value={UnavailableTypeSpecific}>特定日</SelectItem>
              </SelectContent>
            </Select>

            {/* 曜日 or 日付 */}
            {form.unavailableType === UnavailableTypeWeekly ? (
              <Select
                value={form.dayOfWeek}
                onValueChange={(v) => handleFieldChange("dayOfWeek", v)}
              >
                <SelectTrigger className={STYLE.selectCompact} aria-label="曜日">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{DAY_OF_WEEK_SELECT_ITEMS}</SelectContent>
              </Select>
            ) : (
              <input
                type="date"
                aria-label="特定日"
                value={form.specificDate}
                onChange={(e) => handleFieldChange("specificDate", e.target.value)}
                className={`rounded-xxs border ${C.borderMedium} px-2 py-1 text-sm ${C.text} ${C.bgWhite}`}
              />
            )}
          </div>

          {/* 時間帯 */}
          <div className="flex items-center gap-2">
            <FieldHelp
              label="不可時間帯"
              content="予約を受け付けない時間帯の開始〜終了時刻です。"
            />
            <Select value={form.startTime} onValueChange={(v) => handleFieldChange("startTime", v)}>
              <SelectTrigger className={STYLE.selectCompact} aria-label="開始時刻">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{TIME_SELECT_ITEMS}</SelectContent>
            </Select>
            <span className={`text-sm ${C.text50}`}>〜</span>
            <Select value={form.endTime} onValueChange={(v) => handleFieldChange("endTime", v)}>
              <SelectTrigger className={STYLE.selectCompact} aria-label="終了時刻">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{TIME_SELECT_ITEMS}</SelectContent>
            </Select>
            <SubmitButton
              loadingText="追加中..."
              className="h-8 text-sm px-3"
              formAction={formAction}
              disabled={isSpecificDateMissing}
            >
              <Plus className={ICON.smXs} />
              追加
            </SubmitButton>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={handleDeleteCancel}
        onConfirm={handleDeleteConfirm}
        title="予約不可時間を削除しますか？"
        description={
          pendingDelete === null
            ? undefined
            : `「${unavailableTimeLabel(pendingDelete)} ${pendingDelete.startTime}〜${pendingDelete.endTime}」の予約不可時間を削除します。この操作は取り消せません。`
        }
        confirmLabel="削除"
        variant="destructive"
        isPending={deleteMutation.isPending}
      />
    </div>
  );
}
