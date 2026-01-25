export * from './owner';

import { ReactNode } from "react";

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
  ownerName: string;
  phone?: string;
  petNumber?: string; // Optional for mocks that might not have it
  name: string; // Unified name
  species: string;
  breed?: string;
  gender?: string;
  status?: "生存" | "死亡";
  birthDate?: string;
  weight?: string;
  environment?: string;
  lastVisit?: string;
  insuranceName?: string;
  insuranceDetails?: string;
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
  status: "confirmed" | "pending" | "canceled" | "checked_in" | "in_consultation" | "accounting" | "completed" | "cancelled";
  notes?: string;
  petId?: string;
}

// Trimming Types
export interface TrimmingRecord {
  id: string;
  date: string;
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
  code: string;
  name: string;
  category?: string;
  price: number; // Changed from optional to required for TreatmentMaster compatibility
  status: "active" | "inactive";
  description?: string;
  inventoryId?: string;
  defaultQuantity?: number;
}

export type MasterCategory = "examination" | "vaccine" | "medicine" | "staff" | "insurance" | "cage" | "serviceType" | "trimming_course" | "trimming_option";

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
