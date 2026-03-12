/**
 * Backend ↔ Frontend 型変換
 * バックエンドは snake_case、フロントエンドは camelCase
 */
import type { AuthUser, UserType, Permission } from "../types";

export interface BackendClinicMembership {
  clinic_id: string;
  clinic_name: string;
  branch_name: string;
  is_main: boolean;
}

export interface BackendMeResponse {
  id: string;
  email: string;
  display_name: string;
  user_type: string;
  job_title: string | null;
  avatar_url: string | null;
  main_clinic_id: string;
  clinics: BackendClinicMembership[];
  permissions: Record<string, string[]>;
}

export function mapMeToAuthUser(me: BackendMeResponse): AuthUser {
  return {
    id: me.id,
    email: me.email,
    displayName: me.display_name,
    userType: me.user_type as UserType,
    jobTitle: me.job_title as AuthUser["jobTitle"],
    avatarUrl: me.avatar_url,
    mainClinicId: me.main_clinic_id,
    clinics: me.clinics.map((c) => ({
      clinicId: c.clinic_id,
      clinicName: c.clinic_name,
      branchName: c.branch_name,
      isMain: c.is_main,
    })),
    permissions: me.permissions as Record<string, readonly Permission[]>,
  };
}
