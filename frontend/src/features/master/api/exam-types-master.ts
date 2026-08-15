import {
  useMutation,
  useQuery,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";
import type {
  CreateExaminationTypeRequest,
  ReorderTreatmentRequest,
  UpdateExaminationTypeRequest,
} from "@/types/treatment";

export interface ExamReferenceRange {
  id: string;
  examTypeFieldId: string;
  animalSpeciesId: string;
  refMin?: number;
  refMax?: number;
  qualitativeMin?: string;
  qualitativeMax?: string;
}

export interface ExaminationTypeField {
  id: string;
  examTypeId: string;
  name: string;
  inspectionValue: string;
  normalValue: string;
  unit: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
  referenceRanges: ExamReferenceRange[];
}

export interface ExaminationTypeMaster {
  id: string;
  name: string;
  parentId?: string;
  price: number;
  isActive: boolean;
  description: string;
  sortOrder: number;
  isNonInsurance: boolean;
  createdAt: string;
  updatedAt: string;
  items: ExaminationTypeField[];
}

export interface ExamReferenceRangeResponse {
  id: number;
  exam_type_field_id: number;
  animal_species_id: number;
  ref_min?: number;
  ref_max?: number;
  qualitative_min?: string;
  qualitative_max?: string;
}

export interface ExaminationTypeFieldResponse {
  id: number;
  exam_type_id: number;
  name: string;
  inspection_value: string;
  normal_value: string;
  unit: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
  reference_ranges?: ExamReferenceRangeResponse[];
}

export interface ExaminationTypeResponse {
  id: number;
  clinic_id: number;
  name: string;
  price?: number;
  is_active: boolean;
  description: string;
  parent_id?: number;
  sort_order: number;
  is_non_insurance: boolean;
  items?: ExaminationTypeFieldResponse[];
  created_at: string;
  updated_at: string;
}

export interface CreateExaminationTypeFieldRequest {
  name: string;
  inspection_value?: string;
  normal_value?: string;
  unit?: string;
  sort_order?: number;
}

export type UpdateExaminationTypeFieldRequest =
  Partial<CreateExaminationTypeFieldRequest>;

export interface ExamReferenceRangeInput {
  animal_species_id: number;
  ref_min?: number;
  ref_max?: number;
  qualitative_min?: string;
  qualitative_max?: string;
}

function transformReferenceRange(
  data: ExamReferenceRangeResponse,
): ExamReferenceRange {
  return {
    id: String(data.id),
    examTypeFieldId: String(data.exam_type_field_id),
    animalSpeciesId: String(data.animal_species_id),
    refMin: data.ref_min,
    refMax: data.ref_max,
    qualitativeMin: data.qualitative_min,
    qualitativeMax: data.qualitative_max,
  };
}

function transformExaminationTypeField(
  data: ExaminationTypeFieldResponse,
): ExaminationTypeField {
  return {
    id: String(data.id),
    examTypeId: String(data.exam_type_id),
    name: data.name,
    inspectionValue: data.inspection_value,
    normalValue: data.normal_value,
    unit: data.unit,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
    referenceRanges: (data.reference_ranges ?? []).map(transformReferenceRange),
  };
}

export function transformExaminationTypeResponse(
  data: ExaminationTypeResponse,
): ExaminationTypeMaster {
  return {
    id: String(data.id),
    name: data.name,
    parentId: data.parent_id !== undefined ? String(data.parent_id) : undefined,
    price: data.price ?? 0,
    isActive: data.is_active,
    description: data.description ?? "",
    sortOrder: data.sort_order ?? 0,
    isNonInsurance: data.is_non_insurance ?? false,
    createdAt: data.created_at ?? "",
    updatedAt: data.updated_at ?? "",
    items: (data.items ?? []).map(transformExaminationTypeField),
  };
}

const getAllExaminationTypes = async (): Promise<ExaminationTypeMaster[]> => {
  const { data } = await axios.get<ExaminationTypeResponse[]>(
    "/v1/masters/examination-types",
  );
  return data.map(transformExaminationTypeResponse);
};

export async function createExaminationTypeField(
  examTypeId: string,
  req: CreateExaminationTypeFieldRequest,
): Promise<ExaminationTypeField> {
  const { data } = await axios.post<ExaminationTypeFieldResponse>(
    `/v1/masters/examination-types/${examTypeId}/fields`,
    req,
  );
  return transformExaminationTypeField(data);
}

export async function updateExaminationTypeField(
  examTypeId: string,
  fieldId: string,
  req: UpdateExaminationTypeFieldRequest,
): Promise<ExaminationTypeField> {
  const { data } = await axios.patch<ExaminationTypeFieldResponse>(
    `/v1/masters/examination-types/${examTypeId}/fields/${fieldId}`,
    req,
  );
  return transformExaminationTypeField(data);
}

export async function deleteExaminationTypeField(
  examTypeId: string,
  fieldId: string,
): Promise<void> {
  await axios.delete(
    `/v1/masters/examination-types/${examTypeId}/fields/${fieldId}`,
  );
}

export async function reorderExaminationTypeFields(
  examTypeId: string,
  ids: number[],
): Promise<void> {
  await axios.patch(
    `/v1/masters/examination-types/${examTypeId}/fields/reorder`,
    { ids },
  );
}

export async function replaceExamTypeFieldReferenceRanges(
  examTypeId: string,
  fieldId: string,
  ranges: ExamReferenceRangeInput[],
): Promise<ExaminationTypeField> {
  const { data } = await axios.put<ExaminationTypeFieldResponse>(
    `/v1/masters/examination-types/${examTypeId}/fields/${fieldId}/reference-ranges`,
    { ranges },
  );
  return transformExaminationTypeField(data);
}

export const useGetAllExaminationTypes = () =>
  useQuery({
    queryKey: queryKeys.masters.category("examination-types"),
    queryFn: getAllExaminationTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export async function invalidateExaminationTypeFieldQueries(
  queryClient: QueryClient,
  examTypeId: string,
): Promise<void> {
  await Promise.all([
    queryClient.invalidateQueries({
      queryKey: queryKeys.masters.category("examination-types"),
    }),
    queryClient.invalidateQueries({
      queryKey: queryKeys.examinations.typeFields(examTypeId),
    }),
  ]);
}

function useInvalidateExaminationTypeFields() {
  const queryClient = useQueryClient();
  return (examTypeId: string) =>
    invalidateExaminationTypeFieldQueries(queryClient, examTypeId);
}

export const useCreateExaminationType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateExaminationTypeRequest): Promise<ExaminationTypeMaster> => {
      const { data } = await axios.post<ExaminationTypeResponse>(
        "/v1/masters/examination-types",
        req,
      );
      return transformExaminationTypeResponse(data);
    },
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: queryKeys.masters.category("examination-types"),
    }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useUpdateExaminationType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      req,
    }: {
      id: string;
      req: UpdateExaminationTypeRequest;
    }): Promise<ExaminationTypeMaster> => {
      const { data } = await axios.patch<ExaminationTypeResponse>(
        `/v1/masters/examination-types/${id}`,
        req,
      );
      return transformExaminationTypeResponse(data);
    },
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: queryKeys.masters.category("examination-types"),
    }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useDeleteExaminationType = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/examination-types/${id}`),
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: queryKeys.masters.category("examination-types"),
    }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useReorderExaminationTypes = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) =>
      axios.patch("/v1/masters/examination-types/reorder", req),
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: queryKeys.masters.category("examination-types"),
    }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export function useCreateExaminationTypeField() {
  const invalidate = useInvalidateExaminationTypeFields();
  return useMutation({
    mutationFn: ({ examTypeId, req }: {
      examTypeId: string;
      req: CreateExaminationTypeFieldRequest;
    }) => createExaminationTypeField(examTypeId, req),
    onSuccess: (_data, variables) => invalidate(variables.examTypeId),
    onError: (error) => handleApiError(error, "検査項目の作成"),
  });
}

export function useUpdateExaminationTypeField() {
  const invalidate = useInvalidateExaminationTypeFields();
  return useMutation({
    mutationFn: ({ examTypeId, fieldId, req }: {
      examTypeId: string;
      fieldId: string;
      req: UpdateExaminationTypeFieldRequest;
    }) => updateExaminationTypeField(examTypeId, fieldId, req),
    onSuccess: (_data, variables) => invalidate(variables.examTypeId),
    onError: (error) => handleApiError(error, "検査項目の更新"),
  });
}

export function useDeleteExaminationTypeField() {
  const invalidate = useInvalidateExaminationTypeFields();
  return useMutation({
    mutationFn: ({ examTypeId, fieldId }: { examTypeId: string; fieldId: string }) =>
      deleteExaminationTypeField(examTypeId, fieldId),
    onSuccess: (_data, variables) => invalidate(variables.examTypeId),
    onError: (error) => handleApiError(error, "検査項目の削除"),
  });
}

export function useReorderExaminationTypeFields() {
  const invalidate = useInvalidateExaminationTypeFields();
  return useMutation({
    mutationFn: ({ examTypeId, ids }: { examTypeId: string; ids: number[] }) =>
      reorderExaminationTypeFields(examTypeId, ids),
    onSuccess: (_data, variables) => invalidate(variables.examTypeId),
    onError: (error) => handleApiError(error, "検査項目の並び替え"),
  });
}

export function useReplaceExamTypeFieldReferenceRanges() {
  const invalidate = useInvalidateExaminationTypeFields();
  return useMutation({
    mutationFn: ({ examTypeId, fieldId, ranges }: {
      examTypeId: string;
      fieldId: string;
      ranges: ExamReferenceRangeInput[];
    }) => replaceExamTypeFieldReferenceRanges(examTypeId, fieldId, ranges),
    onSuccess: (_data, variables) => invalidate(variables.examTypeId),
    onError: (error) => handleApiError(error, "基準範囲の更新"),
  });
}
