import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

// FE-RC-049: マジックナンバーの named constants 化。
const LAB_DEVICE_DEFAULT_BAUD = 9600;
const LAB_DEVICE_BOARD_REFETCH_MS = 2000;
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

export interface LabDeviceWait {
  petId: number;
  petName: string;
  staffId: number;
  expiresAt: string;
}

interface LabDeviceStation {
  waitTtlSeconds: number;
  slotsJson: string;
}

export interface LabDeviceTodayVisit {
  recordId: number;
  petId: number;
  petName: string;
  ownerName: string;
  species: string;
  doctorName: string;
  visitType: string;
  petIsDeceased?: boolean;
}

interface LabDeviceBoard {
  wait: LabDeviceWait | null;
  unlinked: LabDeviceJobCard[];
  saved: LabDeviceJobCard[];
  received: LabDeviceJobCard[];
  todayVisits: LabDeviceTodayVisit[];
  station: LabDeviceStation;
}

type LabDeviceSlotParity = "none" | "even" | "odd";

export interface LabDeviceSlot {
  key: string;
  sourceType: string;
  deviceHint: string;
  baud: number;
  parity?: LabDeviceSlotParity;
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

interface LabDeviceWaitResponse {
  pet_id: number;
  pet_name: string;
  staff_id: number;
  expires_at: string;
}

interface LabDeviceStationResponse {
  wait_ttl_seconds: number;
  slots_json: string;
}

interface LabDeviceTodayVisitResponse {
  record_id: number;
  pet_id: number;
  pet_name: string;
  owner_name: string;
  species: string;
  doctor_name: string;
  visit_type: string;
  pet_is_deceased?: boolean;
}

interface LabDeviceBoardResponse {
  wait: LabDeviceWaitResponse | null;
  unlinked: LabDeviceJobCardResponse[];
  saved: LabDeviceJobCardResponse[];
  received: LabDeviceJobCardResponse[];
  today_visits: LabDeviceTodayVisitResponse[];
  station: LabDeviceStationResponse;
}

interface LabDeviceFramesResponse {
  results: Array<{ duplicate: boolean; job: LabDeviceJobCardResponse }>;
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

function toCard(card: LabDeviceJobCardResponse): LabDeviceJobCard {
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

function toWait(wait: LabDeviceWaitResponse): LabDeviceWait {
  return {
    petId: wait.pet_id,
    petName: wait.pet_name,
    staffId: wait.staff_id,
    expiresAt: wait.expires_at,
  };
}

function toStation(station: LabDeviceStationResponse): LabDeviceStation {
  return {
    waitTtlSeconds: station.wait_ttl_seconds,
    slotsJson: station.slots_json,
  };
}

function toTodayVisit(visit: LabDeviceTodayVisitResponse): LabDeviceTodayVisit {
  return {
    recordId: visit.record_id,
    petId: visit.pet_id,
    petName: visit.pet_name,
    ownerName: visit.owner_name,
    species: visit.species,
    doctorName: visit.doctor_name,
    visitType: visit.visit_type,
    petIsDeceased: visit.pet_is_deceased,
  };
}

function parseLabDeviceSlotParity(value: unknown): LabDeviceSlotParity | undefined {
  return value === "none" || value === "even" || value === "odd" ? value : undefined;
}

// FE-RC-037: サーバ由来の JSON 文字列を無検証キャストせず、配列であることをまず確認する。
function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function parseLabDeviceSlots(slotsJson: string): LabDeviceSlot[] {
  try {
    const parsed: unknown = JSON.parse(slotsJson);
    if (!Array.isArray(parsed)) {
      return [];
    }
    return parsed.filter(isRecord).map((slot) => ({
      key: String(slot.key ?? ""),
      sourceType: String(slot.source_type ?? ""),
      deviceHint: String(slot.device_hint ?? ""),
      baud: typeof slot.baud === "number" ? slot.baud : LAB_DEVICE_DEFAULT_BAUD,
      parity: parseLabDeviceSlotParity(slot.parity),
    }));
  } catch {
    return [];
  }
}

async function fetchBoard(): Promise<LabDeviceBoard> {
  const { data } = await axios.get<LabDeviceBoardResponse>("/v1/lab-device/board");
  return {
    wait: data.wait ? toWait(data.wait) : null,
    unlinked: (data.unlinked ?? []).map(toCard),
    saved: (data.saved ?? []).map(toCard),
    received: (data.received ?? []).map(toCard),
    todayVisits: (data.today_visits ?? []).map(toTodayVisit),
    station: toStation(data.station),
  };
}

async function fetchUnlinked(): Promise<LabDeviceJobCard[]> {
  const { data } = await axios.get<LabDeviceJobCardResponse[]>("/v1/lab-device/unlinked");
  return (data ?? []).map(toCard);
}

export function useGetLabDeviceBoard(enabled = true) {
  return useQuery({
    queryKey: queryKeys.labDevice.board(),
    queryFn: fetchBoard,
    enabled,
    refetchInterval: LAB_DEVICE_BOARD_REFETCH_MS,
  });
}

export function useGetLabDeviceUnlinked(enabled = true) {
  return useQuery({
    queryKey: queryKeys.labDevice.unlinked(),
    queryFn: fetchUnlinked,
    enabled,
    refetchInterval: LAB_DEVICE_UNLINKED_REFETCH_MS,
  });
}

function invalidateBoard(queryClient: ReturnType<typeof useQueryClient>) {
  void queryClient.invalidateQueries({ queryKey: queryKeys.labDevice.board() });
  void queryClient.invalidateQueries({ queryKey: queryKeys.labDevice.unlinked() });
}

export function usePutLabDeviceWait() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (petId: number) => {
      const { data } = await axios.put<LabDeviceWaitResponse>("/v1/lab-device/wait", { pet_id: petId });
      return toWait(data);
    },
    onSuccess: () => invalidateBoard(queryClient),
    onError: (error) => handleApiError(error, "受診中ペットの選択"),
  });
}

export function useClearLabDeviceWait() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await axios.delete("/v1/lab-device/wait");
    },
    onSuccess: () => invalidateBoard(queryClient),
    onError: (error) => handleApiError(error, "待機の解除"),
  });
}

// FE-RC-012: onFrame (呼び出し元) が status 別のトースト出し分け・重複防止(toast id)・再スロー
// まで一貫して担っているため、ここに handleApiError を追加すると FE-RC-005 で問題視される
// トースト二重表示を再発させる。呼び出し元の catch が唯一のエラー通知経路である。
export function useReceiveLabDeviceFrames() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: { payloadBase64: string; deviceHint: string }) => {
      const { data } = await axios.post<LabDeviceFramesResponse>("/v1/lab-device/frames", {
        payload_base64: input.payloadBase64,
        device_hint: input.deviceHint,
      });
      return data.results.map((row) => ({ duplicate: row.duplicate, job: toCard(row.job) }));
    },
    onSuccess: () => invalidateBoard(queryClient),
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
      return toCard(data);
    },
    onSuccess: () => {
      invalidateBoard(queryClient);
      void queryClient.invalidateQueries({ queryKey: queryKeys.examinations.all() });
    },
    onError: (error) => handleApiError(error, "検査結果の紐付け"),
  });
}

export function useDetachLabDeviceJob() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (jobId: string) => {
      const { data } = await axios.post<LabDeviceJobCardResponse>(`/v1/lab-imports/${jobId}/detach`);
      return toCard(data);
    },
    onSuccess: () => {
      invalidateBoard(queryClient);
      void queryClient.invalidateQueries({ queryKey: queryKeys.examinations.all() });
    },
    onError: (error) => handleApiError(error, "検査結果の紐付け解除"),
  });
}
