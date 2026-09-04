export type InventoryFormGate =
  | { kind: "edit-loading" }
  | { kind: "edit-not-found" }
  | { kind: "edit-error"; retryRead?: () => void };

export function resolveInventoryFormGate(input: {
  isEdit: boolean;
  isLoading: boolean;
  isReadNotFound: boolean;
  isReadError: boolean;
  retryRead?: () => void;
}): InventoryFormGate | null {
  if (!input.isEdit) return null;
  if (input.isLoading) return { kind: "edit-loading" };
  if (input.isReadNotFound) return { kind: "edit-not-found" };
  if (input.isReadError) return { kind: "edit-error", retryRead: input.retryRead };
  return null;
}

export const INVENTORY_FORM_ID = "inventory-form";
