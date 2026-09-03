import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

/** Backend GET /v1/examinations/:id/print-snapshot response (OpenAPI ExaminationPrintSnapshot). */
interface BackendExaminationPrintItem {
  id: number;
  exam_type_field_id?: number | null;
  name: string;
  inspection_value: string;
  normal_value: string;
  result: string;
  unit: string;
  reference_value: string;
  ref_min?: number | null;
  ref_max?: number | null;
  qualitative_min?: string | null;
  qualitative_max?: string | null;
  is_assessed: boolean;
  is_abnormal: boolean;
  status: string;
  sort_order: number;
}

interface BackendExaminationPrintDisplay {
  medical_record_no: string;
  pet_name: string;
  medical_record_owner_name: string;
  pet_owner_name: string;
  species_name: string;
  exam_type_name: string;
  doctor_name: string;
}

interface BackendExaminationPrintSnapshot {
  examination_id: number;
  clinic_id: number;
  version: number;
  kind: "working" | "official";
  status: string;
  print_boundary: "official" | "draft";
  watermark?: string;
  date: string;
  result_summary: string;
  machine: string;
  exam_type_id: number;
  medical_record_id?: number | null;
  pet_id?: number | null;
  doctor_id?: number | null;
  display: BackendExaminationPrintDisplay;
  items: BackendExaminationPrintItem[];
}

interface ExaminationPrintItem {
  id: string;
  examTypeFieldId: string | null;
  name: string;
  inspectionValue: string;
  normalValue: string;
  result: string;
  unit: string;
  referenceValue: string;
  refMin: number | null;
  refMax: number | null;
  qualitativeMin: string | null;
  qualitativeMax: string | null;
  isAssessed: boolean;
  isAbnormal: boolean;
  status: string;
  sortOrder: number;
}

export interface ExaminationPrintSnapshot {
  examinationId: string;
  clinicId: string;
  version: number;
  kind: "working" | "official";
  status: string;
  printBoundary: "official" | "draft";
  watermark: string;
  date: string;
  resultSummary: string;
  machine: string;
  examTypeId: string;
  medicalRecordId: string | null;
  petId: string | null;
  doctorId: string | null;
  display: {
    medicalRecordNo: string;
    petName: string;
    medicalRecordOwnerName: string;
    petOwnerName: string;
    speciesName: string;
    examTypeName: string;
    doctorName: string;
  };
  items: ExaminationPrintItem[];
}

function transformExaminationPrintSnapshot(
  data: BackendExaminationPrintSnapshot,
): ExaminationPrintSnapshot {
  return {
    examinationId: String(data.examination_id),
    clinicId: String(data.clinic_id),
    version: data.version,
    kind: data.kind,
    status: data.status,
    printBoundary: data.print_boundary,
    watermark: data.watermark ?? "",
    date: data.date,
    resultSummary: data.result_summary ?? "",
    machine: data.machine ?? "",
    examTypeId: String(data.exam_type_id),
    medicalRecordId: data.medical_record_id != null ? String(data.medical_record_id) : null,
    petId: data.pet_id != null ? String(data.pet_id) : null,
    doctorId: data.doctor_id != null ? String(data.doctor_id) : null,
    display: {
      medicalRecordNo: data.display?.medical_record_no ?? "",
      petName: data.display?.pet_name ?? "",
      medicalRecordOwnerName: data.display?.medical_record_owner_name ?? "",
      petOwnerName: data.display?.pet_owner_name ?? "",
      speciesName: data.display?.species_name ?? "",
      examTypeName: data.display?.exam_type_name ?? "",
      doctorName: data.display?.doctor_name ?? "",
    },
    items: (data.items ?? []).map((item) => ({
      id: String(item.id),
      examTypeFieldId: item.exam_type_field_id != null ? String(item.exam_type_field_id) : null,
      name: item.name,
      inspectionValue: item.inspection_value ?? "",
      normalValue: item.normal_value ?? "",
      result: item.result ?? "",
      unit: item.unit ?? "",
      referenceValue: item.reference_value ?? "",
      refMin: item.ref_min ?? null,
      refMax: item.ref_max ?? null,
      qualitativeMin: item.qualitative_min ?? null,
      qualitativeMax: item.qualitative_max ?? null,
      // Stored server flag only — never reassess on the FE print path.
      isAssessed: item.is_assessed,
      isAbnormal: item.is_abnormal,
      status: item.status,
      sortOrder: item.sort_order,
    })),
  };
}

const getExaminationPrintSnapshot = async (
  id: string,
  version?: number,
): Promise<ExaminationPrintSnapshot> => {
  const { data } = await axios.get<BackendExaminationPrintSnapshot>(
    `/v1/examinations/${id}/print-snapshot`,
    { params: version != null ? { version } : undefined },
  );
  return transformExaminationPrintSnapshot(data);
};

export const useGetExaminationPrintSnapshot = (id: string | undefined, version?: number) => {
  const examinationId = id ?? "";
  return useQuery({
    queryKey: queryKeys.examinations.printSnapshot(examinationId, version),
    queryFn: () => getExaminationPrintSnapshot(examinationId, version),
    enabled: examinationId.length > 0,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
