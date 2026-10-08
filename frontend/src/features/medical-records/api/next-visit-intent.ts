// EMR-308: PUT/DELETE /v1/reservations/pets/:petId/next-visit-intent の mutation 実体は
// reception（キャンセル後再設定）・reservations（期限超過パネル）も使うため
// src/hooks/use-next-visit-intent.ts へ昇格済み。ここは feature 内既存 import を
// 壊さないための re-export（use-get-reservations 昇格時と同じ先例）。
export { useSetNextVisitIntent, useClearNextVisitIntent } from "@/hooks/use-next-visit-intent";
export type { SetNextVisitIntentVariables } from "@/hooks/use-next-visit-intent";
