// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";

// Types
import type {
    ApiDailyRecord,
    CreateVitalRecordRequest,
    CreateCareLogRecordRequest,
    CreateStaffNoteRecordRequest,
} from "./daily-records-types";

// ---- Fetchers ----

const listDailyRecords = async (hospitalizationId: string): Promise<ApiDailyRecord[]> => {
    const { data } = await axios.get<ApiDailyRecord[]>(
        `/v1/hospitalizations/${hospitalizationId}/daily-records`
    );
    return data;
};

const getDailyRecord = async (
    hospitalizationId: string,
    date: string
): Promise<ApiDailyRecord> => {
    const { data } = await axios.get<ApiDailyRecord>(
        `/v1/hospitalizations/${hospitalizationId}/daily-records/${date}`
    );
    return data;
};

const createDailyRecord = async (
    hospitalizationId: string,
    date: string
): Promise<ApiDailyRecord> => {
    const { data } = await axios.post<ApiDailyRecord>(
        `/v1/hospitalizations/${hospitalizationId}/daily-records`,
        { date }
    );
    return data;
};

const addVitalRecord = async (
    hospitalizationId: string,
    date: string,
    payload: CreateVitalRecordRequest
): Promise<ApiDailyRecord> => {
    const { data } = await axios.post<ApiDailyRecord>(
        `/v1/hospitalizations/${hospitalizationId}/daily-records/${date}/vitals`,
        payload
    );
    return data;
};

const addCareLogRecord = async (
    hospitalizationId: string,
    date: string,
    payload: CreateCareLogRecordRequest
): Promise<ApiDailyRecord> => {
    const { data } = await axios.post<ApiDailyRecord>(
        `/v1/hospitalizations/${hospitalizationId}/daily-records/${date}/care-logs`,
        payload
    );
    return data;
};

const addStaffNoteRecord = async (
    hospitalizationId: string,
    date: string,
    payload: CreateStaffNoteRecordRequest
): Promise<ApiDailyRecord> => {
    const { data } = await axios.post<ApiDailyRecord>(
        `/v1/hospitalizations/${hospitalizationId}/daily-records/${date}/staff-notes`,
        payload
    );
    return data;
};

// ---- Query Keys ----

export const dailyRecordKeys = {
    all: (hospitalizationId: string) =>
        ["hospitalizations", hospitalizationId, "daily-records"] as const,
    byDate: (hospitalizationId: string, date: string) =>
        ["hospitalizations", hospitalizationId, "daily-records", date] as const,
};

// ---- Hooks ----

export function useDailyRecords(hospitalizationId: string) {
    return useQuery({
        queryKey: dailyRecordKeys.all(hospitalizationId),
        queryFn: () => listDailyRecords(hospitalizationId),
        enabled: !!hospitalizationId,
    });
}

export function useDailyRecord(hospitalizationId: string, date: string) {
    return useQuery({
        queryKey: dailyRecordKeys.byDate(hospitalizationId, date),
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
                dailyRecordKeys.byDate(hospitalizationId, dateStr),
                data
            );
            queryClient.invalidateQueries({
                queryKey: dailyRecordKeys.all(hospitalizationId),
            });
        },
    });
}

export function useCreateDailyVital(hospitalizationId: string, date: string) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (payload: CreateVitalRecordRequest) =>
            addVitalRecord(hospitalizationId, date, payload),
        onSuccess: (data) => {
            queryClient.setQueryData(
                dailyRecordKeys.byDate(hospitalizationId, date),
                data
            );
            queryClient.invalidateQueries({
                queryKey: dailyRecordKeys.all(hospitalizationId),
            });
        },
    });
}

export function useCreateCareLog(hospitalizationId: string, date: string) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (payload: CreateCareLogRecordRequest) =>
            addCareLogRecord(hospitalizationId, date, payload),
        onSuccess: (data) => {
            queryClient.setQueryData(
                dailyRecordKeys.byDate(hospitalizationId, date),
                data
            );
            queryClient.invalidateQueries({
                queryKey: dailyRecordKeys.all(hospitalizationId),
            });
        },
    });
}

export function useCreateStaffNote(hospitalizationId: string, date: string) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (payload: CreateStaffNoteRecordRequest) =>
            addStaffNoteRecord(hospitalizationId, date, payload),
        onSuccess: (data) => {
            queryClient.setQueryData(
                dailyRecordKeys.byDate(hospitalizationId, date),
                data
            );
            queryClient.invalidateQueries({
                queryKey: dailyRecordKeys.all(hospitalizationId),
            });
        },
    });
}

export type {
    ApiDailyRecord,
    ApiVitalRecord,
    ApiCareLogRecord,
    ApiStaffNoteRecord,
    CreateVitalRecordRequest,
    CreateCareLogRecordRequest,
    CreateStaffNoteRecordRequest,
} from "./daily-records-types";
