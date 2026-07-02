import { describe, it, expect } from "vitest";

import {
  computeClosingTimeRanges,
  formatRangeText,
  toTimeInputValue,
  UNSET_RANGE_LABEL,
  DEFAULT_CLOSING_AM_START,
} from "./closing-time-ranges";

describe("computeClosingTimeRanges (#151 AM/PM/EMG 導出・#215 越日対応)", () => {
  it("平日: 境界 12:00 / 終了 18:30 / amStart 既定 9:00 から AM・PM・EMG を導出する", () => {
    // Arrange / Act
    const ranges = computeClosingTimeRanges("12:00", "18:30");

    // Assert — #215: AM は 9:00 開始、EMG は翌 amStart−1秒 までの越日レンジ
    expect(ranges.am).toEqual({ start: "09:00:00", end: "11:59:59" });
    expect(ranges.pm).toEqual({ start: "12:00:00", end: "18:29:59" });
    expect(ranges.emg).toEqual({ start: "18:30:00", end: "翌08:59:59" });
  });

  it("PM 終了は終了時刻の 1 秒前である（クライアント例の :29 は誤記、ルールは −1秒=:59）", () => {
    const ranges = computeClosingTimeRanges("12:00", "18:30");
    expect(ranges.pm?.end).toBe("18:29:59");
    expect(ranges.pm?.end).not.toBe("18:29:29");
  });

  it("EMG は越日（end〜翌 amStart−1秒）で表示する（#215 で同日内表示から変更）", () => {
    const ranges = computeClosingTimeRanges("12:00", "18:30");
    expect(ranges.emg?.end).toBe("翌08:59:59");
    expect(ranges.emg?.end).not.toBe("23:59:59");
  });

  it("amStart 明示指定: 08:00 なら AM は 08:00 開始・EMG は翌07:59:59 終点", () => {
    const ranges = computeClosingTimeRanges("12:00", "18:30", "08:00");
    expect(ranges.am).toEqual({ start: "08:00:00", end: "11:59:59" });
    expect(ranges.emg).toEqual({ start: "18:30:00", end: "翌07:59:59" });
  });

  it("amStart が 00:00 のときのみ EMG は同日 23:59:59 終点（越日しない）", () => {
    const ranges = computeClosingTimeRanges("12:00", "18:30", "00:00");
    expect(ranges.am).toEqual({ start: "00:00:00", end: "11:59:59" });
    expect(ranges.emg).toEqual({ start: "18:30:00", end: "23:59:59" });
  });

  it("amStart が境界以上（逆転設定）の場合 AM は null", () => {
    const ranges = computeClosingTimeRanges("09:00", "18:30", "09:00");
    expect(ranges.am).toBeNull();
  });

  it("日曜: 終了 17:30 で PM・EMG が日曜終了時刻に追従する", () => {
    const ranges = computeClosingTimeRanges("12:00", "17:30");
    expect(ranges.pm).toEqual({ start: "12:00:00", end: "17:29:59" });
    expect(ranges.emg).toEqual({ start: "17:30:00", end: "翌08:59:59" });
  });

  it("バックエンドの time 形式 'HH:MM:SS' を受け付ける", () => {
    const ranges = computeClosingTimeRanges("12:00:00", "18:30:00", "09:00:00");
    expect(ranges.am).toEqual({ start: "09:00:00", end: "11:59:59" });
    expect(ranges.pm).toEqual({ start: "12:00:00", end: "18:29:59" });
  });

  it("境界が空文字なら AM・PM は null（部分レンジを描画しない）。EMG は終了と amStart から導出", () => {
    const ranges = computeClosingTimeRanges("", "18:30");
    expect(ranges.am).toBeNull();
    expect(ranges.pm).toBeNull();
    expect(ranges.emg).toEqual({ start: "18:30:00", end: "翌08:59:59" });
  });

  it("終了が空文字なら PM・EMG は null。AM は境界と amStart から導出", () => {
    const ranges = computeClosingTimeRanges("12:00", "");
    expect(ranges.am).toEqual({ start: "09:00:00", end: "11:59:59" });
    expect(ranges.pm).toBeNull();
    expect(ranges.emg).toBeNull();
  });

  it("amStart が不正な形式なら AM・EMG は null（PM は境界と終了から導出）", () => {
    const ranges = computeClosingTimeRanges("12:00", "18:30", "9時");
    expect(ranges.am).toBeNull();
    expect(ranges.pm).toEqual({ start: "12:00:00", end: "18:29:59" });
    expect(ranges.emg).toBeNull();
  });

  it("終了が境界以下（逆転設定）の場合、誤解を招く反転レンジを出さず PM は null", () => {
    const ranges = computeClosingTimeRanges("18:00", "12:00");
    expect(ranges.pm).toBeNull();
  });

  it("境界が 00:00 の場合は AM レンジを成立させない", () => {
    const ranges = computeClosingTimeRanges("00:00", "18:30");
    expect(ranges.am).toBeNull();
  });

  it("不正な形式は全セグメント null", () => {
    const ranges = computeClosingTimeRanges("25:99", "abc");
    expect(ranges).toEqual({ am: null, pm: null, emg: null });
  });

  it("既定 amStart は BE の DB default と同じ 09:00 である", () => {
    expect(DEFAULT_CLOSING_AM_START).toBe("09:00");
  });
});

describe("formatRangeText", () => {
  it("レンジを 'HH:MM:SS～HH:MM:SS' に整形する", () => {
    expect(formatRangeText({ start: "12:00:00", end: "18:29:59" })).toBe("12:00:00～18:29:59");
  });

  it("越日レンジは '翌' 接頭辞つきで整形する（#215）", () => {
    expect(formatRangeText({ start: "18:30:00", end: "翌08:59:59" })).toBe("18:30:00～翌08:59:59");
  });

  it("null は未設定フォールバック文言を返す", () => {
    expect(formatRangeText(null)).toBe(UNSET_RANGE_LABEL);
    expect(formatRangeText(null)).toBe("未設定");
  });
});

describe("toTimeInputValue", () => {
  it("'HH:MM:SS' を <input type=time> 用の 'HH:MM' に正規化する", () => {
    expect(toTimeInputValue("14:00:00")).toBe("14:00");
  });

  it("'HH:MM' はそのまま（ゼロ埋め維持）", () => {
    expect(toTimeInputValue("09:30")).toBe("09:30");
  });

  it("不正値は空文字を返す", () => {
    expect(toTimeInputValue("")).toBe("");
    expect(toTimeInputValue("nope")).toBe("");
  });
});
