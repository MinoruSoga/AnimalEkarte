/**
 * 見積 API の移行 HTTP DTO（非永続）の estimates 側入口。
 * 定義は medical-records と共有するため @/types/estimate-legacy-wire に置く（EMR-527）。
 */
export type {
  LegacyEstimateItemRequest,
  LegacyEstimateItemResponse,
  LegacyEstimateOwnerSummary,
  LegacyEstimatePetSummary,
  LegacyEstimateResponse,
} from "@/types/estimate-legacy-wire";
