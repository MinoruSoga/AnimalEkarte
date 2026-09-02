import { transformCreatePetRequest, PET_STATUS_REVERSE_MAP } from "@/lib/transforms/pet";
import type {
  CreateOwnerPetRequest,
  CreateOwnerRequest,
  UpdateOwnerRequest,
  Owner,
} from "@/types/owner";
import type { OwnerData, PetFormData, MembershipTypeLabel } from "../types";

type CreateOwnerPetRequestWithDangerReason = CreateOwnerPetRequest & {
  danger_reason?: string;
};

export const MEMBERSHIP_TYPE_TO_API: Record<string, string> = {
  "非会員": "non_member",
  "会員": "member",
  "退亡者": "deceased",
  "他診/準": "transferred",
};

export const DEFAULT_OWNER_DATA: OwnerData = {
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

export const DENIED_MUTATION_PERMISSIONS: Readonly<OwnerMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
};

export function mapOwnerToFormData(owner: Owner): OwnerData {
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

export function mapOwnerPetsToFormData(owner: Owner): PetFormData[] {
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

export function mapPendingPetToCreateRequest(
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

export function validateOwnerForm(ownerData: OwnerData): Record<string, string> {
  const errors: Record<string, string> = {};
  if (!ownerData.ownerName.trim()) errors.ownerName = "飼主名を入力してください";
  if (!ownerData.ownerNameKana.trim()) errors.ownerNameKana = "飼主名よみを入力してください";
  if (!ownerData.phone.trim()) {
    errors.phone = "電話番号を入力してください";
  } else if (!/^[\d\-+() ]+$/.test(ownerData.phone.trim())) {
    errors.phone = "電話番号の形式が正しくありません（数字・ハイフンのみ）";
  }
  if (ownerData.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(ownerData.email)) {
    errors.email = "メールアドレスの形式が正しくありません";
  }
  const postalPattern = /^\d{3}-?\d{4}$/;
  if (ownerData.postalCode && !postalPattern.test(ownerData.postalCode.trim())) {
    errors.postalCode = "郵便番号の形式が正しくありません（例: 123-4567）";
  }
  if (ownerData.homePostalCode && !postalPattern.test(ownerData.homePostalCode.trim())) {
    errors.homePostalCode = "郵便番号の形式が正しくありません（例: 123-4567）";
  }
  if (ownerData.discountRate != null && (ownerData.discountRate < 0 || ownerData.discountRate > 100)) {
    errors.discountRate = "値引率は0〜100の範囲で入力してください";
  }
  return errors;
}

export function buildOwnerRequestPayload(ownerData: OwnerData) {
  return {
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
}

export function buildUpdateOwnerRequest(ownerData: OwnerData): UpdateOwnerRequest {
  return {
    ...buildOwnerRequestPayload(ownerData),
    birth_date: ownerData.birthDate || null,
  };
}

export function buildCreateOwnerRequest(
  ownerData: OwnerData,
  pets: PetFormData[],
): CreateOwnerRequest {
  const pendingPets = pets.filter(
    (pet): pet is PetFormData & { animalSpeciesId: string } =>
      pet.isPending === true && Boolean(pet.animalSpeciesId),
  );
  return {
    ...buildOwnerRequestPayload(ownerData),
    birth_date: ownerData.birthDate || undefined,
    clinic_id: ownerData.clinicId ? Number(ownerData.clinicId) : undefined,
    pets: pendingPets.map(mapPendingPetToCreateRequest),
  };
}

export function resolveCreatedOwnerClinicId(
  ownerData: OwnerData,
  newOwner: { clinicId?: string },
  createData: CreateOwnerRequest,
): string | undefined {
  return (
    ownerData.clinicId
    ?? newOwner.clinicId
    ?? (createData.clinic_id != null ? String(createData.clinic_id) : undefined)
  );
}
