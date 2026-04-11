/**
 * Owner UI types (camelCase, string IDs) + API Request types (snake_case)
 * Backend model: {@link import("./generated/models").Owner}
 */
import type { Owner as BackendOwner } from "./generated/models";
import type { Pet } from "./index";

/** UI-facing Owner type (camelCase, string IDs — post-transform) */
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

/**
 * 飼主作成リクエスト — models.ts Owner から導出（snake_case, API-facing）
 * @see {@link BackendOwner}
 */
export type CreateOwnerRequest = Partial<
  Omit<BackendOwner, 'id' | 'clinic_id' | 'created_at' | 'updated_at' | 'pets' | 'birth_date'>
> & {
  name: string;
  birth_date?: string;
};

/**
 * 飼主更新リクエスト — CreateOwnerRequest の全フィールド optional
 * @see {@link BackendOwner}
 */
export type UpdateOwnerRequest = Partial<
  Omit<BackendOwner, 'id' | 'clinic_id' | 'created_at' | 'updated_at' | 'pets'>
>;
