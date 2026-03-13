/**
 * Backend API response types
 * Generated from backend/docs/api.yaml via openapi-typescript
 * DO NOT EDIT manually — run `make codegen` to regenerate
 */
import type { components } from "@/types/generated/api";

export type BackendExaminationItem = components["schemas"]["ExaminationItem"];
export type BackendExamination = components["schemas"]["Examination"];

export interface CreateExaminationRequest {
  medical_record_id: string;
  pet_id?: string | null;
  exam_type_id: string;
  doctor_id?: string | null;
  date: string;
  machine?: string;
  result_summary?: string;
}

export interface UpdateExaminationRequest {
  status?: "依頼中" | "検査中" | "完了";
  result_summary?: string;
  machine?: string;
  date?: string;
}
