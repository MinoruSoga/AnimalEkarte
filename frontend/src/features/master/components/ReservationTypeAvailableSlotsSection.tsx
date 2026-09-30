import { useCallback, useMemo, useState, useTransition } from "react";
import { useNavigate } from "react-router";
import { CalendarDays, Clock, Plus } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { FieldHelp } from "@/components/shared/FieldHelp";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { DAY_OF_WEEK_LABELS } from "@/constants/day-of-week";
import {
  useGetAvailableSlots,
  useCreateAvailableSlot,
  useDeleteAvailableSlot,
} from "../api/reservation-type-available-slots";
import { AvailableSlotTypeWeekly, AvailableSlotTypeSpecific } from "@/types/generated/models";
import { TIME_SELECT_ITEMS } from "./AvailableSlotOptions";
import { DAY_OF_WEEK_SELECT_ITEMS } from "./DayOfWeekSelectItems";
import type {
  CreateAvailableSlotRequest,
  ReservationTypeAvailableSlot,
} from "../api/reservation-type-available-slots";

interface FormState {
  availableType: string;
  dayOfWeek: string;
  specificDate: string;
  startTime: string;
}

const DEFAULT_FORM: FormState = {
  availableType: AvailableSlotTypeWeekly,
  dayOfWeek: "1",
  specificDate: "",
  startTime: "09:45",
};

function availableSlotLabel(item: ReservationTypeAvailableSlot): string {
  return item.availableType === AvailableSlotTypeWeekly
    ? `毎週${DAY_OF_WEEK_LABELS[item.dayOfWeek ?? 0]}曜日`
    : (item.specificDate ?? "");
}

interface Props {
  clinicId: string;
  reservationTypeId: string;
  /** 参照権限のみのパネル表示。追加フォーム・削除ボタン等の mutation UI を描画しない */
  readOnly?: boolean;
}

export function ReservationTypeAvailableSlotsSection({
  clinicId,
  reservationTypeId,
  readOnly = false,
}: Props) {
  const { data: items = [], isLoading } = useGetAvailableSlots(clinicId, reservationTypeId);
  const createMutation = useCreateAvailableSlot(clinicId, reservationTypeId);
  const deleteMutation = useDeleteAvailableSlot(clinicId, reservationTypeId);
  const { mutate } = deleteMutation;

  const [form, setForm] = useState<FormState>(DEFAULT_FORM);
  const navigate = useNavigate();

  const handleFieldChange = useCallback(
    <K extends keyof FormState>(key: K, value: FormState[K]) => {
      setForm((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  // 特定日選択時に日付未入力だと空の specific_date を送るデッドな送信になるため抑止
  const isSpecificDateMissing =
    form.availableType === AvailableSlotTypeSpecific && form.specificDate === "";

  const [isAddPending, startAddTransition] = useTransition();

  const handleAdd = useCallback(() => {
    startAddTransition(async () => {
      try {
        const req: CreateAvailableSlotRequest = {
          available_type: form.availableType,
          start_time: form.startTime,
          is_active: true,
          ...(form.availableType === AvailableSlotTypeWeekly
            ? { day_of_week: Number(form.dayOfWeek) }
            : { specific_date: form.specificDate }),
        };
        await createMutation.mutateAsync(req);
        setForm(DEFAULT_FORM);
      } catch {
        // エラー通知は useCreateAvailableSlot の onError に一本化（二重トースト防止）
      }
    });
  }, [form, createMutation]);

  // 破壊的削除は ConfirmDialog 経由（直行削除禁止）
  const [pendingDelete, setPendingDelete] = useState<ReservationTypeAvailableSlot | null>(null);

  const handleDeleteRequest = useCallback((item: ReservationTypeAvailableSlot) => {
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
          data-testid="available-slot-row"
          className={`flex items-center justify-between gap-2 py-1.5 px-2 rounded-xxs ${C.hoverBgLight} transition-colors group`}
        >
          <Clock className={`${ICON.smXs} ${C.text40} shrink-0`} />
          <span className={`flex-1 text-sm ${C.text}`}>{availableSlotLabel(item)}</span>
          <span className={`text-sm ${C.text50} tabular-nums`}>{item.startTime}</span>
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
    <div data-testid="available-slots-section" className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-3">
        <Clock className={`${ICON.smXs} ${C.text50}`} />
        <p className={`text-xs font-medium ${C.text50}`}>予約可能枠</p>
        <FieldHelp
          label="予約可能枠"
          content="この予約区分で予約を受け付ける時間枠です。未設定の場合は営業時間内の空き枠が使われます。"
        />
        {readOnly ? null : (
          <button
            type="button"
            onClick={() =>
              navigate(`${paths.lineReservation.slots.getHref()}?typeId=${reservationTypeId}`)
            }
            className={`ml-auto flex items-center gap-1 min-h-11 text-xs ${C.text50} ${C.hoverTextBrand} transition-colors`}
          >
            <CalendarDays className={ICON.smXs} />
            カレンダーで編集
          </button>
        )}
      </div>

      {isLoading ? (
        <p className={`text-sm ${C.text60} py-2`}>読み込み中...</p>
      ) : items.length > 0 ? (
        <div className="mb-3 space-y-0.5">{itemList}</div>
      ) : (
        <p className={`text-xs ${C.text60} mb-3`}>未設定の場合は営業時間内の空き枠を使用します</p>
      )}

      {/* BUG-MASTER-RESVTYPE-SLOT-FORM-NESTED (EMR-208): MasterSidePanel が
          コンテンツ全体を <form action> で包むため、ここに <form> を置くとネスト form
          となりブラウザが破棄して送信不能になる。formAction 経由の submitter も
          祖先 form が無い readOnly パネルでは発火しない。form 要素にも祖先 form への
          依存にも頼らず、type="button" + useTransition で mutation を直接呼ぶ */}
      {readOnly ? null : (
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <FieldHelp
              label="スロット種別"
              content="「毎週」は曜日ごとの繰り返し枠、「特定日」は指定した日付だけの枠です。"
            />
            <Select
              value={form.availableType}
              onValueChange={(value) => handleFieldChange("availableType", value)}
            >
              <SelectTrigger className={STYLE.selectCompact} aria-label="スロット種別">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={AvailableSlotTypeWeekly}>毎週</SelectItem>
                <SelectItem value={AvailableSlotTypeSpecific}>特定日</SelectItem>
              </SelectContent>
            </Select>

            {form.availableType === AvailableSlotTypeWeekly ? (
              <Select
                value={form.dayOfWeek}
                onValueChange={(value) => handleFieldChange("dayOfWeek", value)}
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
                title="枠を適用する日付です"
                value={form.specificDate}
                onChange={(event) => handleFieldChange("specificDate", event.target.value)}
                className={`rounded-xxs border ${C.borderMedium} px-2 py-1 text-sm ${C.text} ${C.bgWhite}`}
              />
            )}
          </div>

          <div className="flex items-center gap-2">
            <FieldHelp label="開始時刻" content="予約枠の開始時刻です。" />
            <Select
              value={form.startTime}
              onValueChange={(value) => handleFieldChange("startTime", value)}
            >
              <SelectTrigger className={STYLE.selectCompact} aria-label="開始時刻">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{TIME_SELECT_ITEMS}</SelectContent>
            </Select>
            <Button
              type="button"
              onClick={handleAdd}
              disabled={isAddPending || isSpecificDateMissing}
              className="h-8 text-sm px-3"
            >
              {isAddPending ? (
                "追加中..."
              ) : (
                <>
                  <Plus className={ICON.smXs} />
                  追加
                </>
              )}
            </Button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={handleDeleteCancel}
        onConfirm={handleDeleteConfirm}
        title="予約可能枠を削除しますか？"
        description={
          pendingDelete === null
            ? undefined
            : `「${availableSlotLabel(pendingDelete)} ${pendingDelete.startTime}」の予約可能枠を削除します。この操作は取り消せません。`
        }
        confirmLabel="削除"
        variant="destructive"
        isPending={deleteMutation.isPending}
      />
    </div>
  );
}
