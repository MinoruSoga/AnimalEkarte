import { toJSTWallDate } from "@/lib/jst-date";

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
