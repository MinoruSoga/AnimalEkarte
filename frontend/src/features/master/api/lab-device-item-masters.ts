import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";

export type LabDeviceItemMasterResponse = {
  id: number;
  source_type: string;
  device_item_code: string;
  display_name: string;
  unit: string;
  value_shape: string;
  exam_type_field_id?: number | null;
  sort_order: number;
  is_active: boolean;
};

export type LabDeviceItemMasterEnsureApiResponse = {
  inserted_count: number;
  items: LabDeviceItemMasterResponse[];
};

export type UpdateLabDeviceItemMasterRequest = {
  display_name: string;
  unit: string;
  exam_type_field_id: number | null;
  is_active: boolean;
};

function transformLabDeviceItemMaster(data: LabDeviceItemMasterResponse) {
  return {
    id: String(data.id),
    sourceType: data.source_type,
    deviceItemCode: data.device_item_code,
    displayName: data.display_name,
    unit: data.unit,
    valueShape: data.value_shape,
    examTypeFieldId:
      data.exam_type_field_id === undefined || data.exam_type_field_id === null
        ? null
        : String(data.exam_type_field_id),
    sortOrder: data.sort_order,
    isActive: data.is_active,
  };
}

export type LabDeviceItemMaster = ReturnType<typeof transformLabDeviceItemMaster>;

const queryKey = queryKeys.masters.category("lab-device-item-masters");

const getLabDeviceItemMasters = async (): Promise<LabDeviceItemMaster[]> => {
  const { data } = await axios.get<LabDeviceItemMasterResponse[]>("/v1/lab-device-item-masters");
  return data.map(transformLabDeviceItemMaster);
};

const ensureLabDeviceItemMasters = async (): Promise<{
  insertedCount: number;
  items: LabDeviceItemMaster[];
}> => {
  const { data } = await axios.post<LabDeviceItemMasterEnsureApiResponse>(
    "/v1/lab-device-item-masters/ensure",
  );
  return {
    insertedCount: data.inserted_count,
    items: data.items.map(transformLabDeviceItemMaster),
  };
};

const updateLabDeviceItemMaster = async (
  id: string,
  req: UpdateLabDeviceItemMasterRequest,
): Promise<LabDeviceItemMaster> => {
  const { data } = await axios.patch<LabDeviceItemMasterResponse>(
    `/v1/lab-device-item-masters/${id}`,
    req,
  );
  return transformLabDeviceItemMaster(data);
};

export const useGetLabDeviceItemMasters = () =>
  useQuery({
    queryKey,
    queryFn: getLabDeviceItemMasters,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useEnsureLabDeviceItemMasters = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ensureLabDeviceItemMasters,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey });
    },
    onError: (error) => handleApiError(error, "既定項目の投入"),
  });
};

export const useUpdateLabDeviceItemMaster = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateLabDeviceItemMasterRequest }) =>
      updateLabDeviceItemMaster(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};
