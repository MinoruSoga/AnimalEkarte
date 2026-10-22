/**
 * EMR-523: パッケージ master 編集フォームの純粋ロジック。
 * 契約正本: backend/docs/api.yaml `TreatmentPackage*`（採用 SHA f1ded6ec367459d91baec08414ce780df39e9a3f）。
 * ID/version/quantity/円/tax_rate は文字列のまま保持し、number 化・float 演算・0 補完をしない。
 */
import type {
  TreatmentPackage,
  TreatmentPackageComponentInput,
  TreatmentPackageCreate,
  TreatmentPackageItemKind,
  TreatmentPackageItemRef,
  TreatmentPackageManualDefinition,
  TreatmentPackagePatch,
  TreatmentPackageReorder,
} from "../api/treatment-packages";

const POSITIVE_INT64 = /^[1-9][0-9]*$/;
const DECIMAL = /^-?(0|[1-9][0-9]*)(\.[0-9]*[1-9])?$/;
const QUANTITY = /^([1-9][0-9]*(\.[0-9]*[1-9])?|0\.[0-9]*[1-9])$/;
const YEN = /^(0|[1-9][0-9]*)$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

export const EDIT_VERSION_ERROR = "編集開始時の版を取得できません。再読込してください";

/**
 * EMR-532/Q1: 一覧取得の既存 BE cap（backend/internal/persistence/scope.go の
 * persistence.MaxMasterListRows）に対応する限定定数。raw items がこの件数以上なら
 * 「全件取得を確認できない」として並替を事前抑止する保守的な閾値であり、
 * 全体の登録件数制限を新設するものではない。正確に 10,000 件で全件取得できた
 * 場合も抑止する（保守的制約・PR に明記）。
 */
export const REORDER_FULL_SET_GUARD = 10000;

export type ReorderBlockReason = "pending" | "error" | "no-order-version" | "count-guard";

/**
 * 並替の事前抑止判定（純粋関数）。raw（filter 前）の items.length を使う。
 * pending / error / order_version なし / 件数閾値以上で送信不可を返す。
 */
export function reorderBlockReason(input: {
  isPending: boolean;
  isError: boolean;
  orderVersion: string | undefined;
  rawItemCount: number;
}): ReorderBlockReason | null {
  if (input.isPending) return "pending";
  if (input.isError) return "error";
  if (input.orderVersion === undefined) return "no-order-version";
  if (input.rawItemCount >= REORDER_FULL_SET_GUARD) return "count-guard";
  return null;
}

/** 編集開始時に捕捉した version（canonical 文字列）。以後の refetch 版を混入しない。 */
export function requireEditingVersion(value: unknown): string {
  if (typeof value !== "string" || !POSITIVE_INT64.test(value) || value.length > 19) {
    throw new Error(EDIT_VERSION_ERROR);
  }
  return value;
}

export const ITEM_KIND_LABELS: Record<TreatmentPackageItemKind, string> = {
  consultation: "診察",
  procedure: "処置",
  medicine: "薬剤",
  examination_type: "検査",
  vaccine: "予防接種",
  checkup_type: "定期健診",
  merchandise_item: "物販",
  manual_other: "手入力（その他）",
};

/** typed master ref のフィールド名（kind → FK プロパティ名）。manual_other は FK を持たない。 */
export const ITEM_KIND_ID_FIELD = {
  consultation: "consultation_id",
  procedure: "procedure_id",
  medicine: "medicine_id",
  examination_type: "examination_type_id",
  vaccine: "vaccine_id",
  checkup_type: "checkup_type_id",
  merchandise_item: "merchandise_item_id",
} as const;
export type MasterItemKind = keyof typeof ITEM_KIND_ID_FIELD;

export function isMasterItemKind(kind: TreatmentPackageItemKind): kind is MasterItemKind {
  return kind !== "manual_other";
}

export interface ManualDraft {
  name: string;
  reason: string;
  unitLabel: string;
  /** 非負整数円の文字列。number 化しない。 */
  unitPrice: string;
  taxType: "included" | "excluded" | "exempt";
  /** ExactNumeric 十進文字列。 */
  taxRate: string;
  isNonInsurance: boolean;
}

export interface ComponentDraft {
  /** wire component_key。編集・reorder をまたいで安定（duplicate master は別 key で別行）。 */
  componentKey: string;
  kind: TreatmentPackageItemKind;
  /** master 種別の typed ID 文字列。manual_other では未使用。number 化しない。 */
  masterId: string;
  /** 正数の canonical 十進文字列。生入力を保持し保存時に検証する。 */
  quantity: string;
  manual: ManualDraft;
}

export interface TreatmentPackageDraft {
  name: string;
  description: string;
  isActive: boolean;
  components: ComponentDraft[];
}

export function emptyManualDraft(): ManualDraft {
  return {
    name: "",
    reason: "",
    unitLabel: "",
    unitPrice: "",
    taxType: "included",
    taxRate: "",
    isNonInsurance: false,
  };
}

/** 新規構成行。component_key は新規 canonical UUID（重複 master も別 key の別行）。 */
export function newComponentDraft(kind: TreatmentPackageItemKind): ComponentDraft {
  return {
    componentKey: crypto.randomUUID(),
    kind,
    masterId: "",
    quantity: "1",
    manual: emptyManualDraft(),
  };
}

export function emptyPackageDraft(): TreatmentPackageDraft {
  return {
    name: "",
    description: "",
    isActive: true,
    components: [newComponentDraft("consultation")],
  };
}

/** wire → draft。文字列値はそのまま往復し、ID/数量を number 化しない。 */
export function draftFromPackage(pkg: TreatmentPackage): TreatmentPackageDraft {
  return {
    name: pkg.name,
    description: pkg.description,
    isActive: pkg.is_active,
    components: pkg.components.map((c) => componentDraftFromWire(c)),
  };
}

function componentDraftFromWire(c: TreatmentPackageComponentInput): ComponentDraft {
  const item = c.item;
  const base = {
    componentKey: c.component_key,
    kind: item.kind,
    masterId: "",
    quantity: c.quantity,
    manual: emptyManualDraft(),
  };
  if (item.kind === "manual_other") {
    return {
      ...base,
      manual: {
        name: item.manual.name,
        reason: item.manual.reason,
        unitLabel: item.manual.unit_label ?? "",
        unitPrice: item.manual.unit_price,
        taxType: item.manual.tax_type,
        taxRate: item.manual.tax_rate,
        isNonInsurance: item.manual.is_non_insurance,
      },
    };
  }
  return { ...base, masterId: masterIdFromRef(item) };
}

/** typed master ref → masterId。kind ごとに narrowing して FK を取る。 */
function masterIdFromRef(item: TreatmentPackageItemRef): string {
  switch (item.kind) {
    case "consultation":
      return item.consultation_id;
    case "procedure":
      return item.procedure_id;
    case "medicine":
      return item.medicine_id;
    case "examination_type":
      return item.examination_type_id;
    case "vaccine":
      return item.vaccine_id;
    case "checkup_type":
      return item.checkup_type_id;
    case "merchandise_item":
      return item.merchandise_item_id;
    case "manual_other":
      return "";
  }
}

function buildItemRef(c: ComponentDraft): TreatmentPackageItemRef | null {
  if (c.kind === "manual_other") {
    const m = c.manual;
    const manual: TreatmentPackageManualDefinition = {
      name: m.name.trim(),
      reason: m.reason.trim(),
      unit_label: m.unitLabel === "" ? null : m.unitLabel,
      unit_price: m.unitPrice,
      tax_type: m.taxType,
      tax_rate: m.taxRate,
      is_non_insurance: m.isNonInsurance,
    };
    return { kind: "manual_other", manual };
  }
  if (!POSITIVE_INT64.test(c.masterId)) return null;
  const field = ITEM_KIND_ID_FIELD[c.kind];
  return { kind: c.kind, [field]: c.masterId } as TreatmentPackageItemRef;
}

function componentToWire(c: ComponentDraft): TreatmentPackageComponentInput | null {
  const item = buildItemRef(c);
  if (item === null) return null;
  return { component_key: c.componentKey, item, quantity: c.quantity };
}

export interface DraftValidation {
  /** 保存可能かどうか。 */
  valid: boolean;
  errors: string[];
  /** componentKey → 行内エラー。 */
  componentErrors: ReadonlyMap<string, string[]>;
}

/**
 * 保存前検証。未入力を false/0 で補完せず、エラーとして報告する。
 * master 参照は typed ID 文字列の canonical 形のみ検査（存在・権限・active は BE の責務）。
 */
export function validateDraft(draft: TreatmentPackageDraft): DraftValidation {
  const errors: string[] = [];
  const componentErrors = new Map<string, string[]>();
  const add = (key: string, message: string) => {
    const list = componentErrors.get(key) ?? [];
    list.push(message);
    componentErrors.set(key, list);
  };
  if (draft.name.trim() === "") errors.push("名称は必須です");
  if (draft.name.length > 255) errors.push("名称は255文字以内です");
  if (draft.components.length === 0) errors.push("構成品を1件以上登録してください");
  for (const c of draft.components) {
    if (!UUID.test(c.componentKey)) add(c.componentKey, "component key が不正です");
    if (isMasterItemKind(c.kind)) {
      if (!POSITIVE_INT64.test(c.masterId)) add(c.componentKey, "マスタを選択してください");
    } else {
      const m = c.manual;
      if (m.name.trim() === "") add(c.componentKey, "名称は必須です");
      if (m.reason.trim() === "") add(c.componentKey, "理由は必須です");
      if (!YEN.test(m.unitPrice) || m.unitPrice.length > 19)
        add(c.componentKey, "単価は0以上の整数円です");
      if (!DECIMAL.test(m.taxRate) || m.taxRate === "-0")
        add(c.componentKey, "税率は正確な十進数です");
    }
    if (!QUANTITY.test(c.quantity)) add(c.componentKey, "数量は正の十進数です");
  }
  return { valid: errors.length === 0 && componentErrors.size === 0, errors, componentErrors };
}

export function toCreateBody(draft: TreatmentPackageDraft): TreatmentPackageCreate | null {
  const components = draft.components.map(componentToWire);
  if (components.some((c) => c === null)) return null;
  return {
    name: draft.name.trim(),
    description: draft.description,
    is_active: draft.isActive,
    components: components as TreatmentPackageComponentInput[],
  };
}

function wireOf(pkg: TreatmentPackage): { components: TreatmentPackageComponentInput[] } {
  return { components: pkg.components };
}

function sameComponents(a: TreatmentPackageComponentInput[], b: TreatmentPackageComponentInput[]) {
  return JSON.stringify(a) === JSON.stringify(b);
}

/**
 * PATCH body: version = 編集開始時捕捉版（CAS）。省略フィールドは維持、変更フィールドのみ送る。
 * components は変更時のみ**全構成置換**で送る（部分 merge はしない）。
 * 実変更なし（no-op）では fake revision を起こさないよう null を返す。
 */
export function toPatchBody(
  draft: TreatmentPackageDraft,
  baseline: TreatmentPackage,
): TreatmentPackagePatch | null {
  const body: TreatmentPackagePatch = { version: requireEditingVersion(baseline.version) };
  let changed = false;
  const name = draft.name.trim();
  if (name !== baseline.name) {
    body.name = name;
    changed = true;
  }
  if (draft.description !== baseline.description) {
    body.description = draft.description;
    changed = true;
  }
  if (draft.isActive !== baseline.is_active) {
    body.is_active = draft.isActive;
    changed = true;
  }
  const components = draft.components.map(componentToWire);
  if (components.some((c) => c === null)) return null;
  const wireComponents = components as TreatmentPackageComponentInput[];
  if (!sameComponents(wireComponents, wireOf(baseline).components)) {
    body.components = wireComponents;
    changed = true;
  }
  return changed ? body : null;
}

/** reorder body: 非削除全件（inactive 含む）を一度だけ列挙。 */
export function buildReorderBody(
  orderVersion: string,
  ids: readonly string[],
): TreatmentPackageReorder {
  return { order_version: requireEditingVersion(orderVersion), ids: [...ids] };
}

/** draft が baseline と等しいか（dirty 判定用）。 */
export function isDraftEqual(draft: TreatmentPackageDraft, baseline: TreatmentPackage): boolean {
  return JSON.stringify(draft) === JSON.stringify(draftFromPackage(baseline));
}

/**
 * EMR-532/Q3: 型 → 型別閲覧 resource。backend の treatmentPackageKindViewResource
 * と対応させる（consultation/procedure/medicine/examination_type/vaccine →
 * master-medical、checkup_type → checkups、merchandise_item → master-merchandise）。
 * manual_other は型別 view を持たないため呼び出し側で除外する（MasterItemKind のみ受理）。
 */
export type KindViewResource = "master-medical" | "checkups" | "master-merchandise";

export function treatmentPackageKindViewResource(kind: MasterItemKind): KindViewResource {
  if (kind === "checkup_type") return "checkups";
  if (kind === "merchandise_item") return "master-merchandise";
  return "master-medical";
}

/** 409 比較用: manual の更新可能フィールド（単価・税率）を文字列のまま表現する。number 化しない。 */
function manualPriceSummary(unitPrice: string, taxType: string, taxRate: string): string {
  return ` 単価${unitPrice}円 税${taxType}/${taxRate}`;
}

/** 409 比較用: draft 構成の行ラベル（kind ラベル + 識別子 + 数量 + manual 単価/税率）。number 化しない。 */
export function draftComponentSummary(c: ComponentDraft): string {
  const name = c.kind === "manual_other" ? c.manual.name : c.masterId;
  const base = `${ITEM_KIND_LABELS[c.kind]} ${name} ×${c.quantity}`;
  if (c.kind !== "manual_other") return base;
  return base + manualPriceSummary(c.manual.unitPrice, c.manual.taxType, c.manual.taxRate);
}

/** 409 比較用: wire 構成の行ラベル（manual 単価/税率を含む）。 */
export function wireComponentSummary(c: TreatmentPackageComponentInput): string {
  const item = c.item;
  if (item.kind !== "manual_other") {
    return `${ITEM_KIND_LABELS[item.kind]} ${masterIdFromRef(item)} ×${c.quantity}`;
  }
  const m = item.manual;
  return (
    `${ITEM_KIND_LABELS.manual_other} ${m.name} ×${c.quantity}` +
    manualPriceSummary(m.unit_price, m.tax_type, m.tax_rate)
  );
}

/** 409 比較用: draft と最新 wire の項目別差分フラグ（純粋関数）。 */
export function draftPackageDiff(draft: TreatmentPackageDraft, latest: TreatmentPackage) {
  const wire = draft.components.map(componentToWire);
  return {
    name: draft.name.trim() !== latest.name,
    description: draft.description !== latest.description,
    isActive: draft.isActive !== latest.is_active,
    components:
      wire.some((c) => c === null) ||
      !sameComponents(wire as TreatmentPackageComponentInput[], latest.components),
  };
}

/** 一覧を ids の順に並べ替えた ID 列（全件・inactive 含む）。 */
export function moveId(ids: readonly string[], id: string, direction: -1 | 1): string[] {
  const index = ids.indexOf(id);
  const target = index + direction;
  if (index < 0 || target < 0 || target >= ids.length) return [...ids];
  const next = [...ids];
  const [moved] = next.splice(index, 1);
  next.splice(target, 0, moved!);
  return next;
}
