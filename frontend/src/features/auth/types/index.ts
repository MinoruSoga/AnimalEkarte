/**
 * Authentication & Authorization types.
 */
import type { ReactNode } from "react";

export const USER_TYPE_VALUES = ["system_admin", "clinic_admin", "staff"] as const;
export type UserType = (typeof USER_TYPE_VALUES)[number];

export const USER_TYPE_LABELS: Record<UserType, string> = {
  system_admin: "運営管理者",
  clinic_admin: "医院管理者",
  staff: "スタッフ",
};

export const JOB_TITLE_VALUES = [
  "veterinarian",
  "nurse",
  "trimmer",
  "reception",
  "general_staff",
] as const;
export type JobTitle = (typeof JOB_TITLE_VALUES)[number];

export const JOB_TITLE_LABELS: Record<JobTitle, string> = {
  veterinarian: "医師",
  nurse: "看護師",
  trimmer: "トリマー",
  reception: "受付",
  general_staff: "職員",
};

export const PERMISSION_VALUES = [
  "account_admin",
  "medical",
  "medical_read",
  "trimming",
  "billing",
  "reception",
  "hospitalization",
  "master_admin",
  "shift_admin",
  "inventory",
] as const;
export type Permission = (typeof PERMISSION_VALUES)[number];

export interface ClinicMembership {
  clinicId: string;
  clinicName: string;
  branchName: string;
  isMain: boolean;
}

export type ClinicPermissions = Record<string, readonly Permission[]>;

export interface AuthUser {
  id: string;
  email: string;
  displayName: string;
  userType: UserType;
  jobTitle: JobTitle | null;
  avatarUrl: string | null;
  mainClinicId: string;
  clinics: ClinicMembership[];
  permissions: ClinicPermissions;
}

export interface AuthContextValue {
  user: AuthUser | null;
  currentClinicId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isSwitchingClinic: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  switchClinic: (clinicId: string) => Promise<void>;
  hasPermission: (permission: Permission) => boolean;
  hasAnyPermission: (permissions: readonly Permission[]) => boolean;
}

export interface ProtectedRouteProps {
  permission?: Permission;
  anyOf?: readonly Permission[];
  children: ReactNode;
}
