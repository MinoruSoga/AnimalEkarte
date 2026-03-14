import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type {
  TrimmingCourse as ModelTrimmingCourse,
  TrimmingOption as ModelTrimmingOption,
} from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

export type TargetSize = "small" | "medium" | "large" | "cat";

export const TARGET_SIZE_LABELS: Record<TargetSize, string> = {
  small: "小型",
  medium: "中型",
  large: "大型",
  cat: "猫",
};

export const TARGET_SIZE_OPTIONS: { value: TargetSize; label: string }[] = [
  { value: "small", label: "小型" },
  { value: "medium", label: "中型" },
  { value: "large", label: "大型" },
  { value: "cat", label: "猫" },
];

// ─────────────────────────────────────────────────
// Frontend display types (camelCase)
// ─────────────────────────────────────────────────

export interface TrimmingCourse {
  id: string;
  clinicId: string;
  name: string;
  price: number | null;
  isActive: boolean;
  description: string;
  targetSize: TargetSize | null;
  duration: number | null;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface TrimmingOption {
  id: string;
  clinicId: string;
  name: string;
  price: number | null;
  isActive: boolean;
  description: string;
  duration: number | null;
  combinable: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreateTrimmingCourseRequest {
  name: string;
  price?: number | null;
  description?: string;
  target_size?: TargetSize | null;
  duration?: number | null;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateTrimmingCourseRequest {
  name?: string;
  price?: number | null;
  description?: string;
  target_size?: TargetSize | null;
  duration?: number | null;
  is_active?: boolean;
  sort_order?: number;
}

export interface CreateTrimmingOptionRequest {
  name: string;
  price?: number | null;
  description?: string;
  duration?: number | null;
  combinable?: boolean;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateTrimmingOptionRequest {
  name?: string;
  price?: number | null;
  description?: string;
  duration?: number | null;
  combinable?: boolean;
  is_active?: boolean;
  sort_order?: number;
}

// ─────────────────────────────────────────────────
// Transform functions
// ─────────────────────────────────────────────────

function transformTrimmingCourse(data: ModelTrimmingCourse): TrimmingCourse {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    price: data.price ?? null,
    isActive: data.is_active,
    description: data.description,
    targetSize: (data.target_size as TargetSize) ?? null,
    duration: data.duration ?? null,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

function transformTrimmingOption(data: ModelTrimmingOption): TrimmingOption {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    price: data.price ?? null,
    isActive: data.is_active,
    description: data.description,
    duration: data.duration ?? null,
    combinable: data.combinable,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const TRIMMING_COURSES_KEY = ["masters", "trimming-courses"] as const;
const TRIMMING_OPTIONS_KEY = ["masters", "trimming-options"] as const;

// ─────────────────────────────────────────────────
// API functions - TrimmingCourse
// ─────────────────────────────────────────────────

export async function listTrimmingCourses(): Promise<TrimmingCourse[]> {
  const { data } = await axios.get<ModelTrimmingCourse[]>(
    "/v1/masters/trimming-courses",
  );
  return data.map(transformTrimmingCourse);
}

export async function createTrimmingCourse(
  req: CreateTrimmingCourseRequest,
): Promise<TrimmingCourse> {
  const { data } = await axios.post<ModelTrimmingCourse>(
    "/v1/masters/trimming-courses",
    req,
  );
  return transformTrimmingCourse(data);
}

export async function updateTrimmingCourse(
  id: string,
  req: UpdateTrimmingCourseRequest,
): Promise<TrimmingCourse> {
  const { data } = await axios.patch<ModelTrimmingCourse>(
    `/v1/masters/trimming-courses/${id}`,
    req,
  );
  return transformTrimmingCourse(data);
}

export async function deleteTrimmingCourse(id: string): Promise<void> {
  await axios.delete(`/v1/masters/trimming-courses/${id}`);
}

// ─────────────────────────────────────────────────
// API functions - TrimmingOption
// ─────────────────────────────────────────────────

export async function listTrimmingOptions(): Promise<TrimmingOption[]> {
  const { data } = await axios.get<ModelTrimmingOption[]>(
    "/v1/masters/trimming-options",
  );
  return data.map(transformTrimmingOption);
}

export async function createTrimmingOption(
  req: CreateTrimmingOptionRequest,
): Promise<TrimmingOption> {
  const { data } = await axios.post<ModelTrimmingOption>(
    "/v1/masters/trimming-options",
    req,
  );
  return transformTrimmingOption(data);
}

export async function updateTrimmingOption(
  id: string,
  req: UpdateTrimmingOptionRequest,
): Promise<TrimmingOption> {
  const { data } = await axios.patch<ModelTrimmingOption>(
    `/v1/masters/trimming-options/${id}`,
    req,
  );
  return transformTrimmingOption(data);
}

export async function deleteTrimmingOption(id: string): Promise<void> {
  await axios.delete(`/v1/masters/trimming-options/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks - TrimmingCourse
// ─────────────────────────────────────────────────

export function useListTrimmingCourses() {
  return useQuery({
    queryKey: TRIMMING_COURSES_KEY,
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
      queryClient.invalidateQueries({ queryKey: TRIMMING_COURSES_KEY });
    },
  });
}

export function useUpdateTrimmingCourse() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTrimmingCourseRequest }) =>
      updateTrimmingCourse(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: TRIMMING_COURSES_KEY });
    },
  });
}

export function useDeleteTrimmingCourse() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTrimmingCourse,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: TRIMMING_COURSES_KEY });
    },
  });
}

// ─────────────────────────────────────────────────
// TanStack Query hooks - TrimmingOption
// ─────────────────────────────────────────────────

export function useListTrimmingOptions() {
  return useQuery({
    queryKey: TRIMMING_OPTIONS_KEY,
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
      queryClient.invalidateQueries({ queryKey: TRIMMING_OPTIONS_KEY });
    },
  });
}

export function useUpdateTrimmingOption() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateTrimmingOptionRequest }) =>
      updateTrimmingOption(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: TRIMMING_OPTIONS_KEY });
    },
  });
}

export function useDeleteTrimmingOption() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTrimmingOption,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: TRIMMING_OPTIONS_KEY });
    },
  });
}
