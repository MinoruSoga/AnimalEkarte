import { useState, useCallback, useMemo } from "react";
import { useNavigate, useLocation } from "react-router";
import { toast } from "sonner";
import {
  addHours,
  startOfWeek,
  endOfWeek,
  startOfMonth,
  endOfMonth,
  addDays,
  subDays,
  format,
} from "date-fns";
import { ja } from "date-fns/locale";
import { formatJSTWallDate, toJSTWallDate } from "@/lib/jst-date";
import { useTodayJst } from "@/hooks/use-today-jst";
import { useConfirmTentative } from "@/hooks/use-confirm-tentative";
import { useSetNextVisitIntent, useClearNextVisitIntent } from "@/hooks/use-next-visit-intent";
import type { CalendarView, Reservation, ReservationFormData, NavigationState } from "../types";
import type { ReservationCreateMutations } from "@/types/reservation-create-mutations";

import { useGetReservations } from "../api/get-reservations";
import { isExpiredTentativeReservation, isTentativeReservation } from "../lib/tentative";
import {
  useReservationActions,
  type ReservationMutationPermissions,
  type StatusConfirmTarget,
} from "./use-reservation-actions";
import { useReservationModalState } from "./use-reservation-modal-state";
import { useReservationRecordNavigation } from "./use-reservation-record-navigation";

const EMPTY_RESERVATIONS: Reservation[] = [];

/** EMR-308 A3: 期限超過 tentative 一覧の探索下限。実運用上の最古意図日をカバーする固定下限。 */
const EXPIRED_TENTATIVE_RANGE_START = "2000-01-01";

/** EMR-308 A3/A4: 「次回予定を再設定」ダイアログの対象 */
export interface ResetNextVisitTarget {
  petId: string;
  petName: string;
  /** 直前の意図日（YYYY-MM-DD）。BE はキャンセル後も intended_date を保持する */
  previousIntendedDate?: string;
  /** tentative の作成元カルテ id（source_medical_record_id として引き継ぐ） */
  sourceMedicalRecordId?: string;
  /**
   * Codex PR606 #5: 拠点横断表示では行の clinicId を保持し、
   * PUT intent をその医院へ送る（グローバル選択医院への誤送信を防ぐ）。
   */
  clinicId?: string;
}

interface UseReservationManagementParams {
  currentDate: Date;
  view: CalendarView;
  days: 5 | 7;
  /** #86: 拠点横断表示。複数医院IDのとき clinic_ids をAPIに送信する */
  clinicIds?: string[];
  createMutations: ReservationCreateMutations;
  /** FE-RC-204: 未指定は fail-closed（全拒否）。 */
  permissions?: Readonly<ReservationMutationPermissions>;
}

export function useReservationManagement({
  currentDate,
  view,
  days,
  clinicIds,
  createMutations,
  permissions,
}: UseReservationManagementParams) {
  const navigate = useNavigate();
  const location = useLocation();
  const locationFrom = (location.state as NavigationState | null)?.from ?? null;

  // 表示中の週/月の期間を算出し、その範囲の予約だけを取得する。
  // WeekView (startOfWeek から days 日) / MonthView (startOfWeek(monthStart)〜endOfWeek(monthEnd)) の
  // 日付生成ロジックと一致させ、表示範囲の予約が limit から漏れないようにする (BUG #82)。
  const { startDate, endDate } = useMemo(() => {
    if (view === "month") {
      const start = startOfWeek(startOfMonth(currentDate), { locale: ja });
      const end = endOfWeek(endOfMonth(currentDate), { locale: ja });
      return { startDate: format(start, "yyyy-MM-dd"), endDate: format(end, "yyyy-MM-dd") };
    }
    const start = startOfWeek(currentDate, { locale: ja });
    const end = addDays(start, days - 1);
    return { startDate: format(start, "yyyy-MM-dd"), endDate: format(end, "yyyy-MM-dd") };
  }, [currentDate, view, days]);

  const { data: appointments = EMPTY_RESERVATIONS, isLoading } = useGetReservations({
    startDate,
    endDate,
    clinicIds,
  });

  // EMR-308 A3: 期限超過 tentative の actionable list 用フェッチ。
  // BE の end_date は当日を含むため「昨日」指定で意図日 < 今日 に限定する。
  // （start_time sentinel は意図日の JST 00:00。当日 tentative は範囲外＝期限超過ではない。）
  // status=tentative は BE の素の等値フィルタ（reservation_repository.go FindAll）で有効。
  // Codex PR606 #13: マウント中に JST 暦日を跨ぐと expiredRangeEndDate が古いまま残り、
  // 当日に期限超過化した行がクエリ範囲・client 判定の両方から漏れる。
  // useTodayJst で暦日 tick を監視し、境界と再判定の両方を当日へ追従させる。
  const todayJst = useTodayJst();
  const expiredRangeEndDate = useMemo(
    // "+09:00" 付きで parse して JST 暦日をローカル TZ 非依存に固定する（jst-date 規約）。
    () => formatJSTWallDate(subDays(toJSTWallDate(`${todayJst}T00:00:00+09:00`), 1)),
    [todayJst],
  );
  const { data: tentativeRangeAppointments = EMPTY_RESERVATIONS } = useGetReservations({
    status: "tentative",
    startDate: EXPIRED_TENTATIVE_RANGE_START,
    endDate: expiredRangeEndDate,
    clinicIds,
  });
  // サーバー側フィルタに加えて意図日の JST 暦日で厳密に再判定する（fail-safe）。
  // todayJst を渡して暦日跨ぎでも判定境界が更新されるようにする（#13）。
  const expiredTentativeAppointments = useMemo(
    () => tentativeRangeAppointments.filter((r) => isExpiredTentativeReservation(r, todayJst)),
    [tentativeRangeAppointments, todayJst],
  );

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Reservation | null>(null);
  const [statusConfirmOpen, setStatusConfirmOpen] = useState(false);
  const [statusConfirmTarget, setStatusConfirmTarget] = useState<StatusConfirmTarget | null>(null);

  const {
    isFormOpen,
    editingAppointment,
    editingAppointmentRef,
    handleOpenForm,
    handleCloseForm,
    isDetailOpen,
    detailAppointment,
    setDetailAppointment,
    handleOpenDetail,
    handleCloseDetail,
  } = useReservationModalState({ locationSearch: location.search });

  // ── EMR-308: tentative 確定 / 次回予定の再設定・終了 ──
  const confirmTentativeMutation = useConfirmTentative();
  const setIntentMutation = useSetNextVisitIntent();
  const clearIntentMutation = useClearNextVisitIntent();

  const [confirmTentativeTarget, setConfirmTentativeTarget] = useState<Reservation | null>(null);
  const [resetNextVisitTarget, setResetNextVisitTarget] = useState<ResetNextVisitTarget | null>(
    null,
  );
  const [endTentativeTarget, setEndTentativeTarget] = useState<Reservation | null>(null);

  // tentative 行を確定モーダルで開く（detail modal / 期限超過パネルからの入口）。
  // tentative 以外の行や権限なしは fail-closed。
  const handleOpenConfirmTentative = useCallback(
    (reservation: Reservation) => {
      if (permissions?.canEdit !== true) {
        toast.error("予約を確定する権限がありません");
        return;
      }
      if (!isTentativeReservation(reservation)) return;
      setConfirmTentativeTarget(reservation);
    },
    [permissions],
  );

  const handleCloseConfirmTentative = useCallback(() => setConfirmTentativeTarget(null), []);

  // ReservationFormModal の tentative 確定モードから呼ばれる保存ハンドラ。
  // POST /v1/reservations/:id/confirm-tentative — 同一行を昇格する（別行を作らない）。
  const handleConfirmTentative = useCallback(
    async (tentativeId: string, data: Partial<Reservation>): Promise<string | null> => {
      try {
        await confirmTentativeMutation.mutateAsync({
          reservationId: tentativeId,
          formData: data,
          // Codex PR606 #5: 拠点横断表示の行は行の clinicId で確定する
          clinicId: confirmTentativeTarget?.clinicId,
        });
      } catch {
        // onError(handleApiError) が toast 済み。モーダル内にインライン表示して開いたまま。
        return "予約の確定に失敗しました。入力内容を確認して再度お試しください。";
      }
      toast.success("次回予定を確定しました");
      setConfirmTentativeTarget(null);
      return null;
    },
    [confirmTentativeMutation, confirmTentativeTarget],
  );

  // 「次回予定を再設定」ダイアログを開く（A4 キャンセル後 / A3 期限超過パネルの共通入口）。
  // Codex PR606 #18: PUT intent の BE 権限は「既存 active tentative の上書き =
  // ルート reservations:create + repo CanUpdateExisting（reservations:edit）の両方」
  // 「新規 intent = reservations:create」。対象行が active tentative なら既存行の
  // 更新なので canCreate && canEdit を要求し、キャンセル済み行からの再作成のみ
  // canCreate で通す（#18 再指摘: edit 単独では BE が 403 を返す）。
  const openResetNextVisit = useCallback(
    (reservation: Reservation) => {
      const overwritesActiveTentative = isTentativeReservation(reservation);
      const allowed = overwritesActiveTentative
        ? permissions?.canEdit === true && permissions?.canCreate === true
        : permissions?.canCreate === true;
      if (!allowed) {
        toast.error("次回予定を再設定する権限がありません");
        return;
      }
      if (!reservation.petId) {
        toast.error("ペット情報が取得できないため次回予定を再設定できません");
        return;
      }
      setResetNextVisitTarget({
        petId: reservation.petId,
        petName: reservation.petName,
        previousIntendedDate: reservation.intendedDate,
        sourceMedicalRecordId: reservation.sourceMedicalRecordId,
        clinicId: reservation.clinicId,
      });
    },
    [permissions],
  );

  const handleResetNextVisitSubmit = useCallback(
    async (date: string) => {
      const target = resetNextVisitTarget;
      if (!target) throw new Error("再設定対象の次回予定がありません");
      await setIntentMutation.mutateAsync({
        petId: target.petId,
        intendedDate: date,
        medicalRecordId: target.sourceMedicalRecordId,
        // Codex PR606 #5: 行の clinicId で送る（拠点横断の誤医院送信を防ぐ）
        clinicId: target.clinicId,
      });
      toast.success(`次回予定を ${date} に再設定しました`);
    },
    [resetNextVisitTarget, setIntentMutation],
  );

  const handleResetNextVisitClose = useCallback(() => setResetNextVisitTarget(null), []);

  // A3「終了」: tentative を DELETE next-visit-intent（pets スコープ・冪等）で終わらせる。
  // 実予約のキャンセルや削除ではなく「次回予定の意思表示を閉じる」操作。
  const handleEndTentativeRequest = useCallback(
    (reservation: Reservation) => {
      if (permissions?.canEdit !== true) {
        toast.error("次回予定を終了する権限がありません");
        return;
      }
      if (!isTentativeReservation(reservation) || !reservation.petId) return;
      setEndTentativeTarget(reservation);
    },
    [permissions],
  );

  const executeEndTentative = useCallback(async () => {
    const target = endTentativeTarget;
    if (!target?.petId) return;
    if (!target.updatedAt) {
      toast.error("次回予定の版が取得できません。再読み込みしてください");
      return;
    }
    try {
      // Codex PR606 #5: 行の clinicId で DELETE する（拠点横断の誤医院送信を防ぐ）
      await clearIntentMutation.mutateAsync({
        petId: target.petId,
        clinicId: target.clinicId,
        expectedReservationId: target.id,
        expectedUpdatedAt: target.updatedAt,
      });
      toast.success("次回予定を終了しました", {
        description: `${target.petName}（${target.ownerName}様）`,
      });
      setEndTentativeTarget(null);
    } catch {
      // onError(handleApiError) が toast 済み。ダイアログは開いたまま再試行可能。
    }
  }, [endTentativeTarget, clearIntentMutation]);

  const handleEndTentativeClose = useCallback(() => setEndTentativeTarget(null), []);

  const handleTimeSlotClick = useCallback(
    (date: Date) => {
      const stub: ReservationFormData = {
        start: date,
        end: addHours(date, 1),
        status: "confirmed",
        visitType: "first",
        doctor: "",
        isDesignated: false,
      };
      handleOpenForm(stub);
    },
    [handleOpenForm],
  );

  const {
    handleSave,
    handleReservationUpdate,
    handleStatusChange,
    executeStatusChange,
    executeDelete,
    handleCloseCreateForm,
  } = useReservationActions({
    appointments,
    editingAppointmentRef,
    deleteTarget,
    setDeleteConfirmOpen,
    setDeleteTarget,
    statusConfirmTarget,
    setStatusConfirmOpen,
    setStatusConfirmTarget,
    setDetailAppointment,
    handleCloseForm,
    handleCloseDetail,
    locationFrom,
    navigate,
    createMutations,
    permissions,
    // EMR-308 A4: tentative 昇格済み予約のキャンセル後に「次回予定の再設定」をオプション提示する。
    // 自動再作成はしない（スタッフが明示的に再設定を選んだ場合のみ PUT intent）。
    onTentativeCancelled: openResetNextVisit,
  });

  const handleDelete = useCallback((reservation: Reservation) => {
    setDeleteTarget(reservation);
    setDeleteConfirmOpen(true);
  }, []);

  const {
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handleCreateRecord,
    handlePetSelectConfirm,
  } = useReservationRecordNavigation({ navigate });

  const handleDeleteConfirmClose = useCallback(() => {
    setDeleteConfirmOpen(false);
    setDeleteTarget(null);
  }, []);

  const handleStatusConfirmClose = useCallback(() => {
    setStatusConfirmOpen(false);
    setStatusConfirmTarget(null);
  }, []);

  return {
    // Data
    appointments,
    isLoading,

    // Form modal
    isFormOpen,
    editingAppointment,
    handleOpenForm,
    handleCloseForm: handleCloseCreateForm,
    handleSave,

    // Detail modal
    isDetailOpen,
    detailAppointment,
    handleOpenDetail,
    handleCloseDetail,
    handleStatusChange,
    handleDelete,
    handleCreateRecord,

    // Calendar interactions
    handleTimeSlotClick,
    handleReservationUpdate,

    // Delete confirm
    deleteConfirmOpen,
    deleteTarget,
    executeDelete,
    handleDeleteConfirmClose,

    // Status confirm (BUG-020 destructive statuses)
    statusConfirmOpen,
    statusConfirmTarget,
    executeStatusChange,
    handleStatusConfirmClose,

    // Pet select confirm
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handlePetSelectConfirm,

    // EMR-308 A3: 期限超過 tentative 一覧（actionable panel 用）
    expiredTentativeAppointments,

    // EMR-308 A2/A3: tentative 確定モーダル
    confirmTentativeTarget,
    handleOpenConfirmTentative,
    handleCloseConfirmTentative,
    handleConfirmTentative,

    // EMR-308 A3/A4: 次回予定の再設定（キャンセル後オプション＋期限超過パネル「再設定」）
    resetNextVisitTarget,
    handleResetNextVisitRequest: openResetNextVisit,
    handleResetNextVisitSubmit,
    handleResetNextVisitClose,
    isResetNextVisitSubmitting: setIntentMutation.isPending,

    // EMR-308 A3: 期限超過 tentative の終了確認
    endTentativeTarget,
    handleEndTentativeRequest,
    executeEndTentative,
    handleEndTentativeClose,
    isEndTentativeSubmitting: clearIntentMutation.isPending,
  };
}
