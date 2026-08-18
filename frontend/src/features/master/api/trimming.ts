import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type {
  TrimmingCourse as ModelTrimmingCourse,
  TrimmingOption as ModelTrimmingOption,
  TargetSize,
} from "@/types/generated/models";

export type { TargetSize };

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

export const TARGET_SIZE_LABELS: Record<string, string> = {
  small: "小型",
  medium: "中型",
  large: "大型",
  cat: "猫",
};

export const TARGET_SIZE_OPTIONS: { value: string; label: string }[] = [
  { value: "small", label: "小型" },
  { value: "medium", label: "中型" },
  { value: "large", label: "大型" },
  { value: "cat", label: "猫" },
];

// ─────────────────────────────────────────────────
// Request types (derived from models.ts, with nullable price/target_size/duration overrides)
// ─────────────────────────────────────────────────

type TrimmingCourseBase = Omit<
  ModelTrimmingCourse,
  "id" | "clinic_id" | "created_at" | "updated_at" | "price" | "target_size" | "duration"
> & {
  price?: number | null;
  target_size?: TargetSize | null;
  duration?: number | null;
};

export type CreateTrimmingCourseRequest = Pick<TrimmingCourseBase, "name"> &
  Partial<Omit<TrimmingCourseBase, "name">>;

export type UpdateTrimmingCourseRequest = Partial<TrimmingCourseBase>;

type TrimmingOptionBase = Omit<
  ModelTrimmingOption,
  "id" | "clinic_id" | "created_at" | "updated_at" | "price" | "duration"
> & {
  price?: number | null;
  duration?: number | null;
};

export type CreateTrimmingOptionRequest = Pick<TrimmingOptionBase, "name"> &
  Partial<Omit<TrimmingOptionBase, "name">>;

export type UpdateTrimmingOptionRequest = Partial<TrimmingOptionBase>;

// ─────────────────────────────────────────────────
// Transform functions
// ─────────────────────────────────────────────────

/** キャッシュ汚染時の MasterItem.status も吸収して有効/無効を判定する */
export function resolveTrimmingActiveFlag(item: {
  is_active?: boolean;
  isActive?: boolean;
  status?: string;
}): boolean {
  if (typeof item.is_active === "boolean") return item.is_active;
  if (typeof item.isActive === "boolean") return item.isActive;
  return item.status === "active";
}

function transformTrimmingCourse(data: ModelTrimmingCourse) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    price: data.price ?? null,
    isActive: resolveTrimmingActiveFlag(data),
    description: data.description,
    targetSize: (data.target_size as TargetSize) ?? null,
    courseTypeId: data.course_type_id != null ? String(data.course_type_id) : null,
    duration: data.duration ?? null,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type TrimmingCourse = ReturnType<typeof transformTrimmingCourse>;

function transformTrimmingOption(data: ModelTrimmingOption) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    price: data.price ?? null,
    isActive: resolveTrimmingActiveFlag(data),
    description: data.description,
    duration: data.duration ?? null,
    combinable: data.is_combinable,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type TrimmingOption = ReturnType<typeof transformTrimmingOption>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

// P8: useMasterItems("trimmingCourse") と queryKey を統一（キャッシュ無効化が機能するため）
// ─────────────────────────────────────────────────
// API functions - TrimmingCourse
// ─────────────────────────────────────────────────

async function listTrimmingCourses(): Promise<TrimmingCourse[]> {
  const { data } = await axios.get<ModelTrimmingCourse[]>(
    "/v1/masters/trimming-courses",
  );
  return data.map(transformTrimmingCourse);
}

async function createTrimmingCourse(
  req: CreateTrimmingCourseRequest,
): Promise<TrimmingCourse> {
  const { data } = await axios.post<ModelTrimmingCourse>(
    "/v1/masters/trimming-courses",
    req,
  );
  return transformTrimmingCourse(data);
}

async function updateTrimmingCourse(
  id: string,
  req: UpdateTrimmingCourseRequest,
): Promise<TrimmingCourse> {
  const { data } = await axios.patch<ModelTrimmingCourse>(
    `/v1/masters/trimming-courses/${id}`,
    req,
  );
  return transformTrimmingCourse(data);
}

async function deleteTrimmingCourse(id: string): Promise<void> {
  await axios.delete(`/v1/masters/trimming-courses/${id}`);
}

// ─────────────────────────────────────────────────
// API functions - TrimmingOption
// ─────────────────────────────────────────────────

async function listTrimmingOptions(): Promise<TrimmingOption[]> {
  const { data } = await axios.get<ModelTrimmingOption[]>(
    "/v1/masters/trimming-options",
  );
  return data.map(transformTrimmingOption);
}

async function createTrimmingOption(
  req: CreateTrimmingOptionRequest,
): Promise<TrimmingOption> {
  const { data } = await axios.post<ModelTrimmingOption>(
    "/v1/masters/trimming-options",
    req,
  );
  return transformTrimmingOption(data);
}

async function updateTrimmingOption(
  id: string,
  req: UpdateTrimmingOptionRequest,
): Promise<TrimmingOption> {
  const { data } = await axios.patch<ModelTrimmingOption>(
    `/v1/masters/trimming-options/${id}`,
    req,
  );
  return transformTrimmingOption(data);
}

async function deleteTrimmingOption(id: string): Promise<void> {
  await axios.delete(`/v1/masters/trimming-options/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks - TrimmingCourse
// ─────────────────────────────────────────────────

export function useGetTrimmingCourses() {
  return useQuery({
    queryKey: queryKeys.masters.trimmingCoursesFull(),
    queryFn: listTrimmingCourses,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateTrimmingCourse() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createTrimmingCourse,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("trimmingCourse") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateTrimmingCourse() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTrimmingCourseRequest }) =>
      updateTrimmingCourse(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("trimmingCourse") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteTrimmingCourse() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTrimmingCourse,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("trimmingCourse") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}


// ─────────────────────────────────────────────────
// TanStack Query hooks - TrimmingOption
// ─────────────────────────────────────────────────

export function useGetTrimmingOptions() {
  return useQuery({
    queryKey: queryKeys.masters.category("trimming-options"),
    queryFn: listTrimmingOptions,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateTrimmingOption() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createTrimmingOption,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("trimming-options") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateTrimmingOption() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTrimmingOptionRequest }) =>
      updateTrimmingOption(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("trimming-options") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteTrimmingOption() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTrimmingOption,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("trimming-options") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

