import { memo, useCallback, useEffect, useState, lazy, Suspense } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { usePermission } from "@/hooks/use-permission";
import { C, LAYOUT } from "@/lib/design-tokens";
import { useGetInsurances } from "../api/get-insurances";
import { useAnimalSpecies } from "../hooks/use-animal-species";
import type { PetFormData } from "../types";
import { PetEditModalFields } from "./PetEditModalFields";
import { createPetFormData } from "./pet-form-data";

const OwnerSearchModal = lazy(() =>
  import("@/components/shared/OwnerSearchModal/OwnerSearchModal").then((m) => ({ default: m.OwnerSearchModal }))
);

interface PetEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  ownerName?: string;
  petData?: PetFormData;
  onSave: (data: PetFormData) => void;
  onChangeOwner?: (newOwner: { id: string; name: string; discountRate: number; membershipType: string }) => void;
}

export const PetEditModal = memo(function PetEditModal({
  open,
  onOpenChange,
  ownerName = "飼主名",
  petData,
  onSave,
  onChangeOwner,
}: PetEditModalProps) {
  const { canEdit } = usePermission("owners");
  const { allSpecies, activeSpecies, isLoading: isLoadingSpecies } = useAnimalSpecies({
    includeInactive: !!petData?.id,
  });
  const animalSpeciesList = petData?.id ? allSpecies : activeSpecies;
  const { data: insuranceList = [], isLoading: isLoadingInsurances } = useGetInsurances();

  const [formData, setFormData] = useState<PetFormData>(() => createPetFormData(petData));
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [isOwnerSearchOpen, setIsOwnerSearchOpen] = useState(false);

  const clearFieldError = useCallback((field: string) => {
    setFieldErrors((prev) => {
      const next = { ...prev };
      delete next[field];
      return next;
    });
  }, []);

  useEffect(() => {
    if (!open) return;
    setFormData(createPetFormData(petData));
    setFieldErrors({});
  // petData の各フィールドではなく petData 参照自体の変化（open トリガー）で十分
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const handleSave = () => {
    const errors: Record<string, string> = {};
    if (!formData.petName.trim()) errors.petName = "ペット名を入力してください";
    if (!formData.animalSpeciesId) errors.animalSpeciesId = "動物種を選択してください";
    if (!formData.gender) errors.gender = "性別を選択してください";
    const trimmedDangerReason = formData.dangerReason?.trim() ?? "";
    if (formData.dangerLevel === "高" && !trimmedDangerReason) {
      errors.dangerReason = "危険度が高の場合は理由を入力してください";
    } else if (trimmedDangerReason && Array.from(trimmedDangerReason).length > 500) {
      errors.dangerReason = "危険理由は500文字以内で入力してください";
    }
    if (formData.weight !== "" && formData.weight !== undefined) {
      const weightNum = parseFloat(formData.weight);
      if (!isNaN(weightNum) && weightNum < 0) {
        errors.weight = "体重は0以上の値を入力してください";
      } else if (!isNaN(weightNum) && weightNum > 200) {
        errors.weight = "体重は200kg以下で入力してください";
      }
    }

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      return;
    }

    setFieldErrors({});
    onSave(formData);
    onOpenChange(false);
    if (!petData?.id) {
      toast.success("ペットを追加しました");
    }
  };

  const handleCancel = useCallback(() => {
    onOpenChange(false);
  }, [onOpenChange]);

  const handleOwnerChange = useCallback(
    (newOwner: { id: string; name: string; discountRate: number; membershipType: string }) => {
      setIsOwnerSearchOpen(false);
      onChangeOwner?.(newOwner);
    },
    [onChangeOwner],
  );

  const isEdit = !!petData?.id;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={`${LAYOUT.modal.xl} overflow-y-auto`}>
        <DialogHeader>
          <div className="flex items-center justify-between">
            <div>
              <DialogTitle className={`text-sm font-bold ${C.text}`}>
                {isEdit ? `${ownerName}のペット情報編集` : `${ownerName}のペット新規登録`}
              </DialogTitle>
              <DialogDescription className={`text-sm ${C.text60}`}>
                {isEdit
                  ? "ペットの情報を編集してください。"
                  : "ペットの基本情報を入力してください。"}
              </DialogDescription>
            </div>
            {isEdit && onChangeOwner ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsOwnerSearchOpen(true)}
                className={`h-8 text-xs ${C.borderMedium}`}
              >
                飼主変更
              </Button>
            ) : null}
          </div>
        </DialogHeader>

        <fieldset disabled={!canEdit} className="border-0 p-0 m-0 min-w-0">
          <PetEditModalFields
            formData={formData}
            setFormData={setFormData}
            fieldErrors={fieldErrors}
            clearFieldError={clearFieldError}
            animalSpeciesList={animalSpeciesList}
            isLoadingSpecies={isLoadingSpecies}
            insuranceList={insuranceList}
            isLoadingInsurances={isLoadingInsurances}
            canEdit={canEdit}
            isEdit={isEdit}
          />

          <div className={`flex justify-end gap-2 mt-4 pt-4 border-t ${C.borderDivider}`}>
            <Button
              variant="outline"
              className={`h-11 text-sm ${C.borderMedium}`}
              onClick={handleCancel}
            >
              キャンセル
            </Button>
            {canEdit ? (
              <PrimaryButton onClick={handleSave} className="text-sm px-4">
                {isEdit ? "更新" : "登録"}
              </PrimaryButton>
            ) : null}
          </div>
        </fieldset>
      </DialogContent>

      {isEdit && onChangeOwner ? (
        <Suspense fallback={null}>
          <OwnerSearchModal
            open={isOwnerSearchOpen}
            onOpenChange={setIsOwnerSearchOpen}
            currentOwnerName={ownerName}
            onSelect={handleOwnerChange}
          />
        </Suspense>
      ) : null}
    </Dialog>
  );
});
