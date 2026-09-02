import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";

import type { UpdateLabDeviceItemMasterRequest } from "./lab-device-item-masters";

type LabDeviceResponse = {
  id: number;
  source_type: string;
  name: string;
  exam_type_id?: number | null;
  is_active: boolean;
  sort_order: number;
};

export type CreateLabDeviceRequest = {
  name: string;
  source_type: string;
  exam_type_id: number | null;
  is_active: boolean;
  sort_order: number;
};

export type UpdateLabDeviceRequest = {
  name: string;
  exam_type_id: number | null;
  is_active: boolean;
  sort_order: number;
};

export type SaveLabDeviceConfigurationRequest = {
  device: UpdateLabDeviceRequest;
  items: Array<UpdateLabDeviceItemMasterRequest & { id: number }>;
};

function transformLabDevice(data: LabDeviceResponse) {
  return {
    id: String(data.id),
    sourceType: data.source_type,
    name: data.name,
    examTypeId:
      data.exam_type_id === undefined || data.exam_type_id === null
        ? null
        : String(data.exam_type_id),
    isActive: data.is_active,
    sortOrder: data.sort_order,
  };
}

export type LabDevice = ReturnType<typeof transformLabDevice>;

const queryKey = queryKeys.masters.category("lab-devices");

const getLabDevices = async (): Promise<LabDevice[]> => {
  const { data } = await axios.get<LabDeviceResponse[]>("/v1/lab-devices");
  return data.map(transformLabDevice);
};

const createLabDevice = async (req: CreateLabDeviceRequest): Promise<LabDevice> => {
  const { data } = await axios.post<LabDeviceResponse>("/v1/lab-devices", req);
  return transformLabDevice(data);
};

const saveLabDeviceConfiguration = async (id: string, req: SaveLabDeviceConfigurationRequest): Promise<void> => {
  await axios.put(`/v1/lab-devices/${id}/configuration`, req);
};

export const useGetLabDevices = () =>
  useQuery({
    queryKey,
    queryFn: getLabDevices,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateLabDevice = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createLabDevice,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey });
    },
    onError: (error) => handleApiError(error, "登録"),
  });
};

export const useSaveLabDeviceConfiguration = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: SaveLabDeviceConfigurationRequest }) =>
      saveLabDeviceConfiguration(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey });
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("lab-device-item-masters") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};
