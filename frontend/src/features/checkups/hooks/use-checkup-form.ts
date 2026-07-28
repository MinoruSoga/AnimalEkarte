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

interface CheckupFormState {
  checkupTypeId: string;
  date: string;
  nextDate: string;
  doctorId: string;
  result: string;
}

interface ActionState {
  success: boolean;
  timestamp: number;
}

interface CheckupMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
}

const DENIED_MUTATION_PERMISSIONS: Readonly<CheckupMutationPermissions> = {
  canCreate: false,
  canEdit: false,
};

const DEFAULT_FORM: CheckupFormState = {
  checkupTypeId: "",
  date: "",
  nextDate: "",
  doctorId: "",
  result: "",
};

// useCheckupForm — checkup新規登録フォームのロジックを管理するフック
export function useCheckupForm(
  permissions: Readonly<CheckupMutationPermissions> = DENIED_MUTATION_PERMISSIONS,
) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId") ?? "";

  const { data: pet, isLoading: isPetLoading } = useGetPet(petId);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [localOverrides, setLocalOverrides] = useState<Partial<CheckupFormState>>({});

  const formData: CheckupFormState = { ...DEFAULT_FORM, ...localOverrides };

  // #211: 選択中の健診パッケージのフィールド定義 + 入力値。
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
      const errors: Record<string, string> = {};
      if (!formData.checkupTypeId) errors.checkupTypeId = "健診種別を選択してください";
      if (!formData.date) errors.date = "実施日を入力してください";

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

        // 1. カルテを作成（checkupのサブリソース登録に medical_record_id が必須）
        const medicalRecord = await createMedicalRecordForCheckup({
          pet_id: pet.id,
          owner_id: pet.ownerId,
          visit_date: formData.date,
        });

        // 2. 作成したカルテに健診記録を登録
        if (
          !isMutationAllowed("canCreate")
          || !isMutationAllowed("canEdit")
          || isMutationPetDeceased()
        ) {
          return { success: false, timestamp: Date.now() };
        }
        const checkup = await createCheckupOnMedicalRecord(medicalRecord.id, {
          checkup_type_id: Number(formData.checkupTypeId),
          date: formData.date,
          ...(formData.nextDate ? { next_date: formData.nextDate } : {}),
          ...(formData.doctorId ? { doctor_id: Number(formData.doctorId) } : {}),
          ...(formData.result ? { result: formData.result } : {}),
        });

        // 3. #211 健診パッケージの型付き結果値を保存（入力がある場合のみ）。
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

  // petId未指定時はペット選択画面に戻す
  useEffect(() => {
    if (!petId && !isPetLoading) {
      navigate(paths.checkups.selectPet.getHref());
    }
  }, [petId, isPetLoading, navigate]);

  // 登録成功後に一覧へ遷移
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
    // 健診種別を変えたら、旧パッケージのフィールド値は破棄する。
    setCheckupTypeId: useCallback(
      (v: string) => {
        setField("checkupTypeId", v);
        setFieldValues({});
      },
      [setField],
    ),
    setDate: useCallback((v: string) => setField("date", v), [setField]),
    setNextDate: useCallback((v: string) => setField("nextDate", v), [setField]),
    setDoctorId: useCallback((v: string) => setField("doctorId", v), [setField]),
    setResult: useCallback((v: string) => setField("result", v), [setField]),
  };
}
