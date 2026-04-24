import { useState, useActionState } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import {
  useGetInventoryItem,
  useCreateInventoryItem,
  useUpdateInventoryItem,
} from "../api/inventory";
import type { InventoryItem } from "@/types";
import type {
  CreateInventoryItemRequest,
  UpdateInventoryItemRequest,
} from "../api/types";

interface FormState {
  success: boolean;
  timestamp: number;
  fieldErrors?: Record<string, string>;
}

export function useInventoryForm(id?: string) {
  const isEdit = Boolean(id);

  const { data: existingItem, isLoading } = useGetInventoryItem(id ?? "");
  const createMutation = useCreateInventoryItem();
  const updateMutation = useUpdateInventoryItem();

  const [prevExistingItem, setPrevExistingItem] = useState(existingItem);
  const [category, setCategory] = useState<InventoryItem["category"]>(
    (existingItem?.category as InventoryItem["category"]) ?? "medicine"
  );
  const [expiryDate, setExpiryDate] = useState("");
  const [lastRestocked, setLastRestocked] = useState("");

  // previous value パターン: レンダー中に同期（useEffect を排除）
  if (prevExistingItem !== existingItem) {
    setPrevExistingItem(existingItem);
    if (existingItem?.category) {
      setCategory(existingItem.category as InventoryItem["category"]);
    }
  }

  const resolvedExpiry =
    expiryDate ||
    (existingItem?.expiryDate ? existingItem.expiryDate.slice(0, 10) : "");
  const resolvedLastRestocked =
    lastRestocked ||
    (existingItem?.lastRestocked
      ? existingItem.lastRestocked.slice(0, 10)
      : "");

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: FormState, formData: FormData): Promise<FormState> => {
      const quantityStr = formData.get("quantity") as string;
      const minStockLevelStr = formData.get("minStockLevel") as string;
      const expiryDateStr = formData.get("expiryDate") as string;
      const lastRestockedStr = formData.get("lastRestocked") as string;
      const resolvedCategory = category || "medicine";

      // Validate: minStockLevel must be <= quantity
      const quantity = quantityStr ? Number(quantityStr) : 0;
      const minStockLevel = minStockLevelStr ? Number(minStockLevelStr) : 0;

      if (minStockLevel > quantity) {
        return {
          success: false,
          timestamp: Date.now(),
          fieldErrors: { minStockLevel: "最低在庫数は現在庫数以下で設定してください" },
        };
      }

      try {
        if (isEdit && id) {
          const req: UpdateInventoryItemRequest = {
            name: formData.get("name") as string,
            category: resolvedCategory,
            quantity: quantityStr ? Number(quantityStr) : undefined,
            unit: formData.get("unit") as string,
            min_stock_level: minStockLevelStr
              ? Number(minStockLevelStr)
              : undefined,
            location: (formData.get("location") as string) || undefined,
            expiry_date: expiryDateStr || undefined,
            supplier: (formData.get("supplier") as string) || undefined,
            last_restocked: lastRestockedStr || undefined,
          };
          await updateMutation.mutateAsync({ id, req });
          toast.success("在庫情報を更新しました");
        } else {
          const req: CreateInventoryItemRequest = {
            name: formData.get("name") as string,
            category: resolvedCategory,
            quantity: quantityStr ? Number(quantityStr) : 0,
            unit: formData.get("unit") as string,
            min_stock_level: minStockLevelStr
              ? Number(minStockLevelStr)
              : 0,
            location: (formData.get("location") as string) || undefined,
            expiry_date: expiryDateStr || undefined,
            supplier: (formData.get("supplier") as string) || undefined,
            last_restocked: lastRestockedStr || undefined,
          };
          await createMutation.mutateAsync(req);
          toast.success("在庫情報を登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0, fieldErrors: {} }
  );

  return {
    isEdit,
    isLoading,
    existingItem,
    category,
    setCategory,
    resolvedExpiry,
    setExpiryDate,
    resolvedLastRestocked,
    setLastRestocked,
    formAction,
    formState,
    isPending,
  };
}
