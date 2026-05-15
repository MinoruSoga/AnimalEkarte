import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface AutoManagedPrefix {
  id: number;
  prefix: string;
  category: string;
  description?: string | null;
}

export interface ConditionTagMapping {
  id: number;
  condition_code: string;
  tag_name: string;
  description?: string | null;
}

export interface SendPurposeTagPrefix {
  id: number;
  purpose: string;
  tag_prefix: string;
  description?: string | null;
}

export interface CreateAutoManagedPrefixRequest {
  prefix: string;
  category: string;
  description?: string;
}

export interface CreateConditionTagMappingRequest {
  condition_code: string;
  tag_name: string;
  description?: string;
}

export interface CreateSendPurposeTagPrefixRequest {
  purpose: string;
  tag_prefix: string;
  description?: string;
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const AUTO_MANAGED_PREFIXES_KEY = ["lstep-tag-config", "auto-managed-prefixes"] as const;
const CONDITION_TAG_MAPPINGS_KEY = ["lstep-tag-config", "condition-tag-mappings"] as const;
const SEND_PURPOSE_TAG_PREFIXES_KEY = ["lstep-tag-config", "send-purpose-tag-prefixes"] as const;

// ─────────────────────────────────────────────────
// API helpers
// ─────────────────────────────────────────────────

async function fetchAutoManagedPrefixes(): Promise<AutoManagedPrefix[]> {
  const { data } = await axios.get<AutoManagedPrefix[]>(
    "/v1/lstep-tag-config/auto-managed-prefixes",
  );
  return data;
}

async function createAutoManagedPrefix(
  req: CreateAutoManagedPrefixRequest,
): Promise<AutoManagedPrefix> {
  const { data } = await axios.post<AutoManagedPrefix>(
    "/v1/lstep-tag-config/auto-managed-prefixes",
    req,
  );
  return data;
}

async function deleteAutoManagedPrefix(id: number): Promise<void> {
  await axios.delete(`/v1/lstep-tag-config/auto-managed-prefixes/${id}`);
}

async function fetchConditionTagMappings(): Promise<ConditionTagMapping[]> {
  const { data } = await axios.get<ConditionTagMapping[]>(
    "/v1/lstep-tag-config/condition-tag-mappings",
  );
  return data;
}

async function createConditionTagMapping(
  req: CreateConditionTagMappingRequest,
): Promise<ConditionTagMapping> {
  const { data } = await axios.post<ConditionTagMapping>(
    "/v1/lstep-tag-config/condition-tag-mappings",
    req,
  );
  return data;
}

async function deleteConditionTagMapping(id: number): Promise<void> {
  await axios.delete(`/v1/lstep-tag-config/condition-tag-mappings/${id}`);
}

async function fetchSendPurposeTagPrefixes(): Promise<SendPurposeTagPrefix[]> {
  const { data } = await axios.get<SendPurposeTagPrefix[]>(
    "/v1/lstep-tag-config/send-purpose-tag-prefixes",
  );
  return data;
}

async function createSendPurposeTagPrefix(
  req: CreateSendPurposeTagPrefixRequest,
): Promise<SendPurposeTagPrefix> {
  const { data } = await axios.post<SendPurposeTagPrefix>(
    "/v1/lstep-tag-config/send-purpose-tag-prefixes",
    req,
  );
  return data;
}

async function deleteSendPurposeTagPrefix(id: number): Promise<void> {
  await axios.delete(`/v1/lstep-tag-config/send-purpose-tag-prefixes/${id}`);
}

// ─────────────────────────────────────────────────
// Hooks: auto_managed_prefixes
// ─────────────────────────────────────────────────

export function useGetAutoManagedPrefixes() {
  return useQuery({
    queryKey: AUTO_MANAGED_PREFIXES_KEY,
    queryFn: fetchAutoManagedPrefixes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateAutoManagedPrefix() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createAutoManagedPrefix,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: AUTO_MANAGED_PREFIXES_KEY });
    },
    onError: (error) => handleApiError(error, "自動管理プレフィックスの追加"),
  });
}

export function useDeleteAutoManagedPrefix() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteAutoManagedPrefix,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: AUTO_MANAGED_PREFIXES_KEY });
    },
    onError: (error) => handleApiError(error, "自動管理プレフィックスの削除"),
  });
}

// ─────────────────────────────────────────────────
// Hooks: condition_tag_mappings
// ─────────────────────────────────────────────────

export function useGetConditionTagMappings() {
  return useQuery({
    queryKey: CONDITION_TAG_MAPPINGS_KEY,
    queryFn: fetchConditionTagMappings,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateConditionTagMapping() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createConditionTagMapping,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CONDITION_TAG_MAPPINGS_KEY });
    },
    onError: (error) => handleApiError(error, "慢性疾患タグマッピングの追加"),
  });
}

export function useDeleteConditionTagMapping() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteConditionTagMapping,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CONDITION_TAG_MAPPINGS_KEY });
    },
    onError: (error) => handleApiError(error, "慢性疾患タグマッピングの削除"),
  });
}

// ─────────────────────────────────────────────────
// Hooks: send_purpose_tag_prefixes
// ─────────────────────────────────────────────────

export function useGetSendPurposeTagPrefixes() {
  return useQuery({
    queryKey: SEND_PURPOSE_TAG_PREFIXES_KEY,
    queryFn: fetchSendPurposeTagPrefixes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateSendPurposeTagPrefix() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createSendPurposeTagPrefix,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SEND_PURPOSE_TAG_PREFIXES_KEY });
    },
    onError: (error) => handleApiError(error, "LINE送信目的タグプレフィックスの追加"),
  });
}

export function useDeleteSendPurposeTagPrefix() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteSendPurposeTagPrefix,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SEND_PURPOSE_TAG_PREFIXES_KEY });
    },
    onError: (error) => handleApiError(error, "LINE送信目的タグプレフィックスの削除"),
  });
}
