import type { QueryClient } from "@tanstack/react-query";

import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import { normalizeDoseSpecies } from "@/lib/medicine-dose";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { getVitals } from "../api/vitals";
import { getAllMedicinesMaster } from "@/hooks/use-treatment-master";
import {
  buildDoseCalcInput,
  computeDosePreview,
  fetchMedicineDoseParamsOnce,
  medicineDoseParamsQueryKey,
  DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
} from "../api/medicine-dose-lookup";
import { computeDoseGate, type DoseGateSource } from "./treatment-row-dose-gate";
import { isPerWeightMedicine, resolveLatestVitalWeight } from "../api/medicine-dose-lookup";

/**
 * 選択した薬剤が記録のクリニックで無効・不在のときの固定文言。
 * 再試行では解決しないため非リトライ扱いとする（CodeRabbit 4307602167）。
 */
const INACTIVE_MEDICINE_MESSAGE =
  "選択した薬剤は記録のクリニックで有効ではありません。マスタを確認して再選択してください";

export async function resolveMasterSelectionDose(params: {
  item: TreatmentMasterItem;
  medicalRecordId: string;
  recordClinicId?: string;
  petSpecies?: string | null;
  queryClient: QueryClient;
}): Promise<
  { kind: "ready"; quantity: number } | { kind: "blocked"; message: string; retryable: boolean }
> {
  const { item, medicalRecordId, recordClinicId, petSpecies, queryClient } = params;

  if (item.category !== "薬剤") {
    return { kind: "ready", quantity: 1 };
  }
  if (!item.medicineId) {
    return {
      kind: "blocked",
      message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
      retryable: true,
    };
  }

  const medicineId = item.medicineId;

  // 薬剤マスタを先に引く。用量計算が per_weight の薬剤だけバイタル/用量パラメータを
  // 後続で引く（Codex 4307600239: none 薬の作成を vitals 障害で止めない）。
  let medicines;
  try {
    medicines = await queryClient.fetchQuery({
      queryKey: recordClinicId
        ? queryKeys.masters.categoryForClinic("medicines", recordClinicId)
        : queryKeys.masters.category("medicines"),
      queryFn: () => getAllMedicinesMaster(recordClinicId),
      staleTime: QUERY_STALE_TIMES.STATIC,
    });
  } catch {
    return {
      kind: "blocked",
      message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
      retryable: true,
    };
  }

  if (!Array.isArray(medicines)) {
    return {
      kind: "blocked",
      message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
      retryable: true,
    };
  }
  if (medicines.some((entry) => entry == null || typeof entry !== "object")) {
    return {
      kind: "blocked",
      message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
      retryable: true,
    };
  }

  // 記録のクリニックで有効な薬剤のみ選択可能。無効・不在は再試行では解決しないため
  // 非リトライの再選択案内にする（CodeRabbit 4307602167: リトライ無限ループを避ける）。
  const medicine = medicines.find(
    (m) => String(m.id) === String(medicineId) && m.isActive === true,
  );
  if (!medicine) {
    return {
      kind: "blocked",
      message: INACTIVE_MEDICINE_MESSAGE,
      retryable: false,
    };
  }

  // #201: 薬剤選択時、体重・species・薬マスタの投与量パラメータが揃えば quantity をプリフィルする。
  // いずれか欠けている場合は既定値 1（従来通りの手動入力・保存継続）とする。
  // TASK-025: technical failure（fetch/HTTP/parse）は missing data と型で区別し、通常保存を止める。
  let quantity = 1;
  let doseGateSource: DoseGateSource = { kind: "missing" };
  if (isPerWeightMedicine(medicine)) {
    let vitals;
    try {
      vitals = await queryClient.fetchQuery({
        queryKey: queryKeys.medicalRecords.vitals(medicalRecordId, recordClinicId),
        queryFn: () => getVitals(medicalRecordId, recordClinicId),
        staleTime: QUERY_STALE_TIMES.REALTIME,
      });
    } catch {
      return {
        kind: "blocked",
        message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
        retryable: true,
      };
    }
    if (
      !Array.isArray(vitals) ||
      vitals.some((entry) => entry == null || typeof entry !== "object")
    ) {
      return {
        kind: "blocked",
        message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
        retryable: true,
      };
    }

    // Codex (PR #700 round5): species・体重が揃わないケースは backend も dose params を
    // 参照せず手動数量で受理する（treatment_dose_save.go:39-69）。dose-params の取得は
    // 用量計算の前提（体重＋species 解決）が揃った場合に限り、障害時の過剰ブロックを防ぐ。
    const resolvedWeight = resolveLatestVitalWeight(vitals);
    const species = normalizeDoseSpecies(petSpecies ?? null);
    if (resolvedWeight && species) {
      let doseParams;
      try {
        doseParams = await queryClient.fetchQuery({
          queryKey: medicineDoseParamsQueryKey(medicineId, recordClinicId),
          queryFn: () => fetchMedicineDoseParamsOnce(medicineId, recordClinicId),
          staleTime: QUERY_STALE_TIMES.STATIC,
        });
      } catch {
        return {
          kind: "blocked",
          message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
          retryable: true,
        };
      }
      if (
        !Array.isArray(doseParams) ||
        doseParams.some((entry) => entry == null || typeof entry !== "object")
      ) {
        return {
          kind: "blocked",
          message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
          retryable: true,
        };
      }

      const doseCalcInput = buildDoseCalcInput(
        medicine,
        doseParams,
        petSpecies ?? null,
        resolvedWeight.weightKg,
      );
      doseGateSource = doseCalcInput
        ? { kind: "ready", input: doseCalcInput }
        : { kind: "missing" };
      const preview = computeDosePreview(
        medicine,
        doseParams,
        petSpecies ?? null,
        resolvedWeight.weightKg,
      );
      // healthcare-review-201 MEDIUM: 推奨値自体が丸め up 等で安全域を超えている場合、
      // 未編集のまま保存され得るため、その値ではプリフィルしない
      // （既定値 1 の手動入力とし、実際の保存値を下の物理ブロック判定に通す）。
      if (preview && !preview.exceedsMax && !preview.belowMin) quantity = preview.quantity;
    }
  }
  // #201: 実際に submit される quantity がマスタの絶対上限を超える場合は、
  // ConfirmDialog で解除できない物理ブロックを create 前に適用する。
  // TASK-377: create 経路に inline 理由 UI が無いため、理由必須の quantity では create を止める
  // （行追加後の quantity 編集で理由を入力する導線へ誘導）。
  const gate = computeDoseGate(doseGateSource, quantity);
  if (gate.isBlocked) {
    return { kind: "blocked", message: gate.blockReason, retryable: false };
  }
  if (gate.requiresDeviationReason) {
    return {
      kind: "blocked",
      message: gate.reason
        ? `${gate.reason}。推奨値付近で追加してから数量と逸脱理由を入力してください`
        : "用量逸脱の理由入力が必要です。推奨値付近で追加してから数量と理由を入力してください",
      retryable: false,
    };
  }
  return { kind: "ready", quantity };
}
