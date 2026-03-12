/**
 * Backend ↔ Frontend 型変換
 * バックエンドは snake_case、フロントエンドは camelCase
 */
import { z } from "zod";
import type { AuthUser } from "../types";
import { USER_TYPE_VALUES, JOB_TITLE_VALUES, PERMISSION_VALUES } from "../types";

const clinicMembershipSchema = z.object({
  clinic_id: z.string(),
  clinic_name: z.string(),
  branch_name: z.string(),
  is_main: z.boolean(),
});

// z.enum は mutable tuple を要求するため、readonly const 配列を型キャストして渡す
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const toEnumTuple = <T extends string>(v: readonly T[]): [T, ...T[]] => v as any;

export const backendMeResponseSchema = z.object({
  id: z.string(),
  email: z.string(),
  display_name: z.string(),
  user_type: z.enum(toEnumTuple(USER_TYPE_VALUES)),
  job_title: z.enum(toEnumTuple(JOB_TITLE_VALUES)).nullable(),
  avatar_url: z.string().nullable(),
  main_clinic_id: z.string(),
  clinics: z.array(clinicMembershipSchema),
  permissions: z.record(z.string(), z.array(z.enum(toEnumTuple(PERMISSION_VALUES)))),
});

export type BackendMeResponse = z.infer<typeof backendMeResponseSchema>;

export function mapMeToAuthUser(raw: unknown): AuthUser {
  const me = backendMeResponseSchema.parse(raw);
  return {
    id: me.id,
    email: me.email,
    displayName: me.display_name,
    userType: me.user_type,
    jobTitle: me.job_title,
    avatarUrl: me.avatar_url,
    mainClinicId: me.main_clinic_id,
    clinics: me.clinics.map((c) => ({
      clinicId: c.clinic_id,
      clinicName: c.clinic_name,
      branchName: c.branch_name,
      isMain: c.is_main,
    })),
    permissions: me.permissions,
  };
}
