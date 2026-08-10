/**
 * Owners feature types (UI-facing: 日本語ラベル値)
 * Note: PetGender/AcquisitionType/DangerLevel/MembershipTypeLabel は UI 表示用日本語値。
 * models.ts の英語 enum 値（male/female/unknown 等）とは異なる。
 * FE3-4: 生成型 `MembershipType`（wire 値）と同名衝突していたため `MembershipTypeLabel` へ改名。
 * Backend ↔ Frontend の変換は transform 関数が担当。
 * @see {@link import("@/types/generated/models").PetGender} — Backend enum (英語値)
 * @see {@link import("@/types/generated/models").MembershipType} — Backend enum (英語値)
 */
export const PET_GENDER_VALUES = ["雄", "雌", "不明"] as const;

export const ACQUISITION_TYPE_VALUES = ["購入", "譲渡", "保護", "その他"] as const;
type AcquisitionType = (typeof ACQUISITION_TYPE_VALUES)[number];

export const DANGER_LEVEL_VALUES = ["低", "中", "高"] as const;
type DangerLevel = (typeof DANGER_LEVEL_VALUES)[number];

// Owner-related type constants and definitions
export const MEMBERSHIP_TYPE_VALUES = ["非会員", "会員", "退亡者", "他診/準"] as const;
export type MembershipTypeLabel = (typeof MEMBERSHIP_TYPE_VALUES)[number];

export interface PetFormData {
  id: string;
  /** true = 新規飼主登録モードでローカルに追加済み（飼主保存時に一括API送信） */
  isPending?: boolean;
  petNumber: string;
  petName: string;
  petNameKana?: string;
  status: string;
  species: string;
  /** バックエンドの animal_species.id （UUID文字列）*/
  animalSpeciesId?: string;
  gender: string;
  birthDate: string;
  breed?: string;
  color: string;
  /** #158 レガシー EMR 準拠: ペット血液型 */
  bloodType?: string;
  /** #158 レガシー EMR 準拠: マイクロチップ番号 */
  microchipNumber?: string;
  weight: string;
  neuteredDate?: string;
  acquisitionType?: AcquisitionType;
  dangerLevel?: DangerLevel;
  dangerReason?: string;
  food?: string;
  environment: string;
  remarks: string;
  /** バックエンドの insurance.id （UUID文字列）*/
  insuranceId?: string;
  /** 保険会社名（APIから取得） */
  insuranceName?: string;
  insuranceDetails?: string;
  /** PR#186 P2-2 Bug#1: 死亡記録日時。未死亡 = undefined/null */
  deceasedAt?: string | null;
  /** BUG-003: staff GET /pets/{id} 由来の死亡理由。owner nested には無い */
  deceasedReason?: string | null;
}

// Owner data
export interface OwnerData {
  ownerId: string;
  postalCode: string;
  company: string;
  membershipType: MembershipTypeLabel;
  ownerName: string;
  address1: string;
  ownerNameKana: string;
  address2: string;
  homeAddress1: string;
  homeAddress2: string;
  isDangerous: boolean;
  birthDate: string;
  email: string;
  phone: string;
  companyPhone: string;
  remarks: string;
  discountRate?: number;
  homePostalCode?: string;
  /** #158 レガシー EMR 準拠: DM 送付希望。undefined=未変更 / null=未設定 / true=必要 / false=不要 */
  dmPreference?: boolean | null;
  /** #84: 登録先医院（新規登録時のみ指定可。未指定は現在の医院） */
  clinicId?: string;
}
