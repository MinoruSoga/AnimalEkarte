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

interface IntentState {
  petId: string;
  clinicId: string | undefined;
  token: object;
  draft: string | null;
  serverDate: string;
  pending: number;
  deferredReset: boolean;
}

function createIntentState(
  petId: string,
  clinicId: string | undefined,
  serverDate: string,
): IntentState {
  return { petId, clinicId, token: {}, draft: null, serverDate, pending: 0, deferredReset: false };
}

interface IntentMutationQueue {
  promise: Promise<void>;
  generation: number;
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
  clearIntentMutation: AsyncMutation<string | { petId: string; clinicId: string }, unknown>;
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
  // ルートの create + repo の CanUpdateExisting（reservations:edit）を両方要する
  // （BE が repo ロック内で 403 を返す）。DELETE intent は reservations:edit。
  // 方向別に権限を分ける — fail-closed。
  // Codex PR606 #8: 既存 intent の有無は tentativeId の存在で判定する。
  // kind==="tentative" では実予約と tentative が同居するケース（kind==="appointment"）
  // を見落とし、create-only ユーザーが上書き PUT を発行してしまう。
  // Codex PR606 #8 再指摘: 既存 tentative の上書き PUT は canCreate && canEdit の
  // 両方を要求する（canEdit 単独では BE が 403 を返し fail-closed を破る）。
  const hasExistingIntent = nextVisit.tentativeId !== null;
  const canSetIntent = !isPetDeceased && (hasExistingIntent ? canCreate && canEdit : canCreate);
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

  // Codex PR606 #11: 同一ペットの intent mutation を FIFO 直列化する。
  // 連続入力で PUT→DELETE が逆順完了すると stale な intentDate が復活するため、
  // pet 単位の promise chain に乗せる（active intent はpetごとに1件）。
  // generation counter で「最新の要求だけが失敗時に draft を戻す」を保証する。
  // Codex PR606 #11 再指摘: pending 数でも先行 mutation の refetch による
  // draft リセットを抑制する。render 中に読むため ref ではなく state で持つ
  // （react-hooks/refs: render 中の ref アクセス禁止）。
  const [intentState, setIntentState] = useState(() =>
    createIntentState(petId, clinicId, nextVisit.intentDate),
  );
  const mutationQueuesRef = useRef(new Map<string, IntentMutationQueue>());
  const mutationQueueRef = useRef<{ token: object; queue: IntentMutationQueue } | null>(null);

  // 同じサーバー日付でも、別ペット/医院のdraft・pendingは引き継がない。
  // tokenはA→B→Aでも作り直し、最初のAの完了を再訪後のAへ適用しない。
  if (intentState.petId !== petId || intentState.clinicId !== clinicId) {
    setIntentState(createIntentState(petId, clinicId, nextVisit.intentDate));
  } else if (
    intentState.serverDate !== nextVisit.intentDate ||
    (intentState.pending === 0 && intentState.deferredReset)
  ) {
    // 同一scopeの先行refetchは後続draftを保持し、pending解消後に再同期する。
    setIntentState({
      ...intentState,
      serverDate: nextVisit.intentDate,
      draft: intentState.pending === 0 ? null : intentState.draft,
      deferredReset: intentState.pending > 0,
    });
  }

  const token = intentState.token;
  useLayoutEffect(() => {
    let queue = mutationQueuesRef.current.get(petId);
    if (!queue) {
      queue = { promise: Promise.resolve(), generation: 0 };
      mutationQueuesRef.current.set(petId, queue);
    }
    // UIはscope再訪でもfreshにするが、同一petへの既要求writeのFIFOは保つ。
    mutationQueueRef.current = { token, queue };
    return () => {
      mutationQueueRef.current = null;
    };
  }, [petId, token]);

  const isCurrentScope = useCallback(() => mutationQueueRef.current?.token === token, [token]);
  const updateIntentState = useCallback(
    (update: (state: IntentState) => IntentState) => {
      if (!isCurrentScope()) return;
      setIntentState((state) => (state.token === token ? update(state) : state));
    },
    [isCurrentScope, token],
  );
  const resetIntentDraft = useCallback(
    () => updateIntentState((state) => ({ ...state, draft: null })),
    [updateIntentState],
  );

  const enqueueIntentMutation = useCallback(
    (task: () => Promise<void>): number | null => {
      const active = mutationQueueRef.current;
      if (active?.token !== token) return null;
      const queue = active.queue;
      const gen = ++queue.generation;
      updateIntentState((state) => ({ ...state, pending: state.pending + 1 }));
      // task 内で例外を握り潰す前提 — chain が rejected のまま残らないよう
      // 呼び出し側（runSetIntent/runClearIntent）は必ず try/catch する。
      queue.promise = queue.promise.then(async () => {
        try {
          await task();
        } finally {
          updateIntentState((state) => ({ ...state, pending: state.pending - 1 }));
        }
      });
      return gen;
    },
    [token, updateIntentState],
  );

  const runSetIntent = useCallback(
    (date: string) => {
      const gen = enqueueIntentMutation(async () => {
        try {
          await setIntentMutation.mutateAsync({
            petId,
            intendedDate: date,
            medicalRecordId,
            ...(clinicId !== undefined ? { clinicId } : {}),
          });
          if (isCurrentScope()) toast.success(`次回予定を ${date} に設定しました`);
        } catch {
          // エラー表示は api hook の onError(handleApiError) 側に集約。
          // 失敗時は入力表示をサーバー値へ戻す — ただし後続の要求が
          // 積まれている場合はそちらの draft を上書きしない（#11）。
          if (isCurrentScope() && gen === mutationQueueRef.current?.queue.generation)
            resetIntentDraft();
        }
      });
    },
    [
      enqueueIntentMutation,
      petId,
      medicalRecordId,
      clinicId,
      setIntentMutation,
      isCurrentScope,
      resetIntentDraft,
    ],
  );

  const runClearIntent = useCallback(() => {
    const gen = enqueueIntentMutation(async () => {
      try {
        await clearIntentMutation.mutateAsync(clinicId !== undefined ? { petId, clinicId } : petId);
        if (isCurrentScope()) toast.success("次回予定をクリアしました");
      } catch {
        if (isCurrentScope() && gen === mutationQueueRef.current?.queue.generation)
          resetIntentDraft();
      }
    });
  }, [
    enqueueIntentMutation,
    petId,
    clinicId,
    clearIntentMutation,
    isCurrentScope,
    resetIntentDraft,
  ]);

  // NextVisitDateField の onChange にそのまま渡す。
  // "" はクリア（DELETE）、有効日は upsert（PUT）、無効日は入力表示だけ残し mutation しない。
  const handleIntentDateChange = useCallback(
    (newDate: string) => {
      if (!isCurrentScope()) return;
      if (newDate === "") {
        if (!intentCapsRef.current.clear) return;
        updateIntentState((state) => ({ ...state, draft: newDate }));
        void runClearIntent();
        return;
      }
      if (!intentCapsRef.current.set) return;
      updateIntentState((state) => ({ ...state, draft: newDate }));
      if (validateNextVisitIntentDate(newDate) !== null) return;
      void runSetIntent(newDate);
    },
    [runClearIntent, runSetIntent, isCurrentScope, updateIntentState],
  );

  return {
    nextVisitLabel: nextVisit.label,
    intentFieldValue: intentState.draft ?? nextVisit.intentDate,
    handleIntentDateChange,
    resetIntentDraft,
    canMutateIntent: canSetIntent || canClearIntent,
    canSetIntent,
    canClearIntent,
    /** EMR-308 A1: 確定導線。kind==="tentative" のときのみ true（reservations:edit 必要）。 */
    canConfirmTentative,
    /** 確定モーダルの対象 tentative 行。なければ null。 */
    tentativeReservation,
  };
}
