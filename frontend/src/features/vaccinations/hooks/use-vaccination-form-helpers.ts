import { useState, useEffect, useCallback, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import type { Pet } from "@/types";
import type { EntityReadResult } from "@/lib/entity-read-result";
import type { CreateVaccinationRequest, UpdateVaccinationRequest } from "../api/types";
import type {
  VaccinationFormState,
  VaccinationMutationPermissions,
} from "./use-vaccination-form-model";
import {
  buildCreateVaccinationRequest,
  buildUpdateVaccinationRequest,
  DEFAULT_NEXT_SCHEDULE_TYPE,
  validateVaccinationForm,
  vaccinationOverridesOnDate,
  vaccinationOverridesOnNextDate,
  vaccinationOverridesOnScheduleType,
  vaccinationOverridesOnVaccineId,
} from "./use-vaccination-form-model";

export function useVaccinationFormPetSync(input: {
  isEdit: boolean;
  petId: string | null;
  petFromQuery: Pet | undefined;
  petFromEdit: Pet | undefined;
  isPetLoading: boolean;
  setSelectedPets: (pets: Pet[]) => void;
}) {
  const { isEdit, petId, petFromQuery, petFromEdit, isPetLoading, setSelectedPets } = input;
  const navigate = useNavigate();

  useEffect(() => {
    if (!isEdit) {
      if (petFromQuery) {
        setSelectedPets([petFromQuery]);
      } else if (!petId && !isPetLoading) {
        navigate(paths.vaccinations.selectPet.getHref());
      }
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);

  useEffect(() => {
    if (isEdit && petFromEdit) {
      setSelectedPets([petFromEdit]);
    }
  }, [isEdit, petFromEdit, setSelectedPets]);
}

export function useVaccinationHistoryFilter() {
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const handleClearHistoryFilter = () => {
    setHistorySearchTerm("");
  };
  return {
    filterStartDate,
    setFilterStartDate,
    filterEndDate,
    setFilterEndDate,
    historySearchTerm,
    setHistorySearchTerm,
    sortOrder,
    setSortOrder,
    handleClearHistoryFilter,
  };
}

export function useVaccinationFormFields(input: {
  formData: VaccinationFormState;
  formDataRef: { current: VaccinationFormState };
  vaccinesMaster: readonly { id: string; interval?: string }[];
  vaccineOptions: { value: string; label: string }[];
  doctorName: string;
  setLocalOverrides: (
    updater: (prev: Partial<VaccinationFormState>) => Partial<VaccinationFormState>,
  ) => void;
}) {
  const { formData, formDataRef, vaccinesMaster, vaccineOptions, doctorName, setLocalOverrides } =
    input;
  const setField = useCallback(
    <K extends keyof VaccinationFormState>(key: K, value: VaccinationFormState[K]) => {
      setLocalOverrides((prev) => ({ ...prev, [key]: value }));
    },
    [setLocalOverrides],
  );

  const setVaccineId = useCallback(
    (v: string) => {
      setLocalOverrides((prev) => {
        const selected = vaccinesMaster.find((vac) => vac.id === v);
        const currentDate = prev.date ?? formDataRef.current.date;
        return vaccinationOverridesOnVaccineId(prev, v, currentDate, selected?.interval);
      });
    },
    [vaccinesMaster, formDataRef, setLocalOverrides],
  );

  const setDate = useCallback(
    (v: string) => {
      setLocalOverrides((prev) => {
        const scheduleType =
          prev.nextScheduleType ??
          formDataRef.current.nextScheduleType ??
          DEFAULT_NEXT_SCHEDULE_TYPE;
        return vaccinationOverridesOnDate(prev, v, scheduleType);
      });
    },
    [formDataRef, setLocalOverrides],
  );

  const setNextScheduleType = useCallback(
    (v: string) => {
      setLocalOverrides((prev) => {
        const currentDate = prev.date ?? formDataRef.current.date;
        return vaccinationOverridesOnScheduleType(prev, v, currentDate);
      });
    },
    [formDataRef, setLocalOverrides],
  );

  const setNextDate = useCallback(
    (v: string) => {
      setLocalOverrides((prev) => {
        const base = formDataRef.current;
        const vaccinationDate = prev.date ?? base.date;
        const currentType =
          prev.nextScheduleType ?? base.nextScheduleType ?? DEFAULT_NEXT_SCHEDULE_TYPE;
        return vaccinationOverridesOnNextDate(prev, v, vaccinationDate, currentType);
      });
    },
    [formDataRef, setLocalOverrides],
  );

  const setSupplemental = useCallback((v: string) => setField("supplemental", v), [setField]);
  const setLot1 = useCallback((v: string) => setField("lot1", v), [setField]);
  const setLot2 = useCallback((v: string) => setField("lot2", v), [setField]);
  const setLot3 = useCallback((v: string) => setField("lot3", v), [setField]);
  const setLot4 = useCallback((v: string) => setField("lot4", v), [setField]);
  const setRemarks = useCallback((v: string) => setField("remarks", v), [setField]);

  const form = useMemo(
    () => ({
      doctorName,
      vaccineId: formData.vaccineId,
      setVaccineId,
      vaccineOptions,
      date: formData.date,
      setDate,
      supplemental: formData.supplemental,
      setSupplemental,
      lot1: formData.lot1,
      setLot1,
      lot2: formData.lot2,
      setLot2,
      lot3: formData.lot3,
      setLot3,
      lot4: formData.lot4,
      setLot4,
      nextScheduleType: formData.nextScheduleType,
      setNextScheduleType,
      nextDate: formData.nextDate,
      setNextDate,
      remarks: formData.remarks,
      setRemarks,
    }),
    [
      doctorName,
      formData.vaccineId,
      setVaccineId,
      vaccineOptions,
      formData.date,
      setDate,
      formData.supplemental,
      setSupplemental,
      formData.lot1,
      setLot1,
      formData.lot2,
      setLot2,
      formData.lot3,
      setLot3,
      formData.lot4,
      setLot4,
      formData.nextScheduleType,
      setNextScheduleType,
      formData.nextDate,
      setNextDate,
      formData.remarks,
      setRemarks,
    ],
  );

  return { form };
}

export function createVaccinationDeleteHandler(input: {
  isEdit: boolean;
  id: string | undefined;
  isMutationAllowed: (action: keyof VaccinationMutationPermissions) => boolean;
  isEditPetReady: () => boolean;
  isEditPetDeceased: () => boolean;
  deleteVaccination: (
    id: string,
    opts: { onSuccess: () => void; onError: (error: unknown) => void },
  ) => void;
}): (onSuccess?: () => void) => void {
  return (onSuccess?: () => void) => {
    if (!input.isEdit || !input.id) return;
    if (!input.isMutationAllowed("canDelete")) {
      toast.error("この操作を行う権限がありません");
      return;
    }
    if (!input.isEditPetReady()) {
      toast.error("ペット情報の読み込みが完了してから削除してください");
      return;
    }
    if (input.isEditPetDeceased()) {
      toast.error("死亡したペットの予防接種記録は削除できません");
      return;
    }
    input.deleteVaccination(input.id, {
      onSuccess: () => {
        toast.success("予防接種情報を削除しました");
        onSuccess?.();
      },
      onError: (error) => {
        handleApiError(error, "削除");
      },
    });
  };
}

export function useVaccinationRoutePetId() {
  const [searchParams] = useSearchParams();
  return searchParams.get("petId");
}

interface VaccinationSaveDeps {
  isEdit: boolean;
  id: string | undefined;
  petId: string | null;
  formDataRef: { current: VaccinationFormState };
  entityReadRef: { current: EntityReadResult<import("@/types").VaccinationRecord> };
  selectedPetRef: { current: Pet | undefined };
  queryPetRef: { current: Pet | undefined };
  editPetRef: { current: Pet | undefined };
  isMutationAllowed: (action: keyof VaccinationMutationPermissions) => boolean;
  setFieldErrors: (errors: Record<string, string>) => void;
  updateMutation: {
    mutateAsync: (vars: { id: string; req: UpdateVaccinationRequest }) => Promise<unknown>;
  };
  createMutation: {
    mutateAsync: (req: CreateVaccinationRequest) => Promise<unknown>;
  };
}

export async function runVaccinationSave(deps: VaccinationSaveDeps): Promise<{
  success: boolean;
  timestamp: number;
  fieldErrors?: Record<string, string>;
}> {
  const formData = deps.formDataRef.current;
  const errors = validateVaccinationForm(deps.isEdit, formData);
  if (Object.keys(errors).length > 0) {
    deps.setFieldErrors(errors);
    return { success: false, timestamp: Date.now() };
  }
  deps.setFieldErrors({});

  try {
    if (deps.isEdit && deps.id) {
      if (deps.entityReadRef.current.status !== "found") {
        return { success: false, timestamp: Date.now() };
      }
      const req = buildUpdateVaccinationRequest(formData);
      if (!deps.isMutationAllowed("canEdit")) {
        toast.error("この操作を行う権限がありません");
        return { success: false, timestamp: Date.now() };
      }
      if (!deps.editPetRef.current) {
        toast.error("ペット情報の読み込みが完了してから保存してください");
        return { success: false, timestamp: Date.now() };
      }
      if (deps.editPetRef.current.status === "死亡") {
        toast.error("死亡したペットの予防接種記録は保存できません");
        return { success: false, timestamp: Date.now() };
      }
      await deps.updateMutation.mutateAsync({ id: deps.id, req });
      toast.success("予防接種情報を更新しました");
    } else {
      const pet = deps.petId ? deps.queryPetRef.current : deps.selectedPetRef.current;
      if (!pet) return { success: false, timestamp: Date.now() };
      const req = buildCreateVaccinationRequest(formData, pet.id);
      if (!deps.isMutationAllowed("canCreate")) {
        toast.error("この操作を行う権限がありません");
        return { success: false, timestamp: Date.now() };
      }
      if (pet.status === "死亡") {
        toast.error("死亡したペットの予防接種記録は保存できません");
        return { success: false, timestamp: Date.now() };
      }
      await deps.createMutation.mutateAsync(req);
      toast.success("予防接種を登録しました");
    }
    return { success: true, timestamp: Date.now() };
  } catch (error) {
    handleApiError(error, "保存");
    return { success: false, timestamp: Date.now() };
  }
}
