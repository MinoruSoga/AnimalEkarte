import { useMemo } from "react";
import { toast } from "sonner";
import type { MasterItem } from "@/types";
import {
  useGetMasterItemsByCategory,
  useGetMasterItems,
  useCreateMasterItem,
  useUpdateMasterItem,
  useDeleteMasterItem,
} from "@/features/master/api";
import type { CreateMasterItemRequest, UpdateMasterItemRequest } from "@/features/master/api";

interface MutationCallbacks {
  onSuccess?: () => void;
  onError?: () => void;
}

const CATEGORY_MAP: Record<string, string> = {
  examination: "検査",
  vaccine: "予防",
  medicine: "薬剤",
  consultation: "診察",
  serviceType: "診療内容",
  procedure: "処置",
  hospitalization: "入院",
  staff: "スタッフ",
  insurance: "保険",
  cage: "ケージ",
  trimmingCourse: "トリミングコース",
  trimmingOption: "トリミングオプション",
  diagnosisCategory: "診断カテゴリ",
  diagnosisName: "診断名",
};

export function useMasterItems(category?: string, searchTerm?: string) {
  const backendCategory =
    category && category !== "all"
      ? CATEGORY_MAP[category] || category
      : undefined;

  const { data: allItems = [], isLoading: isLoadingAll } = useGetMasterItems();

  const { data: categoryItems = [], isLoading: isLoadingCategory } =
    useGetMasterItemsByCategory(backendCategory ?? "");

  const createMutation = useCreateMasterItem();
  const updateMutation = useUpdateMasterItem();
  const deleteMutation = useDeleteMasterItem();

  const sourceItems = backendCategory ? categoryItems : allItems;
  const isLoading = backendCategory ? isLoadingCategory : isLoadingAll;

  const filteredItems = useMemo(() => {
    if (!searchTerm) return sourceItems;
    const lower = searchTerm.toLowerCase();
    return sourceItems.filter(
      (i) =>
        i.name.toLowerCase().includes(lower) ||
        i.code.toLowerCase().includes(lower) ||
        (i.category && i.category.toLowerCase().includes(lower))
    );
  }, [sourceItems, searchTerm]);

  const add = (item: Omit<MasterItem, "id">, callbacks?: MutationCallbacks) => {
    const req: CreateMasterItemRequest = {
      code: item.code,
      name: item.name,
      category: item.category ?? "",
      price: item.price ?? 0,
      description: item.description,
      inventory_id: item.inventoryId,
      default_quantity: item.defaultQuantity,
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
      category: updates.category,
      price: updates.price,
      status: updates.status,
      description: updates.description,
      inventory_id: updates.inventoryId,
      default_quantity: updates.defaultQuantity,
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
