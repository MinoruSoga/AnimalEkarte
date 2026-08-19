// CPM (Customer Portfolio Management) ステージの共有定義。
// backend service.CPMStage*（"cpm_encounter" 等）と値域を一致させる。
//
// 集計・分析 (features/aggregation) と健診対象者抽出 (features/lstep/routes/CheckupSyncPage) の
// 双方がこの単一定義を参照し、セグメントのラベル drift を防ぐ（ISSUE-180）。
// ※ feature をまたぐ共有のため lib に置く。feature 内へ deep import しないこと。

export type CPMStage =
  | "cpm_encounter"
  | "cpm_growing"
  | "cpm_core"
  | "cpm_spot"
  | "cpm_noah"
  | "cpm_dormant";

/** 集計ダッシュボード用。6区分に入らない飼主をチップで辿れるようにする (BUG-003)。 */
export type AggregationCPMStage = CPMStage | "cpm_unclassified";

export interface CPMStageOption {
  value: CPMStage;
  label: string;
}

// フルラベル（フィルタのセレクト用）。健診対象者抽出のセレクトと完全一致させる。
export const CPM_STAGE_OPTIONS: readonly CPMStageOption[] = [
  { value: "cpm_encounter", label: "Encounter（初回・新規）" },
  { value: "cpm_growing", label: "Growing（成長中）" },
  { value: "cpm_core", label: "Core（コア顧客）" },
  { value: "cpm_spot", label: "Spot（単発高額）" },
  { value: "cpm_noah", label: "Noah（最上位）" },
  { value: "cpm_dormant", label: "Dormant（休眠）" },
] as const;

// 短縮ラベル（一覧バッジ・人数サマリーチップ用）。
export const CPM_STAGE_SHORT_LABELS: Record<CPMStage, string> = {
  cpm_encounter: "Encounter",
  cpm_growing: "Growing",
  cpm_core: "Core",
  cpm_spot: "Spot",
  cpm_noah: "Noah",
  cpm_dormant: "Dormant",
};

export const AGGREGATION_CPM_STAGE_OPTIONS: readonly {
  value: AggregationCPMStage;
  label: string;
}[] = [
  ...CPM_STAGE_OPTIONS,
  { value: "cpm_unclassified", label: "Unclassified（未分類）" },
] as const;

export const AGGREGATION_CPM_STAGE_SHORT_LABELS: Record<AggregationCPMStage, string> = {
  ...CPM_STAGE_SHORT_LABELS,
  cpm_unclassified: "Unclassified",
};

export const AGGREGATION_CPM_STAGES: readonly AggregationCPMStage[] =
  AGGREGATION_CPM_STAGE_OPTIONS.map((o) => o.value);

// CPMStage 値域の配列（反復・網羅処理用）。OPTIONS から導出して定義の二重化を防ぐ。
export const CPM_STAGES: readonly CPMStage[] = CPM_STAGE_OPTIONS.map((o) => o.value);
