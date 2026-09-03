import type { CarePlanItemType } from "../api/care-plan-items";

/** type ごとに DDL(chk_care_plan_item_ref)が必須とするマスタ参照が要る種別 */
export function requiresRef(type: CarePlanItemType): boolean {
  return type === "medicine" || type === "treatment" || type === "item";
}

/** refId を、現在の type に応じた正しい FK フィールドへ振り分ける */
export function buildRefFields(type: CarePlanItemType, refId: string | null) {
  return {
    medicine_id: type === "medicine" ? refId : null,
    procedure_id: type === "treatment" ? refId : null,
    hospitalization_plan_id: type === "item" ? refId : null,
  };
}
