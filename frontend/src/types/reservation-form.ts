// FE6-17: features/reservations/types/index.ts から正本を移動。
// CODING_RULES §3.6 に従い型定義は src/types/ に置き、feature 側は re-export のみとする。
// components/shared/ReservationFormModal（feature 非依存）の新規飼主インラインフォームが
// 使用する形状のため、feature に依存しない場所へ配置する。
export interface NewOwnerFormData {
  ownerName: string;
  phone: string;
  petName: string;
  chiefComplaint: string;
  animalSpeciesId: number;
}
