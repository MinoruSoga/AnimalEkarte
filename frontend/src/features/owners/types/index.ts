// Pet-related type constants and definitions
export const PET_GENDER_VALUES = ["雄", "雌", "不明"] as const;
export type PetGender = (typeof PET_GENDER_VALUES)[number];

export const ACQUISITION_TYPE_VALUES = ["購入", "譲渡", "保護", "その他"] as const;
export type AcquisitionType = (typeof ACQUISITION_TYPE_VALUES)[number];

export const DANGER_LEVEL_VALUES = ["低", "中", "高"] as const;
export type DangerLevel = (typeof DANGER_LEVEL_VALUES)[number];

export const INSURANCE_COMPANY_VALUES = [
  "アニコム",
  "アイペット",
  "ペット＆ファミリー",
  "楽天ペット保険",
  "アクサダイレクト",
  "SBIいきいき少短",
  "FPC",
  "その他",
] as const;
export type InsuranceCompany = (typeof INSURANCE_COMPANY_VALUES)[number];

export const PET_INSURANCE_RATIO_VALUES = ["50%", "70%", "90%", "100%", "その他"] as const;
export type PetInsuranceRatio = (typeof PET_INSURANCE_RATIO_VALUES)[number];

// Owner-related type constants and definitions
export const MEMBERSHIP_TYPE_VALUES = ["非会員", "会員", "退亡者", "他診/準"] as const;
export type MembershipType = (typeof MEMBERSHIP_TYPE_VALUES)[number];

// Pet data (for PetEditModal)
export interface PetEditModalData {
  petName: string;
  petNameKana: string;
  species: string;
  gender: PetGender;
  birthDate: string;
  breed: string;
  color: string;
  weight: string;
  neuteredDate: string;
  acquisitionType: AcquisitionType;
  dangerLevel: DangerLevel;
  food: string;
  environment: string;
  insuranceName: InsuranceCompany;
  insuranceDetails: string;
  remarks: string;
}

export interface PetFormData {
  id: string;
  petNumber: string;
  petName: string;
  petNameKana?: string;
  status: string;
  species: string;
  gender: string;
  birthDate: string;
  breed?: string;
  color: string;
  weight: string;
  neuteredDate?: string;
  acquisitionType?: AcquisitionType;
  dangerLevel?: DangerLevel;
  food?: string;
  environment: string;
  remarks: string;
  insuranceName?: InsuranceCompany;
  insuranceDetails?: string;
}

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
  breed?: string;
  neuteredDate?: string;
  acquisitionType?: AcquisitionType;
  dangerLevel?: DangerLevel;
  food?: string;
  insuranceName?: InsuranceCompany;
  insuranceDetails?: string;
}

// Owner data
export interface OwnerData {
  ownerId: string;
  postalCode: string;
  company: string;
  membershipType: MembershipType;
  ownerName: string;
  address1: string;
  postalNumber: string;
  ownerNameKana: string;
  address2: string;
  homeAddress1: string;
  homeAddress2: string;
  isDangerous: boolean;
  birthDate: string;
  email: string;
  phone: string;
  companyPhone: string;
  remarks: string;
  discountRate?: number;
  homePostalCode?: string;
}
