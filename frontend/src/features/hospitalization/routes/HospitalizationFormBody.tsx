import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { MasterSelectModal, type MasterSelectItem } from "@/components/shared/MasterSelectModal";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { LAYOUT, STYLE } from "@/lib/design-tokens";
import { ResourceHospitalization } from "@/types/generated/models";
import type { StaffItem } from "@/hooks/use-staffs";
import type { HospitalizationFormData } from "../types";
import { HospitalizationFormHeaderExtra } from "./HospitalizationFormHeaderExtra";
import {
  HospitalizationFormFields,
  type HospitalizationFormFieldsProps,
} from "./HospitalizationFormFields";

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
  /** EMR-412 R4: 保存済みレコードの区分（未取得時 undefined）。見出しはここからのみ取る。 */
  savedHospitalizationType?: "入院" | "ホテル";
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
  savedHospitalizationType,
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
  // EMR-412 N2/R4: 新規＝「入院・ホテル登録」、編集・削除確認＝保存済みの区分名。
  // 未保存のフォーム入力を保存済み区分の代わりにしない（不一致時はこの値が変わらない）。
  const typeLabel = savedHospitalizationType ?? "入院";
  const formTitle = hospitalizationId ? `${typeLabel}編集` : "入院・ホテル登録";
  const deleteTitle = `${typeLabel}を削除しますか？`;
  return (
    <>
      <form action={formAction} className="h-full">
        <PageLayout
          title={formTitle}
          onBack={onBack}
          resource={ResourceHospitalization}
          maxWidth={LAYOUT.pageContentMaxWidth.form}
          headerAction={
            <FormHeaderActions
              onCancel={onBack}
              submitLabel={canSubmit ? "保存" : undefined}
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
          <fieldset
            disabled={!canSubmit}
            className={`border-0 p-0 m-0 min-w-0 ${STYLE.readonlyInputs}`}
          >
            <HospitalizationFormFields {...fields} canSubmit={canSubmit} />
          </fieldset>
        </PageLayout>
      </form>
      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onClose={onCloseDeleteConfirm}
        title={deleteTitle}
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
