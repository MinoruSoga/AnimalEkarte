// Re-export from shared transforms layer
// OwnerApiResponse = OwnerResponse（owner-responses.ts）。models.Owner は使わない。
export {
  transformOwner,
  type OwnerApiResponse,
} from "@/lib/transforms/owner";
export type { OwnerResponse } from "@/types/generated/owner-responses";
