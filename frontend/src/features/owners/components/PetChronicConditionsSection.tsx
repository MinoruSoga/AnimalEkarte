import type { Dispatch, SetStateAction } from "react";

import { HeartPulse, Trash2 } from "lucide-react";

import { DatePicker } from "@/components/shared/DatePicker";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { usePermission, type UsePermissionResult } from "@/hooks/use-permission";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { PetChronicCondition } from "@/types/generated/models";

import {
  usePetChronicConditionsSection,
  type ChronicConditionDraft,
} from "../hooks/use-pet-chronic-conditions-section";

interface PetChronicConditionsSectionProps {
  petId: string;
  canEdit: boolean;
}

/**
 * EMR-248: ペット編集モーダルの慢性疾患セクション。
 * owners:view がなければ何も描画しない（canView ゲートはデータクエリ発行前に行うため
 * 外側コンポーネントで判定し、Content はマウントされない）。
 */
export function PetChronicConditionsSection({ petId, canEdit }: PetChronicConditionsSectionProps) {
  const permissions = usePermission("owners");
  if (!permissions.canView) {
    return null;
  }
  return (
    <PetChronicConditionsSectionContent petId={petId} canEdit={canEdit} permissions={permissions} />
  );
}

interface PetChronicConditionsSectionContentProps extends PetChronicConditionsSectionProps {
  permissions: UsePermissionResult;
}

function PetChronicConditionsSectionContent({
  petId,
  canEdit,
  permissions,
}: PetChronicConditionsSectionContentProps) {
  const {
    conditionsQuery,
    conditions,
    addDraft,
    setAddDraft,
    isAddDirty,
    addState,
    addAction,
    isAddPending,
    editingId,
    editDraft,
    setEditDraft,
    startEdit,
    cancelEdit,
    editState,
    editAction,
    isEditPending,
    handleDeleteCondition,
    isBusy,
    loadError,
    canCreateCondition,
    canEditCondition,
    canDeleteCondition,
  } = usePetChronicConditionsSection(petId, canEdit, permissions);

  const addControlsDisabled =
    !canCreateCondition ||
    isAddPending ||
    conditionsQuery.data === undefined ||
    conditionsQuery.error !== null;
  const rowActionsDisabled =
    isBusy || conditionsQuery.data === undefined || conditionsQuery.error !== null;

  return (
    <section
      aria-labelledby="pet-chronic-conditions-title"
      className="col-span-1 space-y-3 md:col-span-2 lg:col-span-3"
    >
      <h2
        id="pet-chronic-conditions-title"
        className={`flex items-center gap-2 text-sm font-bold ${C.text}`}
      >
        <HeartPulse className={`${ICON.action} ${C.text60}`} aria-hidden="true" />
        慢性疾患
      </h2>

      {loadError !== null ? (
        <p className={`text-sm ${C.danger}`} role="alert">
          慢性疾患情報を取得できませんでした。
        </p>
      ) : null}

      <form action={addAction} className="space-y-2">
        <fieldset disabled={addControlsDisabled} className="m-0 space-y-2 border-0 p-0">
          <div className="grid grid-cols-1 gap-2 md:grid-cols-2">
            <div className="space-y-1">
              <Label htmlFor="pet-chronic-condition-add-name" className={STYLE.sectionLabel}>
                疾患名
              </Label>
              <Input
                id="pet-chronic-condition-add-name"
                value={addDraft.conditionName}
                onChange={(event) =>
                  setAddDraft((prev) => ({ ...prev, conditionName: event.target.value }))
                }
                placeholder="例: 慢性腎臓病"
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="pet-chronic-condition-add-code" className={STYLE.sectionLabel}>
                疾患コード
              </Label>
              <Input
                id="pet-chronic-condition-add-code"
                value={addDraft.conditionCode}
                onChange={(event) =>
                  setAddDraft((prev) => ({ ...prev, conditionCode: event.target.value }))
                }
                placeholder="例: CKD"
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="pet-chronic-condition-add-diagnosed" className={STYLE.sectionLabel}>
                診断日
              </Label>
              <DatePicker
                id="pet-chronic-condition-add-diagnosed"
                value={addDraft.diagnosedAt}
                onChange={(value) => setAddDraft((prev) => ({ ...prev, diagnosedAt: value }))}
                placeholder="診断日を選択…"
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="pet-chronic-condition-add-notes" className={STYLE.sectionLabel}>
                備考
              </Label>
              <Input
                id="pet-chronic-condition-add-notes"
                value={addDraft.notes}
                onChange={(event) =>
                  setAddDraft((prev) => ({ ...prev, notes: event.target.value }))
                }
                placeholder="任意"
              />
            </div>
          </div>
          <div className="flex justify-end">
            <SubmitButton loadingText="慢性疾患を追加中..." className="text-sm">
              慢性疾患を追加
            </SubmitButton>
          </div>
        </fieldset>
      </form>

      {addState.kind === "error" ? (
        <p className={`text-sm ${C.danger}`} role="alert" aria-live="assertive">
          {addState.message}
        </p>
      ) : null}
      {addState.kind === "success" && !isAddDirty ? (
        <p className={`text-sm ${C.textSuccess}`} role="status" aria-live="polite">
          {addState.message}
        </p>
      ) : null}

      <div className={`overflow-hidden rounded-lg border ${C.borderMedium} ${C.bgWhite}`}>
        {conditionsQuery.isLoading ? (
          <p className={`px-4 py-6 text-center text-sm ${C.text60}`}>読み込み中...</p>
        ) : conditions.length === 0 ? (
          <p className={`px-4 py-6 text-center text-sm ${C.text60}`}>
            慢性疾患は登録されていません。
          </p>
        ) : (
          <ul className="divide-y">
            {conditions.map((condition) => (
              <li key={condition.id} className={`p-3 ${C.borderDivider}`}>
                {editingId === condition.id ? (
                  <ChronicConditionEditForm
                    condition={condition}
                    editDraft={editDraft}
                    setEditDraft={setEditDraft}
                    editAction={editAction}
                    cancelEdit={cancelEdit}
                    controlsDisabled={!canEditCondition || isEditPending}
                    editErrorMessage={editState.kind === "error" ? editState.message : null}
                  />
                ) : (
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 space-y-0.5">
                      <p className={`text-sm font-medium ${C.text}`}>
                        {condition.condition_name}
                        <span
                          className={`ml-2 text-xs ${condition.is_active ? C.textSuccess : C.text60}`}
                        >
                          {condition.is_active ? "有効" : "無効"}
                        </span>
                      </p>
                      <p className={`text-xs ${C.text60}`}>
                        {condition.condition_code} / 診断日 {condition.diagnosed_at}
                        {condition.notes ? ` / ${condition.notes}` : ""}
                      </p>
                    </div>
                    <div className="flex shrink-0 gap-1">
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        disabled={!canEditCondition || rowActionsDisabled}
                        aria-label={`慢性疾患 ${condition.condition_name}を編集`}
                        onClick={() => startEdit(condition)}
                      >
                        編集
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        disabled={!canDeleteCondition || rowActionsDisabled}
                        aria-label={`慢性疾患 ${condition.condition_name}を削除`}
                        onClick={() => handleDeleteCondition(condition.id)}
                        className={`${C.danger} ${C.borderDanger}`}
                      >
                        <Trash2 className={ICON.action} aria-hidden="true" />
                        削除
                      </Button>
                    </div>
                  </div>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {editState.kind === "success" && editingId === null ? (
        <p className={`text-sm ${C.textSuccess}`} role="status" aria-live="polite">
          {editState.message}
        </p>
      ) : null}
    </section>
  );
}

interface ChronicConditionEditFormProps {
  condition: PetChronicCondition;
  editDraft: ChronicConditionDraft;
  setEditDraft: Dispatch<SetStateAction<ChronicConditionDraft>>;
  editAction: (formData: FormData) => void;
  cancelEdit: () => void;
  controlsDisabled: boolean;
  editErrorMessage: string | null;
}

function ChronicConditionEditForm({
  condition,
  editDraft,
  setEditDraft,
  editAction,
  cancelEdit,
  controlsDisabled,
  editErrorMessage,
}: ChronicConditionEditFormProps) {
  const name = condition.condition_name;
  return (
    <form action={editAction} className="space-y-2">
      <fieldset disabled={controlsDisabled} className="m-0 space-y-2 border-0 p-0">
        <div className="grid grid-cols-1 gap-2 md:grid-cols-2">
          <div className="space-y-1">
            <Label
              htmlFor={`pet-chronic-condition-edit-name-${condition.id}`}
              className={STYLE.sectionLabel}
            >
              {`疾患名（${name}）`}
            </Label>
            <Input
              id={`pet-chronic-condition-edit-name-${condition.id}`}
              value={editDraft.conditionName}
              onChange={(event) =>
                setEditDraft((prev) => ({ ...prev, conditionName: event.target.value }))
              }
            />
          </div>
          <div className="space-y-1">
            <Label
              htmlFor={`pet-chronic-condition-edit-code-${condition.id}`}
              className={STYLE.sectionLabel}
            >
              {`疾患コード（${name}）`}
            </Label>
            <Input
              id={`pet-chronic-condition-edit-code-${condition.id}`}
              value={editDraft.conditionCode}
              onChange={(event) =>
                setEditDraft((prev) => ({ ...prev, conditionCode: event.target.value }))
              }
            />
          </div>
          <div className="space-y-1">
            <Label
              htmlFor={`pet-chronic-condition-edit-diagnosed-${condition.id}`}
              className={STYLE.sectionLabel}
            >
              {`診断日（${name}）`}
            </Label>
            <DatePicker
              id={`pet-chronic-condition-edit-diagnosed-${condition.id}`}
              value={editDraft.diagnosedAt}
              onChange={(value) => setEditDraft((prev) => ({ ...prev, diagnosedAt: value }))}
              placeholder="診断日を選択…"
            />
          </div>
          <div className="space-y-1">
            <Label
              htmlFor={`pet-chronic-condition-edit-notes-${condition.id}`}
              className={STYLE.sectionLabel}
            >
              {`備考（${name}）`}
            </Label>
            <Input
              id={`pet-chronic-condition-edit-notes-${condition.id}`}
              value={editDraft.notes}
              onChange={(event) => setEditDraft((prev) => ({ ...prev, notes: event.target.value }))}
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Checkbox
            id={`pet-chronic-condition-edit-active-${condition.id}`}
            checked={editDraft.isActive}
            onCheckedChange={(checked) =>
              setEditDraft((prev) => ({ ...prev, isActive: checked === true }))
            }
            touchTarget
          />
          <Label
            htmlFor={`pet-chronic-condition-edit-active-${condition.id}`}
            className={`flex min-h-11 items-center text-sm font-normal cursor-pointer ${C.text}`}
          >
            {`有効（${name}）`}
          </Label>
        </div>
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            aria-label={`慢性疾患 ${name}の編集をキャンセル`}
            onClick={cancelEdit}
          >
            キャンセル
          </Button>
          <SubmitButton
            size="sm"
            className="text-sm"
            aria-label={`慢性疾患 ${name}を保存`}
            loadingText="保存中..."
          >
            保存
          </SubmitButton>
        </div>
      </fieldset>
      {editErrorMessage !== null ? (
        <p className={`text-sm ${C.danger}`} role="alert" aria-live="assertive">
          {editErrorMessage}
        </p>
      ) : null}
    </form>
  );
}
