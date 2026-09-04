// FE-RC-015: 実装は components/shared へ昇格済み（medical-records/CheckupsTab から直接
// 参照可能にするため）。checkups 内部の後方互換 re-export として維持する。テストは
// components/shared/DynamicCheckupFields/DynamicCheckupFields.test.tsx に移設済み。
export {
  DynamicCheckupFields,
  buildCheckupResultsPayload,
  type CheckupFieldValue,
} from "@/components/shared/DynamicCheckupFields/DynamicCheckupFields";
