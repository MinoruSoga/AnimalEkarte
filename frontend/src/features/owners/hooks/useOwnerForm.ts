// React/Framework
import { useState, useEffect } from "react";
import { useNavigate } from "react-router";

// External
import { toast } from "sonner";

// Internal
import { getOwner, createOwner, updateOwner } from "../api";
import { CreateOwnerRequest, UpdateOwnerRequest } from "@/types/owner";

export interface PetInfo {
  id: string;
  petNumber: string;
  petName: string;
  petNameKana?: string;
  status: string;
  species: string;
  gender: string;
  birthDate: string;
  color: string;
  weight: string;
  environment: string;
  remarks: string;
  // Extra fields for modal
  breed?: string;
  neuteredDate?: string;
  acquisitionType?: string;
  dangerLevel?: string;
  food?: string;
  insuranceName?: string;
  insuranceDetails?: string;
}

export interface OwnerData {
  ownerId: string;
  postalCode: string;
  company: string;
  membershipType: string;
  ownerName: string;
  address1: string;
  postalNumber: string;
  ownerNameKana: string;
  address2: string;
  homeAddress1: string;
  isDangerous: boolean;
  birthDate: string;
  email: string;
  homeAddress2: string;
  remarks: string;
  phone: string;
  companyPhone: string;
}

export function useOwnerForm(id?: string) {
  const navigate = useNavigate();
  const isEdit = !!id;
  // Initialize isLoading based on id presence (for async fetch on mount)
  const [isLoading, setIsLoading] = useState(!!id);

  const [ownerData, setOwnerData] = useState<OwnerData>({
    ownerId: "",
    postalCode: "",
    company: "",
    membershipType: "非会員",
    ownerName: "",
    address1: "",
    postalNumber: "",
    ownerNameKana: "",
    address2: "",
    homeAddress1: "",
    isDangerous: false,
    birthDate: "",
    email: "",
    homeAddress2: "",
    remarks: "",
    phone: "",
    companyPhone: "",
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
          membershipType: "会員", // Default or derived
          ownerName: owner.name,
          address1: owner.address || "",
          postalNumber: "",
          ownerNameKana: owner.name_kana || "",
          address2: "",
          homeAddress1: "",
          isDangerous: false,
          birthDate: "",
          email: owner.email || "",
          homeAddress2: "",
          remarks: owner.notes || "",
          phone: owner.phone || "",
          companyPhone: "",
        });

        // Map pets if available
        if (owner.pets) {
          setPets(owner.pets.map(p => ({
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
            insuranceName: p.insuranceName,
            insuranceDetails: p.insuranceDetails
          })));
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

  const handleDeletePet = (id: string) => {
    setPets(pets.filter((pet) => pet.id !== id));
  };

  const handleSavePet = (petData: Partial<PetInfo>) => {
    if (editingPet) {
      // Update existing pet
      setPets(
        pets.map((pet) =>
          pet.id === editingPet.id ? { ...pet, ...petData } : pet
        )
      );
    } else {
      // Add new pet
      const newPet: PetInfo = {
        id: Date.now().toString(),
        petNumber: petData.petNumber || "",
        petName: petData.petName || "",
        status: petData.status || "生存",
        species: petData.species || "",
        gender: petData.gender || "",
        birthDate: petData.birthDate || "",
        color: petData.color || "",
        weight: petData.weight || "",
        environment: petData.environment || "",
        remarks: petData.remarks || "",
        ...petData
      };
      setPets([...pets, newPet]);
    }
  };

  const handleSave = async () => {
    setIsLoading(true);
    try {
      // Combine address fields
      const fullAddress = [
        ownerData.postalCode,
        ownerData.address1,
        ownerData.address2
      ].filter(Boolean).join(" ");

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
      
      // Navigate back after short delay
      setTimeout(() => {
          navigate("/owners");
      }, 800);

    } catch {
      toast.error("保存に失敗しました");
    } finally {
      setIsLoading(false);
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
    handleSave
  };
}
