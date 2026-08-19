import { paths } from "@/config/paths";

export const EXAMINATION_LIST_CHART_TAB = "検査";
export const EXAMINATION_LIST_EXAM_ID_PARAM = "examId";

export function examinationListDetailHref(input: {
  id: string;
  medicalRecordId?: string;
}): string {
  if (input.medicalRecordId) {
    const params = new URLSearchParams({
      tab: EXAMINATION_LIST_CHART_TAB,
      [EXAMINATION_LIST_EXAM_ID_PARAM]: input.id,
    });
    return `${paths.medicalRecords.detail.getHref(input.medicalRecordId)}?${params.toString()}`;
  }
  return paths.examinations.detail.getHref(input.id);
}

export function examinationCreateHref(petId: string): string {
  const params = new URLSearchParams({
    petId,
    tab: EXAMINATION_LIST_CHART_TAB,
  });
  return `${paths.medicalRecords.new.getHref()}?${params.toString()}`;
}
