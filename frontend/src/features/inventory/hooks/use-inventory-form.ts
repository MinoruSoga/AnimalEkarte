import { useState, useActionState, useCallback, useLayoutEffect, useRef } from "react";
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
import type { CreateInventoryItemRequest, UpdateInventoryItemRequest } from "../api/types";
import {
  buildInventoryItemRequest,
  readInventoryFormFields,
  validateInventoryFormFields,
} from "./use-inventory-form-model";

interface FormState {
  success: boolean;
  timestamp: number;
  fieldErrors?: Record<string, string>;
  error?: string;
}

/** FE-RC-220: action 別の最新権限値。mutation 直前に isMutationAllowed() で再検査する。 */
export interface InventoryMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
}

const DENIED: Readonly<InventoryMutationPermissions> = {
  canCreate: false,
  canEdit: false,
};

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

export interface UseInventoryFormArgs {
  permissions?: Readonly<InventoryMutationPermissions>;
}

export function useInventoryForm(id?: string, args: UseInventoryFormArgs = {}) {
  const isEdit = Boolean(id);
  const permissions = args.permissions ?? DENIED;
  const { canCreate, canEdit } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit };
  }, [canCreate, canEdit]);
  const isMutationAllowed = useCallback(
    (action: keyof InventoryMutationPermissions) => permissionsRef.current[action] === true,
    [],
  );

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
    existingItem?.category ?? "medicine",
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
    expiryDate || (existingItem?.expiryDate ? existingItem.expiryDate.slice(0, 10) : "");
  const resolvedLastRestocked =
    lastRestocked || (existingItem?.lastRestocked ? existingItem.lastRestocked.slice(0, 10) : "");

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
          if (!isMutationAllowed("canEdit")) {
            toast.error(PERMISSION_DENIED_MESSAGE);
            return { success: false, error: PERMISSION_DENIED_MESSAGE, timestamp: Date.now() };
          }
          const req: UpdateInventoryItemRequest = buildInventoryItemRequest(
            fields,
            resolvedCategory,
          );
          await updateMutation.mutateAsync({ id, req });
          toast.success("在庫情報を更新しました");
        } else {
          if (!isMutationAllowed("canCreate")) {
            toast.error(PERMISSION_DENIED_MESSAGE);
            return { success: false, error: PERMISSION_DENIED_MESSAGE, timestamp: Date.now() };
          }
          const req: CreateInventoryItemRequest = buildInventoryItemRequest(
            fields,
            resolvedCategory,
          );
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
    { success: false, timestamp: 0, fieldErrors: {} },
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
