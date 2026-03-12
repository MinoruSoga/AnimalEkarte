import { useMemo } from "react";
import { toast } from "sonner";
import type { MasterItem } from "@/types";
import { useGetMasterItemsByCategory } from "../api/get-master-item";
import { useCreateMasterItem } from "../api/create-master-item";
import { useUpdateMasterItem } from "../api/update-master-item";
import { useDeleteMasterItem } from "../api/delete-master-item";
import type { CreateMasterItemRequest, UpdateMasterItemRequest } from "../api/types";

interface MutationCallbacks {
  onSuccess?: () => void;
  onError?: () => void;
}

export function useMasterItems(category?: string, searchTerm?: string) {
  const resolvedCategory = category && category !== "all" ? category : "";

  const { data: categoryItems = [], isLoading } = useGetMasterItemsByCategory(resolvedCategory);

  const createMutation = useCreateMasterItem({ category: resolvedCategory });
  const updateMutation = useUpdateMasterItem({ category: resolvedCategory });
  const deleteMutation = useDeleteMasterItem({ category: resolvedCategory });

  const filteredItems = useMemo(() => {
    if (!searchTerm) return categoryItems;
    const lower = searchTerm.toLowerCase();
    return categoryItems.filter(
      (i) =>
        i.name.toLowerCase().includes(lower) ||
        i.code.toLowerCase().includes(lower) ||
        (i.description && i.description.toLowerCase().includes(lower))
    );
  }, [categoryItems, searchTerm]);

  const add = (item: Omit<MasterItem, "id">, callbacks?: MutationCallbacks) => {
    const req: CreateMasterItemRequest = {
      code: item.code,
      name: item.name,
      category: resolvedCategory,
      price: item.price ?? 0,
      description: item.description,
    };
    createMutation.mutate(req, {
      onSuccess: callbacks?.onSuccess,
      onError: () => {
        toast.error("登録に失敗しました");
        callbacks?.onError?.();
      },
    });
  };

  const update = (
    id: string,
    updates: Partial<MasterItem>,
    callbacks?: MutationCallbacks
  ) => {
    const req: UpdateMasterItemRequest = {
      code: updates.code,
      name: updates.name,
      price: updates.price,
      status: updates.status,
      description: updates.description,
    };
    updateMutation.mutate(
      { id, req },
      {
        onSuccess: callbacks?.onSuccess,
        onError: () => {
          toast.error("更新に失敗しました");
          callbacks?.onError?.();
        },
      }
    );
  };

  const remove = (id: string, callbacks?: MutationCallbacks) => {
    deleteMutation.mutate(id, {
      onSuccess: callbacks?.onSuccess,
      onError: () => {
        toast.error("削除に失敗しました");
        callbacks?.onError?.();
      },
    });
  };

  return { data: filteredItems, add, update, remove, isLoading };
}
