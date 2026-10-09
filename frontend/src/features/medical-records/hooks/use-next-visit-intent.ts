import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { toast } from "sonner";

import { usePermission } from "@/hooks/use-permission";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { todayJSTISO, toJSTWallDate } from "@/lib/jst-date";

import type { SetNextVisitIntentVariables } from "../api/next-visit-intent";
import {
  deriveNextVisitState,
  nextVisitFetchEndDate,
  validateNextVisitIntentDate,
} from "../lib/next-visit";

interface AsyncMutation<TVariables, TData> {
  mutateAsync: (variables: TVariables) => Promise<TData>;
}

interface UseNextVisitIntentParams {
  petId: string;
  /** 保存済みカルテがあればその id。PUT 時に source_medical_record_id として送る。 */
  medicalRecordId?: string;
  /** EMR-308/Codex: 拠点横断カルテではレコード自身の clinicId を指定する。
   *  query の X-Clinic-ID に使い、mutation 側は呼び出し元が同値で構築済みを前提とする。 */
  clinicId?: string;
  isPetDeceased: boolean;
  setIntentMutation: AsyncMutation<SetNextVisitIntentVariables, unknown>;
  clearIntentMutation: AsyncMutation<string, unknown>;
}

/**
 * EMR-308: ヘッダー「次回予定」コントロールの状態・mutation 配線。
 * 表示は GET /v1/reservations（useGetReservations 共有フック）から導出し、
 * set/clear は PUT/DELETE /v1/reservations/pets/:petId/next-visit-intent を呼ぶ。
 * mutation オブジェクトは use-medical-record-auto-create 先例の AsyncMutation
 * インターフェース注入で、テストから差し替え可能にする。
 */
export function useNextVisitIntent({
  petId,
  medicalRecordId,
  clinicId,
  isPetDeceased,
  setIntentMutation,
  clearIntentMutation,
}: UseNextVisitIntentParams) {
  // TASK-444-S1: generated/models 直 import 禁止のためリソースは文字列リテラルで渡す
  const { canCreate, canEdit } = usePermission("reservations");

  const { data: reservations } = useGetReservations({
    petId,
    startDate: todayJSTISO(),
    endDate: nextVisitFetchEndDate(),
    enabled: petId !== "",
    clinicId,
  });

  const nextVisit = deriveNextVisitState(reservations ?? [], toJSTWallDate(new Date()));

  // PUT intent は upsert: 新規は reservations:create、既存 intent の上書きは
  // reservations:edit を要する（BE が repo ロック内で 403 を返す）。DELETE intent は
  // reservations:edit。方向別に権限を分ける — fail-closed。
  // Codex PR606 #8: 既存 intent の有無は tentativeId の存在で判定する。
  // kind==="tentative" では実予約と tentative が同居するケース（kind==="appointment"）
  // を見落とし、create-only ユーザーが上書き PUT を発行してしまう。
  const hasExistingIntent = nextVisit.tentativeId !== null;
  const canSetIntent = !isPetDeceased && (hasExistingIntent ? canEdit : canCreate);
  const canClearIntent = canEdit && !isPetDeceased;
  const intentCapsRef = useRef({ set: canSetIntent, clear: canClearIntent });
  useLayoutEffect(() => {
    intentCapsRef.current = { set: canSetIntent, clear: canClearIntent };
  }, [canSetIntent, canClearIntent]);

  // EMR-308 A1: 確定導線用に tentative 行そのもの（start/visitType/doctor 等を含む
  // transformed Reservation）を露出する。id だけでなく行全体を返すのは、
  // buildTentativeConfirmInitialData が Reservation のフィールドをそのまま使うため。
  const tentativeReservation =
    (nextVisit.tentativeId
      ? (reservations ?? []).find((r) => r.id === nextVisit.tentativeId)
      : undefined) ?? null;
  // POST confirm-tentative は reservations:edit。実予約が既にある場合は
  // tentative を表示だけに留め、確定導線は出さない（二重予約を防ぐ）。
  const canConfirmTentative =
    canEdit && !isPetDeceased && nextVisit.kind === "tentative" && tentativeReservation !== null;

  // Popover 入力のドラフト。通常はサーバー値（tentative の意図日）に従い、
  // 入力中はタイプした値を保持してバリデーション表示を可能にする。
  // サーバー値が変わったらドラフトを破棄する（レンダー中の派生状態リセット、
  // MedicalRecordFormReadyPanels の prev-record-name パターンと同型）。
  const [intentDraft, setIntentDraft] = useState<string | null>(null);
  const [prevIntentDate, setPrevIntentDate] = useState(nextVisit.intentDate);
  if (prevIntentDate !== nextVisit.intentDate) {
    setPrevIntentDate(nextVisit.intentDate);
    setIntentDraft(null);
  }

  const resetIntentDraft = useCallback(() => setIntentDraft(null), []);

  // Codex PR606 #11: 同一ペットの intent mutation を FIFO 直列化する。
  // 連続入力で PUT→DELETE が逆順完了すると stale な intentDate が復活するため、
  // hook インスタンス（=1ペット）単位の promise chain に乗せる。
  // generation counter で「最新の要求だけが失敗時に draft を戻す」を保証する。
  const mutationQueueRef = useRef<Promise<unknown>>(Promise.resolve());
  const mutationGenRef = useRef(0);
  const enqueueIntentMutation = useCallback((task: () => Promise<void>): number => {
    const gen = ++mutationGenRef.current;
    // task 内で例外を握り潰す前提 — chain が rejected のまま残らないよう
    // 呼び出し側（runSetIntent/runClearIntent）は必ず try/catch する。
    mutationQueueRef.current = mutationQueueRef.current.then(task);
    return gen;
  }, []);

  const runSetIntent = useCallback(
    (date: string) => {
      const gen = enqueueIntentMutation(async () => {
        try {
          await setIntentMutation.mutateAsync({ petId, intendedDate: date, medicalRecordId });
          toast.success(`次回予定を ${date} に設定しました`);
        } catch {
          // エラー表示は api hook の onError(handleApiError) 側に集約。
          // 失敗時は入力表示をサーバー値へ戻す — ただし後続の要求が
          // 積まれている場合はそちらの draft を上書きしない（#11）。
          if (gen === mutationGenRef.current) setIntentDraft(null);
        }
      });
    },
    [enqueueIntentMutation, petId, medicalRecordId, setIntentMutation],
  );

  const runClearIntent = useCallback(() => {
    const gen = enqueueIntentMutation(async () => {
      try {
        await clearIntentMutation.mutateAsync(petId);
        toast.success("次回予定をクリアしました");
      } catch {
        if (gen === mutationGenRef.current) setIntentDraft(null);
      }
    });
  }, [enqueueIntentMutation, petId, clearIntentMutation]);

  // NextVisitDateField の onChange にそのまま渡す。
  // "" はクリア（DELETE）、有効日は upsert（PUT）、無効日は入力表示だけ残し mutation しない。
  const handleIntentDateChange = useCallback(
    (newDate: string) => {
      setIntentDraft(newDate);
      if (newDate === "") {
        if (!intentCapsRef.current.clear) return;
        void runClearIntent();
        return;
      }
      if (!intentCapsRef.current.set) return;
      if (validateNextVisitIntentDate(newDate) !== null) return;
      void runSetIntent(newDate);
    },
    [runClearIntent, runSetIntent],
  );

  return {
    nextVisitLabel: nextVisit.label,
    intentFieldValue: intentDraft ?? nextVisit.intentDate,
    handleIntentDateChange,
    resetIntentDraft,
    canMutateIntent: canSetIntent || canClearIntent,
    /** EMR-308 A1: 確定導線。kind==="tentative" のときのみ true（reservations:edit 必要）。 */
    canConfirmTentative,
    /** 確定モーダルの対象 tentative 行。なければ null。 */
    tentativeReservation,
  };
}
