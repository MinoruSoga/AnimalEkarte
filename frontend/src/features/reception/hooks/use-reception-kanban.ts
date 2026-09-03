import { useState, useMemo, useRef, useCallback, useLayoutEffect, useTransition } from "react";
import { toast } from "sonner";
import type { ColumnData, ReservationStatus } from "@/types";
import { PetStatusDeceased } from "@/types/generated/models";
// bundle-barrel-imports: barrel経由ではなく各ファイルから直接import
import { useGetReception, todayISO } from "../api/get-reception";
import { useGetStaffs, buildStaffMap } from "../api/get-staffs";
import { useUpdateAppointmentStatus } from "../api/update-appointment-status";
import { COLUMN_TITLE_TO_STATUS, RECEPTION_COLUMNS } from "../api/transforms";
import type { ReceptionColumn, ReceptionAppointment } from "../api/types";
import {
  captureCardSnapshot,
  findAppointment,
  mergeCard,
  relocateCard,
  removeCard,
  restoreCardSnapshot,
} from "../lib/kanban-columns";
import type { CardSnapshot } from "../lib/kanban-columns";

/** カラム遷移の状態機械: 現在カラム title → 次カラム title。会計済 は terminal（昇格先なし）。 */
const NEXT_COLUMN_TITLE: Record<string, string> = {
  受付予約: "受付済",
  受付済: "診療中",
  診療中: "会計待ち",
  会計待ち: "会計済",
};

/** terminal カラム: ここから advance すると completed 確定 + ローカル除外になる */
const TERMINAL_COLUMN_TITLE = "会計済";

/** 受付済→診療中 の直接ドラッグ警告を再表示するまでの最短間隔（連打での toast 多重表示を防ぐ） */
const INVALID_DRAG_ALERT_THROTTLE_MS = 3000;
/** 上記警告 toast の表示時間 */
const INVALID_DRAG_ALERT_DURATION_MS = 4000;

/** ReceptionAppointment にスタッフ名解決を適用して返す */
function toAppointment(appt: ReceptionAppointment, staffMap: Map<string, string>): ReceptionAppointment {
  return {
    ...appt,
    // doctor_id（UUID）をスタッフ名に変換。未登録IDの場合はUUIDをそのまま表示
    doctor: appt.doctor ? (staffMap.get(appt.doctor) ?? appt.doctor) : undefined,
    doctorId: appt.doctorId || appt.doctor || "",
  };
}

/** ReceptionColumn → ColumnData（@/types）変換 */
function toColumnData(col: ReceptionColumn, staffMap: Map<string, string>): ColumnData {
  return {
    title: col.title,
    appointments: col.appointments.map((appt) => toAppointment(appt, staffMap)),
  };
}

interface UseReceptionKanbanPermissions {
  canEditReservation?: boolean;
  canDeleteReservation?: boolean;
}

type ReservationMutationAction = "edit" | "delete";

export function useReceptionKanban({
  canEditReservation,
  canDeleteReservation,
}: UseReceptionKanbanPermissions) {
  const today = useMemo(() => todayISO(), []);
  const { data: apiColumns, isLoading, isError } = useGetReception(today);
  const { data: staffs } = useGetStaffs();
  const updateStatusMutation = useUpdateAppointmentStatus();
  const { mutateAsync } = updateStatusMutation;
  // rerender-transitions: API mutation の pending 管理に useTransition を使用
  // （useState(false) + setIsPending パターンは try-finally でリセット漏れが起きるため禁止）
  const [isUpdatingStatus, startUpdateStatusTransition] = useTransition();
  const permissionsRef = useRef({ canEditReservation, canDeleteReservation });
  useLayoutEffect(() => {
    permissionsRef.current = { canEditReservation, canDeleteReservation };
  }, [canEditReservation, canDeleteReservation]);

  const hasMutationPermission = useCallback((action: ReservationMutationAction): boolean => {
    return action === "edit"
      ? permissionsRef.current.canEditReservation === true
      : permissionsRef.current.canDeleteReservation === true;
  }, []);

  // staffId → スタッフ名のMap（APIレスポンス変換で使用）。
  // staffs が undefined（ローディング中）の場合は空配列で buildStaffMap を呼ぶ。
  // インライン `= []` デフォルトは毎レンダーで新規参照を生成し useMemo が無限再計算するため禁止。
  const staffMap = useMemo(() => buildStaffMap(staffs ?? []), [staffs]);

  // API データを ColumnData[] に変換。ローディング中は空カラムを表示
  const apiColumnData: ColumnData[] = useMemo(() => {
    if (!apiColumns) {
      return RECEPTION_COLUMNS.map((col) => ({ title: col.title, appointments: [] }));
    }
    return apiColumns.map((col) => toColumnData(col, staffMap));
  }, [apiColumns, staffMap]);

  // ローカル状態: ドラッグ操作のために API データを元に保持
  const [prevApiColumns, setPrevApiColumns] = useState<ColumnData[]>([]);
  const [columns, setColumns] = useState<ColumnData[]>([]);
  const columnsRef = useRef(columns);
  useLayoutEffect(() => {
    columnsRef.current = columns;
  }, [columns]);

  // rerender-derived-state-no-effect: effect を使わずレンダー内でインライン同期。
  // APIデータが更新されたときのみ columns をリセット（参照比較で余分な更新を防ぐ）。
  if (prevApiColumns !== apiColumnData) {
    setPrevApiColumns(apiColumnData);
    setColumns(apiColumnData);
  }

  // Filter States
  const [selectedVisitTypes, setSelectedVisitTypes] = useState<string[]>(["初診", "再診"]);
  const [selectedDoctor, setSelectedDoctor] = useState<string>("all");
  const [isTrimmingOnly, setIsTrimmingOnly] = useState(false);

  const lastAlertRef = useRef(0);

  // js-set-map-lookups: O(1) ルックアップのため Set を事前構築
  const selectedVisitTypeSet = useMemo(() => new Set(selectedVisitTypes), [selectedVisitTypes]);

  // Filter Logic
  const filteredColumns = useMemo(() => {
    return columns.map((col) => ({
      ...col,
      appointments: col.appointments.filter((app) => {
        // 1. Visit Type Filter
        if (!selectedVisitTypeSet.has(app.visitType)) return false;

        // 2. Doctor/Designation Filter
        if (selectedDoctor !== "all") {
          if (selectedDoctor === "医師指名なし") {
            if (app.isDesignated) return false;
          } else if (app.doctor !== selectedDoctor || !app.isDesignated) {
            return false;
          }
        }

        // 3. Trimming Filter
        if (isTrimmingOnly && app.reservationCategory !== "trimming") return false;

        return true;
      }),
    }));
  }, [columns, selectedVisitTypeSet, selectedDoctor, isTrimmingOnly]);

  // rerender-dependencies: filteredColumns（オブジェクト配列）を ref で保持し mutator の deps から除外する
  const filteredColumnsRef = useRef(filteredColumns);
  useLayoutEffect(() => {
    filteredColumnsRef.current = filteredColumns;
  }, [filteredColumns]);

  const toggleVisitType = useCallback((type: string) => {
    setSelectedVisitTypes((prev) =>
      prev.includes(type) ? prev.filter((t) => t !== type) : [...prev, type],
    );
  }, []);

  /**
   * ステータス変更 API を useTransition でラップする共通ラッパ。
   * - 成功時: onSuccess を実行
   * - 失敗時: 対象カードだけを操作前 snapshot へ戻す（通知は useUpdateAppointmentStatus 側の onError に一本化）
   */
  const runStatusMutation = useCallback(
    (
      id: string,
      status: ReservationStatus,
      opts: {
        action: ReservationMutationAction;
        rollbackSnapshot?: CardSnapshot | null;
        onSuccess?: () => void;
      },
    ): Promise<boolean> => {
      if (!hasMutationPermission(opts.action)) return Promise.resolve(false);
      if (findAppointment(columnsRef.current, id)?.petStatus === PetStatusDeceased) {
        return Promise.resolve(false);
      }

      return new Promise<boolean>((resolve) => {
        startUpdateStatusTransition(async () => {
          try {
            await mutateAsync({ id, status });
            opts.onSuccess?.();
            resolve(true);
          } catch {
            // useUpdateAppointmentStatus の onError が handleApiError 済み。ここでは再通知しない。
            const rollbackSnapshot = opts.rollbackSnapshot;
            if (rollbackSnapshot) {
              setColumns((currentColumns) =>
                restoreCardSnapshot(currentColumns, rollbackSnapshot),
              );
            }
            resolve(false);
          }
        });
      });
    },
    [hasMutationPermission, mutateAsync, startUpdateStatusTransition],
  );

  /** カードをドラッグで sourceColumn → targetColumn へ移動する。受付済→診療中 直行は禁止。 */
  const moveCard = useCallback(
    (hoverIndex: number, sourceColumn: string, targetColumn: string, cardId: string): boolean => {
      if (!hasMutationPermission("edit")) return false;

      const sourceColFiltered = filteredColumnsRef.current.find((c) => c.title === sourceColumn);
      const targetColFiltered = filteredColumnsRef.current.find((c) => c.title === targetColumn);
      if (!sourceColFiltered || !targetColFiltered) return false;
      const appointment = sourceColFiltered.appointments.find((item) => item.id === cardId);
      if (!appointment || appointment.petStatus === PetStatusDeceased) return false;

      // 受付済 → 診療中 への直接ドラッグは禁止（カルテ作成が必要）
      if (sourceColumn === "受付済" && targetColumn === "診療中") {
        const now = Date.now();
        if (now - lastAlertRef.current > INVALID_DRAG_ALERT_THROTTLE_MS) {
          toast.error("カルテ作成が必要です", {
            description: "このステータスに変更するには、詳細画面からカルテを作成してください。",
            duration: INVALID_DRAG_ALERT_DURATION_MS,
          });
          lastAlertRef.current = now;
        }
        return false;
      }

      // カラムが変わった場合は API でステータス更新
      if (sourceColumn !== targetColumn) {
        const newStatus = COLUMN_TITLE_TO_STATUS[targetColumn];
        if (newStatus) {
          const rollbackSnapshot = captureCardSnapshot(columnsRef.current, cardId);
          void runStatusMutation(cardId, newStatus, {
            action: "edit",
            rollbackSnapshot,
          });
        }
      }

      // 楽観的 UI 更新。filtered target の参照カードを raw target にマップして挿入位置を決める。
      setColumns((prev) =>
        relocateCard(prev, cardId, sourceColumn, targetColumn, (targetAppointments) => {
          if (hoverIndex < targetColFiltered.appointments.length) {
            const referenceCard = targetColFiltered.appointments[hoverIndex];
            const refIndex = targetAppointments.findIndex((a) => a.id === referenceCard.id);
            if (refIndex !== -1) return refIndex;
          }
          return targetAppointments.length;
        }) ?? prev,
      );
      return true;
    },
    [hasMutationPermission, runStatusMutation],
  );

  /** ボタン操作でカードを次ステータスへ進める。会計済 は completed 確定 + ローカル除外。 */
  const advanceStatus = useCallback(
    (appointment: ReceptionAppointment) => {
      if (!hasMutationPermission("edit")) return;
      const currentColumn = filteredColumnsRef.current.find((column) =>
        column.appointments.some((item) => item.id === appointment.id),
      );
      const currentAppointment = currentColumn?.appointments.find(
        (item) => item.id === appointment.id,
      );
      if (!currentColumn || !currentAppointment) return;
      if (currentAppointment.petStatus === PetStatusDeceased) return;
      const currentTitle = currentColumn.title;

      // 会計済（terminal）: completed で確定し、リストから除外する
      if (currentTitle === TERMINAL_COLUMN_TITLE) {
        const rollbackSnapshot = captureCardSnapshot(columnsRef.current, appointment.id);
        setColumns((prev) => removeCard(prev, appointment.id) ?? prev);
        void runStatusMutation(appointment.id, "completed", {
          action: "edit",
          rollbackSnapshot,
          onSuccess: () => toast.success("手続きを完了し、リストから削除しました"),
        });
        return;
      }

      const nextTitle = NEXT_COLUMN_TITLE[currentTitle];
      if (!nextTitle) return;
      const newStatus = COLUMN_TITLE_TO_STATUS[nextTitle];
      if (!newStatus) return;

      const rollbackSnapshot = captureCardSnapshot(columnsRef.current, appointment.id);
      void runStatusMutation(appointment.id, newStatus, {
        action: "edit",
        rollbackSnapshot,
        onSuccess: () => {
          toast.success(`ステータスを「${nextTitle}」に変更しました`, {
            description: `${appointment.petName}ちゃんのステータスを更新しました。`,
          });
        },
      });

      // currentTitle は filtered で存在確認済みのため relocate は成立する
      setColumns((prev) => relocateCard(prev, appointment.id, currentTitle, nextTitle) ?? prev);
    },
    [hasMutationPermission, runStatusMutation],
  );

  const cancelAppointment = useCallback(
    (appointmentId: string): Promise<boolean> => {
      if (!hasMutationPermission("delete")) return Promise.resolve(false);
      if (
        findAppointment(columnsRef.current, appointmentId)?.petStatus === PetStatusDeceased
      ) return Promise.resolve(false);
      const rollbackSnapshot = captureCardSnapshot(columnsRef.current, appointmentId);
      if (!rollbackSnapshot) return Promise.resolve(false);
      // 楽観的: ローカルからも除外
      setColumns((prev) => removeCard(prev, appointmentId) ?? prev);
      return runStatusMutation(appointmentId, "cancelled", {
        action: "delete",
        rollbackSnapshot,
      });
    },
    [hasMutationPermission, runStatusMutation],
  );

  const updateAppointment = useCallback((updatedAppointment: ReceptionAppointment) => {
    if (!hasMutationPermission("edit")) return;
    if (
      findAppointment(columnsRef.current, updatedAppointment.id)?.petStatus === PetStatusDeceased
    ) return;
    setColumns((prev) => mergeCard(prev, updatedAppointment) ?? prev);
    toast.success("予約情報を更新しました");
  }, [hasMutationPermission]);

  return {
    columns, // ローカル状態（ポーリング影響を避けるためドラッグハンドラで使用）
    filteredColumns,
    isLoading,
    isError,
    isUpdatingStatus,
    staffs: staffs ?? [],
    moveCard,
    advanceStatus,
    cancelAppointment,
    updateAppointment,
    filters: {
      selectedVisitTypes,
      selectedDoctor,
      isTrimmingOnly,
      setSelectedDoctor,
      setIsTrimmingOnly,
      toggleVisitType,
    },
  };
}
