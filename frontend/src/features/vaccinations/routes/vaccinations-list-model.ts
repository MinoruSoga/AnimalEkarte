import { paths } from "@/config/paths";

export const VACCINATION_LIST_CHART_TAB = "予防接種";
export const VACCINATION_LIST_ID_PARAM = "vaccinationId";

export function vaccinationListDetailHref(input: {
  id: string;
  medicalRecordId?: string;
}): string {
  if (input.medicalRecordId) {
    const params = new URLSearchParams({
      tab: VACCINATION_LIST_CHART_TAB,
      [VACCINATION_LIST_ID_PARAM]: input.id,
    });
    return `${paths.medicalRecords.detail.getHref(input.medicalRecordId)}?${params.toString()}`;
  }
  return paths.vaccinations.detail.getHref(input.id);
}

export function vaccinationCreateHref(petId: string): string {
  // BUG-501: 一覧→新規はカルテ未保存ゲートを避ける独立フォームへ（仕様 15 §1.1 / S03#1）
  const params = new URLSearchParams({ petId });
  return `${paths.vaccinations.new.getHref()}?${params.toString()}`;
}
