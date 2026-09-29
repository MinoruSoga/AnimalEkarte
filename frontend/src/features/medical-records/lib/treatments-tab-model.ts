import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import type { CreateTreatmentInput, Treatment, TreatmentItemType } from "../types";

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

// ─────────────────────────────────────────────────────────────────────
// EMR-219: 問診履歴コピー時の治療明細複写
// ─────────────────────────────────────────────────────────────────────

/**
 * 明細複写で参照する現行マスタの最小形（id と現在価格のみ使用）。
 * ConsultationItem / ProcedureItem / Medicine はいずれも { id: string; price: number }
 * を含むためそのまま代入可能。wire 上の治療明細 *_id は number で返る実態に合わせ、
 * マスタ照合は String 正規化で行う（GET treatments 応答は transform を通さない）。
 */
export interface CopyTreatmentMasterRef {
  id: string;
  price: number;
}

export interface CopyTreatmentMasters {
  consultations: CopyTreatmentMasterRef[];
  procedures: CopyTreatmentMasterRef[];
  medicines: CopyTreatmentMasterRef[];
}

function findCopyMaster(
  masters: CopyTreatmentMasterRef[],
  sourceRefId: string | null | undefined,
): CopyTreatmentMasterRef | undefined {
  if (sourceRefId == null || sourceRefId === "") return undefined;
  const id = String(sourceRefId);
  return masters.find((m) => m.id === id);
}

/**
 * wire 応答の参照系 id（*string）を BE create が期待する *uint64（JSON number）へ変換する。
 * 数値化できない値は undefined にして送信自体を省略する（buildMasterSelectionPayload と
 * 同じ ts-review-201 回帰防止。文字列のまま送ると ShouldBindJSON が 400 になる）。
 */
function toNumericId(id: string | null | undefined): number | undefined {
  if (id == null || id === "") return undefined;
  const n = Number(id);
  return Number.isFinite(n) ? n : undefined;
}

/**
 * 複写元 Treatment 行 → 現在カルテへの create ペイロード。
 *
 * - item_type/content/memo/quantity/is_selected/is_insurance/discount_rate/
 *   discount_amount/status はソース行をそのまま引き継ぐ。
 * - unit_price は当時価格を持ち込まず、consultation/procedure/medicine は
 *   対応マスタに一致する行があれば master.id + master.price（現行価格）へ再解決する。
 *   マスタ不一致または item_type=other（無型）行はソースの unit_price・参照を維持する。
 * - 参照系 id（consultation/procedure/medicine/inventory）はいずれも BE が *uint64
 *   （JSON number）を期待するため Number 化する（buildMasterSelectionPayload と同じ
 *   ts-review-201 回帰防止）。wire 応答は *string で返るため文字列のまま送ると 400。
 * - dose_deviation_reason は wire 上 dose_param_snapshot 内に保持される
 *   （backend/internal/model/treatment.go:56 / treatment_request.go:25）ため、
 *   snapshot から取り出して create ペイロードの同キーへ引き継ぐ。同一用量の複写は
 *   同一の逸脱判定を受けるため、reason-required の複写行が 400 で落ちるのを防ぐ。
 * - sort_order は「複写先の現在最大 + 1」を offset としてソース値に加算する。
 */
export function buildCopiedTreatmentPayload(params: {
  source: Treatment;
  masters: CopyTreatmentMasters;
  sortOrderOffset: number;
}): CreateTreatmentInput {
  const { source, masters, sortOrderOffset } = params;
  const deviationReason = source.dose_param_snapshot?.dose_deviation_reason;
  const payload: CreateTreatmentInput = {
    item_type: source.item_type,
    content: source.content,
    memo: source.memo,
    quantity: source.quantity,
    is_selected: source.is_selected,
    is_insurance: source.is_insurance,
    discount_rate: source.discount_rate,
    discount_amount: source.discount_amount,
    status: source.status,
    unit_price: source.unit_price,
    consultation_id: toNumericId(source.consultation_id),
    procedure_id: toNumericId(source.procedure_id),
    inventory_id: toNumericId(source.inventory_id),
    sort_order: source.sort_order + sortOrderOffset,
    dose_deviation_reason: typeof deviationReason === "string" ? deviationReason : undefined,
  };

  if (source.item_type === "consultation") {
    const master = findCopyMaster(masters.consultations, source.consultation_id);
    if (master) {
      payload.consultation_id = Number(master.id);
      payload.unit_price = master.price;
    }
  } else if (source.item_type === "procedure") {
    const master = findCopyMaster(masters.procedures, source.procedure_id);
    if (master) {
      payload.procedure_id = Number(master.id);
      payload.unit_price = master.price;
    }
  } else if (source.item_type === "medicine") {
    const master = findCopyMaster(masters.medicines, source.medicine_id);
    if (master) {
      payload.medicine_id = Number(master.id);
      payload.unit_price = master.price;
    } else if (source.medicine_id != null && source.medicine_id !== "") {
      const numericId = Number(source.medicine_id);
      if (Number.isFinite(numericId)) payload.medicine_id = numericId;
    }
  }
  return payload;
}
