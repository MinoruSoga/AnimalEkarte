export interface HospitalizationFormData {
  hospitalizationType: string;
  ownerName: string;
  species: string;
  petName: string;
  petNumber: string;
  petInsurance: string;
  petDetails: string;
  visit: string;
  nextVisit: string;
  /** EMR-353: 次回予定の選択種別（"3weeks"/"4weeks"/"1year"/"other"）。未選択は undefined。 */
  nextScheduleType?: string;
  weight: string;
  displayDate: string;
  endDate: string;
  memo: string;
  ownerRequest: string;
  staffNotes: string;
  doctorId: string;
  doctorName: string;
  cageId: string;
  isInsurance: boolean;
  insuranceCompanyName: string;
  insuranceNumber: string;
}
