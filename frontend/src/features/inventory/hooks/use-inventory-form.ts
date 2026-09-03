import { useState, useActionState } from "react";
import { toast } from "sonner";
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
import {
  buildInventoryItemRequest,
  readInventoryFormFields,
  validateInventoryFormFields,
} from "./use-inventory-form-model";

interface FormState {
  success: boolean;
  timestamp: number;
  fieldErrors?: Record<string, string>;
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
    existingItem?.category ?? "medicine"
  );
  const [expiryDate, setExpiryDate] = useState("");
  const [lastRestocked, setLastRestocked] = useState("");

  if (prevExistingItem !== existingItem) {
    setPrevExistingItem(existingItem);
    if (existingItem?.category) {
      setCategory(existingItem.category);
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
      if (
        isEdit &&
        (isReadNotFound || isReadError || isReadLoading || entityRead.status !== "found")
      ) {
        return { success: false, timestamp: Date.now() };
      }

      const fields = readInventoryFormFields(formData);
      const resolvedCategory = category || "medicine";
      const fieldErrors = validateInventoryFormFields(fields);
      if (Object.keys(fieldErrors).length > 0) {
        return { success: false, timestamp: Date.now(), fieldErrors };
      }

      try {
        if (isEdit && id) {
          const req: UpdateInventoryItemRequest = buildInventoryItemRequest(fields, resolvedCategory);
          await updateMutation.mutateAsync({ id, req });
          toast.success("在庫情報を更新しました");
        } else {
          const req: CreateInventoryItemRequest = buildInventoryItemRequest(fields, resolvedCategory);
          await createMutation.mutateAsync(req);
          toast.success("在庫情報を登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch {
        // FE-RC-005: useCreateInventoryItem / useUpdateInventoryItem の onError が
        // 既に handleApiError でトースト表示済みのため、ここでは重複させない。
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
