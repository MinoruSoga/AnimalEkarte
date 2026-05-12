export interface BackendCheckupGlobal {
  id: string;
  medical_record_id: string;
  checkup_type_id: string;
  pet_id?: string;
  date: string;
  next_date?: string;
  doctor_id?: string;
  result: string;
  checkup_type?: { id: string; name: string };
  doctor?: { id: string; name: string };
  pet_name: string;
  owner_name: string;
  owner_id?: string;
}

export interface CheckupFilters {
  startDate?: string;
  endDate?: string;
  nextStartDate?: string;
  nextEndDate?: string;
}

// --- Checkup Alerts (GET /v1/checkups/alerts) ---

export interface BackendCheckupAlertItem {
  checkup_id: number;
  pet_id: number;
  pet_name: string;
  owner_name: string;
  checkup_type_name: string;
  last_checkup_date: string | null;
  next_date: string;
  days: number;
}

export interface BackendCheckupAlertCategory {
  count: number;
  items: BackendCheckupAlertItem[];
}

export interface BackendCheckupAlertsResponse {
  overdue: BackendCheckupAlertCategory;
  upcoming: BackendCheckupAlertCategory;
}

export interface CheckupAlertItem {
  checkupId: number;
  petId: number;
  petName: string;
  ownerName: string;
  checkupTypeName: string;
  lastCheckupDate: string | null;
  nextDate: string;
  days: number;
}

export interface CheckupAlertCategory {
  count: number;
  items: CheckupAlertItem[];
}

export interface CheckupAlertsResponse {
  overdue: CheckupAlertCategory;
  upcoming: CheckupAlertCategory;
}
