import { useCallback, useEffect } from "react";
import { useNavigate, useParams } from "react-router";

import { paths } from "@/config/paths";
import { useBackNavigation, useBackPath } from "@/hooks/use-back-navigation";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useInventoryForm } from "../hooks/use-inventory-form";
import { usePermission } from "@/hooks/use-permission";
import type { InventoryItem } from "@/types";
import { resolveInventoryFormGate } from "./inventory-form-model";
import { InventoryFormBody, InventoryFormStatusView } from "./InventoryFormPanels";

export function InventoryForm() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { canEdit, canCreate } = usePermission("inventory");

  const {
    isEdit,
    isLoading,
    isReadNotFound,
    isReadError,
    retryRead,
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
  } = useInventoryForm(id, { permissions: { canCreate, canEdit } });

  const canSubmit = isEdit ? canEdit : canCreate;
  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  const backPath = useBackPath(paths.inventory.getHref());

  useEffect(() => {
    if (formState.success) {
      markClean();
      navigate(backPath);
    }
  }, [formState.success, formState.timestamp, navigate, markClean, backPath]);

  const handleBack = useBackNavigation(paths.inventory.getHref());

  const handleCategoryChange = useCallback(
    (value: InventoryItem["category"]) => {
      setCategory(value);
    },
    [setCategory],
  );

  const handleExpiryChange = useCallback(
    (value: string) => {
      setExpiryDate(value);
    },
    [setExpiryDate],
  );

  const handleLastRestockedChange = useCallback(
    (value: string) => {
      setLastRestocked(value);
    },
    [setLastRestocked],
  );

  const gate = resolveInventoryFormGate({
    isEdit,
    isLoading,
    isReadNotFound,
    isReadError,
    retryRead,
  });
  if (gate) {
    return <InventoryFormStatusView gate={gate} onBack={handleBack} />;
  }

  return (
    <InventoryFormBody
      isEdit={isEdit}
      canSubmit={canSubmit === true}
      isDirty={isDirty}
      isPending={isPending}
      existingItem={existingItem}
      category={category}
      resolvedExpiry={resolvedExpiry}
      resolvedLastRestocked={resolvedLastRestocked}
      fieldErrors={formState.fieldErrors}
      formAction={formAction}
      onBack={handleBack}
      onMarkDirty={markDirty}
      onCategoryChange={handleCategoryChange}
      onExpiryChange={handleExpiryChange}
      onLastRestockedChange={handleLastRestockedChange}
    />
  );
}
