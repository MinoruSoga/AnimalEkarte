import { describe, expect, it, vi } from "vitest";

import { createReceptionAppointmentTimeRange } from "./use-medical-record-form-model";

// FE-RC-027: JST 時刻計算を lib/jst-date.ts の共通ヘルパー（formatJSTTime 等）に
// 委譲するようリファクタリングした際の回帰防止。実行環境の tz に関わらず
// 常に +09:00 オフセットの絶対時刻を返すことを固定する。
describe("createReceptionAppointmentTimeRange FE-RC-027 JST 委譲", () => {
  it("visitDate 指定時、現在時刻の JST 時:分を採用し +09:00 オフセットで返す", () => {
    // 2026-07-20T01:23:00Z = JST 10:23
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-20T01:23:00.000Z"));

    const { startTime, endTime } = createReceptionAppointmentTimeRange(15, "2026-07-25");

    expect(startTime).toBe("2026-07-25T10:23:00+09:00");
    expect(endTime).toBe("2026-07-25T10:38:00+09:00");

    vi.useRealTimers();
  });

  it("visitDate 未指定時は現在時刻そのものを起点に duration 分後を返す", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-20T01:00:00.000Z"));

    const { startTime, endTime } = createReceptionAppointmentTimeRange(30);

    // JST 10:00 開始、30分後は 10:30
    expect(startTime).toBe("2026-07-20T10:00:00+09:00");
    expect(endTime).toBe("2026-07-20T10:30:00+09:00");

    vi.useRealTimers();
  });
});
