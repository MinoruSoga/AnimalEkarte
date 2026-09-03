// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { handleApiError } from "@/lib/handle-api-error";

// Types
import type {
  ApiDailyRecord,
  CreateVitalRecordRequest,
  CreateCareLogRequest,
  CreateStaffNoteRequest,
} from "./daily-records-types";

// ---- Fetchers ----

const getDailyRecord = async (hospitalizationId: string, date: string): Promise<ApiDailyRecord> => {
  const { data } = await axios.get<ApiDailyRecord>(
    `/v1/hospitalizations/${hospitalizationId}/daily-records/${date}`,
  );
  return data;
};

const createDailyRecord = async (
  hospitalizationId: string,
  date: string,
): Promise<ApiDailyRecord> => {
  const { data } = await axios.post<ApiDailyRecord>(
    `/v1/hospitalizations/${hospitalizationId}/daily-records`,
    { date },
  );
  return data;
};

const addVitalRecord = async (
  hospitalizationId: string,
  date: string,
  payload: CreateVitalRecordRequest,
): Promise<ApiDailyRecord> => {
  const { data } = await axios.post<ApiDailyRecord>(
    `/v1/hospitalizations/${hospitalizationId}/daily-records/${date}/vitals`,
    payload,
  );
  return data;
};

const addCareLog = async (
  hospitalizationId: string,
  date: string,
  payload: CreateCareLogRequest,
): Promise<ApiDailyRecord> => {
  const { data } = await axios.post<ApiDailyRecord>(
    `/v1/hospitalizations/${hospitalizationId}/daily-records/${date}/care-logs`,
    payload,
  );
  return data;
};

const addStaffNote = async (
  hospitalizationId: string,
  date: string,
  payload: CreateStaffNoteRequest,
): Promise<ApiDailyRecord> => {
  const { data } = await axios.post<ApiDailyRecord>(
    `/v1/hospitalizations/${hospitalizationId}/daily-records/${date}/staff-notes`,
    payload,
  );
  return data;
};

// ---- Hooks ----

export function useGetDailyRecord(hospitalizationId: string, date: string) {
  return useQuery({
    queryKey: queryKeys.hospitalizations.dailyRecords.byDate(hospitalizationId, date),
    queryFn: () => getDailyRecord(hospitalizationId, date),
    enabled: !!hospitalizationId && !!date,
  });
}

export function useCreateDailyRecord(hospitalizationId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (date: string) => createDailyRecord(hospitalizationId, date),
    onSuccess: (data) => {
      const dateStr = data.date.split("T")[0];
      queryClient.setQueryData(
        queryKeys.hospitalizations.dailyRecords.byDate(hospitalizationId, dateStr),
        data,
      );
      queryClient.invalidateQueries({
        queryKey: queryKeys.hospitalizations.dailyRecords.all(hospitalizationId),
      });
    },
    onError: (error) => handleApiError(error, "日次記録の追加"),
  });
}

export function useCreateDailyVital(hospitalizationId: string, date: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateVitalRecordRequest) =>
      addVitalRecord(hospitalizationId, date, payload),
    onSuccess: (data) => {
      queryClient.setQueryData(
        queryKeys.hospitalizations.dailyRecords.byDate(hospitalizationId, date),
        data,
      );
      queryClient.invalidateQueries({
        queryKey: queryKeys.hospitalizations.dailyRecords.all(hospitalizationId),
      });
    },
    onError: (error) => handleApiError(error, "バイタル追加"),
  });
}

export function useCreateCareLog(hospitalizationId: string, date: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateCareLogRequest) => addCareLog(hospitalizationId, date, payload),
    onSuccess: (data) => {
      queryClient.setQueryData(
        queryKeys.hospitalizations.dailyRecords.byDate(hospitalizationId, date),
        data,
      );
      queryClient.invalidateQueries({
        queryKey: queryKeys.hospitalizations.dailyRecords.all(hospitalizationId),
      });
    },
    onError: (error) => handleApiError(error, "ケアログ追加"),
  });
}

export function useCreateStaffNote(hospitalizationId: string, date: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateStaffNoteRequest) => addStaffNote(hospitalizationId, date, payload),
    onSuccess: (data) => {
      queryClient.setQueryData(
        queryKeys.hospitalizations.dailyRecords.byDate(hospitalizationId, date),
        data,
      );
      queryClient.invalidateQueries({
        queryKey: queryKeys.hospitalizations.dailyRecords.all(hospitalizationId),
      });
    },
    onError: (error) => handleApiError(error, "スタッフノート追加"),
  });
}

export type {
  ApiDailyRecord,
  CreateVitalRecordRequest,
  CreateCareLogRequest,
  CreateStaffNoteRequest,
} from "./daily-records-types";
