import {
  useState,
  useEffect,
  useLayoutEffect,
  useCallback,
  useActionState,
  useRef,
} from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import { useGetPet } from "@/hooks/use-pet";
import {
  createCheckupOnMedicalRecord,
  createMedicalRecordForCheckup,
} from "../api/create-checkup-medical-record";
import { useGetCheckupTypeFields } from "../api/get-checkup-type-fields";
import { replaceCheckupFieldResults } from "../api/replace-checkup-field-results";
import { buildCheckupResultsPayload, type CheckupFieldValue } from "../components/DynamicCheckupFields";
import {
  buildCheckupOnMedicalRecordRequest,
  checkupOverridesOnDate,
  checkupOverridesOnNextDate,
  checkupOverridesOnScheduleType,
  DEFAULT_CHECKUP_FORM,
  DENIED_MUTATION_PERMISSIONS,
  validateCheckupForm,
  type CheckupFormState,
  type CheckupMutationPermissions,
} from "./use-checkup-form-model";

interface ActionState {
  success: boolean;
  timestamp: number;
}

export function useCheckupForm(
  permissions: Readonly<CheckupMutationPermissions> = DENIED_MUTATION_PERMISSIONS,
) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId") ?? "";

  const { data: pet, isLoading: isPetLoading } = useGetPet(petId);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [localOverrides, setLocalOverrides] = useState<Partial<CheckupFormState>>({});
  const formData: CheckupFormState = { ...DEFAULT_CHECKUP_FORM, ...localOverrides };

  const { data: checkupFields = [] } = useGetCheckupTypeFields(formData.checkupTypeId);
  const [fieldValues, setFieldValues] = useState<Record<number, CheckupFieldValue>>({});
  const { canCreate, canEdit } = permissions;
  const permissionsRef = useRef(permissions);
  const petStatusRef = useRef(pet?.status);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit };
  }, [canCreate, canEdit]);
  useLayoutEffect(() => {
    petStatusRef.current = pet?.status;
  }, [pet?.status]);
  const isMutationAllowed = useCallback(
    (action: keyof CheckupMutationPermissions) =>
      permissionsRef.current[action] === true,
    [],
  );
  const isMutationPetDeceased = useCallback(
    () => petStatusRef.current === "死亡",
    [],
  );

  const setField = useCallback(<K extends keyof CheckupFormState>(key: K, value: CheckupFormState[K]) => {
    setLocalOverrides((prev) => ({ ...prev, [key]: value }));
  }, []);

  const setFieldValue = useCallback((fieldId: number, value: CheckupFieldValue) => {
    setFieldValues((prev) => ({ ...prev, [fieldId]: value }));
  }, []);

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const errors = validateCheckupForm(formData);
      if (Object.keys(errors).length > 0) {
        setFieldErrors(errors);
        return { success: false, timestamp: Date.now() };
      }
      setFieldErrors({});

      if (!pet || pet.status === "死亡") {
        return { success: false, timestamp: Date.now() };
      }

      try {
        if (
          !isMutationAllowed("canCreate") ||
          !isMutationAllowed("canEdit") ||
          isMutationPetDeceased()
        ) {
          return { success: false, timestamp: Date.now() };
        }

        const medicalRecord = await createMedicalRecordForCheckup({
          pet_id: pet.id,
          owner_id: pet.ownerId,
          visit_date: formData.date,
        });

        if (
          !isMutationAllowed("canCreate")
          || !isMutationAllowed("canEdit")
          || isMutationPetDeceased()
        ) {
          return { success: false, timestamp: Date.now() };
        }
        const checkup = await createCheckupOnMedicalRecord(
          medicalRecord.id,
          buildCheckupOnMedicalRecordRequest(formData),
        );

        const resultsPayload = buildCheckupResultsPayload(checkupFields, fieldValues);
        if (resultsPayload.length > 0) {
          if (
            !isMutationAllowed("canCreate")
            || !isMutationAllowed("canEdit")
            || isMutationPetDeceased()
          ) {
            return { success: false, timestamp: Date.now() };
          }
          await replaceCheckupFieldResults(medicalRecord.id, checkup.id, resultsPayload);
        }

        toast.success("定期健診を登録しました");
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  useEffect(() => {
    if (!petId && !isPetLoading) {
      navigate(paths.checkups.selectPet.getHref());
    }
  }, [petId, isPetLoading, navigate]);

  useEffect(() => {
    if (formState.success) {
      navigate(paths.checkups.getHref());
    }
  }, [formState.success, formState.timestamp, navigate]);

  return {
    pet,
    isPetLoading,
    form: formData,
    formAction,
    formState,
    isPending,
    fieldErrors,
    checkupFields,
    fieldValues,
    setFieldValue,
    setCheckupTypeId: useCallback(
      (v: string) => {
        setField("checkupTypeId", v);
        setFieldValues({});
      },
      [setField],
    ),
    setDate: useCallback((v: string) => {
      setLocalOverrides((prev) => checkupOverridesOnDate(prev, v));
    }, []),
    setNextScheduleType: useCallback((v: string) => {
      setLocalOverrides((prev) => checkupOverridesOnScheduleType(prev, v));
    }, []),
    setNextDate: useCallback((v: string) => {
      setLocalOverrides((prev) => checkupOverridesOnNextDate(prev, v));
    }, []),
    setDoctorId: useCallback((v: string) => setField("doctorId", v), [setField]),
    setResult: useCallback((v: string) => setField("result", v), [setField]),
  };
}
