import { toJSTWallDate } from "@/lib/jst-date";

/**
 * FE3-9: 年月コンポーネント版の年齢計算。クランプ・未来日ガードなしの生値を返す。
 * 呼び出し元（PatientContextHeader / formatPetAge）は各自の表示フォーマット・ガードを
 * 自分側に残したまま、この計算部だけを共有する。
 * `at`/`birthdate` は本関数の内部で toJSTWallDate による JST 変換を 1 回だけ行う —
 * 呼び出し元で事前に toJSTWallDate 済みの値を渡すと二重変換になるため、生の Date | string を渡すこと。
 */
export function calcAgePartsAt(
  birthdate: Date | string,
  at: Date | string,
): { years: number; months: number } {
  const atWall = toJSTWallDate(at);
  const birthWall = toJSTWallDate(birthdate);

  let years = atWall.getFullYear() - birthWall.getFullYear();
  let months = atWall.getMonth() - birthWall.getMonth();
  if (months < 0) {
    years -= 1;
    months += 12;
  }
  if (atWall.getDate() < birthWall.getDate()) {
    months -= 1;
    if (months < 0) {
      years -= 1;
      months += 12;
    }
  }

  return { years, months };
}

export function calcAgeAt(asOf: Date | string, birthDate: Date | string): number {
  const asOfWall = toJSTWallDate(asOf);
  const birthWall = toJSTWallDate(birthDate);

  let age = asOfWall.getFullYear() - birthWall.getFullYear();
  const hasBirthdayPassed =
    asOfWall.getMonth() > birthWall.getMonth() ||
    (asOfWall.getMonth() === birthWall.getMonth() && asOfWall.getDate() >= birthWall.getDate());
  if (!hasBirthdayPassed) {
    age -= 1;
  }

  return Math.max(0, age);
}
