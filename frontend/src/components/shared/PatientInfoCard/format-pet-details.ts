import { calcAgePartsAt } from "@/lib/calc-age";

export interface PatientPetDemographics {
  birthDate?: string | null;
  gender?: string | null;
  neuteredDate?: string | null;
}

const UNKNOWN = "不明";

/**
 * PatientInfoCard 用の「年齢 / 性別 / 去勢避妊」1 行表示を組み立てる。
 * 欠損は推測せず「不明」。固定ダミー（旧 9才5ヶ月 / メス / 避妊済）は使わない。
 */
export function formatPatientPetDetails(
  pet: PatientPetDemographics,
  now: Date = new Date(),
): string {
  return [formatAgePart(pet.birthDate, now), formatGenderPart(pet.gender), formatNeuteredPart(pet.gender, pet.neuteredDate)].join(
    " / ",
  );
}

function formatAgePart(birthDate: string | null | undefined, now: Date): string {
  if (!birthDate || !birthDate.trim()) return UNKNOWN;
  const dateOnly = birthDate.split("T")[0];
  const [y, m, d] = dateOnly.split("-").map(Number);
  if (
    !Number.isInteger(y) ||
    !Number.isInteger(m) ||
    !Number.isInteger(d) ||
    m < 1 ||
    m > 12 ||
    d < 1 ||
    d > 31
  ) {
    return UNKNOWN;
  }

  const { years, months } = calcAgePartsAt(dateOnly, now);
  if (years < 0) return UNKNOWN;
  return `${years}歳${months}ヶ月`;
}

function formatGenderPart(gender: string | null | undefined): string {
  const g = gender?.trim();
  if (!g) return UNKNOWN;
  return g;
}

function formatNeuteredPart(
  gender: string | null | undefined,
  neuteredDate: string | null | undefined,
): string {
  if (!neuteredDate || !String(neuteredDate).trim()) return UNKNOWN;
  if (gender === "雄") return "去勢済";
  if (gender === "雌") return "避妊済";
  return "避妊・去勢済";
}
