import {
  memo,
  useState,
  useLayoutEffect,
  useRef,
  useCallback,
  useActionState,
  useMemo,
} from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Plus, X } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import type { Shift, ShiftType, ShiftBreakInput } from "../../types";
import { SHIFT_TYPE_LABELS, isShiftType } from "../../types";
import type {
  ShiftContractError,
  ShiftEntry,
  ShiftMutationIntent,
  ShiftWorkBreak,
  ShiftWorkBreakView,
} from "../../api/contract-types";
import { C, ICON } from "@/lib/design-tokens";
import { createShift } from "../../api/create-shift";
import { updateShift } from "../../api/update-shift";
import { useDeleteShift } from "../../api/delete-shift";
import { getShiftById } from "../../api/get-shifts";
import {
  classifyMutationError,
  getShiftOperationResult,
  isUnknownOutcome,
  newOperationKey,
} from "../../api/shift-operation";
import { useGetShiftTemplates } from "../../api/get-shift-templates";
import { handleApiError } from "@/lib/handle-api-error";
import { getFormString } from "@/lib/form-data";
import { queryKeys } from "@/lib/query-keys";
import { formatJSTTime } from "@/lib/jst-date";
import { requiresShiftTimes } from "../../lib/shift-template-form-utils";
import { DEFAULT_BREAK_START, DEFAULT_BREAK_END } from "../../lib/shift-template-form-model";

/**
 * バックエンドから "HH:MM:SS" 形式で来る時刻を "HH:mm" に正規化する。
 */
function normalizeTimeToHHmm(time: string): string {
  if (!time) return "";
  const parts = time.split(":");
  if (parts.length >= 2) {
    return `${parts[0].padStart(2, "0")}:${parts[1].padStart(2, "0")}`;
  }
  return time;
}

/** JST date-time（契約の明示日時。例: 2026-10-21T09:00:00+09:00） */
function buildJSTDateTime(date: string, hhmm: string): string {
  return `${date}T${hhmm}:00+09:00`;
}

function isWorkType(shiftType: ShiftType): boolean {
  // EMR-241: 時刻必須/勤務種別判定は種別の RequiresTimeSlot 判定のみで決める
  return requiresShiftTimes(shiftType);
}

const SHIFT_TYPE_OPTIONS = (Object.entries(SHIFT_TYPE_LABELS) as [ShiftType, string][]).map(
  ([value, label]) => (
    <SelectItem key={value} value={value}>
      {label}
    </SelectItem>
  ),
);

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

interface ShiftFormDialogProps {
  open: boolean;
  onClose: () => void;
  staffId: string;
  staffName: string;
  date: string; // YYYY-MM-DD
  editShift?: Shift;
  canCreate?: boolean;
  canEdit?: boolean;
  canDelete?: boolean;
}

interface FormActionState {
  timeError?: string;
  breakError?: string;
  formError?: string;
}

/** unknown（旧日時未解決）の休憩。訂正確認が済むまで保存要求へ流用しない */
function splitKnownAndUnresolvedBreaks(breaks: ShiftWorkBreakView[]): {
  known: ShiftBreakInput[];
  unresolved: ShiftWorkBreakView[];
} {
  const known: ShiftBreakInput[] = [];
  const unresolved: ShiftWorkBreakView[] = [];
  for (const b of breaks) {
    if (b.interval_state === "known") {
      known.push({
        break_start: formatJSTTime(b.starts_at),
        break_end: formatJSTTime(b.ends_at),
      });
    } else {
      unresolved.push(b);
    }
  }
  return { known, unresolved };
}

export const ShiftFormDialog = memo(function ShiftFormDialog({
  open,
  onClose,
  staffId,
  staffName,
  date,
  editShift,
  canEdit = false,
  canDelete = false,
  canCreate = false,
}: ShiftFormDialogProps) {
  const isEdit = editShift !== undefined;
  // rerender-dependencies: editShift オブジェクトの代わりに primitive id を deps に使用
  const editShiftId = editShift?.id;
  const queryClient = useQueryClient();
  const permissionsRef = useRef({ canCreate, canEdit, canDelete });
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
  }, [canCreate, canEdit, canDelete]);

  // テンプレート選択
  const { data: templates = [] } = useGetShiftTemplates();

  // Controlled state for Select and time inputs (needed for UI feedback and FormData)
  const [shiftType, setShiftType] = useState<ShiftType>(() => editShift?.shift_type ?? "full");
  const [startTime, setStartTime] = useState(() =>
    normalizeTimeToHHmm(editShift?.starts_at ? formatJSTTime(editShift.starts_at) : ""),
  );
  const [endTime, setEndTime] = useState(() =>
    normalizeTimeToHHmm(editShift?.ends_at ? formatJSTTime(editShift.ends_at) : ""),
  );
  const [breaks, setBreaks] = useState<ShiftBreakInput[]>(() =>
    editShift ? splitKnownAndUnresolvedBreaks(editShift.breaks).known : [],
  );
  // React 19 の form action は送信後に uncontrolled 入力を自動リセットするため、
  // 契約の「入力保持」を満すべくメモは controlled で持つ
  const [notes, setNotes] = useState(() => editShift?.notes ?? "");
  // EMR-514: unknown休憩は消去せず保持し、訂正確認まで送信対象にしない
  const [unresolvedBreaks, setUnresolvedBreaks] = useState<ShiftWorkBreakView[]>(() =>
    editShift ? splitKnownAndUnresolvedBreaks(editShift.breaks).unresolved : [],
  );
  // rerender-dependencies: breaks 配列を ref 経由で参照し formAction deps から除外
  const breaksRef = useRef(breaks);
  const unresolvedBreaksRef = useRef(unresolvedBreaks);
  useLayoutEffect(() => {
    breaksRef.current = breaks;
    unresolvedBreaksRef.current = unresolvedBreaks;
  }, [breaks, unresolvedBreaks]);

  // EMR-514: 結果未確定（unknown/通信断/5xx）の操作。新key再送は禁止
  const [pendingOperation, setPendingOperation] = useState<{
    operationKey: string;
    intent: ShiftMutationIntent;
  } | null>(null);
  // EMR-514: 409契約エラー（code分岐）。version_conflictは最新状態の別読取で比較/再確認する
  const [conflict, setConflict] = useState<{
    payload: ShiftContractError;
    latest: ShiftEntry | null;
  } | null>(null);
  const conflictLatestLoadingRef = useRef(false);

  // FE-RC-032: open 切り替え時の state リセットは useEffect ではなく、
  // 呼び出し側（ShiftCalendar.tsx）が dialog セッションごとに異なる key を
  // 付与することでコンポーネントを再マウントさせ、上記の useState 初期化子に
  // 委ねる（ShiftTemplateSidePanel と同様の設計に統一）。

  /** 休憩入力行を契約の ShiftWorkBreak へ変換する（work種別のみ呼ぶ） */
  const buildBreaksPayload = useCallback((): {
    breaks?: ShiftWorkBreak[];
    error?: string;
  } => {
    const rows = breaksRef.current;
    const hasUnresolved = unresolvedBreaksRef.current.length > 0;
    const filled = rows.filter((b) => b.break_start !== "" || b.break_end !== "");
    if (filled.some((b) => b.break_start === "" || b.break_end === "")) {
      return { error: "休憩の開始時刻と終了時刻を入力してください" };
    }
    for (const b of filled) {
      if (b.break_end <= b.break_start) {
        return { error: "休憩の終了時刻は開始時刻より後に設定してください" };
      }
      if (startTime !== "" && endTime !== "") {
        if (b.break_start < startTime || b.break_end > endTime) {
          return { error: "休憩は勤務時間内に収めてください" };
        }
      }
    }
    // 未指定は「維持」、明示したいときだけ配列を送る（PATCH最小差分）
    const breaksChanged =
      !isEdit ||
      JSON.stringify(filled) !==
        JSON.stringify(editShift ? splitKnownAndUnresolvedBreaks(editShift.breaks).known : []);
    if (!breaksChanged) return {};
    if (hasUnresolved) {
      // 配列全置換のため、未解決行の無断消去を避ける（訂正確認まで送信停止）
      return {
        error: "未解決の休憩があります。訂正確認するまで休憩の変更は送信できません",
      };
    }
    return {
      breaks: filled.map((b) => ({
        starts_at: buildJSTDateTime(date, b.break_start),
        ends_at: buildJSTDateTime(date, b.break_end),
      })),
    };
  }, [date, isEdit, editShift, startTime, endTime]);

  const formAction = useCallback(
    async (_prevState: FormActionState, formData: FormData): Promise<FormActionState> => {
      const allowed = isEdit
        ? permissionsRef.current.canEdit === true
        : permissionsRef.current.canCreate === true;
      if (!allowed) {
        toast.error(PERMISSION_DENIED_MESSAGE);
        return {};
      }
      if (pendingOperation !== null) {
        return { formError: "前の操作の結果が未確定です。結果を照会してください" };
      }
      const rawShiftType = formData.get("shiftType");
      const resolvedShiftType = isShiftType(rawShiftType) ? rawShiftType : shiftType;
      const resolvedStartTime = getFormString(formData, "startTime");
      const resolvedEndTime = getFormString(formData, "endTime");
      const workType = isWorkType(resolvedShiftType);
      // 勤務種別では時刻必須。off/paid_leave は入力値を捨てず state 保持のまま、
      // 契約どおり明示NULLを送る（spec: 24-shift-calendar「時刻・休憩入力を無効化して時刻を送信しない」）。
      const hasStartTime = resolvedStartTime !== "";
      const hasEndTime = resolvedEndTime !== "";

      // 片方だけの時刻入力は全カテゴリで拒否（推測補完をしない）
      if (hasStartTime !== hasEndTime) {
        return { timeError: "開始時刻と終了時刻を入力してください" };
      }
      // 勤務種別は時刻必須。ただし旧日時がNULL（unknown）の編集では
      // メモ等の部分修正を時刻再入力なしで許す（未指定=時刻維持）
      const oldTimesKnown = isEdit && editShift !== undefined && editShift.starts_at !== null;
      if (workType && !hasStartTime && (!isEdit || oldTimesKnown)) {
        return { timeError: "開始時刻と終了時刻を入力してください" };
      }
      if (hasStartTime && hasEndTime && resolvedEndTime <= resolvedStartTime) {
        return { timeError: "終了時刻は開始時刻より後に設定してください" };
      }

      const notesValue = formData.has("notes") ? String(formData.get("notes") ?? "") : null;
      const resolvedNotes = notesValue ?? editShift?.notes ?? "";

      // mutation意図ごとに1回発行し、結果照会が確定するまで同じkeyを使う
      const operationKey = newOperationKey();

      try {
        if (isEdit && editShiftId) {
          const editShiftSnapshot = editShift;
          if (!editShiftSnapshot) return {};
          const req: Parameters<typeof updateShift>[1] = {
            operation_key: operationKey,
            expected_version: editShiftSnapshot.version,
          };
          const wasWork = isWorkType(editShiftSnapshot.shift_type);
          if (resolvedShiftType !== editShiftSnapshot.shift_type) {
            req.shift_type = resolvedShiftType;
          }
          if (workType) {
            const breakResult = buildBreaksPayload();
            if (breakResult.error) return { breakError: breakResult.error };
            if (hasStartTime && hasEndTime) {
              const newStartsAt = buildJSTDateTime(date, resolvedStartTime);
              const newEndsAt = buildJSTDateTime(date, resolvedEndTime);
              if (!wasWork) {
                // 非勤務→勤務: shift_type・両日時を明示する
                req.shift_type = resolvedShiftType;
                req.starts_at = newStartsAt;
                req.ends_at = newEndsAt;
                if (breakResult.breaks !== undefined) req.breaks = breakResult.breaks;
              } else {
                // 勤務→勤務: 変更した項目だけ送る（未指定は維持）。
                // 旧値がNULL（旧unknown日時）のときは明示日時への遷移として必ず送る
                if (
                  editShiftSnapshot.starts_at === null ||
                  editShiftSnapshot.starts_at !== newStartsAt
                ) {
                  req.starts_at = newStartsAt;
                }
                if (editShiftSnapshot.ends_at === null || editShiftSnapshot.ends_at !== newEndsAt) {
                  req.ends_at = newEndsAt;
                }
                if (breakResult.breaks !== undefined) req.breaks = breakResult.breaks;
              }
            }
            // 時刻未入力（旧unknown編集の部分修正）のときは日時を送らない（維持）
          } else if (wasWork) {
            // 勤務→非勤務: shift_type・両日時NULL・breaks[] を明示する
            req.shift_type = resolvedShiftType;
            req.starts_at = null;
            req.ends_at = null;
            req.breaks = [];
          }
          // 非勤務→非勤務（off⇔paid_leave）は時刻・休憩を送らない（未指定=維持）
          if (resolvedNotes !== editShiftSnapshot.notes) {
            // 空文字は解除（NULLは400）
            req.notes = resolvedNotes;
          }
          if (Object.keys(req).length <= 2) {
            return { formError: "変更がありません" };
          }
          const receipt = await updateShift(editShiftId, req);
          if (receipt.shift === null || receipt.resulting_version === null) {
            // update成功契約は shift / resulting_version 非NULL
            return { formError: "更新結果の形式が不正です" };
          }
          await queryClient.invalidateQueries({ queryKey: queryKeys.shifts.all() });
          onClose();
          return {};
        }

        // create: 全入力を明示する（NULL/未指定を同一視しない）
        let breaksPayload: ShiftWorkBreak[] = [];
        if (workType) {
          const breakResult = buildBreaksPayload();
          if (breakResult.error) return { breakError: breakResult.error };
          breaksPayload = breakResult.breaks ?? [];
        }
        const createReq: Parameters<typeof createShift>[0] = {
          operation_key: operationKey,
          staff_id: staffId,
          date,
          shift_type: resolvedShiftType,
          starts_at: workType ? buildJSTDateTime(date, resolvedStartTime) : null,
          ends_at: workType ? buildJSTDateTime(date, resolvedEndTime) : null,
          breaks: breaksPayload,
          notes: resolvedNotes,
        };
        const receipt = await createShift(createReq);
        if (receipt.shift === null || receipt.resulting_version === null) {
          return { formError: "作成結果の形式が不正です" };
        }
        await queryClient.invalidateQueries({ queryKey: queryKeys.shifts.all() });
        onClose();
        return {};
      } catch (err) {
        const classified = classifyMutationError(err);
        if (classified.kind === "contract") {
          // 409: 業務変更なし。codeで分岐し比較/再確認へ
          setConflict({ payload: classified.payload, latest: null });
          return {};
        }
        if (isUnknownOutcome(classified)) {
          // 結果応答喪失・5xx: 同じkeyの結果照会へ。新key再送はしない
          const intent: ShiftMutationIntent = isEdit ? "update" : "create";
          setPendingOperation({ operationKey, intent });
          return {};
        }
        handleApiError(err, isEdit ? "シフト更新" : "シフト追加");
        return {};
      }
    },
    // rerender-dependencies: editShift → id（primitive）、breaks → ref 経由
    [
      isEdit,
      editShiftId,
      editShift,
      staffId,
      date,
      shiftType,
      queryClient,
      onClose,
      buildBreaksPayload,
      pendingOperation,
    ],
  );

  // 送信直前の operation_key は formAction 内で発行して catch まで同じローカル変数を使う

  const [state, dispatchFormAction, isPending] = useActionState<FormActionState, FormData>(
    formAction,
    {},
  );

  const { mutateAsync: deleteShiftAsync, isPending: isDeletePending } = useDeleteShift();
  // BUG-093: 削除確認ダイアログの表示状態
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);

  // EMR-514: 結果照会。succeeded ならreceiptで確定、unknown は新key再送せず停止
  const inquirePendingOperation = useCallback(async () => {
    if (pendingOperation === null) return;
    try {
      const result = await getShiftOperationResult(
        pendingOperation.operationKey,
        pendingOperation.intent,
      );
      if (result.state === "succeeded") {
        await queryClient.invalidateQueries({ queryKey: queryKeys.shifts.all() });
        setPendingOperation(null);
        onClose();
        return;
      }
      // unknown: 未実行・失敗・不存在の証明ではない。停止して入力保持
      toast.error("結果はまだ確定していません。時間を置いてもう一度照会してください");
    } catch (err) {
      // 権限失効(403)・障害(500)はunknown成功へ変換しない
      handleApiError(err, "操作結果の照会");
    }
  }, [pendingOperation, queryClient, onClose]);

  /** version_conflict 時の最新状態の別読取（編集案を上書きせず比較用） */
  const loadLatestShift = useCallback(async () => {
    if (!editShiftId || conflictLatestLoadingRef.current) return;
    conflictLatestLoadingRef.current = true;
    try {
      const latest = await getShiftById(editShiftId);
      setConflict((prev) => (prev ? { ...prev, latest } : prev));
    } catch (err) {
      handleApiError(err, "最新の勤務取得");
    } finally {
      conflictLatestLoadingRef.current = false;
    }
  }, [editShiftId]);

  const handleApplyTemplate = useCallback(
    (templateId: string) => {
      const tpl = templates.find((t) => t.id === templateId);
      if (!tpl) return;
      setShiftType(tpl.shift_type);
      setStartTime(tpl.start_time ?? "");
      setEndTime(tpl.end_time ?? "");
      setBreaks(tpl.breaks.map((b) => ({ break_start: b.break_start, break_end: b.break_end })));
    },
    [templates],
  );

  const handleShiftTypeChange = useCallback((value: string) => {
    setShiftType(value as ShiftType);
  }, []);

  const handleStartTimeChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setStartTime(e.target.value);
  }, []);

  const handleEndTimeChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setEndTime(e.target.value);
  }, []);

  // EMR-514: unknown休憩の訂正確認。既知の片側だけが判明している場合はその値を引き継ぐ
  const handleResolveUnresolvedBreak = useCallback((index: number) => {
    setUnresolvedBreaks((prev) => {
      const target = prev[index];
      if (!target) return prev;
      setBreaks((rows) => [
        ...rows,
        {
          break_start: target.starts_at ? formatJSTTime(target.starts_at) : DEFAULT_BREAK_START,
          break_end: target.ends_at ? formatJSTTime(target.ends_at) : DEFAULT_BREAK_END,
        },
      ]);
      return prev.filter((_, i) => i !== index);
    });
  }, []);

  const workType = isWorkType(shiftType);

  // BUG-093: 削除確認後に実際の削除を実行（409はcode分岐、通信断は照会へ）
  const handleDeleteConfirm = useCallback(async () => {
    if (permissionsRef.current.canDelete !== true) {
      toast.error(PERMISSION_DENIED_MESSAGE);
      return;
    }
    if (!editShift) return;
    setIsDeleteConfirmOpen(false);
    if (pendingOperation !== null) return;
    const operationKey = newOperationKey();
    try {
      const receipt = await deleteShiftAsync({
        shiftId: editShift.id,
        expectedVersion: editShift.version,
        operationKey,
      });
      if (receipt.shift !== null || receipt.resulting_version !== null) {
        // 削除の成功契約は shift=NULL / resulting_version=NULL
        toast.error("削除結果の形式が不正です");
        return;
      }
      onClose();
    } catch (err) {
      const classified = classifyMutationError(err);
      if (classified.kind === "contract") {
        setConflict({ payload: classified.payload, latest: null });
        return;
      }
      if (isUnknownOutcome(classified)) {
        setPendingOperation({ operationKey, intent: "delete" });
        return;
      }
      handleApiError(err, "シフト削除");
    }
  }, [editShift, deleteShiftAsync, onClose, pendingOperation]);

  // js-cache-function-results: テンプレート選択肢を useMemo でキャッシュ
  const templateSelectItems = useMemo(
    () =>
      templates
        .filter((t) => t.is_active)
        .map((t) => (
          <SelectItem key={t.id} value={t.id}>
            {t.name}
          </SelectItem>
        )),
    [templates],
  );

  const formattedDate = date
    ? new Date(`${date}T00:00:00+09:00`).toLocaleDateString("ja-JP", {
        timeZone: "Asia/Tokyo",
        month: "long",
        day: "numeric",
        weekday: "short",
      })
    : "";

  const conflictBookings = conflict?.payload.protected_bookings ?? [];

  return (
    <>
      <Dialog
        open={open}
        onOpenChange={(o) => {
          if (!o) onClose();
        }}
      >
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{isEdit ? "シフト編集" : "シフト追加"}</DialogTitle>
            <DialogDescription>
              {staffName} の {formattedDate} のシフト内容を入力します。
            </DialogDescription>
            <p className={`text-sm ${C.text50}`}>
              {staffName} — {formattedDate}
            </p>
          </DialogHeader>

          <form action={dispatchFormAction} className="space-y-4">
            {/* hidden input を経由してコントロール値を FormData に渡す */}
            <input type="hidden" name="shiftType" value={shiftType} />

            {/* テンプレート選択（テンプレートがある場合のみ表示） */}
            {templateSelectItems.length > 0 ? (
              <div className="space-y-1.5">
                <Label htmlFor="shift-template" className={`${C.text50}`}>
                  テンプレートから入力
                </Label>
                <Select onValueChange={handleApplyTemplate}>
                  <SelectTrigger id="shift-template">
                    <SelectValue placeholder="テンプレートを選択..." />
                  </SelectTrigger>
                  <SelectContent>{templateSelectItems}</SelectContent>
                </Select>
              </div>
            ) : null}

            <div className="space-y-1.5">
              <Label htmlFor="shift-type">シフト種別</Label>
              <Select value={shiftType} onValueChange={handleShiftTypeChange}>
                <SelectTrigger id="shift-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{SHIFT_TYPE_OPTIONS}</SelectContent>
              </Select>
            </div>

            {/* 勤務種別は時刻必須。off/paid_leave は終日非勤務のため無効化し、
                保存時に明示NULLを送る（入力値は捨てない） */}
            <div className="space-y-1.5">
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <Label htmlFor="start-time">開始時刻</Label>
                  <Input
                    id="start-time"
                    name="startTime"
                    type="time"
                    value={startTime}
                    onChange={handleStartTimeChange}
                    disabled={!workType}
                    aria-invalid={state.timeError ? true : undefined}
                    aria-describedby={state.timeError ? "shift-time-error" : undefined}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="end-time">終了時刻</Label>
                  <Input
                    id="end-time"
                    name="endTime"
                    type="time"
                    value={endTime}
                    onChange={handleEndTimeChange}
                    disabled={!workType}
                    aria-invalid={state.timeError ? true : undefined}
                    aria-describedby={state.timeError ? "shift-time-error" : undefined}
                  />
                </div>
              </div>
              {state.timeError ? (
                <FormFieldError id="shift-time-error" message={state.timeError} />
              ) : null}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="notes">メモ</Label>
              <Input
                id="notes"
                name="notes"
                placeholder="メモ（任意）"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </div>

            {/* 休憩時間（work種別のみ有効。off/paid_leave は明示 [] を送る） */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>休憩時間</Label>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2 text-xs"
                  disabled={!workType}
                  onClick={() =>
                    setBreaks((prev) => [
                      ...prev,
                      { break_start: DEFAULT_BREAK_START, break_end: DEFAULT_BREAK_END },
                    ])
                  }
                >
                  <Plus className={`${ICON.xxs} mr-1`} />
                  追加
                </Button>
              </div>
              {breaks.map((b, i) => (
                <div key={i} className="flex items-center gap-2">
                  <Input
                    type="time"
                    aria-label={`休憩${i + 1} 開始時刻`}
                    value={b.break_start}
                    onChange={(e) =>
                      setBreaks((prev) =>
                        prev.map((br, j) =>
                          j === i ? { ...br, break_start: e.target.value } : br,
                        ),
                      )
                    }
                    disabled={!workType}
                    className="flex-1"
                  />
                  <span className={`text-xs ${C.text50}`}>〜</span>
                  <Input
                    type="time"
                    aria-label={`休憩${i + 1} 終了時刻`}
                    value={b.break_end}
                    onChange={(e) =>
                      setBreaks((prev) =>
                        prev.map((br, j) => (j === i ? { ...br, break_end: e.target.value } : br)),
                      )
                    }
                    disabled={!workType}
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0"
                    aria-label={`休憩${i + 1}を削除`}
                    disabled={!workType}
                    onClick={() => setBreaks((prev) => prev.filter((_, j) => j !== i))}
                  >
                    <X className={ICON.smXs} />
                  </Button>
                </div>
              ))}
              {unresolvedBreaks.map((b, i) => (
                <div key={`unresolved-${i}`} className="flex items-center gap-2">
                  <span
                    className={`text-xs ${C.text50} flex-1`}
                    aria-label={`休憩（未解決）${i + 1}`}
                  >
                    {b.starts_at ? formatJSTTime(b.starts_at) : "—"}〜
                    {b.ends_at ? formatJSTTime(b.ends_at) : "—"}（時刻未解決）
                  </span>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="h-7 px-2 text-xs"
                    disabled={!workType}
                    aria-label={`休憩（未解決）${i + 1}を訂正する`}
                    onClick={() => handleResolveUnresolvedBreak(i)}
                  >
                    訂正する
                  </Button>
                </div>
              ))}
              {unresolvedBreaks.length > 0 ? (
                <p className={`text-xs ${C.text50}`} data-testid="unresolved-breaks-note">
                  時刻が未解決の休憩があります。訂正確認するまで休憩の変更は送信されません。
                </p>
              ) : null}
              {state.breakError ? (
                <FormFieldError id="shift-break-error" message={state.breakError} />
              ) : null}
            </div>

            {state.formError ? (
              <FormFieldError id="shift-form-error" message={state.formError} />
            ) : null}

            {/* EMR-514: 409契約エラー（codeで分岐。error文字列は分岐に使わない） */}
            {conflict ? (
              <div
                role="alert"
                data-testid="shift-conflict"
                className={`rounded border p-3 text-xs space-y-2 ${
                  conflict.payload.code === "protected_bookings"
                    ? `${C.bgAccentLight} ${C.borderOrangeBadge}`
                    : `${C.bgStatusGray} ${C.borderMuted}`
                }`}
              >
                <p className="font-medium">
                  {conflict.payload.code === "version_conflict"
                    ? "勤務が他の操作で更新されました。最新状態と比較して再確認してください。"
                    : null}
                  {conflict.payload.code === "operation_key_reused"
                    ? "同じ操作キーが別の内容で使われました。再度操作を作り直してください。"
                    : null}
                  {conflict.payload.code === "shift_exists"
                    ? "同じ対象の勤務が既に存在します。"
                    : null}
                  {conflict.payload.code === "source_unresolved"
                    ? "参照元の情報が未解決のため保存できません。"
                    : null}
                  {conflict.payload.code === "cross_clinic_overlap"
                    ? "他院の勤務と時間が重複しています。"
                    : null}
                  {conflict.payload.code === "protected_bookings"
                    ? "この変更で不適格になる既存予約があるため保存できません。予約は自動解除されません。"
                    : null}
                </p>
                {conflict.payload.code === "cross_clinic_overlap" &&
                conflict.payload.conflicting_clinic_name ? (
                  <p data-testid="cross-clinic-overlap">
                    {conflict.payload.conflicting_clinic_name}{" "}
                    {conflict.payload.overlap_starts_at
                      ? formatJSTTime(conflict.payload.overlap_starts_at)
                      : ""}
                    〜
                    {conflict.payload.overlap_ends_at
                      ? formatJSTTime(conflict.payload.overlap_ends_at)
                      : ""}
                  </p>
                ) : null}
                {conflict.payload.code === "protected_bookings" && conflictBookings.length > 0 ? (
                  <ul data-testid="protected-bookings" className="list-disc pl-4">
                    {conflictBookings.map((b) => (
                      <li key={b.appointment_id}>
                        {b.patient_label}（{b.reservation_type_label}） {formatJSTTime(b.starts_at)}
                        〜{formatJSTTime(b.ends_at)}
                      </li>
                    ))}
                  </ul>
                ) : null}
                {conflict.payload.code === "version_conflict" ? (
                  <div className="space-y-1">
                    {conflict.latest ? (
                      <p data-testid="latest-shift-summary">
                        最新: {SHIFT_TYPE_LABELS[conflict.latest.shift_type]}{" "}
                        {conflict.latest.starts_at ? formatJSTTime(conflict.latest.starts_at) : "—"}
                        〜{conflict.latest.ends_at ? formatJSTTime(conflict.latest.ends_at) : "—"}
                        （版 {conflict.latest.version}）
                      </p>
                    ) : (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        className="h-7 px-2 text-xs"
                        onClick={loadLatestShift}
                      >
                        最新の勤務を取得
                      </Button>
                    )}
                    <p className={C.text50}>
                      内容を調整して保存し直すと、最新の版で新しい操作を作ります。
                    </p>
                  </div>
                ) : null}
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2 text-xs"
                  onClick={() => setConflict(null)}
                >
                  入力に戻る
                </Button>
              </div>
            ) : null}

            {/* EMR-514: 結果未確定（unknown）。新key再送は禁止し、同key結果照会のみ許可 */}
            {pendingOperation ? (
              <div
                role="status"
                data-testid="pending-operation"
                className={`rounded border p-3 text-xs space-y-2 ${C.bgStatusGray} ${C.borderMuted}`}
              >
                <p>送信の結果が未確定です。重複を避けるため新しい操作は送信できません。</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-7 px-2 text-xs"
                  onClick={inquirePendingOperation}
                >
                  結果を照会
                </Button>
              </div>
            ) : null}

            <DialogFooter className="gap-2">
              {isEdit && canDelete ? (
                <Button
                  type="button"
                  variant="destructive"
                  size="sm"
                  onClick={() => setIsDeleteConfirmOpen(true)}
                  disabled={isPending || isDeletePending || pendingOperation !== null}
                >
                  削除
                </Button>
              ) : null}
              <Button
                type="button"
                variant="outline"
                onClick={onClose}
                disabled={isPending || isDeletePending}
              >
                キャンセル
              </Button>
              {canEdit ? (
                <SubmitButton size="sm" disabled={isDeletePending || pendingOperation !== null}>
                  保存
                </SubmitButton>
              ) : null}
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* BUG-093: 削除確認ダイアログ */}
      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onClose={() => setIsDeleteConfirmOpen(false)}
        title="このシフトを削除しますか？"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
});
