import type { QueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";

export interface DiagnosisOptionLabels {
  /** 区分（診断種別）の表示名。UI が表示している名前が判明しない場合は null。 */
  categoryName: string | null;
  /** 診断名の表示名。UI が表示している名前が判明しない場合は null。 */
  diagnosisName: string | null;
}

function findOptionName(options: unknown, id: number | null): string | null {
  if (id === null || !Array.isArray(options)) return null;
  for (const option of options) {
    if (
      typeof option === "object" &&
      option !== null &&
      "id" in option &&
      Number(option.id) === id &&
      "name" in option &&
      typeof option.name === "string" &&
      option.name !== ""
    ) {
      return option.name;
    }
  }
  return null;
}

/**
 * 診断セレクタ（get-diagnosis-options.ts の useGetDiagnosisTypes / useGetDiagnosisNames）が
 * react-query に載せている選択肢から、選択中の ID に対応する表示名を引く。
 * セレクタが描画に使うキャッシュと同じキーを読むため、画面に出ている名前と一致する。
 * キーを変えるときは get-diagnosis-options.ts と揃えること。キャッシュに無い ID は null（推測しない）。
 */
export function findCachedDiagnosisLabels(
  queryClient: QueryClient,
  categoryId: number | null,
  nameId: number | null,
): DiagnosisOptionLabels {
  return {
    categoryName: findOptionName(
      queryClient.getQueryData(queryKeys.masters.category("diagnosis-types")),
      categoryId,
    ),
    // 診断名の選択肢は区分 ID ごとに取得・キャッシュされる。
    diagnosisName:
      categoryId === null
        ? null
        : findOptionName(
            queryClient.getQueryData(queryKeys.masters.diagnosisNames(categoryId)),
            nameId,
          ),
  };
}
