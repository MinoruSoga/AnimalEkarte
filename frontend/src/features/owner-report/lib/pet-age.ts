import { toJSTWallDate } from "@/lib/jst-date";

/**
 * #158 飼主レポート: 生年月日から「N歳Mヶ月」表記を算出する純関数。
 *
 * レガシー EMR（Figma 37:142）はペット情報に誕生日と並べて「9オ6ヶ月」を表示する。
 * これは保存値ではなく birthDate からの導出値なので、捏造ではなく算出で再現する。
 *
 * - birthDate は "YYYY-MM-DD" もしくは "YYYY-MM-DDT..."（ISO）を受け付ける。
 * - 不正な値・未来日付は null を返す（表示側で非表示にする）。
 * - タイムゾーン差で月境界がずれないよう、birthDate は数値分解して比較する。
 * - 基準日 now の既定値は JST 壁時計（toJSTWallDate）。ブラウザのローカル TZ に依存して
 *   月境界で 1 日ずれるのを防ぐ（運用は JST 固定）。
 */
export function formatPetAge(birthDate: string, now: Date = toJSTWallDate(new Date())): string | null {
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

  let years = now.getFullYear() - y;
  let months = now.getMonth() + 1 - m;
  if (now.getDate() < d) {
    months -= 1;
  }
  if (months < 0) {
    years -= 1;
    months += 12;
  }
  if (years < 0) {
    return null;
  }

  return `${years}歳${months}ヶ月`;
}
