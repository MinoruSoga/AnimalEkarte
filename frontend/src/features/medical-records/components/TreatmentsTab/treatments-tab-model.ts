import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import type { CreateTreatmentInput, TreatmentItemType } from "../../types";

/** マスタ検索ダイアログのカテゴリ文字列から治療明細の item_type を決定する。 */
export function resolveItemTypeFromCategory(category: string): TreatmentItemType {
  if (category === "薬剤") return "medicine";
  if (category === "処置") return "procedure";
  if (category === "診察") return "consultation";
  return "other";
}

/**
 * マスタ検索ダイアログでの選択から治療明細作成ペイロードを組み立てる。
 *
 * ts-review-201 CRITICAL 回帰防止: `item.medicineId` は FE 表示用の文字列 ID
 * （`Medicine.id = String(data.id)`）だが、BE の `createTreatmentRequest.MedicineID` は
 * `*uint64`（JSON number）を期待する（`backend/internal/medicalrecord/treatment_request.go`）。
 * 文字列のまま送ると `ShouldBindJSON` が失敗し 400 になり、薬剤選択からの治療行追加が
 * 常に失敗する（一度この回帰が発生した）。ここで必ず `Number(...)` に変換すること。
 */
export function buildMasterSelectionPayload(params: {
  item: TreatmentMasterItem;
  quantity: number;
  sortOrder: number;
}): CreateTreatmentInput {
  const { item, quantity, sortOrder } = params;
  return {
    item_type: resolveItemTypeFromCategory(item.category),
    content: item.name,
    unit_price: item.unitPrice,
    quantity,
    medicine_id: item.medicineId ? Number(item.medicineId) : undefined,
    is_selected: true,
    is_insurance: false,
    discount_amount: 0,
    sort_order: sortOrder,
    memo: "",
  };
}
