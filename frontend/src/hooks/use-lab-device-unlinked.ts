import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

/**
 * FE-RC-015 followup3: LabDeviceUnlinkedBanner が共有する unlinked / attach / detach の実体。
 * features → hooks は許可方向。components は本モジュールを参照（features deep import 禁止）。
 */

const LAB_DEVICE_UNLINKED_REFETCH_MS = 5000;

interface LabDeviceJobItem {
  deviceItemCode: string;
  valueRaw: string;
  unit: string;
  flag: string;
  examTypeFieldId?: number;
  needsReview: boolean;
  sortOrder: number;
}

export interface LabDeviceJobCard {
  jobId: string;
  sourceType: string;
  deviceHint: string;
  status: string;
  petId?: number;
  petName?: string;
  measuredAt?: string;
  receivedAt?: string;
  specimenIdRaw: string;
  itemCount: number;
  unmappedItemCount: number;
  clockSkew: boolean;
  items: LabDeviceJobItem[];
  /** needs_review のサーバ原因コード。旧ジョブの廃止済みコードを返す場合もある。 */
  reviewReason?: string;
}

interface LabDeviceJobItemResponse {
  device_item_code: string;
  value_raw: string;
  unit: string;
  flag: string;
  exam_type_field_id?: number;
  needs_review: boolean;
  sort_order: number;
}

interface LabDeviceJobCardResponse {
  job_id: string;
  source_type: string;
  device_hint: string;
  status: string;
  pet_id?: number;
  pet_name?: string;
  measured_at?: string;
  received_at?: string;
  specimen_id_raw: string;
  item_count: number;
  unmapped_item_count: number;
  clock_skew?: boolean;
  items: LabDeviceJobItemResponse[];
  review_reason?: string;
}

function toItem(item: LabDeviceJobItemResponse): LabDeviceJobItem {
  return {
    deviceItemCode: item.device_item_code,
    valueRaw: item.value_raw,
    unit: item.unit,
    flag: item.flag,
    examTypeFieldId: item.exam_type_field_id,
    needsReview: item.needs_review,
    sortOrder: item.sort_order,
  };
}

/** Board / frames 側でも共有する wire→UI 変換（単一実装）。 */
export function toLabDeviceJobCard(card: LabDeviceJobCardResponse): LabDeviceJobCard {
  return {
    jobId: card.job_id,
    sourceType: card.source_type,
    deviceHint: card.device_hint,
    status: card.status,
    petId: card.pet_id,
    petName: card.pet_name,
    measuredAt: card.measured_at,
    receivedAt: card.received_at,
    specimenIdRaw: card.specimen_id_raw,
    itemCount: card.item_count,
    unmappedItemCount: card.unmapped_item_count,
    clockSkew: Boolean(card.clock_skew),
    items: (card.items ?? []).map(toItem),
    reviewReason: card.review_reason,
  };
}

async function fetchLabDeviceUnlinked(): Promise<LabDeviceJobCard[]> {
  const { data } = await axios.get<LabDeviceJobCardResponse[]>("/v1/lab-device/unlinked");
  return (data ?? []).map(toLabDeviceJobCard);
}

export function invalidateLabDeviceBoardQueries(
  queryClient: ReturnType<typeof useQueryClient>,
) {
  void queryClient.invalidateQueries({ queryKey: queryKeys.labDevice.board() });
  void queryClient.invalidateQueries({ queryKey: queryKeys.labDevice.unlinked() });
}

export function useGetLabDeviceUnlinked(enabled = true) {
  return useQuery({
    queryKey: queryKeys.labDevice.unlinked(),
    queryFn: fetchLabDeviceUnlinked,
    enabled,
    refetchInterval: LAB_DEVICE_UNLINKED_REFETCH_MS,
  });
}

export function useAttachLabDeviceJob() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: { jobId: string; petId: number }) => {
      const { data } = await axios.post<LabDeviceJobCardResponse>(
        `/v1/lab-imports/${input.jobId}/attach`,
        { pet_id: input.petId },
      );
      return toLabDeviceJobCard(data);
    },
    onSuccess: () => {
      invalidateLabDeviceBoardQueries(queryClient);
      void queryClient.invalidateQueries({ queryKey: queryKeys.examinations.all() });
    },
    onError: (error) => handleApiError(error, "検査結果の紐付け"),
  });
}

export function useDetachLabDeviceJob() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (jobId: string) => {
      const { data } = await axios.post<LabDeviceJobCardResponse>(
        `/v1/lab-imports/${jobId}/detach`,
      );
      return toLabDeviceJobCard(data);
    },
    onSuccess: () => {
      invalidateLabDeviceBoardQueries(queryClient);
      void queryClient.invalidateQueries({ queryKey: queryKeys.examinations.all() });
    },
    onError: (error) => handleApiError(error, "検査結果の紐付け解除"),
  });
}
