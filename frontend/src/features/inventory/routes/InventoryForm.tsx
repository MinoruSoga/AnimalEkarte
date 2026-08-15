// React/Framework
import { useCallback, useEffect } from "react";
import { useNavigate, useParams } from "react-router";

// External
import { Package, Save } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";

// Relative
import { BasicInfoSection, StockInfoSection, SupplierInfoSection } from "../components/InventoryFormSections";
import { useInventoryForm } from "../hooks/use-inventory-form";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { InventoryItem } from "@/types";
import { ResourceInventory } from "@/types/generated/models";

const INVENTORY_FORM_ID = "inventory-form";

// ─── InventoryForm ────────────────────────────────────────────────────────────

export function InventoryForm() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { canEdit, canCreate } = usePermission("inventory");

  const {
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
  } = useInventoryForm(id);

  const canSubmit = isEdit ? canEdit : canCreate;

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // React 19 Action の成功を検知して遷移
  useEffect(() => {
    if (formState.success) {
      markClean();
      navigate(paths.inventory.getHref());
    }
  }, [formState.success, formState.timestamp, navigate, markClean]);

  const handleBack = useCallback(() => {
    navigate(paths.inventory.getHref());
  }, [navigate]);

  // rerender-functional-setstate: setCategory は stable setter なので useCallback 内 deps 不要
  const handleCategoryChange = useCallback((value: string) => {
    setCategory(value as InventoryItem["category"]);
  }, [setCategory]);

  // rerender-functional-setstate: setExpiryDate は stable setter なので useCallback 内 deps 不要
  const handleExpiryChange = useCallback((value: string) => {
    setExpiryDate(value);
  }, [setExpiryDate]);

  // rerender-functional-setstate: setLastRestocked は stable setter なので useCallback 内 deps 不要
  const handleLastRestockedChange = useCallback((value: string) => {
    setLastRestocked(value);
  }, [setLastRestocked]);

  if (isEdit && isLoading) {
    return (
      <PageLayout
        title="在庫編集"
        icon={<Package className={`${ICON.page} ${C.text}`} />}
        resource={ResourceInventory}
        maxWidth="max-w-3xl"
      >
        <div className={`text-sm ${C.text60}`}>読み込み中...</div>
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title={isEdit ? "在庫編集" : "在庫登録"}
      resource={ResourceInventory}
      icon={<Package className={`${ICON.page} ${C.text}`} />}
      onBack={handleBack}
      headerAction={
        canSubmit ? (
          <SubmitButton size="sm" form={INVENTORY_FORM_ID}>
            <Save className={`mr-1.5 ${ICON.action}`} />
            {isEdit ? "更新" : "登録"}
          </SubmitButton>
        ) : null
      }
      maxWidth="max-w-3xl"
    >
      {/* FE6-8: jsx-no-leaked-render は非型認識のため isDirty を boolean と静的に断定できず !! で明示する */}
      <NavigationBlocker when={!!isDirty && !isPending} />
      {!canSubmit ? (
        <div
          role="status"
          aria-label="閲覧専用モード"
          className={`rounded-md border px-4 py-2.5 text-sm font-medium ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
        >
          閲覧専用 — 編集権限がないため変更できません
        </div>
      ) : null}
      <form id={INVENTORY_FORM_ID} action={formAction} onChange={markDirty} className="space-y-6">
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        <BasicInfoSection
          defaultName={existingItem?.name}
          defaultUnit={existingItem?.unit}
          category={category}
          existingCategory={existingItem?.category as InventoryItem["category"]}
          onCategoryChange={handleCategoryChange}
          onMarkDirty={markDirty}
        />

        <StockInfoSection
          defaultQuantity={existingItem?.quantity}
          defaultMinStockLevel={existingItem?.minStockLevel}
          defaultLocation={existingItem?.location}
          resolvedExpiry={resolvedExpiry}
          onExpiryChange={handleExpiryChange}
          onMarkDirty={markDirty}
          quantityError={formState.fieldErrors?.quantity}
          minStockLevelError={formState.fieldErrors?.minStockLevel}
        />

        <SupplierInfoSection
          defaultSupplier={existingItem?.supplier}
          resolvedLastRestocked={resolvedLastRestocked}
          onLastRestockedChange={handleLastRestockedChange}
          onMarkDirty={markDirty}
        />

        </fieldset>
        {/* Actions */}
        <div className="flex justify-end gap-3">
          {canSubmit ? (
            <SubmitButton>
              <Save className={`mr-1.5 ${ICON.action}`} />
              {isEdit ? "更新" : "登録"}
            </SubmitButton>
          ) : null}
        </div>
      </form>
    </PageLayout>
  );
}
