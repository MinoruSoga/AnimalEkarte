import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner@2.0.3";

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
    if (isEdit && id) {
      // Mock Data Loading
      if (id === "30042") {
        setOwnerData({
          ownerId: "30042",
          postalCode: "150-0001",
          company: "サンプル株式会社",
          membershipType: "会員",
          ownerName: "林　文明",
          address1: "東京都渋谷区",
          postalNumber: "150-0001",
          ownerNameKana: "ハヤシ　フミアキ",
          address2: "神宮前1-2-3",
          homeAddress1: "",
          isDangerous: false,
          birthDate: "1980-05-15",
          email: "hayashi@example.com",
          homeAddress2: "",
          remarks: "定期検診を希望",
          phone: "090-1234-5678",
          companyPhone: "03-1234-5678",
        });
        setPets([
          {
            id: "1",
            petNumber: "30042-008",
            petName: "Iris",
            status: "死亡",
            species: "犬",
            gender: "雄",
            birthDate: "2015/04/14",
            color: "茶色",
            weight: "26.5kg",
            environment: "室内(散歩する)",
            remarks: "",
          },
          {
            id: "2",
            petNumber: "30042-009",
            petName: "Max",
            status: "生存",
            species: "犬",
            gender: "雄",
            birthDate: "2018/06/20",
            color: "ゴールデン",
            weight: "15.2kg",
            environment: "室内(散歩する)",
            remarks: "",
          },
        ]);
      } else if (id === "30043") {
         setOwnerData({
            ownerId: "30043",
            postalCode: "160-0022",
            company: "",
            membershipType: "非会員",
            ownerName: "田中　花子",
            address1: "東京都新宿区",
            postalNumber: "160-0022",
            ownerNameKana: "タナカ　ハナコ",
            address2: "新宿1-1-1",
            homeAddress1: "",
            isDangerous: false,
            birthDate: "1985-03-20",
            email: "tanaka@example.com",
            homeAddress2: "",
            remarks: "",
            phone: "080-9876-5432",
            companyPhone: "",
         });
         setPets([
            {
                id: "3",
                petNumber: "30043-001",
                petName: "ミケ",
                status: "生存",
                species: "猫",
                gender: "雌",
                birthDate: "2020/03/10",
                color: "三毛",
                weight: "4.2kg",
                environment: "室内",
                remarks: "",
            },
         ]);
      } else {
          setOwnerData(prev => ({...prev, ownerId: id}));
      }
    }
  }, [isEdit, id]);

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

  const handleSavePet = (petData: any) => {
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
        petNumber: petData.petNumber,
        petName: petData.petName,
        status: "生存", // Default
        species: petData.species,
        gender: petData.gender,
        birthDate: petData.birthDate,
        color: petData.color,
        weight: "",
        environment: "",
        remarks: petData.remarks,
        ...petData
      };
      setPets([...pets, newPet]);
    }
  };

  const handleSave = () => {
    toast.success(isEdit ? "飼主情報を更新しました" : "飼主情報を登録しました");
    setTimeout(() => {
        navigate("/owners");
    }, 800);
  };

  return {
    isEdit,
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
