export const HOSPITALIZATION_STATUS = {
  ACTIVE: "入院中",
  DISCHARGED: "退院済",
  RESERVED: "予約",
  TEMP_DISCHARGE: "一時帰宅",
} as const;

export const HOSPITALIZATION_FILTER_STATUS = {
  ALL: "all",
  ACTIVE: "active",
  DISCHARGED: "discharged",
  RESERVED: "reserved",
} as const;

export type HospitalizationFilterStatus = typeof HOSPITALIZATION_FILTER_STATUS[keyof typeof HOSPITALIZATION_FILTER_STATUS];

/** BE wire status（GET /v1/hospitalizations?status=）。`active` タブだけ BE の `admitted` と綴りが異なる。 */
export type HospitalizationWireStatus = "admitted" | "reserved" | "discharged";

/**
 * タブ filter → BE status query の型固定 mapping。
 * `all` は status param を送らない（呼び出し側で omit）。
 */
const HOSPITALIZATION_FILTER_TO_WIRE_STATUS = {
  [HOSPITALIZATION_FILTER_STATUS.ACTIVE]: "admitted",
  [HOSPITALIZATION_FILTER_STATUS.RESERVED]: "reserved",
  [HOSPITALIZATION_FILTER_STATUS.DISCHARGED]: "discharged",
} as const satisfies Record<
  Exclude<HospitalizationFilterStatus, typeof HOSPITALIZATION_FILTER_STATUS.ALL>,
  HospitalizationWireStatus
>;

/** タブ値を BE wire status へ写す。`all` は undefined（param なし）。 */
export function toHospitalizationWireStatus(
  filter: HospitalizationFilterStatus,
): HospitalizationWireStatus | undefined {
  if (filter === HOSPITALIZATION_FILTER_STATUS.ALL) return undefined;
  return HOSPITALIZATION_FILTER_TO_WIRE_STATUS[filter];
}

/** httpapi.ParsePagination の既定値（page 未指定→1、limit/per_page 未指定→20）。 */
export const HOSPITALIZATION_LIST_DEFAULT_PAGE = 1;
export const HOSPITALIZATION_LIST_DEFAULT_LIMIT = 20;

/**
 * 臨床安全境界1（frontend/CLAUDE.md）: 死亡ペットは render 側でも操作要素自体を出さない。
 * callback 側の拒否（二重防壁）と併せて使用し、非表示にした理由をユーザーに提示する。
 */
export const HOSPITALIZATION_DECEASED_BLOCK_MESSAGE = {
  CARE_PLAN: "死亡したペットのため、ケアプランの追加・編集・削除はできません",
  DAILY_RECORD: "死亡したペットのため、デイリーカルテの記録・追加はできません",
  CHECK_IN: "死亡したペットのため、チェックインできません",
} as const;
