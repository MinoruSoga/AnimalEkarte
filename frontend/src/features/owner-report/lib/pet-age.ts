import { calcAgePartsAt } from "@/lib/calc-age";

/**
 * #158 飼主レポート: 生年月日から「N歳Mヶ月」表記を算出する純関数。
 *
 * レガシー EMR（Figma 37:142）はペット情報に誕生日と並べて「9オ6ヶ月」を表示する。
 * これは保存値ではなく birthDate からの導出値なので、捏造ではなく算出で再現する。
 *
 * - birthDate は "YYYY-MM-DD" もしくは "YYYY-MM-DDT..."（ISO）を受け付ける。
 * - 不正な値・未来日付は null を返す（表示側で非表示にする）。
 * - 年月の計算部は calcAgePartsAt（lib/calc-age）に委譲（FE3-9）。バリデーション・
 *   未来日ガード・「N歳Mヶ月」表記はこちらに残す。
 * - 基準日 now の既定値は生の `new Date()`。JST 変換は calcAgePartsAt が内部で 1 回だけ
 *   行うため、ここで事前に toJSTWallDate すると二重変換になり時刻帯によって日付がずれる。
 */
export function formatPetAge(birthDate: string, now: Date = new Date()): string | null {
  const [y, m, d] = birthDate.split("T")[0].split("-").map(Number);
  if (
    !Number.isInteger(y) ||
    !Number.isInteger(m) ||
    !Number.isInteger(d) ||
    m < 1 ||
    m > 12 ||
    d < 1 ||
    d > 31
  ) {
    return null;
  }

  const { years, months } = calcAgePartsAt(birthDate, now);
  if (years < 0) {
    return null;
  }

  return `${years}歳${months}ヶ月`;
}
