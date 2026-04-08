/**
 * Backend ↔ Frontend 型変換
 * バックエンドは snake_case、フロントエンドは camelCase
 */
import { z } from "zod";
import type { AuthUser, AuthClinic } from "../types";

const clinicMembershipSchema = z.object({
  clinic_id: z.string(),
  clinic_name: z.string(),
  is_main: z.boolean(),
});

const meClinicInfoSchema = z.object({
  id: z.string(),
  name: z.string(),
  postal_code: z.string().default(""),
  address: z.string().default(""),
  phone_number: z.string().default(""),
  fax_number: z.string().default(""),
  registration_number: z.string().default(""),
  director_name: z.string().default(""),
  email: z.string().default(""),
  website: z.string().default(""),
  logo_url: z.string().nullable().default(null),
});

const resourcePermissionSchema = z.object({
  view: z.boolean(),
  create: z.boolean(),
  edit: z.boolean(),
  delete: z.boolean(),
});

export const backendMeResponseSchema = z.object({
  id: z.string(),
  email: z.string(),
  display_name: z.string(),
  is_system_admin: z.boolean().default(false),
  // occupation は職種マスタ名
  occupation: z.string().nullable().optional(),
  avatar_url: z.string().nullable().optional(),
  main_clinic_id: z.string(),
  // clinic は /me レスポンスのメイン医院詳細。未所属の場合は null
  clinic: meClinicInfoSchema.nullable().optional(),
  clinics: z.array(clinicMembershipSchema),
  // permissions: resource → CRUD（BEがUNION計算済みのフラット構造）
  permissions: z.record(z.string(), resourcePermissionSchema),
});

export type BackendMeResponse = z.infer<typeof backendMeResponseSchema>;

function mapMeClinicInfo(raw: z.infer<typeof meClinicInfoSchema>): AuthClinic {
  return {
    id: raw.id,
    name: raw.name,
    postalCode: raw.postal_code,
    address: raw.address,
    phoneNumber: raw.phone_number,
    faxNumber: raw.fax_number,
    registrationNumber: raw.registration_number,
    directorName: raw.director_name,
    email: raw.email,
    website: raw.website,
    logoUrl: raw.logo_url,
  };
}

export function mapMeToAuthUser(raw: unknown): AuthUser {
  const me = backendMeResponseSchema.parse(raw);
  return {
    id: me.id,
    email: me.email,
    displayName: me.display_name,
    isSystemAdmin: me.is_system_admin,
    avatarUrl: me.avatar_url,
    mainClinicId: me.main_clinic_id,
    clinic: me.clinic ? mapMeClinicInfo(me.clinic) : null,
    clinics: me.clinics.map((c) => ({
      clinicId: c.clinic_id,
      clinicName: c.clinic_name,
      isMain: c.is_main,
    })),
    permissions: me.permissions,
  };
}
