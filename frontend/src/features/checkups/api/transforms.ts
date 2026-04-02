import type { BackendCheckupGlobal } from "./types";

export interface CheckupRecord {
  id: string;
  medicalRecordId: string;
  checkupTypeId: string;
  petId?: string;
  date: string;
  nextDate?: string;
  result: string;
  checkupTypeName: string;
  doctorName: string;
  petName: string;
  ownerName: string;
  ownerId?: string;
}

export function transformCheckupGlobal(data: BackendCheckupGlobal): CheckupRecord {
  return {
    id: data.id,
    medicalRecordId: data.medical_record_id,
    checkupTypeId: data.checkup_type_id,
    petId: data.pet_id,
    date: data.date ? data.date.slice(0, 10) : "",
    nextDate: data.next_date ? data.next_date.slice(0, 10) : undefined,
    result: data.result,
    checkupTypeName: data.checkup_type?.name ?? "",
    doctorName: data.doctor?.name ?? "",
    petName: data.pet_name,
    ownerName: data.owner_name,
    ownerId: data.owner_id,
  };
}
