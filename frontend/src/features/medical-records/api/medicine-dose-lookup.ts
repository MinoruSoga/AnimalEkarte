import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { calculateDose, normalizeDoseSpecies, normalizeWeightToKg } from "@/lib/medicine-dose";
import type { DoseCalcInput, DoseCalcResult } from "@/lib/medicine-dose";
import { MedicineCalculationTypePerWeight } from "@/types/generated/models";
import type { MedicineDoseParam } from "@/types/generated/models";
import type { Medicine } from "@/lib/transforms/medicine";
import type { Vital } from "../types";

// #201: 薬剤 × 種の投与量パラメータの読み取り専用参照（カルテ側）。
// 書込・薬マスタ UI は features/master/api/medicine-dose-params.ts が別途持つ（本ファイルは GET のみ）。
// queryKey は queryKeys.masters.medicineDoseParams(medicineId) を共有し、マスタ側の
// upsert/delete mutation の invalidate 対象と一致させる（キャッシュ無効化の一貫性）。
export const medicineDoseParamsQueryKey = queryKeys.masters.medicineDoseParams;

const getMedicineDoseParams = async (medicineId: string): Promise<MedicineDoseParam[]> => {
  const { data } = await axios.get<MedicineDoseParam[]>(
    `/v1/masters/medicines/${medicineId}/dose-params`
  );
  return data;
};

/** カルテの治療行が継続的に参照するための React Query 版（キャッシュ共有・自動再取得）。 */
export const useMedicineDoseParams = (medicineId: string | null | undefined) => {
  return useQuery({
    queryKey: medicineDoseParamsQueryKey(medicineId ?? ""),
    queryFn: () => getMedicineDoseParams(medicineId as string),
    enabled: !!medicineId,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

/**
 * #201 TASK-025: dose-params 権威の分岐（BE evaluateDoseForSave の (eval,nil)/(nil,nil)/(nil,err) に対応）。
 * - success: 取得成功。params が空でも missing data（手動入力可）であり technical failure ではない。
 * - failed: fetch/HTTP/parse の technical failure。保存停止対象。
 * - pending: 取得中／未取得。権威データが無いだけで、成功でも失敗でもない。
 * - idle: medicineId 無し等で query が無効。
 *
 * pending を success として符号化しないこと。「取得成功したが当該 species の param が無い」と
 * 区別できなくなり、authority.status === "success" を権威データの根拠に使う実装が壊れる。
 *
 * 利用者向け文言は固定。upstream の response body / stack を転記しない。
 */
export type DoseParamsAuthority =
  | { status: "success"; params: MedicineDoseParam[] }
  | { status: "failed" }
  | { status: "pending" }
  | { status: "idle" };

/** technical failure 時の固定文言（upstream body を含めない）。 */
export const DOSE_PARAMS_LOOKUP_FAILED_MESSAGE =
  "投与量パラメータを取得できなかったため保存できません。再試行してください。";

/**
 * React Query 結果から DoseParamsAuthority を組み立てる。
 * キャッシュ済み data がある場合は background refetch 失敗でも success を優先する
 * （評価可能な権威データが手元にあるため）。data が無く isError のときだけ failed。
 */
export function toDoseParamsAuthority(
  medicineId: string | null | undefined,
  query: { data: MedicineDoseParam[] | undefined; isError: boolean },
): DoseParamsAuthority {
  if (!medicineId) return { status: "idle" };
  if (query.data !== undefined) return { status: "success", params: query.data };
  if (query.isError) return { status: "failed" };
  // 取得中 / 未取得: まだ権威データが無いが technical failure ではない。
  // gate 上は missing data と同じ扱い（手動入力継続）だが、型では success と区別する。
  return { status: "pending" };
}

/**
 * 薬剤追加の瞬間（TreatmentSearchDialog 選択直後）に quantity をプリフィルするための
 * 一度きりの取得。React Query キャッシュにも載せる（同一 queryKey）ので、直後に
 * useMedicineDoseParams がマウントされてもキャッシュヒットする。
 */
export const fetchMedicineDoseParamsOnce = getMedicineDoseParams;

export interface ResolvedVitalWeight {
  weightKg: number;
  /** BE dose_weight_source と同一形式（"vital_records:<id>"）。表示・監査用。 */
  source: string;
}

/**
 * 「当日 vital」= この medical_record に紐づく vital のうち体重(>0)が記録された最新のもの。
 * BE treatment_dose_save.go の resolveDoseWeight と同じ選定ロジック（最新 recorded_at・weight<=0 除外）。
 * pets.weight は使わない（stale な体重での誤計算を避けるため #201 で明示禁止）。
 */
export function resolveLatestVitalWeight(vitals: Vital[]): ResolvedVitalWeight | null {
  let chosen: Vital | null = null;
  let chosenTime = Number.NEGATIVE_INFINITY;
  for (const v of vitals) {
    if (v.weight == null || v.weight <= 0) continue;
    const time = new Date(v.recorded_at).getTime();
    // healthcare-review-201 LOW: recorded_at が不正な場合 NaN になり `>` 比較が常に false になる
    // （= 不正な日時の行を誤って「最新」として採用しない一方、有効な最新行を見逃す恐れがあった）。
    // 不正な recorded_at の行は「時刻不明」として最も古い扱いにし、有効な最新行を確実に選ぶ。
    if (!Number.isNaN(time) && (chosen === null || time > chosenTime)) {
      chosen = v;
      chosenTime = time;
    }
  }
  if (!chosen) return null;
  const weightKg = normalizeWeightToKg(chosen.weight as number, chosen.weight_unit);
  if (weightKg == null) return null;
  return { weightKg, source: `vital_records:${chosen.id}` };
}

/**
 * 薬剤・species別パラメータ・体重・pet種(free-text) から DoseCalcInput を組み立てる。
 * species がマップ不能・体重未記録・当該 species の param 未設定の場合は null
 * （= 手動入力へフェイルクローズ）。calculateDose（プリフィル）・evaluateSubmittedDose
 * （現在値の安全域チェック）の両方がこの入力を共有する。
 */
export function buildDoseCalcInput(
  medicine: Medicine | undefined,
  doseParams: MedicineDoseParam[] | undefined,
  speciesFreeText: string | null | undefined,
  weightKg: number | null,
): DoseCalcInput | null {
  if (!medicine || medicine.calculationType !== MedicineCalculationTypePerWeight) return null;
  if (weightKg == null || weightKg <= 0) return null;
  const species = normalizeDoseSpecies(speciesFreeText);
  if (!species) return null;
  const param = (doseParams ?? []).find((p) => p.species === species);
  if (!param) return null;

  return {
    calculationType: medicine.calculationType,
    medicineUnit: medicine.medicineUnit,
    strength: medicine.strength,
    doseBasis: param.dose_basis,
    frequencyPerDay: medicine.frequencyPerDay,
    dosePerKg: param.dose_per_kg,
    minMgPerKg: param.min_mg_per_kg,
    maxMgPerKg: param.max_mg_per_kg,
    absoluteMaxDose: param.absolute_max_dose,
    roundingStep: param.rounding_step,
    roundingMode: param.rounding_mode,
    weightKg,
  };
}

/**
 * 薬剤・species別パラメータ・体重・pet種(free-text) から #201 の投与量プレビュー（推奨値）を算出する。
 * eligible=false（medicine 未設定・per_weight でない・species マップ不能・体重未記録・
 * 当該 species の param 未設定）の場合は null（= 手動入力へフェイルクローズ、UI は既定挙動のまま）。
 * BE の treatment 保存時再検証（treatment_dose_save.go）とは独立した UX 専用プレビューであり、
 * 安全性の最終権威は BE 側にある。
 */
export function computeDosePreview(
  medicine: Medicine | undefined,
  doseParams: MedicineDoseParam[] | undefined,
  speciesFreeText: string | null | undefined,
  weightKg: number | null,
): DoseCalcResult | null {
  const input = buildDoseCalcInput(medicine, doseParams, speciesFreeText, weightKg);
  if (!input) return null;
  const result = calculateDose(input);
  return result.eligible ? result : null;
}

/**
 * TreatmentsTab → TreatmentRow へ橋渡しする #201 投与量計算コンテキスト。
 * 個々の行が自分の medicine_id に対応する薬剤を探し、`useMedicineDoseParams` で
 * species パラメータを取得して `computeDosePreview` に渡す。
 */
export interface MedicineDoseContext {
  medicines: Medicine[] | undefined;
  /** ペットの free-text 種別（selectedPet.species）。未設定なら計算はスキップ。 */
  petSpecies: string | null;
  /** 当日 vital から解決した体重(kg)。未記録なら null（手動入力にフェイルクローズ）。 */
  weightKg: number | null;
}
