/**
 * NO32: ペット単位トリミング一覧 query の実体は @/hooks/use-pet-trimmings。
 * features/trimming から共有層へ昇格済み（cross-feature import 禁止のため）。
 * 本ファイルは順方向 re-export のみ（実装を持たない。get-medical-records.ts と同型）。
 */
export { useGetTrimmingsByPetId } from "@/hooks/use-pet-trimmings";
