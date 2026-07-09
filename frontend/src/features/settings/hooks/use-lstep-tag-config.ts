import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import {
  fetchAutoManagedPrefixes,
  createAutoManagedPrefix,
  deleteAutoManagedPrefix,
  fetchConditionTagMappings,
  createConditionTagMapping,
  deleteConditionTagMapping,
  fetchSendPurposeTagPrefixes,
  createSendPurposeTagPrefix,
  deleteSendPurposeTagPrefix,
} from "../api/lstep-tag-config";

export type {
  AutoManagedPrefix,
  ConditionTagMapping,
  SendPurposeTagPrefix,
  CreateAutoManagedPrefixRequest,
  CreateConditionTagMappingRequest,
  CreateSendPurposeTagPrefixRequest,
} from "../api/lstep-tag-config";

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const AUTO_MANAGED_PREFIXES_KEY = ["lstep-tag-config", "auto-managed-prefixes"] as const;
const CONDITION_TAG_MAPPINGS_KEY = ["lstep-tag-config", "condition-tag-mappings"] as const;
const SEND_PURPOSE_TAG_PREFIXES_KEY = ["lstep-tag-config", "send-purpose-tag-prefixes"] as const;

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
