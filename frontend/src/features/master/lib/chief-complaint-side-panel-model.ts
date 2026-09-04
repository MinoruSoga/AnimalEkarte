import type { ChiefComplaintType } from "../api/chief-complaint-types";

export interface ChiefComplaintFormData {
  name: string;
  description: string;
  isActive: boolean;
}

export function chiefComplaintToFormData(item: ChiefComplaintType | null): ChiefComplaintFormData {
  return {
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  };
}
