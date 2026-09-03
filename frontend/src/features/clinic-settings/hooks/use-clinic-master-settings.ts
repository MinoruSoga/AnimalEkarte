import { useActionState, useCallback, useDeferredValue, useEffect, useLayoutEffect, useMemo, useRef, useState, useTransition } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";

import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { usePermission } from "@/hooks/use-permission";
import { handleApiError } from "@/lib/handle-api-error";
import { ResourceHospitalSettings } from "@/types/generated/models";
import {
  DEFAULT_CLINIC_FORM_DATA,
  buildCreateClinicRequest,
  buildUpdateClinicRequest,
  clinicToFormData,
  filterClinics,
  type ClinicFormData,
} from "../components/clinic-master-settings-model";
import {
  useCreateClinic,
  useDeleteClinic,
  useGetClinics,
  useUpdateClinic,
  type Clinic,
} from "../api/clinics";

interface FormState {
  success: boolean;
  timestamp: number;
  nameError?: string;
}

export function useClinicMasterSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceHospitalSettings);
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<Clinic | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [formData, setFormData] = useState<ClinicFormData>(DEFAULT_CLINIC_FORM_DATA);
  const [pendingDelete, setPendingDelete] = useState<Clinic | null>(null);
  const [isDeletePending, startDeleteTransition] = useTransition();
  const deferredSearch = useDeferredValue(searchTerm);
  // BUG-018/019: useActionState は stale な formData/selectedItem を掴む。入力に name が無く FormData も空。
  const formDataRef = useRef(formData);
  const selectedItemRef = useRef(selectedItem);
  useLayoutEffect(() => {
    formDataRef.current = formData;
    selectedItemRef.current = selectedItem;
  }, [formData, selectedItem]);

  const { data: rawClinics, isPending, isError } = useGetClinics();
  const createMutation = useCreateClinic();
  const updateMutation = useUpdateClinic();
  const deleteMutation = useDeleteClinic();

  const [formState, formAction] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      const fd = formDataRef.current;
      const selected = selectedItemRef.current;
      if (!fd.name) {
        return { success: false, timestamp: Date.now(), nameError: "院名は必須です" };
      }
      try {
        if (selected?.id) {
          await updateMutation.mutateAsync({ id: selected.id, req: buildUpdateClinicRequest(fd) });
          toast.success("更新しました");
        } else {
          await createMutation.mutateAsync(buildCreateClinicRequest(fd));
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

  const filteredItems = useMemo(
    () => filterClinics(rawClinics ?? [], deferredSearch, activeFilters),
    [rawClinics, deferredSearch, activeFilters],
  );

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

  const emptyMessage = isError
    ? "医院一覧の取得に失敗しました"
    : isPending
      ? "読み込み中..."
      : "医院が登録されていません";

  return {
    navigate,
    canCreate,
    canEdit,
    canDelete,
    isEditing,
    selectedItem,
    searchTerm,
    setSearchTerm,
    activeFilters,
    setActiveFilters,
    formData,
    setFormData,
    pendingDelete,
    setPendingDelete,
    isDeletePending,
    isPending,
    isError,
    formState,
    formAction,
    filteredItems,
    handleEdit,
    handleCreate,
    handleCloseEdit,
    handleDeleteConfirm,
    emptyMessage,
  };
}
