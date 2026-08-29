import { useState, useActionState } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
  type EntityReadResult,
} from "@/lib/entity-read-result";
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

function readFormString(formData: FormData, key: string): string {
  const value = formData.get(key);
  return typeof value === "string" ? value : "";
}

export function useInventoryForm(id?: string) {
  const isEdit = Boolean(id);

  const {
    data: inventoryData,
    isLoading,
    isError,
    error: inventoryError,
    refetch: refetchInventory,
  } = useGetInventoryItem(id ?? "");
  const createMutation = useCreateInventoryItem();
  const updateMutation = useUpdateInventoryItem();

  // BUG-507: classify read failures; never fold missing entity into blank editable defaults
  const entityRead: EntityReadResult<InventoryItem> = resolveEntityReadResult({
    id: isEdit ? id : undefined,
    data: inventoryData,
    isLoading,
    isError,
    error: inventoryError,
    refetch: refetchInventory,
  });
  const existingItem = entityRead.status === "found" ? entityRead.data : undefined;
  const isReadLoading = entityRead.status === "loading";
  const isReadNotFound = isNonDisclosureReadStatus(entityRead.status);
  const isReadError = entityRead.status === "error";
  const retryRead = entityRead.status === "error" ? entityRead.retry : undefined;

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
      // Fail closed: do not mutate while edit load is missing / errored / in-flight
      if (
        isEdit &&
        (isReadNotFound || isReadError || isReadLoading || entityRead.status !== "found")
      ) {
        return { success: false, timestamp: Date.now() };
      }

      const name = readFormString(formData, "name").trim();
      const unit = readFormString(formData, "unit").trim();
      const quantityStr = readFormString(formData, "quantity");
      const minStockLevelStr = readFormString(formData, "minStockLevel");
      const expiryDateStr = readFormString(formData, "expiryDate");
      const lastRestockedStr = readFormString(formData, "lastRestocked");
      const location = readFormString(formData, "location") || undefined;
      const supplier = readFormString(formData, "supplier") || undefined;
      const resolvedCategory = category || "medicine";

      const quantity = Number(quantityStr);
      const minStockLevel = Number(minStockLevelStr);
      const fieldErrors = {
        ...(name === "" ? { name: "品名を入力してください" } : {}),
        ...(unit === "" ? { unit: "単位を入力してください" } : {}),
        ...(quantityStr.trim() === "" || !Number.isInteger(quantity) || quantity < 0
          ? { quantity: "現在庫数は0以上の整数で入力してください" }
          : {}),
        ...(minStockLevelStr.trim() === "" || !Number.isInteger(minStockLevel) || minStockLevel < 0
          ? { minStockLevel: "最低在庫数は0以上の整数で入力してください" }
          : {}),
      };

      if (Object.keys(fieldErrors).length > 0) {
        return {
          success: false,
          timestamp: Date.now(),
          fieldErrors,
        };
      }

      try {
        if (isEdit && id) {
          const req: UpdateInventoryItemRequest = {
            name,
            category: resolvedCategory,
            quantity,
            unit,
            min_stock_level: minStockLevel,
            location,
            expiry_date: expiryDateStr || undefined,
            supplier,
            last_restocked: lastRestockedStr || undefined,
          };
          await updateMutation.mutateAsync({ id, req });
          toast.success("在庫情報を更新しました");
        } else {
          const req: CreateInventoryItemRequest = {
            name,
            category: resolvedCategory,
            quantity,
            unit,
            min_stock_level: minStockLevel,
            location,
            expiry_date: expiryDateStr || undefined,
            supplier,
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
    isLoading: isReadLoading,
    existingItem,
    entityRead,
    isReadNotFound,
    isReadError,
    retryRead,
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
