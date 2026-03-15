/**
 * Backend ↔ Frontend 型変換
 * バックエンドは snake_case、フロントエンドは camelCase
 */
import { z } from "zod";
import type { AuthUser, AuthClinic, StaffRole } from "../types";
import { USER_TYPE_VALUES, STAFF_ROLE_VALUES, PERMISSION_VALUES } from "../types";

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

// z.enum は mutable tuple を要求するため、readonly const 配列を型キャストして渡す
// [T, ...T[]] 形式のタプル型へのキャストで any を回避する
const toEnumTuple = <T extends string>(v: readonly T[]): [T, ...T[]] =>
  v as unknown as [T, ...T[]];

export const backendMeResponseSchema = z.object({
  id: z.string(),
  email: z.string(),
  display_name: z.string(),
  user_type: z.enum(toEnumTuple(USER_TYPE_VALUES)),
  // staff_role はバックエンドの staff_role ENUM（スタッフ紐づけがない場合は null）
  staff_role: z.string().nullable(),
  // job_title は後方互換のため残存（現在は staff_role を優先使用）
  job_title: z.string().nullable().optional(),
  avatar_url: z.string().nullable(),
  main_clinic_id: z.string(),
  // clinic は /me レスポンスのメイン医院詳細。未所属の場合は null
  clinic: meClinicInfoSchema.nullable().optional(),
  clinics: z.array(clinicMembershipSchema),
  permissions: z.record(z.string(), z.array(z.enum(toEnumTuple(PERMISSION_VALUES)))),
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
    userType: me.user_type,
    // staff_role が STAFF_ROLE_VALUES に含まれる値のみ受け入れ、それ以外は null
    staffRole: (STAFF_ROLE_VALUES as readonly string[]).includes(me.staff_role ?? "")
      ? (me.staff_role as StaffRole)
      : null,
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
