import type { Occupation } from "../api/occupations";

export interface OccupationFormData {
  name: string;
  description: string;
  isActive: boolean;
}

export function occupationToFormData(item: Occupation | null): OccupationFormData {
  return {
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  };
}
