import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { formatJSTWallDate, formatJSTWallTime, todayJSTISO } from "@/lib/jst-date";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { usePermission } from "@/hooks/use-permission";
import type { Reservation } from "@/lib/transforms/reservation";

import { nextVisitFetchEndDate } from "../lib/next-visit";
import { useNextVisitIntent } from "./use-next-visit-intent";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(),
}));

/** JST wall-clock の Date を days 後に作る（transformReservation の start と同じ規約） */
function wallDateInDays(days: number, hour = 10, minute = 0): Date {
  const d = new Date();
  d.setDate(d.getDate() + days);
  d.setHours(hour, minute, 0, 0);
  return d;
}

/** バリデーションを通る未来日（2ヶ月後）を YYYY-MM-DD で返す */
function validFutureDate(): string {
  const d = new Date();
  d.setMonth(d.getMonth() + 2);
  return formatJSTWallDate(d);
}

/** テスト用の最小予約。deriveNextVisitState が読む status/start のみを持つ。 */
function reservationOf(status: string, start: Date): Reservation {
  return { status, start } as unknown as Reservation;
}

function mockReservations(items: Reservation[]) {
  vi.mocked(useGetReservations).mockReturnValue({
    data: items,
  } as unknown as ReturnType<typeof useGetReservations>);
}

interface RenderParams {
  petId?: string;
  medicalRecordId?: string;
  isPetDeceased?: boolean;
}

function renderIntent({ petId = "10", medicalRecordId, isPetDeceased = false }: RenderParams = {}) {
  const setIntentMutation = { mutateAsync: vi.fn().mockResolvedValue({}) };
  const clearIntentMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
  const hook = renderHook(() =>
    useNextVisitIntent({
      petId,
      medicalRecordId,
      isPetDeceased,
      setIntentMutation,
      clearIntentMutation,
    }),
  );
  return { ...hook, setIntentMutation, clearIntentMutation };
}

beforeEach(() => {
  vi.clearAllMocks();
  mockReservations([]);
  vi.mocked(usePermission).mockReturnValue({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: false,
  });
});

describe("useNextVisitIntent — 表示導出（EMR-308 A1）", () => {
  it("useGetReservations を petId + 今日〜2年レンジで呼ぶ", () => {
    renderIntent();

    expect(useGetReservations).toHaveBeenCalledWith({
      petId: "10",
      startDate: todayJSTISO(),
      endDate: nextVisitFetchEndDate(),
      enabled: true,
    });
  });

  it("未来の pending 実予約は日時表示", () => {
    const start = wallDateInDays(30, 14, 30);
    mockReservations([reservationOf("pending", start)]);
    const { result } = renderIntent();

    expect(result.current.nextVisitLabel).toBe(
      `${formatJSTWallDate(start)} ${formatJSTWallTime(start)}`,
    );
  });

  it("tentative のみなら意図日を日付のみで表示し、Popover の入力値にも入る", () => {
    const start = wallDateInDays(45, 0, 0);
    mockReservations([reservationOf("tentative", start)]);
    const { result } = renderIntent();

    expect(result.current.nextVisitLabel).toBe(formatJSTWallDate(start));
    expect(result.current.intentFieldValue).toBe(formatJSTWallDate(start));
  });

  it("実予約 + tentative では実予約優先、intentFieldValue は tentative の日付", () => {
    const appointment = wallDateInDays(30, 14, 30);
    const tentative = wallDateInDays(45, 0, 0);
    mockReservations([
      reservationOf("tentative", tentative),
      reservationOf("confirmed", appointment),
    ]);
    const { result } = renderIntent();

    expect(result.current.nextVisitLabel).toBe(
      `${formatJSTWallDate(appointment)} ${formatJSTWallTime(appointment)}`,
    );
    expect(result.current.intentFieldValue).toBe(formatJSTWallDate(tentative));
  });

  it("過去の実予約のみなら未設定", () => {
    mockReservations([reservationOf("confirmed", wallDateInDays(-3, 10, 0))]);
    const { result } = renderIntent();

    expect(result.current.nextVisitLabel).toBe("");
  });
});

describe("useNextVisitIntent — tentative mutation（EMR-308 A2）", () => {
  it("日付変更は PUT 経路（intended_date + source_medical_record_id）を呼ぶ", async () => {
    const { result, setIntentMutation, clearIntentMutation } = renderIntent({
      medicalRecordId: "99",
    });
    const intendedDate = validFutureDate();

    await act(async () => {
      result.current.handleIntentDateChange(intendedDate);
    });

    expect(setIntentMutation.mutateAsync).toHaveBeenCalledWith({
      petId: "10",
      intendedDate,
      medicalRecordId: "99",
    });
    expect(clearIntentMutation.mutateAsync).not.toHaveBeenCalled();
    await waitFor(() => expect(result.current.intentFieldValue).toBe(intendedDate));
  });

  it("medicalRecordId なしでも PUT 経路を呼ぶ（source_medical_record_id は省略される）", async () => {
    const { result, setIntentMutation } = renderIntent();
    const intendedDate = validFutureDate();

    await act(async () => {
      result.current.handleIntentDateChange(intendedDate);
    });

    expect(setIntentMutation.mutateAsync).toHaveBeenCalledWith({
      petId: "10",
      intendedDate,
      medicalRecordId: undefined,
    });
  });

  it("空文字への変更は DELETE 経路（petId 指定）を呼び PUT は呼ばない", async () => {
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0))]);
    const { result, setIntentMutation, clearIntentMutation } = renderIntent();

    await act(async () => {
      result.current.handleIntentDateChange("");
    });

    expect(clearIntentMutation.mutateAsync).toHaveBeenCalledWith("10");
    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("無効日（過去日）は入力表示だけ残し mutation しない", async () => {
    const { result, setIntentMutation, clearIntentMutation } = renderIntent();

    await act(async () => {
      result.current.handleIntentDateChange("2000-01-01");
    });

    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(clearIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(result.current.intentFieldValue).toBe("2000-01-01");
  });

  it("resetIntentDraft で入力表示をサーバー値へ戻す", async () => {
    mockReservations([reservationOf("tentative", wallDateInDays(45, 0, 0))]);
    const { result } = renderIntent();

    await act(async () => {
      result.current.handleIntentDateChange("2000-01-01");
    });
    act(() => {
      result.current.resetIntentDraft();
    });

    expect(result.current.intentFieldValue).toBe(formatJSTWallDate(wallDateInDays(45)));
  });
});

describe("useNextVisitIntent — ガード", () => {
  it("reservations:create 権限が無い場合 canMutateIntent=false で mutation しない", async () => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: false,
      canEdit: true,
      canDelete: false,
    });
    const { result, setIntentMutation, clearIntentMutation } = renderIntent();

    await act(async () => {
      result.current.handleIntentDateChange(validFutureDate());
      result.current.handleIntentDateChange("");
    });

    expect(result.current.canMutateIntent).toBe(false);
    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(clearIntentMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("死亡ペットは canMutateIntent=false で mutation しない", async () => {
    const { result, setIntentMutation, clearIntentMutation } = renderIntent({
      isPetDeceased: true,
    });

    await act(async () => {
      result.current.handleIntentDateChange(validFutureDate());
      result.current.handleIntentDateChange("");
    });

    expect(result.current.canMutateIntent).toBe(false);
    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(clearIntentMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("set mutation が失敗した場合はドラフトを破棄してサーバー値に戻す", async () => {
    const tentative = wallDateInDays(45, 0, 0);
    mockReservations([reservationOf("tentative", tentative)]);
    const setIntentMutation = { mutateAsync: vi.fn().mockRejectedValue(new Error("500")) };
    const { result } = renderHook(() =>
      useNextVisitIntent({
        petId: "10",
        isPetDeceased: false,
        setIntentMutation,
        clearIntentMutation: { mutateAsync: vi.fn().mockResolvedValue(undefined) },
      }),
    );

    await act(async () => {
      result.current.handleIntentDateChange(validFutureDate());
    });

    await waitFor(() => expect(result.current.intentFieldValue).toBe(formatJSTWallDate(tentative)));
  });
});
