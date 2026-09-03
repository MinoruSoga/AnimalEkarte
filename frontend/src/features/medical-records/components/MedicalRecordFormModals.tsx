import { Suspense } from "react";
import type { ComponentProps } from "react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { MedicalRecordDeleteDialog } from "./MedicalRecordFormActions";
import { OwnerSearchModal, StaffSelectionModal, VitalsModal } from "./MedicalRecordLazyModals";

// FE6-9: MedicalRecordForm.tsx:339-390 のモーダル塊を抽出（411→~330 行目安）。
// props は塊内で参照していた自由識別子を逐語的に列挙している（ロジック変更なし）。
interface MedicalRecordFormModalsProps {
  isDeleteConfirmOpen: boolean;
  selectedPetName: string;
  isDeleting: boolean;
  onCloseDeleteConfirm: () => void;
  onConfirmDelete: () => void;
  isNewRecord: boolean;
  recordId: string | undefined;
  isVitalsOpen: boolean;
  onVitalsOpenChange: ComponentProps<typeof VitalsModal>["onOpenChange"];
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
  isPetDeceased?: boolean;
  isStaffModalOpen: boolean;
  staffName: string;
  onSelectStaff: ComponentProps<typeof StaffSelectionModal>["onSelect"];
  onStaffModalOpenChange: ComponentProps<typeof StaffSelectionModal>["onOpenChange"];
  isOwnerSearchOpen: boolean;
  onOwnerSearchOpenChange: ComponentProps<typeof OwnerSearchModal>["onOpenChange"];
  selectedPetOwnerName: string | undefined;
  onSelectOwner: ComponentProps<typeof OwnerSearchModal>["onSelect"];
  pendingOwnerChange: unknown;
  onCancelOwnerChange: () => void;
  onConfirmOwnerChange: () => void;
}

export function MedicalRecordFormModals({
  isDeleteConfirmOpen,
  selectedPetName,
  isDeleting,
  onCloseDeleteConfirm,
  onConfirmDelete,
  isNewRecord,
  recordId,
  isVitalsOpen,
  onVitalsOpenChange,
  recordClinicId,
  isPetDeceased = false,
  isStaffModalOpen,
  staffName,
  onSelectStaff,
  onStaffModalOpenChange,
  isOwnerSearchOpen,
  onOwnerSearchOpenChange,
  selectedPetOwnerName,
  onSelectOwner,
  pendingOwnerChange,
  onCancelOwnerChange,
  onConfirmOwnerChange,
}: MedicalRecordFormModalsProps) {
  return (
    <>
      <MedicalRecordDeleteDialog
        open={isDeleteConfirmOpen}
        petName={selectedPetName}
        isDeleting={isDeleting}
        onClose={onCloseDeleteConfirm}
        onConfirm={onConfirmDelete}
      />
      {/* Vitals Modal */}
      <Suspense fallback={null}>
        {!isNewRecord && recordId ? (
          <VitalsModal
            open={isVitalsOpen}
            onOpenChange={onVitalsOpenChange}
            medicalRecordId={recordId}
            recordClinicId={recordClinicId}
            isPetDeceased={isPetDeceased}
          />
        ) : null}
      </Suspense>

      {/* Staff Selection Modal */}
      <Suspense fallback={null}>
        <StaffSelectionModal
          open={isStaffModalOpen}
          selectedStaffName={staffName}
          onSelect={onSelectStaff}
          onOpenChange={onStaffModalOpenChange}
        />
      </Suspense>

      {/* Owner Search Modal (edit mode only) */}
      {!isNewRecord && recordId ? (
        <Suspense fallback={null}>
          <OwnerSearchModal
            open={isOwnerSearchOpen}
            onOpenChange={onOwnerSearchOpenChange}
            currentOwnerName={selectedPetOwnerName}
            onSelect={onSelectOwner}
          />
        </Suspense>
      ) : null}

      {/* BUG-373: 飼主変更 確認ダイアログ (discount_rate/membership_type 不一致時のみ表示) */}
      <ConfirmDialog
        open={!!pendingOwnerChange}
        onClose={onCancelOwnerChange}
        onConfirm={onConfirmOwnerChange}
        title="飼主変更の確認"
        description="飼主によって値引率や会員区分が異なるため、今後の会計金額が変動する可能性があります。変更を続行してよろしいですか?"
        confirmLabel="続行"
        cancelLabel="キャンセル"
      />
    </>
  );
}
