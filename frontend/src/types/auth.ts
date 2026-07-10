/**
 * Authentication & Authorization types (shared layer).
 * Used across all features for RBAC.
 * Backend types: {@link Account}, {@link StaffClinicAssignment}, {@link Clinic} from models.ts
 */
import type { Resource } from "@/types/generated/models";

export type { Resource };

/** 実効権限の CRUD アクション */
export type ResourceAction = "view" | "create" | "edit" | "delete";

/** 1リソースに対する CRUD 権限 */
interface ResourcePermission {
  view: boolean;
  create: boolean;
  edit: boolean;
  delete: boolean;
}

/** resource → CRUD（バックエンドが UNION 計算済みのフラット実効権限） */
type ResourcePermissions = Record<string, ResourcePermission>;

/** @see {@link import("@/types/generated/models").StaffClinicAssignment} */
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
  /** BUG-367: インボイス帳票の税率別内訳計算用 */
  standardTaxRate: number;
  reducedTaxRate: number;
  accountingDocumentShowLogo: boolean;
  accountingDocumentShowRegistrationWarning: boolean;
  accountingDocumentShowItemCategory: boolean;
  accountingDocumentFooterNote: string;
  /** #190: セクション表示/非表示トグルと表示順 */
  accountingDocumentShowClinicHeader: boolean;
  accountingDocumentShowOwnerPetInfo: boolean;
  accountingDocumentShowItemsTable: boolean;
  accountingDocumentShowPaymentSummary: boolean;
  accountingDocumentSectionOrder: string[];
}

/** @see {@link import("@/types/generated/models").Account} */
export interface AuthUser {
  id: string;
  email: string;
  displayName: string;
  /** true の場合、クロスクリニック権限を持つ運営管理者 */
  isSystemAdmin: boolean;
  avatarUrl: string | null;
  mainClinicId: string;
  /** メイン医院の詳細情報。/me レスポンスから取得。null の場合は未所属 */
  clinic: AuthClinic | null;
  clinics: ClinicMembership[];
  permissions: ResourcePermissions;
}

export interface AuthContextValue {
  user: AuthUser | null;
  currentClinicId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isSwitchingClinic: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  switchClinic: (clinicId: string) => void;
  hasPermission: (resource: Resource, action: ResourceAction) => boolean;
  /** /me を再取得してユーザー（権限含む）を更新する */
  refreshPermissions: () => Promise<void>;
}
