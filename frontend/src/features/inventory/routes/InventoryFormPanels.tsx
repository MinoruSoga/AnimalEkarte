import { ErrorFallback } from "@/components/shared/DataStates";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { Button } from "@/components/ui/button";
import { C, LAYOUT } from "@/lib/design-tokens";
import { ResourceInventory } from "@/types/generated/models";
import type { InventoryItem } from "@/types";
import {
  BasicInfoSection,
  StockInfoSection,
  SupplierInfoSection,
} from "../components/InventoryFormSections";
import { INVENTORY_FORM_ID, type InventoryFormGate } from "./inventory-form-model";

export function InventoryFormStatusView({
  gate,
  onBack,
}: {
  gate: InventoryFormGate;
  onBack: () => void;
}) {
  if (gate.kind === "edit-loading") {
    return (
      <PageLayout
        title="在庫編集"
        resource={ResourceInventory}
        maxWidth={LAYOUT.pageContentMaxWidth.formNarrow}
      >
        <div className={`text-sm ${C.text60}`}>読み込み中...</div>
      </PageLayout>
    );
  }

  if (gate.kind === "edit-not-found") {
    return (
      <PageLayout
        title="在庫"
        onBack={onBack}
        resource={ResourceInventory}
        maxWidth={LAYOUT.pageContentMaxWidth.formNarrow}
      >
        <ErrorFallback message="在庫情報が見つかりません" />
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title="在庫"
      onBack={onBack}
      resource={ResourceInventory}
      maxWidth={LAYOUT.pageContentMaxWidth.formNarrow}
    >
      <div className="space-y-3">
        <ErrorFallback message="在庫情報の取得に失敗しました" />
        {gate.retryRead ? (
          <Button type="button" variant="outline" size="sm" onClick={gate.retryRead}>
            再試行
          </Button>
        ) : null}
      </div>
    </PageLayout>
  );
}

interface InventoryFormBodyProps {
  isEdit: boolean;
  canSubmit: boolean;
  isDirty: boolean;
  isPending: boolean;
  existingItem: InventoryItem | undefined;
  category: InventoryItem["category"];
  resolvedExpiry: string;
  resolvedLastRestocked: string;
  fieldErrors: Record<string, string> | undefined;
  formAction: (payload: FormData) => void;
  onBack: () => void;
  onMarkDirty: () => void;
  onCategoryChange: (value: InventoryItem["category"]) => void;
  onExpiryChange: (value: string) => void;
  onLastRestockedChange: (value: string) => void;
}

export function InventoryFormBody({
  isEdit,
  canSubmit,
  isDirty,
  isPending,
  existingItem,
  category,
  resolvedExpiry,
  resolvedLastRestocked,
  fieldErrors,
  formAction,
  onBack,
  onMarkDirty,
  onCategoryChange,
  onExpiryChange,
  onLastRestockedChange,
}: InventoryFormBodyProps) {
  return (
    <PageLayout
      title={isEdit ? "在庫編集" : "在庫登録"}
      resource={ResourceInventory}
      onBack={onBack}
      headerAction={
        <FormHeaderActions
          onCancel={onBack}
          submitLabel={canSubmit ? "保存" : undefined}
          submitFormId={INVENTORY_FORM_ID}
        />
      }
      maxWidth={LAYOUT.pageContentMaxWidth.formNarrow}
    >
      <NavigationBlocker when={isDirty ? !isPending : false} />
      {!canSubmit ? (
        <div
          role="status"
          aria-label="閲覧専用モード"
          className={`rounded-md border px-4 py-2.5 text-sm font-medium ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
        >
          閲覧専用 — 編集権限がないため変更できません
        </div>
      ) : null}
      <form
        id={INVENTORY_FORM_ID}
        action={formAction}
        noValidate
        onChange={onMarkDirty}
        className="space-y-6"
      >
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
          <BasicInfoSection
            defaultName={existingItem?.name}
            defaultUnit={existingItem?.unit}
            category={category}
            existingCategory={existingItem?.category as InventoryItem["category"]}
            onCategoryChange={onCategoryChange}
            onMarkDirty={onMarkDirty}
            nameError={fieldErrors?.name}
            unitError={fieldErrors?.unit}
          />

          <StockInfoSection
            defaultQuantity={existingItem?.quantity}
            defaultMinStockLevel={existingItem?.minStockLevel}
            defaultLocation={existingItem?.location}
            resolvedExpiry={resolvedExpiry}
            onExpiryChange={onExpiryChange}
            onMarkDirty={onMarkDirty}
            quantityError={fieldErrors?.quantity}
            minStockLevelError={fieldErrors?.minStockLevel}
          />

          <SupplierInfoSection
            defaultSupplier={existingItem?.supplier}
            resolvedLastRestocked={resolvedLastRestocked}
            onLastRestockedChange={onLastRestockedChange}
            onMarkDirty={onMarkDirty}
          />
        </fieldset>
      </form>
    </PageLayout>
  );
}
