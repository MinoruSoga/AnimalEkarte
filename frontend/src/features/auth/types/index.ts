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

/** バックエンドの staff_role ENUM と一致させる */
export const STAFF_ROLE_VALUES = [
  "veterinarian",
  "nurse",
  "trimmer",
  "reception",
  "manager",
] as const;
export type StaffRole = (typeof STAFF_ROLE_VALUES)[number];

export const STAFF_ROLE_LABELS: Record<StaffRole, string> = {
  veterinarian: "医師",
  nurse: "看護師",
  trimmer: "トリマー",
  reception: "受付",
  manager: "管理職",
};

/** @deprecated JOB_TITLE_VALUES は STAFF_ROLE_VALUES に統合。後方互換のため残す */
export const JOB_TITLE_VALUES = STAFF_ROLE_VALUES;
/** @deprecated JobTitle は StaffRole に統合。後方互換のため残す */
export type JobTitle = StaffRole;
/** @deprecated JOB_TITLE_LABELS は STAFF_ROLE_LABELS に統合。後方互換のため残す */
export const JOB_TITLE_LABELS = STAFF_ROLE_LABELS;

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
  isMain: boolean;
}

/** /me レスポンスに含まれるメイン医院の詳細情報 */
export interface AuthClinic {
  id: string;
  name: string;
  postalCode: string;
  address: string;
  phoneNumber: string;
  faxNumber: string;
  registrationNumber: string;
  directorName: string;
  email: string;
  website: string;
  logoUrl: string | null;
}

export type ClinicPermissions = Record<string, readonly Permission[]>;

export interface AuthUser {
  id: string;
  email: string;
  displayName: string;
  userType: UserType;
  /** バックエンドの staff_role ENUM 値。StaffIDが紐づくスタッフが存在する場合のみ非null */
  staffRole: StaffRole | null;
  avatarUrl: string | null;
  mainClinicId: string;
  /** メイン医院の詳細情報。/me レスポンスから取得。null の場合は未所属 */
  clinic: AuthClinic | null;
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
