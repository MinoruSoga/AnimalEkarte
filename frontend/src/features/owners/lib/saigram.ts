export interface SaigramResult {
  kosei: number;
  typeNo: number;
  typeCode: string;
  ha: "☆楽観派" | "★慎重派";
}

const KOSEI_TO_TYPE = [
  5, 4, 8, 6, 3, 11, 5, 4, 8, 6, 1, 7, 9, 2, 8, 6, 1, 7, 9, 2, 12, 12, 2, 9, 9, 2, 12, 12, 2, 9, 7,
  1, 6, 8, 2, 9, 7, 1, 6, 8, 4, 5, 11, 3, 6, 8, 4, 5, 11, 3, 10, 10, 3, 11, 11, 3, 10, 10, 3, 11,
] as const;

const MS_PER_DAY = 86_400_000;
const BASE_DATE_MS = Date.UTC(1999, 10, 8);
const BIRTH_DATE_PATTERN = /^(\d{4})-(\d{2})-(\d{2})$/;

export function calculateSaigram(birthDate: string): SaigramResult | null {
  const match = BIRTH_DATE_PATTERN.exec(birthDate);
  if (!match) return null;

  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const date = new Date(0);
  date.setUTCHours(0, 0, 0, 0);
  date.setUTCFullYear(year, month - 1, day);

  if (
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month - 1 ||
    date.getUTCDate() !== day
  ) {
    return null;
  }

  const diffDays = (date.getTime() - BASE_DATE_MS) / MS_PER_DAY;
  const kosei = (((diffDays % 60) + 60) % 60) + 1;
  const typeNo = KOSEI_TO_TYPE[kosei - 1];
  const category = typeNo <= 4 ? "F" : typeNo <= 8 ? "A" : "M";

  return {
    kosei,
    typeNo,
    typeCode: `${category}${typeNo}`,
    ha: typeNo % 2 === 1 ? "☆楽観派" : "★慎重派",
  };
}
