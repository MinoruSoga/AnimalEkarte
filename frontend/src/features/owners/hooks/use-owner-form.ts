import {
  useState,
  useCallback,
  useActionState,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import { transformCreatePetRequest, PET_STATUS_REVERSE_MAP } from "@/lib/transforms/pet";
import type {
  CreateOwnerPetRequest,
  CreateOwnerRequest,
  UpdateOwnerRequest,
  Owner,
} from "@/types/owner";
import type { PetMutations } from "@/types/pet";
import type { OwnerData, PetFormData, MembershipTypeLabel } from "../types";
import { createOwner } from "../api/create-owner";
import { updateOwner } from "../api/update-owner";
import { usePetFormListState } from "./use-pet-form-list-state";

type CreateOwnerPetRequestWithDangerReason = CreateOwnerPetRequest & {
  danger_reason?: string;
};

const MEMBERSHIP_TYPE_TO_API: Record<string, string> = {
  "非会員": "non_member",
  "会員": "member",
  "退亡者": "deceased",
  "他診/準": "transferred",
};

const DEFAULT_OWNER_DATA: OwnerData = {
  ownerId: "",
  postalCode: "",
  company: "",
  membershipType: "非会員" as MembershipTypeLabel,
  ownerName: "",
  address1: "",
  ownerNameKana: "",
  address2: "",
  homeAddress1: "",
  homeAddress2: "",
  isDangerous: false,
  birthDate: "",
  email: "",
  phone: "",
  companyPhone: "",
  remarks: "",
};

export interface OwnerMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

const DENIED_MUTATION_PERMISSIONS: Readonly<OwnerMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
};

function mapOwnerToFormData(owner: Owner): OwnerData {
  return {
    ownerId: owner.id,
    postalCode: owner.postalCode,
    company: owner.company,
    membershipType: (owner.membershipType as MembershipTypeLabel) || "非会員",
    ownerName: owner.ownerName,
    address1: owner.address1,
    ownerNameKana: owner.ownerNameKana || "",
    address2: owner.address2,
    homeAddress1: owner.homeAddress1,
    homeAddress2: owner.homeAddress2,
    homePostalCode: owner.homePostalCode,
    isDangerous: owner.isDangerous,
    birthDate: owner.birthDate || "",
    email: owner.email,
    phone: owner.phone,
    companyPhone: owner.companyPhone,
    remarks: owner.remarks,
    discountRate: owner.discountRate,
    dmPreference: owner.dmPreference,
  };
}

function mapOwnerPetsToFormData(owner: Owner): PetFormData[] {
  if (!owner.pets) return [];
  return owner.pets.map((backendPet): PetFormData => ({
    id: backendPet.id,
    petNumber: backendPet.petNumber || "",
    petName: backendPet.name,
    petNameKana: backendPet.petNameKana || "",
    status: backendPet.status || "生存",
    species: backendPet.species,
    animalSpeciesId: backendPet.animalSpeciesId,
    gender: backendPet.gender || "",
    birthDate: backendPet.birthDate || "",
    color: backendPet.color || "",
    bloodType: backendPet.bloodType || "",
    microchipNumber: backendPet.microchipNumber || "",
    weight: backendPet.weight || "",
    food: backendPet.food || "",
    environment: backendPet.environment || "",
    neuteredDate: backendPet.neuteredDate || "",
    acquisitionType: (backendPet.acquisitionType as PetFormData["acquisitionType"]) || "購入",
    dangerLevel: (backendPet.dangerLevel as PetFormData["dangerLevel"]) || "低",
    dangerReason: backendPet.dangerReason || "",
    remarks: backendPet.remarks || "",
    breed: backendPet.breed,
    insuranceId: backendPet.insuranceId,
    insuranceName: undefined,
    insuranceDetails: backendPet.insuranceDetails,
    deceasedAt: backendPet.deceasedAt,
    deceasedReason: backendPet.deceasedReason,
  }));
}

function mapPendingPetToCreateRequest(
  pet: PetFormData & { animalSpeciesId: string },
): CreateOwnerPetRequestWithDangerReason {
  const request = transformCreatePetRequest({
    ownerId: "0",
    name: pet.petName || "",
    animalSpeciesId: pet.animalSpeciesId,
    petNameKana: pet.petNameKana,
    breed: pet.breed,
    color: pet.color,
    bloodType: pet.bloodType,
    microchipNumber: pet.microchipNumber,
    gender: pet.gender,
    birthDate: pet.birthDate,
    weight: pet.weight,
    food: pet.food,
    environment: pet.environment,
    neuteredDate: pet.neuteredDate,
    acquisitionType: pet.acquisitionType,
    dangerLevel: pet.dangerLevel,
    dangerReason: pet.dangerReason,
    status: PET_STATUS_REVERSE_MAP[pet.status],
    insuranceId: pet.insuranceId,
    remarks: pet.remarks,
  });

  return {
    name: request.name,
    animal_species_id: request.animal_species_id,
    name_kana: request.name_kana,
    breed: request.breed,
    color: request.color,
    blood_type: request.blood_type,
    microchip_number: request.microchip_number,
    gender: request.gender,
    status: request.status,
    birth_date: request.birth_date,
    weight: request.weight,
    neutered_date: request.neutered_date,
    acquisition_type: request.acquisition_type,
    danger_level: request.danger_level,
    danger_reason: request.danger_reason,
    food: request.food,
    environment: request.environment,
    insurance_id: request.insurance_id,
    remarks: request.remarks,
  };
}

export function useOwnerForm(
  id?: string,
  initialOwner?: Owner,
  petMutations?: PetMutations,
  permissions: Readonly<OwnerMutationPermissions> = DENIED_MUTATION_PERMISSIONS,
) {
  const isEdit = !!id;
  const queryClient = useQueryClient();
  const { canCreate, canEdit, canDelete } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
  }, [canCreate, canDelete, canEdit]);
  const isMutationAllowed = useCallback(
    (action: keyof OwnerMutationPermissions) =>
      permissionsRef.current[action] === true,
    [],
  );

  const [ownerData, setOwnerData] = useState<OwnerData>(
    () => initialOwner ? mapOwnerToFormData(initialOwner) : DEFAULT_OWNER_DATA
  );

  const {
    pets,
    setPets,
    petModalOpen,
    setPetModalOpen,
    editingPet,
    handleAddPet,
    handleEditPet,
    handleDeletePet,
    handleSavePet,
    handlePetLifecycleChange,
  } = usePetFormListState({
    id,
    initialPets: initialOwner ? mapOwnerPetsToFormData(initialOwner) : [],
    petMutations,
    permissions,
  });

  const [manualErrors, setManualErrors] = useState<Record<string, string>>({});

  /**
   * React 19 useActionState を使用したフォームアクション。
   * バリデーションエラーや保存の成否を状態として管理する。
   */
  const [formState, formAction, isPending] = useActionState(
    async (prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const errors: Record<string, string> = {};
      if (!ownerData.ownerName.trim()) errors.ownerName = "飼主名を入力してください";
      if (!ownerData.ownerNameKana.trim()) errors.ownerNameKana = "飼主名よみを入力してください";
      if (!ownerData.phone.trim()) {
        errors.phone = "電話番号を入力してください";
      } else if (!/^[\d\-+() ]+$/.test(ownerData.phone.trim())) {
        // BUG-066: 電話番号は数字・ハイフン・括弧・スペースのみ許容
        errors.phone = "電話番号の形式が正しくありません（数字・ハイフンのみ）";
      }
      // BUG-066: メールアドレス形式チェック
      if (ownerData.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(ownerData.email)) {
        errors.email = "メールアドレスの形式が正しくありません";
      }
      // BUG-066: 郵便番号フォーマットチェック（ハイフンあり: 123-4567 / なし: 1234567）
      const postalPattern = /^\d{3}-?\d{4}$/;
      if (ownerData.postalCode && !postalPattern.test(ownerData.postalCode.trim())) {
        errors.postalCode = "郵便番号の形式が正しくありません（例: 123-4567）";
      }
      if (ownerData.homePostalCode && !postalPattern.test(ownerData.homePostalCode.trim())) {
        errors.homePostalCode = "郵便番号の形式が正しくありません（例: 123-4567）";
      }
      // BUG-066: 値引率は 0〜100 の範囲
      if (ownerData.discountRate != null && (ownerData.discountRate < 0 || ownerData.discountRate > 100)) {
        errors.discountRate = "値引率は0〜100の範囲で入力してください";
      }

      if (Object.keys(errors).length > 0) {
        return { success: false, fieldErrors: errors, timestamp: Date.now() };
      }

      try {
        const ownerRequestPayload = {
          owner_name: ownerData.ownerName,
          owner_name_kana: ownerData.ownerNameKana || undefined,
          company: ownerData.company,
          postal_code: ownerData.postalCode,
          address1: ownerData.address1,
          address2: ownerData.address2,
          home_postal_code: ownerData.homePostalCode || "",
          home_address1: ownerData.homeAddress1,
          home_address2: ownerData.homeAddress2,
          phone: ownerData.phone,
          company_phone: ownerData.companyPhone,
          email: ownerData.email,
          remarks: ownerData.remarks,
          is_dangerous: ownerData.isDangerous,
          discount_rate: ownerData.discountRate,
          membership_type: MEMBERSHIP_TYPE_TO_API[ownerData.membershipType] ?? ownerData.membershipType,
          dm_preference: ownerData.dmPreference,
        };

        if (isEdit && id) {
          const updateData: UpdateOwnerRequest = {
            ...ownerRequestPayload,
            birth_date: ownerData.birthDate || null,
          };
          if (!isMutationAllowed("canEdit")) {
            return { success: false, timestamp: Date.now() };
          }
          await updateOwner(id, updateData);
          await queryClient.invalidateQueries({ queryKey: queryKeys.owners.all() });
          toast.success("飼主情報を更新しました");
          return { success: true, timestamp: Date.now() };
        } else {
          const pendingPets = pets.filter(
            (pet): pet is PetFormData & { animalSpeciesId: string } =>
              pet.isPending === true && Boolean(pet.animalSpeciesId),
          );
          const createData: CreateOwnerRequest = {
            ...ownerRequestPayload,
            birth_date: ownerData.birthDate || undefined,
            // #84: 登録先医院の指定（未選択時は undefined → サーバ側で現在の医院）
            clinic_id: ownerData.clinicId ? Number(ownerData.clinicId) : undefined,
            pets: pendingPets.map(mapPendingPetToCreateRequest),
          };
          if (!isMutationAllowed("canCreate")) {
            return { success: false, timestamp: Date.now() };
          }
          const newOwner = await createOwner(createData);
          await queryClient.invalidateQueries({ queryKey: queryKeys.owners.all() });

          toast.success("飼主情報を登録しました");
          return { success: true, data: newOwner.id, timestamp: Date.now() };
        }
      } catch (error) {
        handleApiError(error, "保存");
        return { ...prevState, success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  // Sync manual errors when formState changes (new submission)
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- ActionState のエラーをフォームフィールドに同期するパターン
    setManualErrors(formState.fieldErrors || {});
  }, [formState.fieldErrors, formState.timestamp]);

  const clearFieldError = useCallback((field: string) => {
    setManualErrors((prev) => {
      if (!prev[field]) return prev;
      const next = { ...prev };
      delete next[field];
      return next;
    });
  }, []);

  return {
    isEdit,
    isLoading: isPending,
    ownerData,
    setOwnerData,
    pets,
    setPets,
    petModalOpen,
    setPetModalOpen,
    editingPet,
    handleAddPet,
    handleEditPet,
    handleDeletePet,
    handleSavePet,
    handlePetLifecycleChange,
    formAction,
    formState,
    fieldErrors: manualErrors,
    clearFieldError,
  };
}
