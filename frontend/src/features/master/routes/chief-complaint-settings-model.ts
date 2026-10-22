import type {
  CreateChiefComplaintTypeRequest,
  UpdateChiefComplaintTypeRequest,
} from "../api/chief-complaint-types";
import type { ChiefComplaintFormData } from "../lib/chief-complaint-side-panel-model";

export function buildChiefComplaintCreateRequest(
  data: ChiefComplaintFormData,
): CreateChiefComplaintTypeRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}

export function buildChiefComplaintUpdateRequest(
  data: ChiefComplaintFormData,
): UpdateChiefComplaintTypeRequest {
  // EMR-447: 更新では空の備考も送り、既存備考を空へ変更できるようにする（Create 既定は変えない）。
  return {
    name: data.name,
    description: data.description,
    is_active: data.isActive,
  };
}
