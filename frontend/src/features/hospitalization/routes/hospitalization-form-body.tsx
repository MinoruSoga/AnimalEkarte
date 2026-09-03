import { FileText } from "lucide-react";

import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import {
  MasterSelectModal,
  type MasterSelectItem,
} from "@/components/shared/MasterSelectModal";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceHospitalization } from "@/types/generated/models";
import type { StaffItem } from "@/hooks/use-staffs";
import type { HospitalizationFormData } from "../types";
import { HospitalizationFormHeaderExtra } from "./hospitalization-form-header-extra";
import { HospitalizationFormFields, type HospitalizationFormFieldsProps } from "./hospitalization-form-fields";

interface HospitalizationFormBodyProps {
  hospitalizationId: string | undefined;
  canSubmit: boolean;
  canShowDelete: boolean;
  isDirty: boolean;
  isDeleteConfirmOpen: boolean;
  isDeletePending: boolean;
  staffModalOpen: boolean;
  doctorStaffItems: StaffItem[];
  formData: HospitalizationFormData;
  formAction: (payload: FormData) => void;
  fields: HospitalizationFormFieldsProps;
  onBack: () => void;
  onOpenDetail: () => void;
  onOpenDeleteConfirm: () => void;
  onCloseDeleteConfirm: () => void;
  onConfirmDelete: () => void;
  onStaffModalOpenChange: (open: boolean) => void;
  onSelectDoctor: (item: MasterSelectItem) => void;
}

export function HospitalizationFormBody({
  hospitalizationId,
  canSubmit,
  canShowDelete,
  isDirty,
  isDeleteConfirmOpen,
  isDeletePending,
  staffModalOpen,
  doctorStaffItems,
  formData,
  formAction,
  fields,
  onBack,
  onOpenDetail,
  onOpenDeleteConfirm,
  onCloseDeleteConfirm,
  onConfirmDelete,
  onStaffModalOpenChange,
  onSelectDoctor,
}: HospitalizationFormBodyProps) {
  return (
    <>
    <form action={formAction} className="h-full">
    <PageLayout
      title={hospitalizationId ? "入院編集" : "入院登録"}
      onBack={onBack}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      resource={ResourceHospitalization}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
      headerAction={
        <FormHeaderActions
          onCancel={onBack}
          submitLabel={canSubmit ? (hospitalizationId ? "更新" : "保存") : undefined}
          extra={
            <HospitalizationFormHeaderExtra
              hospitalizationId={hospitalizationId}
              canShowDelete={canShowDelete}
              onOpenDetail={onOpenDetail}
              onOpenDeleteConfirm={onOpenDeleteConfirm}
            />
          }
        />
      }
    >
        <NavigationBlocker when={isDirty} />
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        <HospitalizationFormFields {...fields} canSubmit={canSubmit} />
        </fieldset>
    </PageLayout>
    </form>
    <ConfirmDialog
        open={isDeleteConfirmOpen}
        onClose={onCloseDeleteConfirm}
        title="入院を削除しますか？"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onConfirmDelete}
        isPending={isDeletePending}
      />
    <MasterSelectModal
      open={staffModalOpen}
      onOpenChange={onStaffModalOpenChange}
      title="担当医を選択"
      items={doctorStaffItems}
      selectedValue={formData.doctorId}
      matchBy="id"
      searchPlaceholder="担当医を検索..."
      onSelect={onSelectDoctor}
    />
    </>
  );
}
