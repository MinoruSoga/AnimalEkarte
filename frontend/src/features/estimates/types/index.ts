/**
 * Estimates feature types (UI-facing: camelCase, string IDs)
 * Backend types: {@link Estimate as BackendEstimate}, {@link EstimateItem as BackendEstimateItem} from models.ts
 */

/**
 * 手書き literal union（FE4-2）。
 * tygo 生成定数は `export const X: EstimateStatus = "draft"` 形式（EstimateStatus = string）のため
 * `typeof 生成定数` は string に退化する。生成側は編集禁止のため、手書き literal union +
 * ランタイム値集合の drift テスト（./union-drift.test.ts）で生成定数の値集合との一致を機械固定する。
 * @see {@link import("@/types/generated/models").EstimateStatus}
 */
export type EstimateStatus = "draft" | "sent" | "approved" | "rejected";

export type { Estimate, EstimateLineItem } from "../api/transforms";
