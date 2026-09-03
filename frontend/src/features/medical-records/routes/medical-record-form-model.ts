/** 来院種別の正本は BE enum `first` / `revisit` のみ（docs/spec/reservation-to-record-flow.md）。 */
export const VISIT_TYPE_OPTIONS = ["初診", "再診"] as const;

const MEDICAL_RECORD_TABS = [
  "問診",
  "診察/治療プラン",
  "治療",
  "予防接種",
  "定期健診",
  "検査",
  "画像",
  "見積書",
  "会計(医師確認)",
];

export const MEDICAL_RECORD_TAB_ITEMS = MEDICAL_RECORD_TABS.map((tab) => ({
  value: tab,
  label: tab,
}));

const MEDICAL_RECORD_TAB_ALIASES: Record<string, (typeof MEDICAL_RECORD_TABS)[number]> = {
  checkup: "定期健診",
  checkups: "定期健診",
  定期健診: "定期健診",
  examinations: "検査",
  exam: "検査",
  検査: "検査",
  vaccination: "予防接種",
  vaccinations: "予防接種",
  予防接種: "予防接種",
  estimate: "見積書",
  estimates: "見積書",
  見積書: "見積書",
  interview: "問診",
  問診: "問診",
  treatment: "治療",
  治療: "治療",
  image: "画像",
  images: "画像",
  画像: "画像",
  accounting: "会計(医師確認)",
  "会計(医師確認)": "会計(医師確認)",
  "診察/治療プラン": "診察/治療プラン",
  plan: "診察/治療プラン",
};

export function initialMedicalRecordTab(raw: string | null | undefined): string {
  if (!raw) return "問診";
  const aliased = MEDICAL_RECORD_TAB_ALIASES[raw];
  if (aliased) return aliased;
  return MEDICAL_RECORD_TABS.includes(raw as (typeof MEDICAL_RECORD_TABS)[number]) ? raw : "問診";
}

export type MedicalRecordFormGate =
  | { kind: "read-loading" }
  | { kind: "not-found" }
  | { kind: "read-error"; retryRead?: () => void }
  | { kind: "pet-loading" }
  | { kind: "missing-pet" }
  | { kind: "empty" };

export function resolveMedicalRecordFormGate(input: {
  isReadLoading: boolean;
  notFound: boolean;
  isReadNotFound: boolean;
  isReadError: boolean;
  retryRead?: () => void;
  isPetLoading: boolean;
  isNewRecord: boolean;
  hasSelectedPet: boolean;
}): MedicalRecordFormGate | null {
  if (input.isReadLoading) return { kind: "read-loading" };
  if (input.notFound || input.isReadNotFound) return { kind: "not-found" };
  if (input.isReadError) return { kind: "read-error", retryRead: input.retryRead };
  if (input.isPetLoading) return { kind: "pet-loading" };
  if (!input.hasSelectedPet) {
    return input.isNewRecord ? { kind: "empty" } : { kind: "missing-pet" };
  }
  return null;
}
