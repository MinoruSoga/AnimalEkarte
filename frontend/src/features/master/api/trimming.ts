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

/**
 * EMR-457 の任意属性。GORM 由来の生成型には含まれない（model 側 json:"-"）ため
 * wire 契約（backend/docs/api.yaml の TrimmingCourse）に合わせてここで定義する。
 * allowed_weekdays は 0=日〜6=土、null/空配列は全曜日。
 */
interface TrimmingCourseAttributeWire {
  short_name?: string | null;
  image_object_key?: string | null;
  image_url?: string | null;
  allowed_weekdays?: number[] | null;
}

/** 書込: 省略は保持、"" と [] は未登録（全曜日）へ戻す。image_url は読取専用。 */
interface TrimmingCourseAttributeInput {
  short_name?: string;
  image_object_key?: string;
  allowed_weekdays?: number[];
}

type TrimmingCourseWire = ModelTrimmingCourse & TrimmingCourseAttributeWire;

export type CreateTrimmingCourseRequest = Pick<TrimmingCourseBase, "name"> &
  Partial<Omit<TrimmingCourseBase, "name">> &
  TrimmingCourseAttributeInput;

export type UpdateTrimmingCourseRequest = Partial<TrimmingCourseBase> &
  TrimmingCourseAttributeInput;

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

function transformTrimmingCourse(data: TrimmingCourseWire) {
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
    shortName: data.short_name ?? null,
    imageObjectKey: data.image_object_key ?? null,
    imageUrl: data.image_url ?? null,
    allowedWeekdays: data.allowed_weekdays ?? null,
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

// P8: useGetMasterItems("trimmingCourse") と queryKey を統一（キャッシュ無効化が機能するため）
// ─────────────────────────────────────────────────
// API functions - TrimmingCourse
// ─────────────────────────────────────────────────

async function listTrimmingCourses(): Promise<TrimmingCourse[]> {
  const { data } = await axios.get<TrimmingCourseWire[]>("/v1/masters/trimming-courses");
  return data.map(transformTrimmingCourse);
}

async function createTrimmingCourse(req: CreateTrimmingCourseRequest): Promise<TrimmingCourse> {
  const { data } = await axios.post<TrimmingCourseWire>("/v1/masters/trimming-courses", req);
  return transformTrimmingCourse(data);
}

async function updateTrimmingCourse(
  id: string,
  req: UpdateTrimmingCourseRequest,
): Promise<TrimmingCourse> {
  const { data } = await axios.patch<TrimmingCourseWire>(`/v1/masters/trimming-courses/${id}`, req);
  return transformTrimmingCourse(data);
}

async function deleteTrimmingCourse(id: string): Promise<void> {
  await axios.delete(`/v1/masters/trimming-courses/${id}`);
}

// ─────────────────────────────────────────────────
// API functions - TrimmingOption
// ─────────────────────────────────────────────────

async function listTrimmingOptions(): Promise<TrimmingOption[]> {
  const { data } = await axios.get<ModelTrimmingOption[]>("/v1/masters/trimming-options");
  return data.map(transformTrimmingOption);
}

async function createTrimmingOption(req: CreateTrimmingOptionRequest): Promise<TrimmingOption> {
  const { data } = await axios.post<ModelTrimmingOption>("/v1/masters/trimming-options", req);
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

// BE の署名付き画像 URL の TTL は 15 分（trimmingImageSignedURLTTL）。
// staleTime=STATIC(30分)・refetchOnWindowFocus=false のため、TTL 未満で
// 定期再取得しないと後から開いたパネルのプレビュー URL が失効する。
const TRIMMING_COURSE_LIST_REFETCH_INTERVAL = 10 * 60 * 1000;

export function useGetTrimmingCourses() {
  return useQuery({
    queryKey: queryKeys.masters.trimmingCoursesFull(),
    queryFn: listTrimmingCourses,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    refetchInterval: TRIMMING_COURSE_LIST_REFETCH_INTERVAL,
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
