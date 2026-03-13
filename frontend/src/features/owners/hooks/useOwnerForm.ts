// React/Framework
import { useState, useTransition } from "react";

// External
import { toast } from "sonner";

// Internal - direct imports (bundle-barrel-imports 準拠)
import { createOwner } from "../api/create-owner";
import { updateOwner } from "../api/update-owner";
import { useCreatePet } from "@/features/pets/api/create-pet";
import { useUpdatePet } from "@/features/pets/api/update-pet";
import { useDeletePet } from "@/features/pets/api/delete-pet";
import { createPet } from "@/features/pets/api/create-pet";
import { transformCreatePetRequest, transformUpdatePetRequest } from "@/features/pets/api/transforms";
import type { CreateOwnerRequest, UpdateOwnerRequest } from "@/types/owner";
import type { Owner } from "@/types/owner";
import type { OwnerData, PetFormData, MembershipType } from "../types";

const PET_STATUS_TO_API: Record<string, "alive" | "deceased" | undefined> = {
  "生存": "alive",
  "死亡": "deceased",
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
    petNameKana: "",
    status: backendPet.status || "生存",
    species: backendPet.species,
    animalSpeciesId: backendPet.animalSpeciesId,
    gender: backendPet.gender || "",
    birthDate: backendPet.birthDate || "",
    color: "",
    weight: backendPet.weight || "",
    environment: backendPet.environment || "",
    remarks: backendPet.remarks || "",
    breed: backendPet.breed,
    insuranceId: backendPet.insuranceId,
    insuranceName: undefined,
    insuranceDetails: backendPet.insuranceDetails,
    microchipId: backendPet.microchipId,
  }));
}

export function useOwnerForm(id?: string, initialOwner?: Owner) {
  const isEdit = !!id;
  // isSavePending: true while owner create/update API call is in-flight
  const [isSavePending, startSaveTransition] = useTransition();

  // Pet API mutations（既存飼主へのペット即時追加・更新・削除）
  const { mutate: createPetMutate } = useCreatePet();
  const { mutate: updatePetMutate } = useUpdatePet();
  const { mutate: deletePetMutate } = useDeletePet();

  const [ownerData, setOwnerData] = useState<OwnerData>(
    () => initialOwner ? mapOwnerToFormData(initialOwner) : DEFAULT_OWNER_DATA
  );

  const [pets, setPets] = useState<PetFormData[]>(
    () => initialOwner ? mapOwnerPetsToFormData(initialOwner) : []
  );
  const [petModalOpen, setPetModalOpen] = useState(false);
  const [editingPet, setEditingPet] = useState<PetFormData | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const handleAddPet = () => {
    setEditingPet(null);
    setPetModalOpen(true);
  };

  const handleEditPet = (pet: PetFormData) => {
    setEditingPet(pet);
    setPetModalOpen(true);
  };

  const handleDeletePet = (petId: string) => {
    // isPending: 新規飼主モードでローカルに追加したペット → API呼び出し不要
    const target = pets.find(p => p.id === petId);
    if (target?.isPending) {
      setPets(prev => prev.filter(p => p.id !== petId));
      toast.success("ペットを削除しました");
      return;
    }

    deletePetMutate(petId, {
      onSuccess: () => {
        setPets(prev => prev.filter((pet) => pet.id !== petId));
        toast.success("ペットを削除しました");
      },
      onError: () => {
        toast.error("ペットの削除に失敗しました");
      },
    });
  };

  const handleSavePet = (petData: PetFormData) => {
    if (editingPet) {
      // ─── 編集 ──────────────────────────────────────────
      if (editingPet.isPending) {
        // pending ペット（新規飼主モード）: ローカル state のみ更新
        setPets(prev =>
          prev.map(p =>
            p.id === editingPet.id
              ? { ...petData, id: editingPet.id, isPending: true }
              : p
          )
        );
        return;
      }

      // 既存 API 保存済みペット: 即時 PATCH
      const updateRequest = transformUpdatePetRequest({
        petNumber: petData.petNumber,
        name: petData.petName,
        animalSpeciesId: petData.animalSpeciesId,
        gender: petData.gender,
        birthDate: petData.birthDate,
        breed: petData.breed,
        weight: petData.weight,
        environment: petData.environment,
        status: PET_STATUS_TO_API[petData.status],
        insuranceId: petData.insuranceId,
        remarks: petData.remarks,
      });

      updatePetMutate(
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
          onError: () => {
            toast.error("ペット情報の更新に失敗しました");
          },
        }
      );
    } else {
      // ─── 新規追加 ──────────────────────────────────────
      if (!id) {
        // 新規飼主モード: API呼び出しせずローカル state に追加
        // 飼主登録（handleSave）時に Promise.all で一括送信する
        if (!petData.animalSpeciesId) {
          toast.error("動物種を選択してください");
          return;
        }
        const tempId = `temp-${Date.now()}`;
        setPets(prev => [...prev, { ...petData, id: tempId, isPending: true }]);
        return;
      }

      // 既存飼主へのペット追加: 即時 POST
      if (!petData.animalSpeciesId) {
        toast.error("動物種を選択してください");
        return;
      }

      const createRequest = transformCreatePetRequest({
        ownerId: id,
        name: petData.petName || "",
        animalSpeciesId: petData.animalSpeciesId,
        petNumber: petData.petNumber,
        breed: petData.breed,
        gender: petData.gender,
        birthDate: petData.birthDate,
        weight: petData.weight,
        environment: petData.environment,
        status: PET_STATUS_TO_API[petData.status],
        insuranceId: petData.insuranceId,
        remarks: petData.remarks,
      });

      createPetMutate(createRequest, {
        onSuccess: (newPetData) => {
          const newPet: PetFormData = {
            id: newPetData.id,
            petNumber: newPetData.petNumber || "",
            petName: newPetData.name,
            status: newPetData.status || "生存",
            species: newPetData.species,
            animalSpeciesId: newPetData.animalSpeciesId,
            gender: newPetData.gender || "",
            birthDate: newPetData.birthDate || "",
            color: "",
            weight: newPetData.weight || "",
            environment: newPetData.environment || "",
            remarks: newPetData.remarks || "",
            breed: newPetData.breed,
            insuranceId: newPetData.insuranceId,
            insuranceName: undefined,
            insuranceDetails: newPetData.insuranceDetails,
          };
          setPets(prev => [...prev, newPet]);
          toast.success("ペットを追加しました");
        },
        onError: () => {
          toast.error("ペットの追加に失敗しました");
        },
      });
    }
  };

  const clearFieldError = (field: string) => {
    setFieldErrors((prev) => {
      const next = { ...prev };
      delete next[field];
      return next;
    });
  };

  const handleSave = (onSuccess?: () => void): void => {
    // バリデーション（同期 — transition 前に実行）
    const errors: Record<string, string> = {};
    if (!ownerData.ownerName.trim()) errors.ownerName = "飼主名を入力してください";
    if (!ownerData.ownerNameKana.trim()) errors.ownerNameKana = "飼主名（カナ）を入力してください";
    if (!ownerData.phone.trim()) errors.phone = "電話番号を入力してください";

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      toast.error("必須項目が未入力です");
      return;
    }
    setFieldErrors({});

    startSaveTransition(async () => {
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
          membership_type: ownerData.membershipType,
        };

        if (isEdit && id) {
          const updateData: UpdateOwnerRequest = ownerRequestPayload;
          await updateOwner(id, updateData);
          toast.success("飼主情報を更新しました");
        } else {
          const createData: CreateOwnerRequest = {
            ...ownerRequestPayload,
            owner_name: ownerData.ownerName,
          };
          const newOwner = await createOwner(createData);

          // pending ペットを飼主 ID を付けて一括 POST
          const pendingPets = pets.filter(p => p.isPending && p.animalSpeciesId);
          if (pendingPets.length > 0) {
            await Promise.all(
              pendingPets.map(pet =>
                createPet(
                  transformCreatePetRequest({
                    ownerId: newOwner.id,
                    name: pet.petName || "",
                    animalSpeciesId: pet.animalSpeciesId!,
                    petNumber: pet.petNumber,
                    breed: pet.breed,
                    gender: pet.gender,
                    birthDate: pet.birthDate,
                    weight: pet.weight,
                    environment: pet.environment,
                    status: PET_STATUS_TO_API[pet.status],
                    insuranceId: pet.insuranceId,
                    remarks: pet.remarks,
                  })
                )
              )
            );
          }

          toast.success("飼主情報を登録しました");
        }

        onSuccess?.();
      } catch {
        toast.error("保存に失敗しました");
      }
    });
  };

  return {
    isEdit,
    isLoading: isSavePending,
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
    handleSave,
    fieldErrors,
    clearFieldError,
  };
}
