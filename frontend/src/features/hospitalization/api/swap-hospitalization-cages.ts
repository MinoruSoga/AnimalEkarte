import { axios } from "@/lib/axios";
import { transformHospitalization } from "./transforms";
import type { Hospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

interface SwapCagesResponse {
  hospitalizations: BackendHospitalization[];
}

// EMR-306 (Codex review P1): ボードの入れ替えは2本の独立 PATCH では片方だけ commit
// され得る（preflight と PATCH の間に別セッションが連結すると半スワップ状態になる）。
// cage-swap エンドポイントは両方の更新を単一トランザクションで行い、
// 構成ケージ宛は 422 / 占有グループ宛は 409 で両方ロールバックする。
// expectedCageIds は入院IDと同じ順序の表示時ケージID。不一致は409で拒否される。
export const swapHospitalizationCages = async (
  hospitalizationIdA: string,
  hospitalizationIdB: string,
  expectedCageIds: readonly [string, string],
): Promise<Hospitalization[]> => {
  const { data } = await axios.post<SwapCagesResponse>(`/v1/hospitalizations/cage-swap`, {
    hospitalization_ids: [Number(hospitalizationIdA), Number(hospitalizationIdB)],
    expected_cage_ids: expectedCageIds.map(Number),
  });
  return data.hospitalizations.map(transformHospitalization);
};
