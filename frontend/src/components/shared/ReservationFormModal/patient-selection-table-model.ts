export interface PatientSearchParams {
  search: string;
  ownerId: string;
  species: string;
}

export const INITIAL_SEARCH_PARAMS: PatientSearchParams = {
  search: "",
  ownerId: "",
  species: "",
};

export const PAGE_SIZE = 20;

export const ALL_SPECIES_VALUE = "__all__";

export const TEXT_FIELDS = [
  {
    key: "search",
    label: "検索（ペット名・飼主名・よみ・電話）",
    placeholder: "例: もも、山田、090",
  },
  { key: "ownerId", label: "飼主No", placeholder: "例: 30042" },
] as const;

export function resolvePatientSelectionText(
  selectedCount: number,
  offPageSelectedCount: number,
): string {
  if (selectedCount === 0) return "";
  if (offPageSelectedCount > 0) {
    return `選択中: ${selectedCount}件（別ページの${offPageSelectedCount}件を含む）`;
  }
  return `選択中: ${selectedCount}件`;
}

export function resolvePatientStatusText(args: {
  hasSearchConditions: boolean;
  showCachedRange: boolean;
  rangeText: string;
  error: unknown;
  isBusy: boolean;
  isCountTrustworthy: boolean;
  totalCount: number;
}): string {
  if (!args.hasSearchConditions) return "";
  if (args.showCachedRange) return `前回取得: ${args.rangeText}`;
  if (args.error || args.isBusy || !args.isCountTrustworthy) return "";
  if (args.totalCount === 0) return "0件";
  return args.rangeText;
}
