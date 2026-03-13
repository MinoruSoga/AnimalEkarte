/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Pet } from "@/types/generated/models";

export type BackendPet = Pet;

// リクエスト型は共有 types に移管。後方互換のため re-export
export type { CreatePetRequest, UpdatePetRequest } from "@/types/pet";
