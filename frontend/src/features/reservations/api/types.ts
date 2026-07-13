/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
// R-F2-S12/S13: UpdateReservationRequest / CreateReservationRequest は shared hooks へ単一ソース化。
// ここは feature 内既存 import (`./types`) を壊さないための re-export（S8 パターン踏襲）。
export type { CreateReservationRequest } from "@/hooks/use-create-reservation";
