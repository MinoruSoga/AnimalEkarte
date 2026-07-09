export { CheckupsList } from "./routes/CheckupsList";
export { CheckupPetSelection } from "./routes/CheckupPetSelection";
export { CheckupForm } from "./routes/CheckupForm";
export { CheckupAlertBadge } from "@/components/shared/CheckupAlertBadge/CheckupAlertBadge";

// #211 健診パッケージ（型付きフィールド機構）
export { DynamicCheckupFields, buildCheckupResultsPayload } from "./components/DynamicCheckupFields";
export type { CheckupFieldValue } from "./components/DynamicCheckupFields";
export { useGetCheckupTypeFields } from "./api/get-checkup-type-fields";
export type {
  CheckupTypeFieldRow,
  CheckupFieldType,
  CheckupFieldOption,
} from "./api/get-checkup-type-fields";
export { replaceCheckupFieldResults } from "./api/replace-checkup-field-results";
export type { CheckupFieldResultInput } from "./api/replace-checkup-field-results";
export { useGetPetCheckupResults } from "./api/get-pet-checkup-results";
export type { PetCheckupResult } from "./api/get-pet-checkup-results";
