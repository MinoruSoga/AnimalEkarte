/**
 * 締め時間設定（午前・午後区切り / 終了時刻）から AM / PM / EMG の
 * 表示用レンジを導出する純粋ユーティリティ（Issue #151）。
 *
 * 導出ルール（#151 本文で確定）:
 *   AM : 00:00:00 〜 (boundary − 1秒)
 *   PM : boundary  〜 (end − 1秒)
 *   EMG: end       〜 23:59:59（同日内）
 *
 * クライアント仕様の「9:00 開始フロア / 越日 EMG (18:30→翌08:59)」の完全再現は
 * AM 開始時刻フィールド (#156) 依存のため本ユーティリティの対象外。本表示は
 * 同日内レンジ（00:00 始点 / 23:59:59 終点）で充足する。
 */

/** 単一の時間帯レンジ。値はいずれも "HH:MM:SS"。 */
export interface ClosingTimeRange {
  start: string;
  end: string;
}

/** 締め境界設定から導出した AM / PM / EMG レンジ。 */
export interface ClosingTimeRanges {
  am: ClosingTimeRange | null;
  pm: ClosingTimeRange | null;
  emg: ClosingTimeRange | null;
}

/** 値が未設定・不正・逆転しているときに表示するフォールバック文言。 */
export const UNSET_RANGE_LABEL = "未設定";

const SECONDS_PER_DAY = 86_400;
const END_OF_DAY_SECONDS = SECONDS_PER_DAY - 1; // 23:59:59
const TIME_PATTERN = /^(\d{1,2}):(\d{2})(?::(\d{2}))?$/;

const pad2 = (value: number): string => String(value).padStart(2, "0");

/**
 * "HH:MM" / "HH:MM:SS" を 0 時起点の秒数へ変換する。
 * 形式不正・範囲外は null を返し、呼び出し側で安全にフォールバックできるようにする。
 */
function parseTimeToSeconds(value: string): number | null {
  const match = TIME_PATTERN.exec(value.trim());
  if (!match) return null;
  const hours = Number(match[1]);
  const minutes = Number(match[2]);
  const seconds = match[3] === undefined ? 0 : Number(match[3]);
  if (hours > 23 || minutes > 59 || seconds > 59) return null;
  return hours * 3600 + minutes * 60 + seconds;
}

/** 0 時起点の秒数を "HH:MM:SS" に整形する。 */
function formatSeconds(total: number): string {
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const seconds = total % 60;
  return `${pad2(hours)}:${pad2(minutes)}:${pad2(seconds)}`;
}

/**
 * 締め境界時刻と終了時刻から AM / PM / EMG レンジを導出する。
 * 各セグメントは必要な入力が欠落・不正・逆転している場合に null を返し、
 * 誤解を招く部分レンジを描画しない（呼び出し側は UNSET_RANGE_LABEL を表示する）。
 *
 * @param boundary 午前・午後の区切り時刻 ("HH:MM" / "HH:MM:SS")
 * @param end      対象曜日の終了時刻 ("HH:MM" / "HH:MM:SS")
 */
export function computeClosingTimeRanges(boundary: string, end: string): ClosingTimeRanges {
  const boundarySec = parseTimeToSeconds(boundary);
  const endSec = parseTimeToSeconds(end);

  const am =
    boundarySec !== null && boundarySec >= 1
      ? { start: formatSeconds(0), end: formatSeconds(boundarySec - 1) }
      : null;

  // end が boundary 以下（逆転・同値）の場合 PM は成立しない。
  const pm =
    boundarySec !== null && endSec !== null && endSec - 1 >= boundarySec
      ? { start: formatSeconds(boundarySec), end: formatSeconds(endSec - 1) }
      : null;

  const emg =
    endSec !== null
      ? { start: formatSeconds(endSec), end: formatSeconds(END_OF_DAY_SECONDS) }
      : null;

  return { am, pm, emg };
}

/** レンジを "HH:MM:SS～HH:MM:SS" に整形する。null は UNSET_RANGE_LABEL を返す。 */
export function formatRangeText(range: ClosingTimeRange | null): string {
  if (range === null) return UNSET_RANGE_LABEL;
  return `${range.start}～${range.end}`;
}

/**
 * バックエンドの time 値 ("HH:MM:SS") を <input type="time"> 用の "HH:MM" に正規化する。
 * 形式不正は空文字（未入力）を返す。
 */
export function toTimeInputValue(value: string): string {
  const seconds = parseTimeToSeconds(value);
  if (seconds === null) return "";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${pad2(hours)}:${pad2(minutes)}`;
}
