import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { Shift, UpdateShiftInput } from "../types";
import { transformShift } from "./transforms";
import type { BackendShift } from "./types";

export async function updateShift(id: string, input: UpdateShiftInput): Promise<Shift> {
  const { data } = await axios.patch<BackendShift>(`/v1/shifts/${id}`, input);
  return transformShift(data);
}

export function useUpdateShift() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateShiftInput }) =>
      updateShift(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["shifts"] });
    },
    onError: (error) => handleApiError(error, "シフトの更新"),
  });
}
