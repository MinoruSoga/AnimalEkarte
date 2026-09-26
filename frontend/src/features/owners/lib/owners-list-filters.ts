import { Heart, PawPrint, Stethoscope } from "lucide-react";
import type {
  ActiveFilter,
  FilterCondition,
  FilterOption,
  FilterProperty,
} from "@/components/shared/PropertyFilter/types";

// #266: 一覧をサーバサイドページネーション化するにあたり、ペット行単位のソート（旧 SortableHeader 列）は
// 撤去した — 1ページ分にしかソートが効かず「効いているように見えて壊れる」サイレント破損になるため。
// フィルタは species/include_deceased とも condition="is" のみ backend が受け付ける
// （pet_repository.go FindAll 参照）。is_not/is_empty を選ばせると黙って無視されるため、
// フィルタUIの条件選択肢自体を is のみに絞る。
// react-refresh/only-export-components: OwnersListTable.tsx / OwnersList.tsx はどちらも
// route/component ファイルのため、共有する非コンポーネント export（定数・純粋関数）は
// この専用ファイルに集約する（petToFormData と同型のルール、OwnersList.tsx 冒頭コメント参照）。
const SERVER_FILTER_CONDITIONS: FilterCondition[] = ["is"];

const INCLUDE_DECEASED_OPTIONS: FilterOption[] = [
  { value: "false", label: "生存のみ（既定）" },
  { value: "true", label: "死亡ペットも含める" },
];

// EMR-197-01: 健診受診履歴フィルタの enum 選択肢。値は backend の CheckupHistoryFilter
// 列挙と一致させる（pet_request.go → checkup_history.go 参照）。within_Ny は JST 当日起点
// N 年（包含境界）内の受診あり、not_within_Ny は同窓内の受診なし、none は受診履歴なし。
const CHECKUP_HISTORY_OPTIONS: FilterOption[] = [
  { value: "within_1y", label: "1年以内に受診" },
  { value: "within_2y", label: "2年以内に受診" },
  { value: "within_3y", label: "3年以内に受診" },
  { value: "not_within_1y", label: "1年以上受診なし" },
  { value: "not_within_2y", label: "2年以上受診なし" },
  { value: "not_within_3y", label: "3年以上受診なし" },
  { value: "none", label: "健診受診履歴なし" },
];

/**
 * #266: species フィルタは pets.animal_species_id (数値ID) で backend に渡す
 * （owner_repository.go 時代の種別名文字列とは異なる契約 — pet_handler.go 参照）。
 * マスタ取得は非同期のため、選択肢は呼び出し側（OwnersList.tsx）が useAnimalSpecies() から
 * 都度組み立てて渡す。ここでは組み立てのみ行い、フェッチはしない（純粋関数を維持する）。
 */
export function buildSpeciesFilterOptions(species: { id: number; name: string }[]): FilterOption[] {
  return species.map((s) => ({ value: String(s.id), label: s.name }));
}

export function buildOwnerFilterProperties(speciesOptions: FilterOption[]): FilterProperty[] {
  return [
    {
      key: "species",
      label: "種",
      type: "select",
      icon: PawPrint,
      conditions: SERVER_FILTER_CONDITIONS,
      options: speciesOptions,
    },
    {
      key: "include_deceased",
      label: "生死",
      type: "select",
      icon: Heart,
      conditions: SERVER_FILTER_CONDITIONS,
      options: INCLUDE_DECEASED_OPTIONS,
    },
    {
      key: "checkup_history",
      label: "健診受診履歴",
      type: "select",
      icon: Stethoscope,
      conditions: SERVER_FILTER_CONDITIONS,
      options: CHECKUP_HISTORY_OPTIONS,
    },
  ];
}

// #266 既知の制約: FilterAddPopover は FilterProperty.conditions の上書きを条件選択ステップで
// 参照せず、type=select の既定4条件（次と一致/次と不一致/空/空でない）を常に提示する
// （共有コンポーネント側の既存ギャップ・本チケットのスコープ外）。pet_repository.go の
// species/include_deceased/checkup_history は「次と一致」相当の完全一致にしか対応していないため、ここで
// condition==="is" のみを転送対象とし、is_not/空/空でない が選ばれた場合は黙って別解釈で転送しない
// （is_not の value をそのまま "is" として送ると絞り込みの意味が反転するサイレントバグになる）。
function isSupportedFilter(
  filter: ActiveFilter,
  key: string,
): filter is ActiveFilter & { value: string } {
  return (
    filter.key === key &&
    filter.condition === "is" &&
    typeof filter.value === "string" &&
    filter.value !== ""
  );
}

export function activeFiltersToParams(filters: ActiveFilter[]): {
  species?: string;
  include_deceased?: string;
  checkup_history?: string;
} {
  const speciesValue = filters.find((f) => isSupportedFilter(f, "species"))?.value;
  const includeDeceasedValue = filters.find((f) => isSupportedFilter(f, "include_deceased"))?.value;
  const checkupHistoryValue = filters.find((f) =>
    isSupportedFilter(f, "checkup_history"),
  )?.value;
  return {
    species: speciesValue,
    // 既定値 (false) は URL に残さない — 明示的に true を選んだ場合のみ転送する。
    include_deceased: includeDeceasedValue === "true" ? "true" : undefined,
    checkup_history: checkupHistoryValue,
  };
}

export function paramsToActiveFilters(
  searchParams: URLSearchParams,
  speciesOptions: FilterOption[],
): ActiveFilter[] {
  const filters: ActiveFilter[] = [];
  const species = searchParams.get("species");
  if (species) {
    const label = speciesOptions.find((o) => o.value === species)?.label ?? species;
    filters.push({ key: "species", condition: "is", value: species, displayValue: label });
  }
  if (searchParams.get("include_deceased") === "true") {
    filters.push({
      key: "include_deceased",
      condition: "is",
      value: "true",
      displayValue: "死亡ペットも含める",
    });
  }
  const checkupHistory = searchParams.get("checkup_history");
  if (checkupHistory) {
    // CHECKUP_HISTORY_OPTIONS 由来の値のみ復元する — enum 外は backend が 400 にする値であり、
    // UI 側で chip 化して正当化しない（loader は URL をそのまま転送するため 400 が見える）。
    const option = CHECKUP_HISTORY_OPTIONS.find((o) => o.value === checkupHistory);
    if (option) {
      filters.push({
        key: "checkup_history",
        condition: "is",
        value: checkupHistory,
        displayValue: option.label,
      });
    }
  }
  return filters;
}
