import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import {
  invalidateLabDeviceBoardQueries,
  toLabDeviceJobCard,
  type LabDeviceJobCard,
} from "@/hooks/use-lab-device-unlinked";

export {
  useAttachLabDeviceJob,
  useDetachLabDeviceJob,
  useGetLabDeviceUnlinked,
  type LabDeviceJobCard,
} from "@/hooks/use-lab-device-unlinked";

// FE-RC-049: マジックナンバーの named constants 化。
const LAB_DEVICE_DEFAULT_BAUD = 9600;
const LAB_DEVICE_BOARD_REFETCH_MS = 2000;

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

interface LabDeviceAgentConsumerResponse {
  agent_consumer_token: string;
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
    unlinked: (data.unlinked ?? []).map(toLabDeviceJobCard),
    saved: (data.saved ?? []).map(toLabDeviceJobCard),
    received: (data.received ?? []).map(toLabDeviceJobCard),
    todayVisits: (data.today_visits ?? []).map(toTodayVisit),
    station: toStation(data.station),
  };
}

async function fetchLabDeviceAgentConsumerToken(): Promise<string> {
  const { data } = await axios.get<LabDeviceAgentConsumerResponse>("/v1/lab-device/agent-consumer");
  if (typeof data.agent_consumer_token !== "string" || data.agent_consumer_token === "") {
    throw new Error("invalid lab device agent consumer response");
  }
  return data.agent_consumer_token;
}

export function useGetLabDeviceBoard(enabled = true) {
  return useQuery({
    queryKey: queryKeys.labDevice.board(),
    queryFn: fetchBoard,
    enabled,
    refetchInterval: LAB_DEVICE_BOARD_REFETCH_MS,
  });
}

export function useGetLabDeviceAgentConsumer(enabled = true) {
  return useQuery({
    queryKey: queryKeys.labDevice.agentConsumer(),
    queryFn: fetchLabDeviceAgentConsumerToken,
    enabled,
    staleTime: Infinity,
  });
}

export function usePutLabDeviceWait() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (petId: number) => {
      const { data } = await axios.put<LabDeviceWaitResponse>("/v1/lab-device/wait", { pet_id: petId });
      return toWait(data);
    },
    onSuccess: () => invalidateLabDeviceBoardQueries(queryClient),
    onError: (error) => handleApiError(error, "受診中ペットの選択"),
  });
}

export function useClearLabDeviceWait() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await axios.delete("/v1/lab-device/wait");
    },
    onSuccess: () => invalidateLabDeviceBoardQueries(queryClient),
    onError: (error) => handleApiError(error, "待機の解除"),
  });
}

// FE-RC-012 followup: unhandled rejection を防ぐため mutation 側にも onError を持たせる。
// onFrame (呼び出し元 LabDeviceBoard.tsx) は status 別のトースト出し分け・重複防止(toast id)・
// 再スローも行うため、400 系はここでの汎用メッセージのみで足り caller 側は再通知しない。
// 401/500 系は機器の再送要否など安全上重要な案内が異なるため、caller 側で
// toast.dismiss() によりこの汎用トーストを差し替える（FE-RC-005 の二重通知を実質回避）。
export function useReceiveLabDeviceFrames() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: { payloadBase64: string; deviceHint: string }) => {
      const { data } = await axios.post<LabDeviceFramesResponse>("/v1/lab-device/frames", {
        payload_base64: input.payloadBase64,
        device_hint: input.deviceHint,
      });
      return data.results.map((row) => ({ duplicate: row.duplicate, job: toLabDeviceJobCard(row.job) }));
    },
    onSuccess: () => invalidateLabDeviceBoardQueries(queryClient),
    onError: (error) => handleApiError(error, "検査機器電文の受信"),
  });
}
