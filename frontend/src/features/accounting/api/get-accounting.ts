import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { AccountingRecord } from "@/types";
import type { Accounting } from "../types";
import { transformAccounting, transformToAccounting } from "./transforms";
import type { BackendAccounting } from "./types";

export const getAccounting = async (id: string): Promise<AccountingRecord> => {
  const { data } = await axios.get<BackendAccounting>(`/v1/accountings/${id}`);
  return transformAccounting(data);
};

export const useGetAccounting = (id: string) => {
  return useQuery({
    queryKey: ["accounting", id],
    queryFn: () => getAccounting(id),
    enabled: !!id,
  });
};

// Accounting 詳細型（明細含む）を取得するhook
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
  petId: string
): Promise<AccountingRecord[]> => {
  const { data } = await axios.get<BackendAccounting[]>(
    `/v1/pets/${petId}/accountings`
  );
  return data.map(transformAccounting);
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
  ownerId: string
): Promise<AccountingRecord[]> => {
  const { data } = await axios.get<BackendAccounting[]>(
    `/v1/owners/${ownerId}/accountings`
  );
  return data.map(transformAccounting);
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
  status: string
): Promise<AccountingRecord[]> => {
  const { data } = await axios.get<BackendAccounting[]>(
    `/v1/accountings/status/${status}`
  );
  return data.map(transformAccounting);
};

export const useGetAccountingsByStatus = (status: string) => {
  return useQuery({
    queryKey: ["accountings", "status", status],
    queryFn: () => getAccountingsByStatus(status),
    enabled: !!status,
  });
};
