import { useActionState, useCallback, useDeferredValue, useEffect, useMemo, useState, useTransition } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";

import { ErrorFallback, LoadingFallback } from "@/components/shared/DataStates";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker/NavigationBlocker";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { handleApiError } from "@/lib/handle-api-error";
import { normalizeKana } from "@/lib/normalize-kana";
import { ResourceHospitalSettings } from "@/types/generated/models";
import {
  ClinicDeleteDialog,
  ClinicMasterList,
  ClinicMasterSidePanel,
} from "../components/ClinicMasterSettingsPanels";
import { CompanyInvoiceSection } from "../components/CompanyInvoiceSection";
import {
  DEFAULT_CLINIC_FORM_DATA,
  clinicToFormData,
  type ClinicFormData,
} from "../components/clinic-master-settings-model";
import {
  useCreateClinic,
  useDeleteClinic,
  useGetClinics,
  useUpdateClinic,
  type Clinic,
  type CreateClinicRequest,
  type UpdateClinicRequest,
} from "../api/clinics";

interface FormState {
  success: boolean;
  timestamp: number;
  nameError?: string;
}

export function ClinicMasterSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceHospitalSettings);
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<Clinic | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] = useState<ClinicFormData>(DEFAULT_CLINIC_FORM_DATA);
  const [pendingDelete, setPendingDelete] = useState<Clinic | null>(null);
  const [isDeletePending, startDeleteTransition] = useTransition();
  const deferredSearch = useDeferredValue(searchTerm);

  const { data: rawClinics, isPending, isError } = useGetClinics();
  const createMutation = useCreateClinic();
  const updateMutation = useUpdateClinic();
  const deleteMutation = useDeleteClinic();

  const [formState, formAction] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      const fd = formData;
      if (!fd.name) {
        return { success: false, timestamp: Date.now(), nameError: "院名は必須です" };
      }

      try {
        if (selectedItem?.id) {
          const req: UpdateClinicRequest = {
            name: fd.name,
            postal_code: fd.postal_code || undefined,
            address: fd.address || undefined,
            phone_number: fd.phone_number || undefined,
            fax_number: fd.fax_number || undefined,
            registration_number: fd.registration_number || undefined,
            director_name: fd.director_name || undefined,
            email: fd.email || undefined,
            website: fd.website || undefined,
            is_active: fd.is_active,
            standard_tax_rate: fd.standard_tax_rate,
            reduced_tax_rate: fd.reduced_tax_rate,
            accounting_document_show_logo: fd.accounting_document_show_logo,
            accounting_document_show_registration_warning: fd.accounting_document_show_registration_warning,
            accounting_document_show_item_category: fd.accounting_document_show_item_category,
            accounting_document_footer_note: fd.accounting_document_footer_note,
            // #190: セクション表示/非表示トグルと表示順
            accounting_document_show_clinic_header: fd.accounting_document_show_clinic_header,
            accounting_document_show_owner_pet_info: fd.accounting_document_show_owner_pet_info,
            accounting_document_show_items_table: fd.accounting_document_show_items_table,
            accounting_document_show_payment_summary: fd.accounting_document_show_payment_summary,
            accounting_document_section_order: fd.accounting_document_section_order,
          };
          await updateMutation.mutateAsync({ id: selectedItem.id, req });
          toast.success("更新しました");
        } else {
          const req: CreateClinicRequest = {
            name: fd.name,
            postal_code: fd.postal_code || undefined,
            address: fd.address || undefined,
            phone_number: fd.phone_number || undefined,
            fax_number: fd.fax_number || undefined,
            registration_number: fd.registration_number || undefined,
            director_name: fd.director_name || undefined,
            email: fd.email || undefined,
            website: fd.website || undefined,
          };
          await createMutation.mutateAsync(req);
          toast.success("登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  useEffect(() => {
    if (formState.success) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- ActionState の成功を検知してUIをリセットするパターン
      setIsEditing(false);
      setSelectedItem(null);
      setFormData(DEFAULT_CLINIC_FORM_DATA);
    }
  }, [formState.success, formState.timestamp]);

  const filteredItems = useMemo(() => {
    const clinics = rawClinics ?? [];
    if (!deferredSearch) return clinics;
    const lower = normalizeKana(deferredSearch).toLowerCase();
    return clinics.filter(
      (clinic) =>
        normalizeKana(clinic.name).toLowerCase().includes(lower) ||
        clinic.phoneNumber.toLowerCase().includes(lower) ||
        clinic.email.toLowerCase().includes(lower),
    );
  }, [rawClinics, deferredSearch]);

  const handleEdit = useCallback((item: Clinic) => {
    setSelectedItem(item);
    setFormData(clinicToFormData(item));
    setIsEditing(true);
  }, []);

  const handleCreate = useCallback(() => {
    setSelectedItem(null);
    setFormData(DEFAULT_CLINIC_FORM_DATA);
    setIsEditing(true);
  }, []);

  const handleCloseEdit = useCallback(() => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData(DEFAULT_CLINIC_FORM_DATA);
  }, []);

  const pendingDeleteId = pendingDelete?.id ?? null;
  const { mutate: deleteClinicMasterFn } = deleteMutation;
  const handleDeleteConfirm = useCallback(() => {
    if (pendingDeleteId === null) return;
    startDeleteTransition(() => {
      deleteClinicMasterFn(pendingDeleteId, {
        onSuccess: () => {
          setPendingDelete(null);
          setIsEditing(false);
          toast.success("削除しました");
        },
        onError: (error) => {
          handleApiError(error, "削除");
        },
      });
    });
  }, [pendingDeleteId, deleteClinicMasterFn]);

  if (isError) {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <ErrorFallback message="医院一覧の取得に失敗しました" />
      </div>
    );
  }

  if (isPending) {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <LoadingFallback />
      </div>
    );
  }

  return (
    <>
      <NavigationBlocker when={isEditing} />
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <ClinicMasterList
            canCreate={canCreate}
            canEdit={canEdit}
            topSection={<CompanyInvoiceSection canEdit={canEdit} />}
            items={filteredItems}
            searchTerm={searchTerm}
            onSearchChange={setSearchTerm}
            onBack={() => navigate(paths.settings.getHref())}
            onCreate={handleCreate}
            onEdit={handleEdit}
          />
        </div>

        {isEditing ? (
          <ClinicMasterSidePanel
            selectedItem={selectedItem}
            formData={formData}
            setFormData={setFormData}
            formAction={formAction}
            nameError={formState.nameError}
            canEdit={canEdit}
            canDelete={canDelete}
            onClose={handleCloseEdit}
            onDeleteClick={setPendingDelete}
          />
        ) : null}
      </div>

      <ClinicDeleteDialog
        pendingDelete={pendingDelete}
        isPending={isDeletePending}
        onClose={() => setPendingDelete(null)}
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
