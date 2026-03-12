// React/Framework
import { useState, useTransition } from "react";

// External
import { toast } from "sonner";

// Internal
import { createOwner, updateOwner } from "../api";
import { useCreatePet, useUpdatePet, useDeletePet } from "@/features/pets/api";
import { transformCreatePetRequest, transformUpdatePetRequest } from "@/features/pets/api/transforms";
import { CreateOwnerRequest, UpdateOwnerRequest } from "@/types/owner";
import type { Owner } from "@/types/owner";
import { OwnerData, PetFormData, MembershipType, INSURANCE_COMPANY_VALUES, PET_INSURANCE_RATIO_VALUES } from "../types";
import { isOneOf } from "@/lib/type-utils";

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
  return owner.pets.map((backendPet) => {
    const petFormData: PetFormData = {
      id: backendPet.id,
      petNumber: backendPet.petNumber || "",
      petName: backendPet.name,
      petNameKana: "", // Not in Pet interface from backend yet?
      status: backendPet.status || "生存",
      species: backendPet.species,
      gender: backendPet.gender || "",
      birthDate: backendPet.birthDate || "",
      color: "", // Not in Pet interface
      weight: backendPet.weight || "",
      environment: backendPet.environment || "",
      remarks: backendPet.remarks || "",
    };
    if (backendPet.insuranceName && isOneOf(backendPet.insuranceName, INSURANCE_COMPANY_VALUES)) {
      petFormData.insuranceName = backendPet.insuranceName;
    }
    if (backendPet.insuranceDetails && isOneOf(backendPet.insuranceDetails, PET_INSURANCE_RATIO_VALUES)) {
      petFormData.insuranceDetails = backendPet.insuranceDetails;
    }
    return petFormData;
  });
}

export function useOwnerForm(id?: string, initialOwner?: Owner) {
  const isEdit = !!id;
  // isSavePending: true while owner create/update API call is in-flight
  const [isSavePending, startSaveTransition] = useTransition();

  // Pet API mutations
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
      // Update existing pet via API
      const updateRequest = transformUpdatePetRequest({
        petNumber: petData.petNumber,
        name: petData.petName,
        species: petData.species,
        gender: petData.gender,
        birthDate: petData.birthDate,
        breed: petData.breed,
        weight: petData.weight,
        environment: petData.environment,
        status: petData.status,
        insuranceName: petData.insuranceName,
        insuranceDetails: petData.insuranceDetails,
        notes: petData.remarks,
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
      // Add new pet via API (requires owner_id)
      if (!id) {
        toast.error("飼主IDが見つかりません");
        return;
      }

      const createRequest = transformCreatePetRequest({
        ownerId: id,
        name: petData.petName || "",
        species: petData.species || "",
        petNumber: petData.petNumber,
        breed: petData.breed,
        gender: petData.gender,
        birthDate: petData.birthDate,
        weight: petData.weight,
        environment: petData.environment,
        status: petData.status,
        insuranceName: petData.insuranceName,
        insuranceDetails: petData.insuranceDetails,
        notes: petData.remarks,
      });

      createPetMutate(createRequest, {
        onSuccess: (newPetData) => {
          // Transform backend response to frontend PetFormData
          const newPet: PetFormData = {
            id: newPetData.id,
            petNumber: newPetData.petNumber || "",
            petName: newPetData.name,
            status: newPetData.status || "生存",
            species: newPetData.species,
            gender: newPetData.gender || "",
            birthDate: newPetData.birthDate || "",
            color: "", // Not in backend response
            weight: newPetData.weight || "",
            environment: newPetData.environment || "",
            remarks: newPetData.remarks || "",
            breed: newPetData.breed,
            insuranceName: newPetData.insuranceName && isOneOf(newPetData.insuranceName, INSURANCE_COMPANY_VALUES) ? newPetData.insuranceName : undefined,
            insuranceDetails: newPetData.insuranceDetails && isOneOf(newPetData.insuranceDetails, PET_INSURANCE_RATIO_VALUES) ? newPetData.insuranceDetails : undefined,
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
    // Validate required fields (synchronous — runs before transition)
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
          await createOwner(createData);
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
