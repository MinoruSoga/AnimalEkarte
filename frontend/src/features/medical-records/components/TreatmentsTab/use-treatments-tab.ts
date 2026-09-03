import { useState, useCallback, useMemo } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { usePermission } from "@/hooks/use-permission";
import { useClinicTaxRates } from "@/hooks/use-clinic-tax-rates";
import { useGetAllMedicinesMaster } from "@/hooks/use-treatment-master";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

import { useGetTreatments } from "../../api/treatments";
import { useCreateTreatment } from "../../api/treatments";
import { useUpdateTreatment } from "../../api/treatments";
import { useDeleteTreatment } from "../../api/treatments";
import { useReorderTreatments } from "../../api/treatments";
import { useGetVitals } from "../../api/vitals";
import {
  buildDoseCalcInput,
  computeDosePreview,
  DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
  fetchMedicineDoseParamsOnce,
  medicineDoseParamsQueryKey,
  resolveLatestVitalWeight,
} from "../../api/medicine-dose-lookup";
import type { TreatmentItemType, UpdateTreatmentInput } from "../../types";
import { buildMasterSelectionPayload } from "./treatments-tab-model";
import { computeDoseGate, type DoseGateSource } from "./treatment-row-dose-gate";

export function useTreatmentsTab({
  medicalRecordId,
  ownerDiscountRate = 0,
  petSpecies,
  recordClinicId,
}: {
  medicalRecordId: string;
  ownerDiscountRate?: number;
  petSpecies?: string | null;
  recordClinicId?: string;
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
  const { mutate: updateTreatmentFn } = updateMutation;
  const deleteMutation = useDeleteTreatment(medicalRecordId, recordClinicId);
  const { mutate: deleteTreatmentFn } = deleteMutation;
  const reorderMutation = useReorderTreatments(medicalRecordId, recordClinicId);
  const { mutate: reorderTreatmentsFn } = reorderMutation;
  const queryClient = useQueryClient();

  // #201: 投与量自動計算の入力（当日 vital の体重・薬マスタ一覧）。
  const { data: vitals } = useGetVitals(medicalRecordId, recordClinicId);
  const { data: medicines } = useGetAllMedicinesMaster();
  const resolvedWeight = useMemo(
    () => (vitals ? resolveLatestVitalWeight(vitals) : null),
    [vitals]
  );
  // react-review-201 HIGH-1: オブジェクトリテラルを直接 props に渡すと毎レンダー新規参照になり
  // TreatmentRow の React.memo が無効化される（全行が無関係な state 変更でも再レンダーされる）。
  const doseContext = useMemo(
    () => ({
      medicines,
      petSpecies: petSpecies ?? null,
      weightKg: resolvedWeight?.weightKg ?? null,
    }),
    [medicines, petSpecies, resolvedWeight]
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
    const sub = selected.reduce(
      (sum, t) => sum + t.unit_price * t.quantity - t.discount_amount,
      0
    );
    // 飼主割引適用
    const ownerDiscount = Math.floor(sub * (ownerDiscountRate / 100));
    const afterDiscount = sub - ownerDiscount;
    const tax = Math.floor(afterDiscount * standardTaxRate);

    return {
      selectedSubtotal: sub,
      selectedCount: selected.length,
      finalTotal: afterDiscount + tax
    };
  }, [sortedTreatments, ownerDiscountRate, standardTaxRate]);

  // 全明細の合計
  const totalSubtotal = useMemo(
    () =>
      sortedTreatments.reduce(
        (sum, t) => sum + t.unit_price * t.quantity - t.discount_amount,
        0
      ),
    [sortedTreatments]
  );

  // ── handlers ──

  const handleUpdate = useCallback(
    (treatmentId: string, input: UpdateTreatmentInput) => {
      if (!canEdit) return;
      updateTreatmentFn({ treatmentId, input });
    },
    [canEdit, updateTreatmentFn]
  );

  const handleDelete = useCallback(
    (treatmentId: string) => {
      if (!canDelete) return;
      deleteTreatmentFn(treatmentId);
    },
    [canDelete, deleteTreatmentFn]
  );

  const handleMoveUp = useCallback(
    (treatmentId: string) => {
      if (!canEdit) return;
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
    [canEdit, sortedTreatments, reorderTreatmentsFn]
  );

  const handleMoveDown = useCallback(
    (treatmentId: string) => {
      if (!canEdit) return;
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
    [canEdit, sortedTreatments, reorderTreatmentsFn]
  );

  // rerender-memo: TreatmentRow(TreatmentsTable 経由)/TreatmentAddControls へ渡す
  // ハンドラを useCallback で安定化する（inline arrow は毎レンダー新規参照になる）。
  const handleAutoFocusDone = useCallback(() => {
    setFocusLastRow(false);
  }, []);

  const handleOpenSearch = useCallback(() => {
    setIsSearchOpen(true);
  }, []);

  const handleStartAdding = useCallback(() => {
    setIsAdding(true);
  }, []);

  const handleAddSubmit = useCallback(() => {
    if (!canCreate || !addContent.trim()) return;
    const nextOrder =
      sortedTreatments.length > 0
        ? sortedTreatments[sortedTreatments.length - 1].sort_order + 1
        : 0;
    // 薬品の場合、投与方法を memo に付記する（BE-MEDI-010: 専用カラム実装まで）
    const memoWithRoute =
      addItemType === "medicine" && addAdminRoute
        ? `[投与方法: ${addAdminRoute}]`
        : "";
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
      }
    );
  }, [canCreate, addItemType, addContent, addAdminRoute, sortedTreatments, createTreatmentFn]);

  const handleAddCancel = useCallback(() => {
    setAddContent("");
    setAddAdminRoute("");
    setIsAdding(false);
  }, []);

  const handleAddItemTypeChange = useCallback((itemType: TreatmentItemType) => {
    setAddItemType(itemType);
    setAddAdminRoute("");
  }, []);

  const handleSelectFromMaster = useCallback(async (item: TreatmentMasterItem) => {
    if (!canCreate) return;
    const nextOrder =
      sortedTreatments.length > 0
        ? sortedTreatments[sortedTreatments.length - 1].sort_order + 1
        : 0;

    // #201: 薬剤選択時、体重・species・薬マスタの投与量パラメータが揃えば quantity をプリフィルする。
    // いずれか欠けている場合は既定値 1（従来通りの手動入力・保存継続）とする。
    // TASK-025: technical failure（fetch/HTTP/parse）は missing data と型で区別し、通常保存を止める。
    let quantity = 1;
    let doseGateSource: DoseGateSource = { kind: "missing" };
    if (item.medicineId && resolvedWeight) {
      const medicine = medicines?.find((m) => m.id === item.medicineId);
      if (medicine) {
        try {
          const doseParams = await queryClient.fetchQuery({
            queryKey: medicineDoseParamsQueryKey(item.medicineId),
            queryFn: () => fetchMedicineDoseParamsOnce(item.medicineId as string),
            staleTime: QUERY_STALE_TIMES.STATIC,
          });
          const doseCalcInput = buildDoseCalcInput(
            medicine,
            doseParams,
            petSpecies,
            resolvedWeight.weightKg
          );
          doseGateSource = doseCalcInput
            ? { kind: "ready", input: doseCalcInput }
            : { kind: "missing" };
          const preview = computeDosePreview(
            medicine,
            doseParams,
            petSpecies,
            resolvedWeight.weightKg
          );
          // healthcare-review-201 MEDIUM: 推奨値自体が丸め up 等で安全域を超えている場合、
          // 未編集のまま保存され得るため、その値ではプリフィルしない
          // （既定値 1 の手動入力とし、実際の保存値を下の物理ブロック判定に通す）。
          if (preview && !preview.exceedsMax && !preview.belowMin) quantity = preview.quantity;
        } catch {
          // technical failure: quantity=1 の silent fallback で create しない。
          // upstream body は画面に出さない（固定文言のみ）。
          setMasterDoseBlockReason(DOSE_PARAMS_LOOKUP_FAILED_MESSAGE);
          setPendingMasterLookupItem(item);
          return;
        }
      }
    }

    const payload = buildMasterSelectionPayload({ item, quantity, sortOrder: nextOrder });

    // #201: 実際に submit される quantity がマスタの絶対上限を超える場合は、
    // ConfirmDialog で解除できない物理ブロックを create 前に適用する。
    // TASK-377: create 経路に inline 理由 UI が無いため、理由必須の quantity では create を止める
    // （行追加後の quantity 編集で理由を入力する導線へ誘導）。
    const gate = computeDoseGate(doseGateSource, quantity);
    if (gate.isBlocked) {
      setMasterDoseBlockReason(gate.blockReason);
      setPendingMasterLookupItem(null);
      return;
    }
    if (gate.requiresDeviationReason) {
      setMasterDoseBlockReason(
        gate.reason
          ? `${gate.reason}。推奨値付近で追加してから数量と逸脱理由を入力してください`
          : "用量逸脱の理由入力が必要です。推奨値付近で追加してから数量と理由を入力してください"
      );
      setPendingMasterLookupItem(null);
      return;
    }

    setMasterDoseBlockReason("");
    setPendingMasterLookupItem(null);
    createTreatmentFn(payload, { onSuccess: () => setFocusLastRow(true) });
  }, [canCreate, sortedTreatments, createTreatmentFn, resolvedWeight, medicines, petSpecies, queryClient]);

  const handleRetryMasterDoseLookup = useCallback(() => {
    if (!pendingMasterLookupItem) return;
    const item = pendingMasterLookupItem;
    void (async () => {
      // 失敗キャッシュで retry が即 reject しないよう、再試行時にだけ error を捨てる。
      // 通常選択経路では実行しない — 共有 queryKey を毎回リセットすると
      // ①STATIC staleTime が無効化され ②同一薬剤の TreatmentRow が一往復のあいだ
      // data 無しに落ちて行の投与量ゲートが一時的に開く。
      if (item.medicineId) {
        await queryClient.resetQueries({
          queryKey: medicineDoseParamsQueryKey(item.medicineId),
        });
      }
      await handleSelectFromMaster(item);
    })();
  }, [pendingMasterLookupItem, handleSelectFromMaster, queryClient]);

  return {
    isLoading,
    canCreate,
    canDelete,
    canEditDiscount,
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
