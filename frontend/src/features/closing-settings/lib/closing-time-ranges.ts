/**
 * 締め時間設定（AM 開始 / 午前・午後区切り / 終了時刻）から AM / PM / EMG の
 * 表示用レンジを導出する純粋ユーティリティ（Issue #151 / #215）。
 *
 * 導出ルール（#151 で確定・#215 で越日対応）:
 *   AM : amStart 〜 (boundary − 1秒)      … amStart 既定 09:00
 *   PM : boundary 〜 (end − 1秒)
 *   EMG: end 〜 翌 (amStart − 1秒)        … 越日レンジ（例: 18:30:00～翌08:59:59）。
 *        amStart が 00:00 のときのみ同日 23:59:59 終点（越日しない）。
 *
 * 深夜 0:00〜amStart の会計は前日の EMG に帰属する（BE resolvePeriodRange と同一規則）。
 */

/** 単一の時間帯レンジ。start は "HH:MM:SS"、end は "HH:MM:SS" または越日の "翌HH:MM:SS"。 */
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

/** AM 開始時刻の既定値（BE clinic_settings.closing_am_start の DB default と一致）。 */
export const DEFAULT_CLOSING_AM_START = "09:00";

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
 * @param amStart  AM 開始時刻 ("HH:MM" / "HH:MM:SS")。省略時は 09:00（#215）
 */
export function computeClosingTimeRanges(
  boundary: string,
  end: string,
  amStart: string = DEFAULT_CLOSING_AM_START,
): ClosingTimeRanges {
  const boundarySec = parseTimeToSeconds(boundary);
  const endSec = parseTimeToSeconds(end);
  const amStartSec = parseTimeToSeconds(amStart);

  // amStart が boundary 以上（逆転・同値）の場合 AM は成立しない。
  const am =
    amStartSec !== null && boundarySec !== null && boundarySec - 1 >= amStartSec
      ? { start: formatSeconds(amStartSec), end: formatSeconds(boundarySec - 1) }
      : null;

  // end が boundary 以下（逆転・同値）の場合 PM は成立しない。
  const pm =
    boundarySec !== null && endSec !== null && endSec - 1 >= boundarySec
      ? { start: formatSeconds(boundarySec), end: formatSeconds(endSec - 1) }
      : null;

  // EMG は end〜翌 amStart−1秒の越日レンジ（#215）。amStart=00:00 のときのみ同日終点。
  const emg =
    endSec !== null && amStartSec !== null
      ? {
          start: formatSeconds(endSec),
          end:
            amStartSec >= 1
              ? `翌${formatSeconds(amStartSec - 1)}`
              : formatSeconds(END_OF_DAY_SECONDS),
        }
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
