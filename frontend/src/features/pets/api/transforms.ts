import type { Pet } from "@/types";
import type { BackendPet } from "./types";
import type { CreatePetRequest, UpdatePetRequest } from "@/types/pet";
import {
  transformBackendPetToFrontend,
  transformCreatePetRequest,
  transformUpdatePetRequest,
  PET_STATUS_MAP,
  PET_GENDER_MAP,
  PET_GENDER_REVERSE_MAP,
} from "@/lib/transforms/pet";

// 後方互換のため re-export（他のファイルが import できるように）
export {
  transformBackendPetToFrontend,
  transformCreatePetRequest,
  transformUpdatePetRequest,
  PET_STATUS_MAP,
  PET_GENDER_MAP,
  PET_GENDER_REVERSE_MAP,
};

export type { Pet, BackendPet, CreatePetRequest, UpdatePetRequest };
