import type { CreateOccupationRequest, UpdateOccupationRequest } from "../api/occupations";
import type { OccupationFormData } from "../lib/occupation-side-panel-model";

export function buildOccupationCreateRequest(data: OccupationFormData): CreateOccupationRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    is_active: data.isActive,
  };
}

export function buildOccupationUpdateRequest(data: OccupationFormData): UpdateOccupationRequest {
  return {
    ...buildOccupationCreateRequest(data),
    // EMR-448: 備考を空へ変更できるよう、空文字も省略せず送る（BE は "" を空で保存）。
    description: data.description,
  };
}
