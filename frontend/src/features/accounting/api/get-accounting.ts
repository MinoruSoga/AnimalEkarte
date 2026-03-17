import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Accounting } from "../types";
import { transformToAccounting } from "./transforms";
import type { BackendAccounting } from "./types";

interface AccountingListResponse {
  data: BackendAccounting[];
  total: number;
  page: number;
  limit: number;
}

export const getAccounting = async (id: string): Promise<Accounting> => {
  const { data } = await axios.get<BackendAccounting>(`/v1/accountings/${id}`);
  return transformToAccounting(data);
};

export const useGetAccounting = (id: string) => {
  return useQuery({
    queryKey: ["accounting", id],
    queryFn: () => getAccounting(id),
    enabled: !!id,
  });
};

// Accounting 詳細型（明細含む）を取得する hook
export const getAccountingDetail = async (id: string): Promise<Accounting> => {
  const { data } = await axios.get<BackendAccounting>(`/v1/accountings/${id}`);
  return transformToAccounting(data);
};

export const useGetAccountingDetail = (id: string | undefined) => {
  return useQuery({
    queryKey: ["accounting-detail", id],
    queryFn: () => getAccountingDetail(id!),
    enabled: !!id,
  });
};

// Fetch accountings by pet ID
export const getAccountingsByPetId = async (
  petId: string,
): Promise<Accounting[]> => {
  const { data } = await axios.get<AccountingListResponse>("/v1/accountings", {
    params: { pet_id: petId },
  });
  return data.data.map(transformToAccounting);
};

export const useGetAccountingsByPetId = (petId: string) => {
  return useQuery({
    queryKey: ["accountings", "pet", petId],
    queryFn: () => getAccountingsByPetId(petId),
    enabled: !!petId,
  });
};

// Fetch accountings by owner ID
export const getAccountingsByOwnerId = async (
  ownerId: string,
): Promise<Accounting[]> => {
  const { data } = await axios.get<AccountingListResponse>("/v1/accountings", {
    params: { owner_id: ownerId },
  });
  return data.data.map(transformToAccounting);
};

export const useGetAccountingsByOwnerId = (ownerId: string) => {
  return useQuery({
    queryKey: ["accountings", "owner", ownerId],
    queryFn: () => getAccountingsByOwnerId(ownerId),
    enabled: !!ownerId,
  });
};

// Fetch accountings by status
export const getAccountingsByStatus = async (
  status: string,
): Promise<Accounting[]> => {
  const { data } = await axios.get<AccountingListResponse>(
    `/v1/accountings/status/${status}`,
  );
  return data.data.map(transformToAccounting);
};

export const useGetAccountingsByStatus = (status: string) => {
  return useQuery({
    queryKey: ["accountings", "status", status],
    queryFn: () => getAccountingsByStatus(status),
    enabled: !!status,
  });
};
