// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";

// ---- Types ----

// Backend API types (snake_case, matches care_plan_item_response.go / model/hospitalization.go)
export type CarePlanItemType = 'food' | 'medicine' | 'treatment' | 'instruction' | 'item';
export type CarePlanItemStatus = 'active' | 'completed' | 'discontinued';
export type CarePlanTiming = 'morning' | 'noon' | 'night';

export interface CarePlanItem {
    id: string;
    hospitalization_id: string;
    type: CarePlanItemType;
    name: string;
    description: string;
    timing: CarePlanTiming[];
    status: CarePlanItemStatus;
    notes: string;
    medicine_id?: string | null;
    procedure_id?: string | null;
    hospitalization_plan_id?: string | null;
    unit_price: number;
    category: string;
    sort_order: number;
    created_at: string;
    updated_at: string;
}

export interface CreateCarePlanItemInput {
    type: CarePlanItemType;
    name: string;
    description?: string;
    timing?: CarePlanTiming[];
    status?: CarePlanItemStatus;
    notes?: string;
    medicine_id?: string | null;
    procedure_id?: string | null;
    hospitalization_plan_id?: string | null;
    unit_price?: number;
    category?: string;
    sort_order?: number;
}

export interface UpdateCarePlanItemInput {
    type?: CarePlanItemType;
    name?: string;
    description?: string;
    timing?: CarePlanTiming[];
    status?: CarePlanItemStatus;
    notes?: string;
    medicine_id?: string | null;
    procedure_id?: string | null;
    hospitalization_plan_id?: string | null;
    unit_price?: number;
    category?: string;
    sort_order?: number;
}

// ---- Query Keys ----

export const carePlanItemKeys = {
    all: (hospitalizationId: string) =>
        ["hospitalizations", hospitalizationId, "care-plan-items"] as const,
};

// ---- Fetchers ----

const listCarePlanItems = async (hospitalizationId: string): Promise<CarePlanItem[]> => {
    const { data } = await axios.get<CarePlanItem[]>(
        `/v1/hospitalizations/${hospitalizationId}/care-plan-items`
    );
    return data;
};

const createCarePlanItem = async (
    hospitalizationId: string,
    input: CreateCarePlanItemInput
): Promise<CarePlanItem> => {
    const { data } = await axios.post<CarePlanItem>(
        `/v1/hospitalizations/${hospitalizationId}/care-plan-items`,
        input
    );
    return data;
};

const updateCarePlanItem = async (
    hospitalizationId: string,
    itemId: string,
    input: UpdateCarePlanItemInput
): Promise<CarePlanItem> => {
    const { data } = await axios.patch<CarePlanItem>(
        `/v1/hospitalizations/${hospitalizationId}/care-plan-items/${itemId}`,
        input
    );
    return data;
};

const deleteCarePlanItem = async (
    hospitalizationId: string,
    itemId: string
): Promise<void> => {
    await axios.delete(`/v1/hospitalizations/${hospitalizationId}/care-plan-items/${itemId}`);
};

// ---- Hooks ----

export function useCarePlanItems(hospitalizationId: string) {
    return useQuery({
        queryKey: carePlanItemKeys.all(hospitalizationId),
        queryFn: () => listCarePlanItems(hospitalizationId),
        enabled: !!hospitalizationId,
        staleTime: QUERY_STALE_TIMES.REALTIME,
        gcTime: QUERY_GC_TIMES.SHORT,
    });
}

export function useCreateCarePlanItem(hospitalizationId: string) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (input: CreateCarePlanItemInput) =>
            createCarePlanItem(hospitalizationId, input),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: carePlanItemKeys.all(hospitalizationId),
            });
        },
        onError: (error: unknown) => {
            handleApiError(error, "作成");
        },
    });
}

export function useUpdateCarePlanItem(hospitalizationId: string) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ itemId, input }: { itemId: string; input: UpdateCarePlanItemInput }) =>
            updateCarePlanItem(hospitalizationId, itemId, input),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: carePlanItemKeys.all(hospitalizationId),
            });
        },
        onError: (error: unknown) => {
            handleApiError(error, "更新");
        },
    });
}

export function useDeleteCarePlanItem(hospitalizationId: string) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (itemId: string) => deleteCarePlanItem(hospitalizationId, itemId),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: carePlanItemKeys.all(hospitalizationId),
            });
        },
        onError: (error: unknown) => {
            handleApiError(error, "削除");
        },
    });
}
