import { useState, useCallback, useLayoutEffect, useMemo, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { usePermission } from "@/hooks/use-permission";
import { useClinicTaxRates } from "@/hooks/use-clinic-tax-rates";
import { useGetAllMedicinesMaster } from "@/hooks/use-treatment-master";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

import { useGetTreatments } from "../api/treatments";
import { useCreateTreatment } from "../api/treatments";
import { useUpdateTreatment } from "../api/treatments";
import type { TreatmentUpdateOutcome } from "../types";
import { useDeleteTreatment } from "../api/treatments";
import { useReorderTreatments } from "../api/treatments";
import { useGetVitals } from "../api/vitals";
import { medicineDoseParamsQueryKey, resolveLatestVitalWeight } from "../api/medicine-dose-lookup";
import { resolveMasterSelectionDose } from "../lib/master-selection-dose";
import type { TreatmentItemType, UpdateTreatmentInput } from "../types";
import { buildMasterSelectionPayload } from "../lib/treatments-tab-model";

export function useTreatmentsTab({
  medicalRecordId,
  ownerDiscountRate = 0,
  petSpecies,
  recordClinicId,
  isWriteBlocked = false,
}: {
  medicalRecordId: string;
  ownerDiscountRate?: number;
  petSpecies?: string | null;
  recordClinicId?: string;
  /** PR679 F5: 確定/ロック/死亡による書込ロック。フック内の全 mutation 経路を fail-closed で止める。 */
  isWriteBlocked?: boolean;
}) {
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  // BUG-372: 割引権限（値引額編集制御）
  const { canEdit: canEditDiscount } = usePermission("discount");
  // FE-RC-048: 消費税率はハードコード 0.1 ではなく病院マスタ設定を正本にする。
  const { standardTaxRate } = useClinicTaxRates();
  const { data: treatments, isLoading } = useGetTreatments(medicalRecordId, recordClinicId);
  const createMutation = useCreateTreatment(medicalRecordId, recordClinicId);
  const { mutate: createTreatmentFn } = createMutation;
  const updateMutation = useUpdateTreatment(medicalRecordId, recordClinicId);
  const { mutateAsync: updateTreatmentFn } = updateMutation;
  const deleteMutation = useDeleteTreatment(medicalRecordId, recordClinicId);
  const { mutate: deleteTreatmentFn } = deleteMutation;
  const reorderMutation = useReorderTreatments(medicalRecordId, recordClinicId);
  const { mutate: reorderTreatmentsFn } = reorderMutation;
  const queryClient = useQueryClient();

  // PR679 F5: useCallback の stale closure 対策。コミット直後に最新の権限/ロックを
  // ref へ同期し、await 後やキャプチャ済みコールバックからも最新値で弾く。
  const latestWriteGateRef = useRef({
    canCreate,
    canEdit,
    canDelete,
    isWriteBlocked,
    medicalRecordId,
    recordClinicId,
    petSpecies,
    nextSortOrder: 0,
  });
  useLayoutEffect(() => {
    latestWriteGateRef.current = {
      canCreate,
      canEdit,
      canDelete,
      isWriteBlocked,
      medicalRecordId,
      recordClinicId,
      petSpecies,
      nextSortOrder: treatments?.length ? Math.max(...treatments.map((t) => t.sort_order)) + 1 : 0,
    };
  });

  // #201: 投与量自動計算の入力（当日 vital の体重・薬マスタ一覧）。
  const { data: vitals } = useGetVitals(medicalRecordId, recordClinicId);
  const { data: medicines } = useGetAllMedicinesMaster({ clinicId: recordClinicId });
  const resolvedWeight = useMemo(
    () => (vitals ? resolveLatestVitalWeight(vitals) : null),
    [vitals],
  );
  // react-review-201 HIGH-1: オブジェクトリテラルを直接 props に渡すと毎レンダー新規参照になり
  // TreatmentRow の React.memo が無効化される（全行が無関係な state 変更でも再レンダーされる）。
  const doseContext = useMemo(
    () => ({
      medicines,
      petSpecies: petSpecies ?? null,
      weightKg: resolvedWeight?.weightKg ?? null,
      clinicId: recordClinicId,
    }),
    [medicines, petSpecies, resolvedWeight, recordClinicId],
  );

  // マスタ検索ダイアログ
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  // 追加フォームの状態
  const [addItemType, setAddItemType] = useState<TreatmentItemType>("consultation");
  const [addContent, setAddContent] = useState("");
  const [addAdminRoute, setAddAdminRoute] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  // 直前の追加後に最終行へ自動フォーカスするフラグ
  const [focusLastRow, setFocusLastRow] = useState(false);

  // #201: マスタ選択時に絶対上限超過を検出した場合のインライン表示。
  const [masterDoseBlockReason, setMasterDoseBlockReason] = useState("");
  // TASK-025: technical failure 時のみ再試行対象のマスタ選択を保持する。
  // 上限超過ブロックでは再試行しても同一結果になるため保持しない。
  const [pendingMasterLookupItem, setPendingMasterLookupItem] =
    useState<TreatmentMasterItem | null>(null);

  // sort_order 昇順でソート済みリスト
  const sortedTreatments = useMemo(() => {
    if (!treatments) return [];
    return [...treatments].sort((a, b) => a.sort_order - b.sort_order);
  }, [treatments]);

  // 合計金額 (selected のみ)
  const { selectedSubtotal, selectedCount, finalTotal } = useMemo(() => {
    const selected = sortedTreatments.filter((t) => t.is_selected);
    const sub = selected.reduce((sum, t) => sum + t.unit_price * t.quantity - t.discount_amount, 0);
    // 飼主割引適用
    const ownerDiscount = Math.floor(sub * (ownerDiscountRate / 100));
    const afterDiscount = sub - ownerDiscount;
    const tax = Math.floor(afterDiscount * standardTaxRate);

    return {
      selectedSubtotal: sub,
      selectedCount: selected.length,
      finalTotal: afterDiscount + tax,
    };
  }, [sortedTreatments, ownerDiscountRate, standardTaxRate]);

  // 全明細の合計
  const totalSubtotal = useMemo(
    () =>
      sortedTreatments.reduce((sum, t) => sum + t.unit_price * t.quantity - t.discount_amount, 0),
    [sortedTreatments],
  );

  // ── handlers ──

  // EMR-488: PATCH 結果を呼出し元へ返す（行内エディタが解決後に閉じ、失敗時に下書きを保持する）。
  // mutateAsync の reject をここで "failed" に畳むため、fire-and-forget の呼出し元でも
  // unhandled rejection が発生しない（Codex review P2）。エラートーストは mutation の onError が表示する。
  const handleUpdate = useCallback(
    (
      treatmentId: string,
      input: UpdateTreatmentInput,
    ): Promise<TreatmentUpdateOutcome> | undefined => {
      const gate = latestWriteGateRef.current;
      if (gate.isWriteBlocked || !gate.canEdit) return Promise.resolve("skipped" as const);
      const target = sortedTreatments.find((t) => t.id === treatmentId);
      // UAT-R2-EXCLUSIVE-LOCK: version 未確定のまま送ると BE は CAS 照合をスキップするため
      // fail-closed で拒否する（clinical_plan の保存アクションと同じ方針）。
      if (typeof target?.version !== "number") {
        toast.error("治療明細の読み込みが完了してから保存してください");
        return Promise.resolve("skipped" as const);
      }
      return updateTreatmentFn({ treatmentId, input: { ...input, version: target.version } }).then(
        () => "saved" as const,
        () => "failed" as const,
      );
    },
    [sortedTreatments, updateTreatmentFn],
  );

  const handleDelete = useCallback(
    (treatmentId: string) => {
      const gate = latestWriteGateRef.current;
      if (gate.isWriteBlocked || !gate.canDelete) return;
      deleteTreatmentFn(treatmentId);
    },
    [deleteTreatmentFn],
  );

  const handleMoveUp = useCallback(
    (treatmentId: string) => {
      const gate = latestWriteGateRef.current;
      if (gate.isWriteBlocked || !gate.canEdit) return;
      const list = sortedTreatments;
      const idx = list.findIndex((t) => t.id === treatmentId);
      if (idx <= 0) return;
      const newList = list.map((t, i) => {
        if (i === idx - 1) return { id: t.id, sort_order: list[idx].sort_order };
        if (i === idx) return { id: t.id, sort_order: list[idx - 1].sort_order };
        return { id: t.id, sort_order: t.sort_order };
      });
      reorderTreatmentsFn({ treatments: newList });
    },
    [sortedTreatments, reorderTreatmentsFn],
  );

  const handleMoveDown = useCallback(
    (treatmentId: string) => {
      const gate = latestWriteGateRef.current;
      if (gate.isWriteBlocked || !gate.canEdit) return;
      const list = sortedTreatments;
      const idx = list.findIndex((t) => t.id === treatmentId);
      if (idx < 0 || idx >= list.length - 1) return;
      const newList = list.map((t, i) => {
        if (i === idx) return { id: t.id, sort_order: list[idx + 1].sort_order };
        if (i === idx + 1) return { id: t.id, sort_order: list[idx].sort_order };
        return { id: t.id, sort_order: t.sort_order };
      });
      reorderTreatmentsFn({ treatments: newList });
    },
    [sortedTreatments, reorderTreatmentsFn],
  );

  // rerender-memo: TreatmentRow(TreatmentsTable 経由)/TreatmentAddControls へ渡す
  // ハンドラを useCallback で安定化する（inline arrow は毎レンダー新規参照になる）。
  const handleAutoFocusDone = useCallback(() => {
    setFocusLastRow(false);
  }, []);

  const handleOpenSearch = useCallback(() => {
    if (latestWriteGateRef.current.isWriteBlocked) return;
    setIsSearchOpen(true);
  }, []);

  const handleStartAdding = useCallback(() => {
    if (latestWriteGateRef.current.isWriteBlocked) return;
    setIsAdding(true);
  }, []);

  const handleAddSubmit = useCallback(() => {
    const gate = latestWriteGateRef.current;
    if (gate.isWriteBlocked || !gate.canCreate || !addContent.trim()) return;
    const nextOrder =
      sortedTreatments.length > 0
        ? sortedTreatments[sortedTreatments.length - 1].sort_order + 1
        : 0;
    // 薬品の場合、投与方法を memo に付記する（BE-MEDI-010: 専用カラム実装まで）
    const memoWithRoute =
      addItemType === "medicine" && addAdminRoute ? `[投与方法: ${addAdminRoute}]` : "";
    createTreatmentFn(
      {
        item_type: addItemType,
        content: addContent.trim(),
        unit_price: 0,
        quantity: 1,
        is_selected: true,
        is_insurance: false,
        discount_amount: 0,
        sort_order: nextOrder,
        memo: memoWithRoute,
      },
      {
        onSuccess: () => {
          setFocusLastRow(true);
          setAddContent("");
          setAddAdminRoute("");
          setIsAdding(false);
        },
      },
    );
  }, [addItemType, addContent, addAdminRoute, sortedTreatments, createTreatmentFn]);

  const handleAddCancel = useCallback(() => {
    setAddContent("");
    setAddAdminRoute("");
    setIsAdding(false);
  }, []);

  const handleAddItemTypeChange = useCallback((itemType: TreatmentItemType) => {
    setAddItemType(itemType);
    setAddAdminRoute("");
  }, []);

  const handleSelectFromMaster = useCallback(
    async (item: TreatmentMasterItem) => {
      if (latestWriteGateRef.current.isWriteBlocked || !latestWriteGateRef.current.canCreate) {
        return;
      }
      if (
        latestWriteGateRef.current.medicalRecordId !== medicalRecordId ||
        latestWriteGateRef.current.recordClinicId !== recordClinicId ||
        latestWriteGateRef.current.petSpecies !== petSpecies
      )
        return;
      const doseResult = await resolveMasterSelectionDose({
        item,
        medicalRecordId: medicalRecordId ?? "",
        recordClinicId,
        petSpecies,
        queryClient,
      });
      const earlyGate = latestWriteGateRef.current;
      if (
        earlyGate.isWriteBlocked ||
        !earlyGate.canCreate ||
        earlyGate.medicalRecordId !== medicalRecordId ||
        earlyGate.recordClinicId !== recordClinicId ||
        earlyGate.petSpecies !== petSpecies
      ) {
        return;
      }
      if (doseResult.kind === "blocked") {
        setMasterDoseBlockReason(doseResult.message);
        setPendingMasterLookupItem(doseResult.retryable ? item : null);
        return;
      }
      const quantity = doseResult.quantity;
      setMasterDoseBlockReason("");
      setPendingMasterLookupItem(null);
      // PR679 F5: await(dose params)の間に確定/ロック/死亡/権限喪失が起き得るため
      // create 直前に最新値を再検査する（コンポーネント wrapper だけでは防げない）。
      const latestGate = latestWriteGateRef.current;
      if (
        latestGate.isWriteBlocked ||
        !latestGate.canCreate ||
        latestGate.medicalRecordId !== medicalRecordId ||
        latestGate.recordClinicId !== recordClinicId ||
        latestGate.petSpecies !== petSpecies
      ) {
        return;
      }
      try {
        const payload = buildMasterSelectionPayload({
          item,
          quantity,
          sortOrder: latestGate.nextSortOrder,
        });
        createTreatmentFn(payload, { onSuccess: () => setFocusLastRow(true) });
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "保険設定を確認してください");
        return;
      }
    },
    [createTreatmentFn, petSpecies, queryClient, medicalRecordId, recordClinicId],
  );

  const handleRetryMasterDoseLookup = useCallback(() => {
    if (!pendingMasterLookupItem) return;
    // PR679 F5: retry はコンポーネントの死亡/権限 wrapper を通らないため hook 層で弾く。
    // resetQueries 後の再選択は handleSelectFromMaster 入口の最新値検査で再度弾かれる。
    if (latestWriteGateRef.current.isWriteBlocked || !latestWriteGateRef.current.canCreate) {
      return;
    }
    const item = pendingMasterLookupItem;
    void (async () => {
      // 失敗キャッシュで retry が即 reject しないよう、再試行時にだけ error を捨てる。
      // 通常選択経路では実行しない — 共有 queryKey を毎回リセットすると
      // ①STATIC staleTime が無効化され ②同一薬剤の TreatmentRow が一往復のあいだ
      // data 無しに落ちて行の投与量ゲートが一時的に開く。
      if (item.medicineId) {
        await queryClient.resetQueries({
          queryKey: medicineDoseParamsQueryKey(item.medicineId, recordClinicId),
        });
      }
      await handleSelectFromMaster(item);
    })();
  }, [pendingMasterLookupItem, handleSelectFromMaster, queryClient, recordClinicId]);

  return {
    isLoading,
    canCreate,
    canDelete,
    // PR679 F5: 値引編集は権限に加えて確定/ロック/死亡でも fail-closed にする。
    canEditDiscount: Boolean(canEditDiscount && !isWriteBlocked),
    sortedTreatments,
    isMutating: updateMutation.isPending || deleteMutation.isPending || reorderMutation.isPending,
    createIsPending: createMutation.isPending,
    focusLastRow,
    doseContext,
    handleUpdate,
    handleDelete,
    handleMoveUp,
    handleMoveDown,
    handleAutoFocusDone,
    isAdding,
    addItemType,
    addContent,
    addAdminRoute,
    setAddContent,
    setAddAdminRoute,
    handleAddItemTypeChange,
    handleAddSubmit,
    handleAddCancel,
    handleOpenSearch,
    handleStartAdding,
    isSearchOpen,
    setIsSearchOpen,
    handleSelectFromMaster,
    masterDoseBlockReason,
    pendingMasterLookupItem,
    handleRetryMasterDoseLookup,
    totalSubtotal,
    selectedCount,
    selectedSubtotal,
    finalTotal,
  };
}
