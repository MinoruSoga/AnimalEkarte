import { axios } from "@/lib/axios";

/** PUT 一括置換に送る健診結果値 1 件分。status / is_abnormal はサーバ側で導出する。 */
export interface CheckupFieldResultInput {
  checkup_type_field_id: number;
  value_number?: number | null;
  value_text?: string;
  value_bool?: boolean | null;
  value_list?: string[];
}

// PUT /v1/medical-records/:medicalRecordId/checkups/:checkupId/field-results
// 既存全削除→一括登録の PUT セマンティクス。
export const replaceCheckupFieldResults = async (
  medicalRecordId: string | number,
  checkupId: string | number,
  results: CheckupFieldResultInput[],
): Promise<void> => {
  await axios.put(
    `/v1/medical-records/${medicalRecordId}/checkups/${checkupId}/field-results`,
    { results },
  );
};
