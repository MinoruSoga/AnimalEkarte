import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { axios } from "@/lib/axios";
import type { AccountingRecord } from "@/types";
import { transformAccounting } from "./transforms";
import type { BackendAccounting, UpdateAccountingRequest } from "./types";

export const updateAccounting = async (
  id: string,
  req: UpdateAccountingRequest
): Promise<AccountingRecord> => {
  const { data } = await axios.patch<BackendAccounting>(
    `/v1/accountings/${id}`,
    req
  );
  return transformAccounting(data);
};

export const useUpdateAccounting = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdateAccountingRequest;
    }) => updateAccounting(id, req),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["accountings"] });
      queryClient.invalidateQueries({ queryKey: ["accounting", id] });
      queryClient.invalidateQueries({ queryKey: ["accounting-detail", id] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "操作に失敗しました");
    },
  });
};
