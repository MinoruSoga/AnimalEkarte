import { axios } from "@/lib/axios";

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
// API functions
// ─────────────────────────────────────────────────

export async function fetchAutoManagedPrefixes(): Promise<AutoManagedPrefix[]> {
  const { data } = await axios.get<AutoManagedPrefix[]>(
    "/v1/lstep-tag-config/auto-managed-prefixes",
  );
  return data;
}

export async function createAutoManagedPrefix(
  req: CreateAutoManagedPrefixRequest,
): Promise<AutoManagedPrefix> {
  const { data } = await axios.post<AutoManagedPrefix>(
    "/v1/lstep-tag-config/auto-managed-prefixes",
    req,
  );
  return data;
}

export async function deleteAutoManagedPrefix(id: number): Promise<void> {
  await axios.delete(`/v1/lstep-tag-config/auto-managed-prefixes/${id}`);
}

export async function fetchConditionTagMappings(): Promise<ConditionTagMapping[]> {
  const { data } = await axios.get<ConditionTagMapping[]>(
    "/v1/lstep-tag-config/condition-tag-mappings",
  );
  return data;
}

export async function createConditionTagMapping(
  req: CreateConditionTagMappingRequest,
): Promise<ConditionTagMapping> {
  const { data } = await axios.post<ConditionTagMapping>(
    "/v1/lstep-tag-config/condition-tag-mappings",
    req,
  );
  return data;
}

export async function deleteConditionTagMapping(id: number): Promise<void> {
  await axios.delete(`/v1/lstep-tag-config/condition-tag-mappings/${id}`);
}

export async function fetchSendPurposeTagPrefixes(): Promise<SendPurposeTagPrefix[]> {
  const { data } = await axios.get<SendPurposeTagPrefix[]>(
    "/v1/lstep-tag-config/send-purpose-tag-prefixes",
  );
  return data;
}

export async function createSendPurposeTagPrefix(
  req: CreateSendPurposeTagPrefixRequest,
): Promise<SendPurposeTagPrefix> {
  const { data } = await axios.post<SendPurposeTagPrefix>(
    "/v1/lstep-tag-config/send-purpose-tag-prefixes",
    req,
  );
  return data;
}

export async function deleteSendPurposeTagPrefix(id: number): Promise<void> {
  await axios.delete(`/v1/lstep-tag-config/send-purpose-tag-prefixes/${id}`);
}
