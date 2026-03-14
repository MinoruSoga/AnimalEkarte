// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// External
import { toast } from "sonner";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

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
        onError: () => {
            toast.error("ケアプラン項目の作成に失敗しました");
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
        onError: () => {
            toast.error("ケアプラン項目の更新に失敗しました");
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
        onError: () => {
            toast.error("ケアプラン項目の削除に失敗しました");
        },
    });
}
