// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";

// ---- Types ----

// Backend API types (snake_case, matches care_plan_item_response.go / model/hospitalization.go)
export type CarePlanItemType = 'food' | 'medicine' | 'treatment' | 'instruction' | 'item';
type CarePlanItemStatus = 'active' | 'completed' | 'discontinued';
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

type CarePlanItemRequestReferenceId = number | null;

interface CarePlanItemRequestReferenceIds {
    medicine_id?: CarePlanItemRequestReferenceId;
    procedure_id?: CarePlanItemRequestReferenceId;
    hospitalization_plan_id?: CarePlanItemRequestReferenceId;
}

type CarePlanItemRequestInput<
    T extends CreateCarePlanItemInput | UpdateCarePlanItemInput
> = Omit<T, keyof CarePlanItemRequestReferenceIds> & CarePlanItemRequestReferenceIds;

function toRequestReferenceId(
    value: string | null | undefined
): CarePlanItemRequestReferenceId | undefined {
    if (value === null || value === undefined) {
        return value;
    }
    if (value === "") {
        return undefined;
    }
    if (!/^[1-9]\d*$/.test(value)) {
        throw new Error("Care plan reference ID must be a positive safe integer");
    }

    const numericValue = Number(value);
    if (!Number.isSafeInteger(numericValue)) {
        throw new Error("Care plan reference ID must be a positive safe integer");
    }
    return numericValue;
}

function toCarePlanItemRequestInput<
    T extends CreateCarePlanItemInput | UpdateCarePlanItemInput
>(input: T): CarePlanItemRequestInput<T> {
    const {
        medicine_id,
        procedure_id,
        hospitalization_plan_id,
        ...requestInput
    } = input;
    const medicineId = toRequestReferenceId(medicine_id);
    const procedureId = toRequestReferenceId(procedure_id);
    const hospitalizationPlanId = toRequestReferenceId(hospitalization_plan_id);

    return {
        ...requestInput,
        ...(medicineId !== undefined ? { medicine_id: medicineId } : {}),
        ...(procedureId !== undefined ? { procedure_id: procedureId } : {}),
        ...(hospitalizationPlanId !== undefined
            ? { hospitalization_plan_id: hospitalizationPlanId }
            : {}),
    };
}

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
        toCarePlanItemRequestInput(input)
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
        toCarePlanItemRequestInput(input)
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

export function useGetCarePlanItems(hospitalizationId: string) {
    return useQuery({
        queryKey: queryKeys.hospitalizations.carePlanItems(hospitalizationId),
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
                queryKey: queryKeys.hospitalizations.carePlanItems(hospitalizationId),
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
                queryKey: queryKeys.hospitalizations.carePlanItems(hospitalizationId),
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
                queryKey: queryKeys.hospitalizations.carePlanItems(hospitalizationId),
            });
        },
        onError: (error: unknown) => {
            handleApiError(error, "削除");
        },
    });
}
