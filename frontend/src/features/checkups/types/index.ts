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
  petId?: string;
  startDate?: string;
  endDate?: string;
  nextStartDate?: string;
  nextEndDate?: string;
  /** X-16②: 実サーバページング移行 */
  page?: number;
  limit?: number;
}
