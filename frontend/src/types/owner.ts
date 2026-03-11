import { Pet } from "./index";

export interface Owner {
  id: string;
  ownerName: string;
  ownerNameKana?: string;
  company: string;
  postalCode: string;
  address1: string;
  address2: string;
  homePostalCode: string;
  homeAddress1: string;
  homeAddress2: string;
  birthDate?: string;
  phone: string;
  companyPhone: string;
  email: string;
  remarks: string;
  isDangerous: boolean;
  discountRate: number;
  membershipType: string;
  createdAt: string;
  updatedAt: string;
  pets?: Pet[];
}

export interface CreateOwnerRequest {
  owner_name: string;
  owner_name_kana?: string;
  company?: string;
  postal_code?: string;
  address1?: string;
  address2?: string;
  home_postal_code?: string;
  home_address1?: string;
  home_address2?: string;
  phone?: string;
  company_phone?: string;
  email?: string;
  remarks?: string;
  is_dangerous?: boolean;
  discount_rate?: number;
  membership_type?: string;
}

export interface UpdateOwnerRequest {
  owner_name?: string;
  owner_name_kana?: string;
  company?: string;
  postal_code?: string;
  address1?: string;
  address2?: string;
  home_postal_code?: string;
  home_address1?: string;
  home_address2?: string;
  phone?: string;
  company_phone?: string;
  email?: string;
  remarks?: string;
  is_dangerous?: boolean;
  discount_rate?: number;
  membership_type?: string;
}
