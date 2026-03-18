/**
 * Authentication & Authorization types.
 * Backend types: {@link UserAccount}, {@link UserClinicMembership}, {@link Clinic} from models.ts
 */
import type { ReactNode } from "react";
import {
  UserTypeSystemAdmin,
  UserTypeClinicAdmin,
  UserTypeStaff,
  StaffRoleVeterinarian,
  StaffRoleNurse,
  StaffRoleTrimmer,
  StaffRoleReception,
  StaffRoleManager,
  PermissionAccountAdmin,
  PermissionMedical,
  PermissionMedicalRead,
  PermissionTrimming,
  PermissionBilling,
  PermissionReception,
  PermissionHospitalization,
  PermissionMasterAdmin,
  PermissionShiftAdmin,
  PermissionInventory,
} from "@/types/generated/models";

/** @see {@link import("@/types/generated/models").UserType} */
export const USER_TYPE_VALUES = [UserTypeSystemAdmin, UserTypeClinicAdmin, UserTypeStaff] as const;
export type UserType = (typeof USER_TYPE_VALUES)[number];

export const USER_TYPE_LABELS: Record<UserType, string> = {
  [UserTypeSystemAdmin]: "運営管理者",
  [UserTypeClinicAdmin]: "医院管理者",
  [UserTypeStaff]: "スタッフ",
};

/** @see {@link import("@/types/generated/models").StaffRole} */
export const STAFF_ROLE_VALUES = [
  StaffRoleVeterinarian,
  StaffRoleNurse,
  StaffRoleTrimmer,
  StaffRoleReception,
  StaffRoleManager,
] as const;
export type StaffRole = (typeof STAFF_ROLE_VALUES)[number];

export const STAFF_ROLE_LABELS: Record<StaffRole, string> = {
  [StaffRoleVeterinarian]: "医師",
  [StaffRoleNurse]: "看護師",
  [StaffRoleTrimmer]: "トリマー",
  [StaffRoleReception]: "受付",
  [StaffRoleManager]: "管理職",
};

/** @deprecated JOB_TITLE_VALUES は STAFF_ROLE_VALUES に統合。後方互換のため残す */
export const JOB_TITLE_VALUES = STAFF_ROLE_VALUES;
/** @deprecated JobTitle は StaffRole に統合。後方互換のため残す */
export type JobTitle = StaffRole;
/** @deprecated JOB_TITLE_LABELS は STAFF_ROLE_LABELS に統合。後方互換のため残す */
export const JOB_TITLE_LABELS = STAFF_ROLE_LABELS;

/** @see {@link import("@/types/generated/models").PermissionType} */
export const PERMISSION_VALUES = [
  PermissionAccountAdmin,
  PermissionMedical,
  PermissionMedicalRead,
  PermissionTrimming,
  PermissionBilling,
  PermissionReception,
  PermissionHospitalization,
  PermissionMasterAdmin,
  PermissionShiftAdmin,
  PermissionInventory,
] as const;
export type Permission = (typeof PERMISSION_VALUES)[number];

/** @see {@link import("@/types/generated/models").UserClinicMembership} */
export interface ClinicMembership {
  clinicId: string;
  clinicName: string;
  isMain: boolean;
}

/** @see {@link import("@/types/generated/models").Clinic} */
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

/** @see {@link import("@/types/generated/models").UserAccount} */
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
