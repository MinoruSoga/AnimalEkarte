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
  it("実予約 + tentative 同居: create のみ（edit なし）では上書き set を送らない", async () => {
    // Codex PR606 #8: kind==="appointment" でも tentative が同居していれば
    // PUT は既存 intent の上書き（BE: reservations:edit）になる。
    // create-only ユーザーに送らせない — fail-closed。
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: false,
      canDelete: false,
    });
    mockReservations([
      {
        id: "t-1",
        status: "tentative",
        start: wallDateInDays(45, 0, 0),
      } as unknown as Reservation,
      reservationOf("confirmed", wallDateInDays(30, 14, 30)),
    ]);
    const { result, setIntentMutation, clearIntentMutation } = renderIntent();

    await act(async () => {
      result.current.handleIntentDateChange(validFutureDate());
    });

    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(result.current.canMutateIntent).toBe(false);
    expect(result.current.intentFieldValue).not.toBe(validFutureDate());
    // edit なしなので clear も送らない
    await act(async () => {
      result.current.handleIntentDateChange("");
    });
    expect(clearIntentMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("実予約 + tentative 同居: edit のみ（create なし）では上書き set を送らない", async () => {
    // Codex PR606 #8 再指摘: 既存 tentative 上書き PUT はルート reservations:create
    // + repo reservations:edit の両方を要求する。edit 単独では BE が 403 を返すため
    // 送信させない — fail-closed。
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: false,
      canEdit: true,
      canDelete: false,
    });
    mockReservations([
      {
        id: "t-1",
        status: "tentative",
        start: wallDateInDays(45, 0, 0),
      } as unknown as Reservation,
      reservationOf("confirmed", wallDateInDays(30, 14, 30)),
    ]);
    const { result, setIntentMutation, clearIntentMutation } = renderIntent();
    const initialValue = result.current.intentFieldValue;

    await act(async () => {
      result.current.handleIntentDateChange(validFutureDate());
    });

    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(result.current.intentFieldValue).toBe(initialValue);
    // DELETE intent（reservations:edit 単独で可）は引き続き送れる
    await act(async () => {
      result.current.handleIntentDateChange("");
    });
    expect(clearIntentMutation.mutateAsync).toHaveBeenCalledWith("10");
  });

  it("create なし・edit のみ: set（PUT）は送らず clear（DELETE）は送る", async () => {
    // PUT intent は reservations:create、DELETE intent は reservations:edit の
    // 方向別契約（routes.go）。canEdit だけでも intentFieldValue の消去は動く。
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: false,
      canEdit: true,
      canDelete: false,
    });
    const { result, setIntentMutation, clearIntentMutation } = renderIntent();

    await act(async () => {
      result.current.handleIntentDateChange(validFutureDate());
    });
    expect(setIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(result.current.intentFieldValue).toBe("");
    expect(result.current.canSetIntent).toBe(false);
    expect(result.current.canClearIntent).toBe(true);

    await act(async () => {
      result.current.handleIntentDateChange("");
    });
    expect(result.current.canMutateIntent).toBe(true);
    expect(clearIntentMutation.mutateAsync).toHaveBeenCalledWith("10");
  });

  it("edit なし・create のみ: set（PUT）は送り clear（DELETE）は送らない", async () => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: false,
      canDelete: false,
    });
    const { result, setIntentMutation, clearIntentMutation } = renderIntent();

    await act(async () => {
      result.current.handleIntentDateChange(validFutureDate());
    });
    expect(result.current.canMutateIntent).toBe(true);
    expect(setIntentMutation.mutateAsync).toHaveBeenCalled();
    expect(result.current.intentFieldValue).toBe(validFutureDate());

    await act(async () => {
      result.current.handleIntentDateChange("");
    });
    expect(clearIntentMutation.mutateAsync).not.toHaveBeenCalled();
    expect(result.current.intentFieldValue).toBe(validFutureDate());
  });

  it("create/edit 両方無い場合 canMutateIntent=false で mutation しない", async () => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: false,
      canEdit: false,
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

describe("useNextVisitIntent — mutation 直列化（Codex PR606 #11）", () => {
  it("set 実行中に clear を要求すると、set 完了まで clear は発行されない", async () => {
    let resolveSet!: () => void;
    const setIntentMutation = {
      mutateAsync: vi.fn().mockImplementation(
        () =>
          new Promise<void>((resolve) => {
            resolveSet = resolve;
          }),
      ),
    };
    const clearIntentMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    const { result } = renderHook(() =>
      useNextVisitIntent({
        petId: "10",
        isPetDeceased: false,
        setIntentMutation,
        clearIntentMutation,
      }),
    );

    await act(async () => {
      result.current.handleIntentDateChange(validFutureDate());
    });
    // set が in-flight の間に clear を要求
    await act(async () => {
      result.current.handleIntentDateChange("");
    });
    // FIFO: set が未解決なので clear はまだ発行されない
    expect(clearIntentMutation.mutateAsync).not.toHaveBeenCalled();

    await act(async () => {
      resolveSet();
    });
    await waitFor(() => expect(clearIntentMutation.mutateAsync).toHaveBeenCalledWith("10"));
    // 発行順序: set → clear（逆順完了による stale intentDate 復活を防ぐ）
    const setOrder = setIntentMutation.mutateAsync.mock.invocationCallOrder[0];
    const clearOrder = clearIntentMutation.mutateAsync.mock.invocationCallOrder[0];
    expect(setOrder).toBeLessThan(clearOrder);
  });

  it("先の set が失敗しても後続要求の draft は消さない（generation guard）", async () => {
    const date1 = formatJSTWallDate(wallDateInDays(60, 0, 0));
    const date2 = formatJSTWallDate(wallDateInDays(75, 0, 0));
    const setIntentMutation = {
      mutateAsync: vi.fn().mockRejectedValueOnce(new Error("500")).mockResolvedValueOnce({}),
    };
    const { result } = renderHook(() =>
      useNextVisitIntent({
        petId: "10",
        isPetDeceased: false,
        setIntentMutation,
        clearIntentMutation: { mutateAsync: vi.fn().mockResolvedValue(undefined) },
      }),
    );

    await act(async () => {
      result.current.handleIntentDateChange(date1); // 失敗する
      result.current.handleIntentDateChange(date2); // 直列化されて後で成功
    });

    await waitFor(() => expect(setIntentMutation.mutateAsync).toHaveBeenCalledTimes(2));
    // 先の失敗が最新 draft（date2）を巻き戻さない
    expect(result.current.intentFieldValue).toBe(date2);
  });

  it("先行 mutation の成功 refetch は後続の入力 draft を消さない（pending 中は draft リセット抑制）", async () => {
    // Codex PR606 #11 再指摘: PUT#1 が成功して起こした refetch が date1（先行値）を
    // 返しても、キュー待ちの PUT#2 に対応する手入力 draft（date2）を保持する。
    const date1 = formatJSTWallDate(wallDateInDays(60, 0, 0));
    const date2 = formatJSTWallDate(wallDateInDays(75, 0, 0));

    const resolvers: Array<() => void> = [];
    const setIntentMutation = {
      mutateAsync: vi.fn().mockImplementation(
        () =>
          new Promise<void>((resolve) => {
            resolvers.push(resolve);
          }),
      ),
    };
    const { result, rerender } = renderHook(() =>
      useNextVisitIntent({
        petId: "10",
        isPetDeceased: false,
        setIntentMutation,
        clearIntentMutation: { mutateAsync: vi.fn().mockResolvedValue(undefined) },
      }),
    );

    // PUT#1（date1）発行 → in-flight のままユーザーが date2 を入力
    await act(async () => {
      result.current.handleIntentDateChange(date1);
      result.current.handleIntentDateChange(date2);
    });
    expect(result.current.intentFieldValue).toBe(date2);

    // PUT#1 が成功 → 直列化で PUT#2 発行。その成功に伴う refetch が
    // 「date1 の tentative」を返す（PUT#2 未反映のサーバー値）。
    await act(async () => {
      resolvers[0]();
    });
    await waitFor(() => expect(setIntentMutation.mutateAsync).toHaveBeenCalledTimes(2));

    mockReservations([reservationOf("tentative", wallDateInDays(60, 0, 0))]);
    act(() => {
      rerender();
    });

    // pending の PUT#2 がある限り、先行 refetch の古い値で draft を消さない
    expect(result.current.intentFieldValue).toBe(date2);

    // PUT#2 完了 → 後続 refetch が date2 を返し、draft は通常どおり同期される
    await act(async () => {
      resolvers[1]();
    });
    mockReservations([reservationOf("tentative", wallDateInDays(75, 0, 0))]);
    act(() => {
      rerender();
    });
    expect(result.current.intentFieldValue).toBe(date2);

    // pending 解消後は抑制が残らず、サーバー値の変化で通常どおりリセットされる
    mockReservations([]);
    act(() => {
      rerender();
    });
    expect(result.current.intentFieldValue).toBe("");
  });
});
