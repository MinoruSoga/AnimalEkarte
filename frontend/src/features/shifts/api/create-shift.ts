import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { Shift, CreateShiftInput } from "../types";
import { transformShift } from "./transforms";
import type { BackendShift } from "./types";

export async function createShift(input: CreateShiftInput): Promise<Shift> {
  const { data } = await axios.post<BackendShift>("/v1/shifts", {
    ...input,
    staff_id: Number(input.staff_id),
  });
  return transformShift(data);
}

export function useCreateShift() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createShift,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["shifts"] });
    },
    onError: (error) => handleApiError(error, "シフトの作成"),
  });
}
