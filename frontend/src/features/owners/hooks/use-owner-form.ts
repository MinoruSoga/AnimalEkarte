// React/Framework
import { useState, useCallback, useActionState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import axios from "axios";

// External
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";

// Internal - shared (feature 間 import 禁止のため @/features/pets は使用不可)
import { transformCreatePetRequest, transformUpdatePetRequest, PET_STATUS_REVERSE_MAP } from "@/lib/transforms/pet";
import type { CreateOwnerRequest, UpdateOwnerRequest } from "@/types/owner";
import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";
import type { PetMutations } from "@/types/pet";
import type { OwnerData, PetFormData, MembershipType } from "../types";

// feature 内
import { createOwner } from "../api/create-owner";
import { updateOwner } from "../api/update-owner";

// BUG-067: NULL バイト・制御文字を除去して安全なテキストを返す（多層防衛）
// eslint-disable-next-line no-control-regex
const CONTROL_CHAR_RE = /[\x00-\x1F\x7F]/g;
const sanitizeText = (value: string): string =>
  value.replace(CONTROL_CHAR_RE, "").trim();

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
  membershipType: "非会員" as MembershipType,
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

function mapOwnerToFormData(owner: Owner): OwnerData {
  return {
    ownerId: owner.id,
    postalCode: owner.postalCode,
    company: owner.company,
    membershipType: (owner.membershipType as MembershipType) || "非会員",
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
    weight: backendPet.weight || "",
    food: backendPet.food || "",
    environment: backendPet.environment || "",
    neuteredDate: backendPet.neuteredDate || "",
    acquisitionType: (backendPet.acquisitionType as PetFormData["acquisitionType"]) || "購入",
    dangerLevel: (backendPet.dangerLevel as PetFormData["dangerLevel"]) || "低",
    remarks: backendPet.remarks || "",
    breed: backendPet.breed,
    insuranceId: backendPet.insuranceId,
    insuranceName: undefined,
    insuranceDetails: backendPet.insuranceDetails,
  }));
}

interface FormState {
  fieldErrors: Record<string, string>;
  success: boolean;
  timestamp: number;
  /** 新規登録成功時の飼主ID（詳細ページへのリダイレクト用）*/
  createdOwnerId?: string;
}

export function useOwnerForm(
  id?: string,
  initialOwner?: Owner,
  petMutations?: PetMutations
) {
  const isEdit = !!id;
  const queryClient = useQueryClient();

  const [ownerData, setOwnerData] = useState<OwnerData>(
    () => initialOwner ? mapOwnerToFormData(initialOwner) : DEFAULT_OWNER_DATA
  );

  const [pets, setPets] = useState<PetFormData[]>(
    () => initialOwner ? mapOwnerPetsToFormData(initialOwner) : []
  );
  const [petModalOpen, setPetModalOpen] = useState(false);
  const [editingPet, setEditingPet] = useState<PetFormData | null>(null);

  /**
   * React 19 useActionState を使用したフォームアクション。
   * バリデーションエラーや保存の成否を状態として管理する。
   */
  const [formState, formAction, isPending] = useActionState(
    async (prevState: FormState, _formData: FormData): Promise<FormState> => {
      const errors: Record<string, string> = {};
      if (!ownerData.ownerName.trim()) errors.ownerName = "飼主名を入力してください";
      if (!ownerData.ownerNameKana.trim()) errors.ownerNameKana = "飼主名（カナ）を入力してください";
      if (!ownerData.phone.trim()) errors.phone = "電話番号を入力してください";

      // BUG-066: 電話番号フォーマットバリデーション
      if (ownerData.phone.trim() && !/^[0-9\-+()]{7,20}$/.test(ownerData.phone.trim())) {
        errors.phone = "電話番号の形式が正しくありません（例: 090-1234-5678）";
      }

      // BUG-066: メールアドレスフォーマットバリデーション
      if (ownerData.email.trim() && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(ownerData.email.trim())) {
        errors.email = "メールアドレスの形式が正しくありません";
      }

      // BUG-066: 値引率バリデーション
      if (ownerData.discountRate !== undefined && ownerData.discountRate !== null) {
        const rate = Number(ownerData.discountRate);
        if (!isNaN(rate) && (rate < 0 || rate > 100)) {
          errors.discountRate = "値引率は0〜100の範囲で入力してください";
        }
      }

      if (Object.keys(errors).length > 0) {
        toast.error("必須項目が未入力です");
        return { fieldErrors: errors, success: false, timestamp: Date.now() };
      }

      try {
        // BUG-067: NULL バイト・制御文字を除去してから送信
        const ownerRequestPayload = {
          owner_name: sanitizeText(ownerData.ownerName),
          owner_name_kana: ownerData.ownerNameKana ? sanitizeText(ownerData.ownerNameKana) : undefined,
          company: sanitizeText(ownerData.company),
          postal_code: sanitizeText(ownerData.postalCode),
          address1: sanitizeText(ownerData.address1),
          address2: sanitizeText(ownerData.address2),
          home_postal_code: ownerData.homePostalCode ? sanitizeText(ownerData.homePostalCode) : "",
          home_address1: sanitizeText(ownerData.homeAddress1),
          home_address2: sanitizeText(ownerData.homeAddress2),
          phone: sanitizeText(ownerData.phone),
          company_phone: sanitizeText(ownerData.companyPhone),
          email: sanitizeText(ownerData.email),
          remarks: ownerData.remarks.replace(CONTROL_CHAR_RE, ""),
          is_dangerous: ownerData.isDangerous,
          discount_rate: ownerData.discountRate,
          membership_type: MEMBERSHIP_TYPE_TO_API[ownerData.membershipType] ?? ownerData.membershipType,
        };

        if (isEdit && id) {
          const updateData: UpdateOwnerRequest = ownerRequestPayload;
          await updateOwner(id, updateData);
          await queryClient.invalidateQueries({ queryKey: ["owners"] });
          toast.success("飼主情報を更新しました");
          return { fieldErrors: {}, success: true, timestamp: Date.now() };
        } else {
          const createData: CreateOwnerRequest = {
            ...ownerRequestPayload,
            owner_name: ownerData.ownerName,
          };
          const newOwner = await createOwner(createData);
          await queryClient.invalidateQueries({ queryKey: ["owners"] });

          const pendingPets = pets.filter(p => p.isPending && p.animalSpeciesId);
          if (pendingPets.length > 0 && petMutations) {
            const results = await Promise.allSettled(
              pendingPets.map(pet =>
                petMutations.createPetFn(
                  transformCreatePetRequest({
                    ownerId: newOwner.id,
                    name: pet.petName || "",
                    animalSpeciesId: pet.animalSpeciesId!,
                    petNumber: pet.petNumber,
                    petNameKana: pet.petNameKana,
                    breed: pet.breed,
                    color: pet.color,
                    gender: pet.gender,
                    birthDate: pet.birthDate,
                    weight: pet.weight,
                    food: pet.food,
                    environment: pet.environment,
                    neuteredDate: pet.neuteredDate,
                    acquisitionType: pet.acquisitionType,
                    dangerLevel: pet.dangerLevel,
                    status: PET_STATUS_REVERSE_MAP[pet.status],
                    insuranceId: pet.insuranceId,
                    remarks: pet.remarks,
                  })
                )
              )
            );
            const failedCount = results.filter(r => r.status === "rejected").length;
            if (failedCount > 0) {
              toast.warning(`${failedCount}件のペット追加に失敗しました`);
            }
          }

          toast.success("飼主情報を登録しました");
          // BUG-065: 新規登録後に詳細ページへリダイレクトするため createdOwnerId を返す
          return { fieldErrors: {}, success: true, timestamp: Date.now(), createdOwnerId: newOwner.id };
        }
      } catch (error) {
        // BUG-064: 409 Conflict はメールアドレス重複エラーとして扱う
        if (axios.isAxiosError(error) && error.response?.status === 409) {
          const emailError = "このメールアドレスはすでに登録されています";
          toast.error(emailError);
          return {
            ...prevState,
            fieldErrors: { ...prevState.fieldErrors, email: emailError },
            success: false,
            timestamp: Date.now(),
          };
        }
        handleApiError(error, "保存");
        return { ...prevState, success: false, timestamp: Date.now() };
      }
    },
    { fieldErrors: {}, success: false, timestamp: 0 }
  );

  const handleAddPet = () => {
    setEditingPet(null);
    setPetModalOpen(true);
  };

  const handleEditPet = (pet: PetFormData) => {
    setEditingPet(pet);
    setPetModalOpen(true);
  };

  const handleDeletePet = (petId: string) => {
    const target = pets.find(p => p.id === petId);
    if (target?.isPending) {
      setPets(prev => prev.filter(p => p.id !== petId));
      toast.success("ペットを削除しました");
      return;
    }

    petMutations?.deletePetMutate(petId, {
      onSuccess: () => {
        setPets(prev => prev.filter((pet) => pet.id !== petId));
        toast.success("ペットを削除しました");
      },
      onError: (error: unknown) => {
        handleApiError(error, "削除");
      },
    });
  };

  const handleSavePet = (petData: PetFormData) => {
    if (editingPet) {
      if (editingPet.isPending) {
        setPets(prev =>
          prev.map(p =>
            p.id === editingPet.id
              ? { ...petData, id: editingPet.id, isPending: true }
              : p
          )
        );
        return;
      }

      const updateRequest = transformUpdatePetRequest({
        petNumber: petData.petNumber,
        name: petData.petName,
        petNameKana: petData.petNameKana,
        animalSpeciesId: petData.animalSpeciesId,
        gender: petData.gender,
        birthDate: petData.birthDate,
        breed: petData.breed,
        color: petData.color,
        weight: petData.weight,
        food: petData.food,
        environment: petData.environment,
        neuteredDate: petData.neuteredDate,
        acquisitionType: petData.acquisitionType,
        dangerLevel: petData.dangerLevel,
        status: PET_STATUS_REVERSE_MAP[petData.status],
        insuranceId: petData.insuranceId,
        remarks: petData.remarks,
      });

      petMutations?.updatePetMutate(
        { id: editingPet.id, req: updateRequest },
        {
          onSuccess: () => {
            setPets(prev =>
              prev.map((pet) =>
                pet.id === editingPet.id ? { ...pet, ...petData } : pet
              )
            );
            toast.success("ペット情報を更新しました");
          },
          onError: (error: unknown) => {
            handleApiError(error, "更新");
          },
        }
      );
    } else {
      if (!id) {
        if (!petData.animalSpeciesId) {
          toast.error("動物種を選択してください");
          return;
        }
        const tempId = `temp-${Date.now()}`;
        setPets(prev => [...prev, { ...petData, id: tempId, isPending: true }]);
        return;
      }

      if (!petData.animalSpeciesId) {
        toast.error("動物種を選択してください");
        return;
      }

      const createRequest = transformCreatePetRequest({
        ownerId: id,
        name: petData.petName || "",
        animalSpeciesId: petData.animalSpeciesId,
        petNumber: petData.petNumber,
        petNameKana: petData.petNameKana,
        breed: petData.breed,
        color: petData.color,
        gender: petData.gender,
        birthDate: petData.birthDate,
        weight: petData.weight,
        food: petData.food,
        environment: petData.environment,
        neuteredDate: petData.neuteredDate,
        acquisitionType: petData.acquisitionType,
        dangerLevel: petData.dangerLevel,
        status: PET_STATUS_REVERSE_MAP[petData.status],
        insuranceId: petData.insuranceId,
        remarks: petData.remarks,
      });

      petMutations?.createPetMutate(createRequest, {
        onSuccess: (newPetData: Pet) => {
          const newPet: PetFormData = {
            id: newPetData.id,
            petNumber: newPetData.petNumber || "",
            petName: newPetData.name,
            petNameKana: newPetData.petNameKana || "",
            status: newPetData.status || "生存",
            species: newPetData.species,
            animalSpeciesId: newPetData.animalSpeciesId,
            gender: newPetData.gender || "",
            birthDate: newPetData.birthDate || "",
            color: newPetData.color || "",
            weight: newPetData.weight || "",
            food: newPetData.food || "",
            environment: newPetData.environment || "",
            neuteredDate: newPetData.neuteredDate || "",
            acquisitionType: (newPetData.acquisitionType as PetFormData["acquisitionType"]) || "購入",
            dangerLevel: (newPetData.dangerLevel as PetFormData["dangerLevel"]) || "低",
            remarks: newPetData.remarks || "",
            breed: newPetData.breed,
            insuranceId: newPetData.insuranceId,
            insuranceName: undefined,
            insuranceDetails: newPetData.insuranceDetails,
          };
          setPets(prev => [...prev, newPet]);
          toast.success("ペットを追加しました");
        },
        onError: (error: unknown) => {
          handleApiError(error, "追加");
        },
      });
    }
  };

  // rerender-functional-setstate: setFieldErrors は stable setter のため useCallback で安定化可能
  // → OwnerInfoSection memo の onClearError prop を stable に保つための前提条件
  const clearFieldError = useCallback((field: string) => {
    // フォームのアクション状態で管理されているエラーをクリアする場合は、
    // コンポーネント側で状態を意識する必要がある。
    // ここでは formState.fieldErrors は read-only。
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
    formAction,
    formState,
    fieldErrors: formState.fieldErrors,
    clearFieldError,
    createdOwnerId: formState.createdOwnerId,
  };
}
