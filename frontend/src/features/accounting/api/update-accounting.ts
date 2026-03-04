import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { AccountingRecord } from "@/types";
import { transformAccounting } from "./transforms";
import type { BackendAccounting, UpdateAccountingRequest } from "./types";

export const updateAccounting = async (
  id: string,
  req: UpdateAccountingRequest
): Promise<AccountingRecord> => {
  const { data } = await axios.put<BackendAccounting>(
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["accountings"] });
    },
  });
};
