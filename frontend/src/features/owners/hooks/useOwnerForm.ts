// React/Framework
import { useState, useEffect } from "react";
import { useNavigate } from "react-router";

// External
import { toast } from "sonner";

// Internal
import { getOwner, createOwner, updateOwner, deleteOwner } from "../api";
import { useCreatePet, useUpdatePet, useDeletePet } from "@/features/pets/api";
import { transformCreatePetRequest, transformUpdatePetRequest } from "@/features/pets/api/transforms";
import { CreateOwnerRequest, UpdateOwnerRequest } from "@/types/owner";
import { OwnerData, PetInfo, MembershipType, INSURANCE_COMPANY_VALUES, PET_INSURANCE_RATIO_VALUES } from "../types";
import { isOneOf } from "@/lib/type-utils";

export function useOwnerForm(id?: string) {
  const navigate = useNavigate();
  const isEdit = !!id;
  // Initialize isLoading based on id presence (for async fetch on mount)
  const [isLoading, setIsLoading] = useState(!!id);

  // Pet API mutations
  const { mutate: createPetMutate } = useCreatePet();
  const { mutate: updatePetMutate } = useUpdatePet();
  const { mutate: deletePetMutate } = useDeletePet();

  const [ownerData, setOwnerData] = useState<OwnerData>({
    ownerId: "",
    postalCode: "",
    company: "",
    membershipType: "非会員" as MembershipType,
    ownerName: "",
    address1: "",
    postalNumber: "",
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
  });

  const [pets, setPets] = useState<PetInfo[]>([]);
  const [petModalOpen, setPetModalOpen] = useState(false);
  const [editingPet, setEditingPet] = useState<PetInfo | null>(null);

  useEffect(() => {
    const fetchOwner = async () => {
      if (!id) return;

      try {
        const owner = await getOwner(id);

        // Map backend Owner to frontend OwnerData
        // Note: Backend has simplified address structure
        setOwnerData({
          ownerId: owner.id,
          postalCode: "", // Not separately stored in backend
          company: "", // Not stored
          membershipType: "会員" as MembershipType, // Default or derived
          ownerName: owner.name,
          address1: owner.address || "",
          postalNumber: "",
          ownerNameKana: owner.name_kana || "",
          address2: "",
          homeAddress1: "",
          homeAddress2: "",
          isDangerous: false,
          birthDate: "",
          email: owner.email || "",
          phone: owner.phone || "",
          companyPhone: "",
          remarks: owner.notes || "",
        });

        // Map pets if available
        if (owner.pets) {
          setPets(
            owner.pets.map((p) => {
              const petInfo: PetInfo = {
                id: p.id,
                petNumber: p.petNumber || "",
                petName: p.name,
                petNameKana: "", // Not in Pet interface from backend yet?
                status: p.status || "生存",
                species: p.species,
                gender: p.gender || "",
                birthDate: p.birthDate || "",
                color: "", // Not in Pet interface
                weight: p.weight || "",
                environment: p.environment || "",
                remarks: p.remarks || "",
              };
              if (p.insuranceName && isOneOf(p.insuranceName, INSURANCE_COMPANY_VALUES)) {
                petInfo.insuranceName = p.insuranceName;
              }
              if (p.insuranceDetails && isOneOf(p.insuranceDetails, PET_INSURANCE_RATIO_VALUES)) {
                petInfo.insuranceDetails = p.insuranceDetails;
              }
              return petInfo;
            })
          );
        }
      } catch {
        toast.error("飼主情報の取得に失敗しました");
        navigate("/owners");
      } finally {
        setIsLoading(false);
      }
    };

    if (isEdit && id) {
      fetchOwner();
    }
  }, [isEdit, id, navigate]);

  const handleAddPet = () => {
    setEditingPet(null);
    setPetModalOpen(true);
  };

  const handleEditPet = (pet: PetInfo) => {
    setEditingPet(pet);
    setPetModalOpen(true);
  };

  const handleDeletePet = (petId: string) => {
    deletePetMutate(petId, {
      onSuccess: () => {
        setPets(pets.filter((pet) => pet.id !== petId));
        toast.success("ペットを削除しました");
      },
      onError: () => {
        toast.error("ペットの削除に失敗しました");
      },
    });
  };

  const handleSavePet = (petData: Partial<PetInfo>) => {
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
            setPets(
              pets.map((pet) =>
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
          // Transform backend response to frontend PetInfo
          const newPet: PetInfo = {
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
          setPets([...pets, newPet]);
          toast.success("ペットを追加しました");
        },
        onError: () => {
          toast.error("ペットの追加に失敗しました");
        },
      });
    }
  };

  const handleSave = async (): Promise<boolean> => {
    // Validate required fields
    if (!ownerData.ownerName.trim()) {
      toast.error("飼主名を入力してください");
      return false;
    }
    if (!ownerData.ownerNameKana.trim()) {
      toast.error("飼主名（かな）を入力してください");
      return false;
    }
    if (!ownerData.phone.trim()) {
      toast.error("電話番号を入力してください");
      return false;
    }

    setIsLoading(true);
    try {
      // Combine address fields
      const fullAddress = [
        ownerData.postalCode,
        ownerData.address1,
        ownerData.address2,
      ]
        .filter(Boolean)
        .join(" ");

      const commonData = {
        name: ownerData.ownerName,
        name_kana: ownerData.ownerNameKana,
        phone: ownerData.phone,
        email: ownerData.email,
        address: fullAddress,
        notes: ownerData.remarks,
      };

      if (isEdit && id) {
        const updateData: UpdateOwnerRequest = commonData;
        await updateOwner(id, updateData);
        toast.success("飼主情報を更新しました");
      } else {
        const createData: CreateOwnerRequest = commonData;
        await createOwner(createData);
        toast.success("飼主情報を登録しました");
      }

      return true;
    } catch {
      toast.error("保存に失敗しました");
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async (ownerId: string) => {
    try {
      await deleteOwner(ownerId);
      toast.success("飼主を削除しました");
      navigate("/owners");
    } catch {
      toast.error("削除に失敗しました");
    }
  };

  return {
    isEdit,
    isLoading,
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
    handleDelete,
  };
}
