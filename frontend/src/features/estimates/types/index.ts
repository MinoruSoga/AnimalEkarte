/**
 * Estimates feature types (UI-facing: camelCase, string IDs)
 * Backend types: {@link Estimate as BackendEstimate}, {@link EstimateItem as BackendEstimateItem} from models.ts
 */

// FE6-3: tygo enum_style: "union"（FE6-1/FE6-2）により生成定数が真の literal union になったため、
// 手書き union を生成型からの re-export へ移行した。drift テストは union-drift.test.ts から削除済み。
export type { EstimateStatus } from "@/types/generated/models";

export type { Estimate, EstimateLineItem } from "../api/transforms";
