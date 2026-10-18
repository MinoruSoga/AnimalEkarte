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
 *
 * 診察/処置カテゴリも同様に `consultation_id` / `procedure_id` を有限数値へ変換して
 * 紐付ける（複写時の現行マスタ価格再解決に必要）。数値化できない ID は参照を送らない。
 * 他カテゴリには診察/処置の参照を設定せず、薬剤の既存 `medicine_id` は維持する。
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
    consultation_id: item.category === "診察" ? toNumericId(item.id) : undefined,
    procedure_id: item.category === "処置" ? toNumericId(item.id) : undefined,
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
interface CopyTreatmentMasterRef {
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
 *   複写先の現行マスタと ID 一致した場合だけ master.id + master.price（現行価格）へ
 *   再解決する。参照欠落・不一致・非有限の typed 行、参照を持つ other 行、未知の
 *   item_type は例外で拒否し、呼び出し側は create 前に全行検証して中断する。
 *   参照を持たない手入力 other 行のみソースの単価・業務値を維持する。
 * - 一致した typed 参照のみ number 化して送る（BE は *uint64=JSON number を期待）。
 *   other/未一致の参照は送信自体を省略し、inventory 等の無関係参照は伝播しない。
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
    sort_order: source.sort_order + sortOrderOffset,
    dose_deviation_reason: typeof deviationReason === "string" ? deviationReason : undefined,
  };

  const REJECT = "複写先のマスタを確認し、治療明細を再選択してください";
  const resolveTypedRef = (
    refId: string | null | undefined,
    list: CopyTreatmentMasterRef[],
    refKey: "consultation_id" | "procedure_id" | "medicine_id",
  ) => {
    const master = findCopyMaster(list, refId);
    const numericId = master ? Number(master.id) : NaN;
    if (!master || !Number.isFinite(numericId) || !Number.isFinite(master.price)) {
      throw new Error(REJECT);
    }
    payload[refKey] = numericId;
    payload.unit_price = master.price;
  };

  if (source.item_type === "consultation") {
    resolveTypedRef(source.consultation_id, masters.consultations, "consultation_id");
  } else if (source.item_type === "procedure") {
    resolveTypedRef(source.procedure_id, masters.procedures, "procedure_id");
  } else if (source.item_type === "medicine") {
    resolveTypedRef(source.medicine_id, masters.medicines, "medicine_id");
  } else if (source.item_type === "other") {
    const hasRef = (v: string | null | undefined) => v != null && v !== "";
    if (
      hasRef(source.consultation_id) ||
      hasRef(source.procedure_id) ||
      hasRef(source.medicine_id) ||
      hasRef(source.inventory_id)
    ) {
      throw new Error(REJECT);
    }
  } else {
    throw new Error(REJECT);
  }
  return payload;
}
