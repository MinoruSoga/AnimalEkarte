/**
 * Checkup package field type — single source for FE dynamic forms.
 * Keep in sync with BE model.CheckupFieldType (FE-RC-020).
 * Hand-written (TASK-444-S1 freezes new generated-models imports).
 */
export type CheckupFieldType =
  | "number"
  | "single_select"
  | "multi_select"
  | "boolean"
  | "checklist"
  | "text";
