export type { Owner, CreateOwnerRequest, UpdateOwnerRequest } from './owner';

import { ReactNode } from "react";

// Clinic Types
export interface ClinicInfo {
  name: string;
  postalCode: string;
  address: string;
  phoneNumber: string;
  faxNumber?: string;
  registrationNumber?: string;
  directorName?: string;
  email?: string;
  website?: string;
  logoUrl?: string;
}

// Sidebar Types
export interface MenuItem {
    icon?: ReactNode;
    label: string;
    path?: string;
    subItems?: MenuItem[];
}

// Dashboard Types
export interface Appointment {
  id: string;
  time: string;
  ownerName: string;
  petType: string;
  petName: string;
  visitType: "初診" | "再診";
  serviceType: string;
  nextAppointment?: "次回予約無" | "次回予約済" | "精算未確認" | "精算確認済";
  isDesignated?: boolean;
  doctor?: string;
  petId?: string;
  ownerId?: string;
}

export interface ColumnData {
  title: string;
  appointments: Appointment[];
}

// Medical Record Types
export interface MedicalRecord {
  id: string;
  recordNo: string;
  date: string;
  ownerId?: string;
  ownerName: string;
  petId?: string;
  petName: string;
  species: string;
  chiefComplaint: string;
  doctor: string;
  status: "作成中" | "確定済";
  // SOAPS フィールド
  subjective?: string;
  objective?: string;
  assessment?: string;
  plan?: string;
  surgeryNotes?: string;
  diagnosis?: string;
  treatment?: string;
  prescription?: string;
  notes?: string;
  visitType?: string;
  accountingId?: string;  // 関連会計レコードID
}

// Hospitalization Types
export interface Hospitalization {
  id: string;
  hospitalizationNo: string;
  ownerName: string;
  petName: string;
  species: string;
  hospitalizationType: "入院" | "ホテル";
  startDate: string;
  endDate: string;
  status: "入院中" | "退院済" | "予約" | "一時帰宅";
  cageId?: string;
}

export interface CarePlanItem {
    id: string;
    hospitalizationId: string;
    type: "food" | "medicine" | "instruction" | "item" | "treatment"; // treatmentを追加
    name: string;
    description: string; // Dosage, amount, etc.
    timing: string[]; // ["morning", "noon", "night"] etc.
    status: "active" | "completed" | "discontinued";
    notes?: string;
    
    // Link to Treatment Master
    masterId?: string; // code
    unitPrice?: number;
    category?: string; // Master category
}

export interface DailyRecord {
    id: string;
    hospitalizationId: string;
    date: string; // YYYY-MM-DD
    vitals: {
        id: string;
        time: string;
        temperature?: number;
        heartRate?: number;
        respirationRate?: number;
        weight?: number;
        notes?: string;
        staff: string;
    }[];
    careLogs: {
        id: string;
        time: string;
        type: "food" | "excretion" | "medicine" | "other" | "treatment";
        status: "completed" | "partial" | "skipped";
        value?: string; // "100%", "Normal Stool", etc.
        staff: string;
        notes?: string;
    }[];
    staffNotes: {
        id: string;
        time: string;
        content: string;
        staff: string;
    }[];
}

// Common Form Types
export interface TreatmentPlan {
  id: string;
  treatmentContent: string;
  memo: string;
  insurance: boolean;
  unitPrice: number;
  quantity: number;
  discount: number;
  discountAmount: number;
  subtotal: number;
}

// Pet & Owner Types
export interface Pet {
  id: string; // Unified ID
  ownerId: string;
  ownerNumber?: number; // 飼主番号（表示用連番）
  ownerName: string;
  phone?: string;
  petNumber?: string; // Optional for mocks that might not have it
  name: string; // Unified name
  petNameKana?: string;
  species: string;
  /** バックエンドの animal_species.id */
  animalSpeciesId?: string;
  breed?: string;
  gender?: string;
  status?: "生存" | "死亡";
  birthDate?: string;
  weight?: string;
  environment?: string;
  lastVisit?: string;
  /** バックエンドの insurance.id */
  insuranceId?: string;
  insuranceName?: string;
  insuranceDetails?: string;
  remarks?: string;
  /** マイクロチップ番号 */
  microchipId?: string;
}

// Reservation Types
export interface ReservationAppointment {
  id: string;
  start: Date;
  end: Date;
  ownerName: string;
  petName: string;
  visitType: "first" | "revisit";
  type: string;
  doctor: string;
  isDesignated: boolean;
  status: "confirmed" | "pending" | "checked_in" | "in_consultation" | "accounting" | "completed" | "cancelled";
  notes?: string;
  petId?: string;
}

export type ReservationStatus = ReservationAppointment["status"];
export type VisitType = "first" | "revisit";
export type CalendarView = "month" | "week";
export type ReservationType = string;
export type NavigationState = { from?: string };

export const CALENDAR_VIEW_VALUES = ["month", "week"] as const;
export const RESERVATION_STATUS_VALUES = [
  "confirmed",
  "pending",
  "checked_in",
  "in_consultation",
  "accounting",
  "completed",
  "cancelled",
] as const;
export const RESERVATION_STATUS_LABELS: Record<ReservationStatus, string> = {
  confirmed: "予約確定",
  pending: "仮予約",
  checked_in: "受付済",
  in_consultation: "診療中",
  accounting: "会計待ち",
  completed: "完了",
  cancelled: "キャンセル",
};
export const RESERVATION_TYPE_VALUES = ["診療", "検診", "手術", "トリミング", "ワクチン", "入院"] as const;

// Trimming Types
export interface TrimmingRecord {
  id: string;
  date: string;
  petId?: string;
  ownerId?: string;
  petNumber: string;
  petName: string;
  ownerName: string;
  species: string;
  weight: string;
  styleRequest: string;
  staff: string;
  status: "完了" | "予約" | "進行中";
}

// Examination Types
export interface ExaminationRecord {
  id: string;
  date: string;
  ownerName: string;
  petName: string;
  testType: string;
  doctor: string;
  status: "依頼中" | "検査中" | "完了";
  resultSummary?: string;
  machine?: string;
  items?: ExaminationItem[];
}

export interface ExaminationItem {
  id: string;
  name: string;
  result?: string;
  unit?: string;
  referenceRange?: string;
}

// Accounting Types
export interface AccountingRecord {
  id: string;
  date: string;
  ownerName: string;
  petName: string;
  amount: number;
  method: "現金" | "クレジットカード" | "電子マネー" | "-";
  status: "未収" | "回収済" | "キャンセル";
  note?: string;
}

// Vaccination Types
export interface VaccinationRecord {
  id: string;
  ownerName: string;
  petName: string;
  vaccineName: string;
  date: string;
  nextDate: string;
}

// Settings Types
export interface MasterItem {
  id: string;
  name: string;
  category?: string;
  price: number; // Changed from optional to required for TreatmentMaster compatibility
  status: "active" | "inactive";
  description?: string;
  inventoryId?: string;
  defaultQuantity?: number;
}

// MasterCategory: master_items テーブルに残るカテゴリのみ
// 分離されたカテゴリ (staff, cage, medicine, insurance, trimming_course, trimming_option, examination)
// は専用テーブル・専用APIに移行済み
export type MasterCategory = "vaccine" | "serviceType" | "consultation" | "procedure" | "hospitalization" | "diagnosis_category" | "diagnosis_name" | "checkup";

// Staff Member Types
export interface StaffMember {
  id: string;
  code: string;
  name: string;
  nameKana?: string;
  role: StaffRole;
  licenseNumber?: string;
  email?: string;
  phone?: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export type StaffRole = 'veterinarian' | 'nurse' | 'trimmer' | 'reception' | 'manager';

export const STAFF_ROLE_LABELS: Record<StaffRole, string> = {
  veterinarian: '獣医師',
  nurse: '看護師',
  trimmer: 'トリマー',
  reception: '受付',
  manager: '管理職',
};

// Cage Types
export interface Cage {
  id: string;
  code: string;
  name: string;
  cageType: CageType;
  cageSize: CageSize;
  bodySize?: BodySize;
  billingUnit: BillingUnit;
  price: number;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export type CageType = 'icu' | 'dog' | 'cat' | 'general';
export type CageSize = 'small' | 'medium' | 'large';
export type BodySize = 'small' | 'medium' | 'large';
export type BillingUnit = 'per_day' | 'per_night';

// Medicine Types
export interface Medicine {
  id: string;
  code: string;
  name: string;
  dosageForm: DosageForm;
  medicineUnit: MedicineUnit;
  price: number;
  defaultQuantity?: number;
  inventoryId?: string;
  description: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export type DosageForm = 'tablet' | 'liquid' | 'injection' | 'topical' | 'powder';
export type MedicineUnit = 'per_tablet' | 'per_ml' | 'per_dose' | 'per_gram';

// InsuranceCompany Types
export interface InsuranceCompany {
  id: string;
  code: string;
  name: string;
  coverageRate: CoverageRate;
  contactPhone: string;
  description: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export type CoverageRate = '50' | '70' | '80' | '100';

// TrimmingCourse Types
export interface TrimmingCourse {
  id: string;
  code: string;
  name: string;
  targetSize: TargetSize;
  duration?: string;
  price: number;
  parentId?: string;
  description: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

// TrimmingOption Types
export interface TrimmingOption {
  id: string;
  code: string;
  name: string;
  targetSize?: TargetSize;
  combinable: Combinable;
  price: number;
  parentId?: string;
  description: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export type TargetSize = 'small' | 'medium' | 'large' | 'cat';
export type Combinable = 'yes' | 'no';

// ExaminationTypeInspection Types
export interface ExaminationTypeInspection {
  id: string;
  examinationTypeId: string;
  name: string;
  inspectionValue: string;
  normalValue: string;
  sortOrder: number;
  createdAt: string;
}

// ExaminationType Types
export interface ExaminationType {
  id: string;
  code: string;
  name: string;
  color?: string;
  price: number;
  parentId?: string;
  description: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
  inspections?: ExaminationTypeInspection[];
  children?: ExaminationType[];
}

// Inventory Types
export interface InventoryItem {
  id: string;
  name: string;
  category: "medicine" | "consumable" | "food" | "other";
  quantity: number;
  unit: string;
  minStockLevel: number;
  location?: string;
  expiryDate?: string;
  supplier?: string;
  lastRestocked?: string;
  status: "sufficient" | "low" | "out_of_stock";
}

// Sorting Types
export const SORT_ORDER_VALUES = ["desc", "asc"] as const;
export type SortOrder = (typeof SORT_ORDER_VALUES)[number];
