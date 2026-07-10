import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { calculateNextDate, useVaccinationForm } from "./use-vaccination-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetVaccination } from "../api/get-vaccination";
import { useCreateVaccination } from "../api/create-vaccination";
import { useUpdateVaccination } from "../api/update-vaccination";
import { useDeleteVaccination } from "../api/delete-vaccination";

// ──────────────────────────────────────────────────────────
// モック定義
// ──────────────────────────────────────────────────────────

const mockNavigate = vi.fn();
let mockSearchParams = new URLSearchParams();

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams, vi.fn()],
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({ data: undefined, isLoading: false })),
}));
vi.mock("../api/get-vaccination", () => ({
  useGetVaccination: vi.fn(() => ({ data: undefined })),
}));
vi.mock("../api/create-vaccination", () => ({
  useCreateVaccination: vi.fn(() => ({ mutateAsync: vi.fn().mockResolvedValue({}), isPending: false })),
}));
vi.mock("../api/update-vaccination", () => ({
  useUpdateVaccination: vi.fn(() => ({ mutateAsync: vi.fn().mockResolvedValue({}), isPending: false })),
}));
vi.mock("../api/delete-vaccination", () => ({
  useDeleteVaccination: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

function runFormAction(action: (payload: FormData) => void) {
  act(() => {
    startTransition(() => {
      action(new FormData());
    });
  });
}

// ──────────────────────────────────────────────────────────
// calculateNextDate（BUG-026 回帰）
// ──────────────────────────────────────────────────────────

describe("calculateNextDate (BUG-026 回帰)", () => {
  const cases: Array<{
    name: string;
    date: string;
    scheduleType: string;
    expected: string;
  }> = [
    { name: "3weeks: 接種日から3週間後を返す", date: "2026-01-01", scheduleType: "3weeks", expected: "2026-01-22" },
    { name: "4weeks: 接種日から4週間後を返す", date: "2026-01-01", scheduleType: "4weeks", expected: "2026-01-29" },
    { name: "1year: 接種日から1年後を返す", date: "2026-01-01", scheduleType: "1year", expected: "2027-01-01" },
    { name: "other: 空文字を返す（手入力に委ねる）", date: "2026-01-01", scheduleType: "other", expected: "" },
    { name: "不明な scheduleType: 空文字を返す（default分岐）", date: "2026-01-01", scheduleType: "unknown", expected: "" },
    { name: "不正な日付文字列: 空文字を返す（isNaN ガード）", date: "not-a-date", scheduleType: "1year", expected: "" },
    { name: "空の接種日: 空文字を返す（早期リターン）", date: "", scheduleType: "1year", expected: "" },
    {
      name: "うるう年境界: 2024-02-29 + 1year は 2025-02-28 に丸められる（date-fns addYears 仕様）",
      date: "2024-02-29",
      scheduleType: "1year",
      expected: "2025-02-28",
    },
    {
      name: "うるう年境界: 2024-02-29 + 4weeks は月をまたいで 2024-03-28 になる",
      date: "2024-02-29",
      scheduleType: "4weeks",
      expected: "2024-03-28",
    },
  ];

  it.each(cases)("$name", ({ date, scheduleType, expected }) => {
    expect(calculateNextDate(date, scheduleType)).toBe(expected);
  });
});

// ──────────────────────────────────────────────────────────
// useVaccinationForm — バリデーション（BUG-024/074/096）
// ──────────────────────────────────────────────────────────

describe("useVaccinationForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
    // Date のみ偽装する。setTimeout/setInterval を実タイマーに残さないと
    // waitFor の内部ポーリングが進まず全 async テストが 5000ms でタイムアウトする。
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-07-10T01:00:00.000Z")); // JST 2026-07-10 10:00
    vi.mocked(useGetPet).mockReturnValue({ data: undefined, isLoading: false } as ReturnType<typeof useGetPet>);
    vi.mocked(useGetVaccination).mockReturnValue({ data: undefined } as ReturnType<typeof useGetVaccination>);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe("新規登録時のバリデーション", () => {
    it("BUG-074: vaccineId 未選択 → fieldErrors.vaccineId がセットされる", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      const { result } = renderHook(() => useVaccinationForm());

      act(() => { result.current.form.setDate("2026-07-01"); });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.vaccineId).toBe("ワクチン種別を選択してください");
      });
    });

    it("vaccineId が '0' の場合も未選択扱いでエラーになる", async () => {
      const { result } = renderHook(() => useVaccinationForm());
      act(() => {
        result.current.form.setVaccineId("0");
        result.current.form.setDate("2026-07-01");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.vaccineId).toBe("ワクチン種別を選択してください");
      });
    });

    it("接種日未入力 → fieldErrors.date がセットされる", async () => {
      const { result } = renderHook(() => useVaccinationForm());
      act(() => { result.current.form.setVaccineId("1"); });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.date).toBe("接種日を入力してください");
      });
    });

    it("BUG-024: 接種日が未来日 → fieldErrors.date がセットされる", async () => {
      const { result } = renderHook(() => useVaccinationForm());
      act(() => {
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-11"); // today=2026-07-10 JST
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.date).toBe("接種日は今日以前の日付を入力してください");
      });
    });

    it("BUG-096: 次回予定日が本日より前 → fieldErrors.nextDate がセットされる", async () => {
      const { result } = renderHook(() => useVaccinationForm());
      act(() => {
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-01");
        result.current.form.setNextDate("2026-07-05"); // < today(07-10) だが > date(07-01)
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.nextDate).toBe("次回予定日は本日以降の日付を入力してください");
      });
    });

    it("BUG-024: 次回予定日が接種日以前（同日） → fieldErrors.nextDate がセットされる", async () => {
      const { result } = renderHook(() => useVaccinationForm());
      act(() => {
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-10"); // = today
        result.current.form.setNextDate("2026-07-10"); // = date（本日以降チェックは通過）
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.nextDate).toBe("次回予定日は接種日より後の日付を入力してください");
      });
    });
  });

  describe("編集時のバリデーション", () => {
    it("接種日未入力 → fieldErrors.date がセットされる（vaccineId チェックはスキップ）", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: {
          id: "10",
          petId: "5",
          vaccineId: "1",
          date: "2026-07-01",
          nextDate: "",
          nextScheduleType: "1year",
        },
      } as ReturnType<typeof useGetVaccination>);
      const { result } = renderHook(() => useVaccinationForm("10"));

      act(() => { result.current.form.setDate(""); });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.date).toBe("接種日を入力してください");
        expect(result.current.fieldErrors.vaccineId).toBeUndefined();
      });
    });

    it("BUG-024: 接種日が未来日 → fieldErrors.date がセットされる", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: { id: "10", petId: "5", vaccineId: "1", date: "2026-07-01", nextDate: "", nextScheduleType: "1year" },
      } as ReturnType<typeof useGetVaccination>);
      const { result } = renderHook(() => useVaccinationForm("10"));

      act(() => { result.current.form.setDate("2026-07-11"); });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.date).toBe("接種日は今日以前の日付を入力してください");
      });
    });

    it("BUG-096 は編集時には適用されない（次回予定日が過去日でもエラーにならない）", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: { id: "10", petId: "5", vaccineId: "1", date: "2026-01-01", nextDate: "", nextScheduleType: "1year" },
      } as ReturnType<typeof useGetVaccination>);
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);

      const { result } = renderHook(() => useVaccinationForm("10"));
      act(() => {
        result.current.form.setDate("2026-01-01");
        result.current.form.setNextDate("2026-02-01"); // 過去日だが date より後なので唯一の nextDate 制約は満たす
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.nextDate).toBeUndefined();
      });
    });

    it("BUG-024: 次回予定日が接種日以前 → fieldErrors.nextDate がセットされる", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: { id: "10", petId: "5", vaccineId: "1", date: "2026-07-01", nextDate: "", nextScheduleType: "1year" },
      } as ReturnType<typeof useGetVaccination>);
      const { result } = renderHook(() => useVaccinationForm("10"));

      act(() => {
        result.current.form.setDate("2026-07-05");
        result.current.form.setNextDate("2026-07-01"); // 接種日より前
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.nextDate).toBe("次回予定日は接種日より後の日付を入力してください");
      });
    });

    it("正常系: バリデーション通過 → fieldErrors が空になり updateMutation が呼ばれる", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: { id: "10", petId: "5", vaccineId: "1", date: "2026-07-01", nextDate: "", nextScheduleType: "1year" },
      } as ReturnType<typeof useGetVaccination>);
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);

      const { result } = renderHook(() => useVaccinationForm("10"));
      act(() => {
        result.current.form.setDate("2026-07-05");
        result.current.form.setNextDate("2026-07-20");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });
      expect(result.current.fieldErrors).toEqual({});
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ id: "10" })
      );
    });
  });

  // ──────────────────────────
  // BUG-026: ワクチン種別/接種日変更時の次回予定日自動計算
  // ──────────────────────────
  describe("次回予定日の自動計算（BUG-026）", () => {
    it("setVaccineId で狂犬病ワクチン(id=2)を選択 → nextScheduleType=1year かつ nextDate が自動計算される", async () => {
      const { result } = renderHook(() => useVaccinationForm());
      act(() => { result.current.form.setDate("2026-07-01"); });
      await waitFor(() => expect(result.current.form.date).toBe("2026-07-01"));

      act(() => { result.current.form.setVaccineId("2"); });

      await waitFor(() => {
        expect(result.current.form.nextScheduleType).toBe("1year");
        expect(result.current.form.nextDate).toBe("2027-07-01");
      });
    });

    it("setDate で接種日を変更 → 現在の nextScheduleType に基づき nextDate が再計算される", async () => {
      const { result } = renderHook(() => useVaccinationForm());
      act(() => { result.current.form.setNextScheduleType("3weeks"); });
      await waitFor(() => expect(result.current.form.nextScheduleType).toBe("3weeks"));

      act(() => { result.current.form.setDate("2026-07-01"); });

      await waitFor(() => {
        expect(result.current.form.nextDate).toBe("2026-07-22");
      });
    });

    it("setNextScheduleType で種別を変更 → 既存の接種日を基準に nextDate が再計算される", async () => {
      const { result } = renderHook(() => useVaccinationForm());
      act(() => { result.current.form.setDate("2026-07-01"); });
      await waitFor(() => expect(result.current.form.date).toBe("2026-07-01"));

      act(() => { result.current.form.setNextScheduleType("4weeks"); });

      await waitFor(() => {
        expect(result.current.form.nextDate).toBe("2026-07-29");
      });
    });
  });

  // ──────────────────────────
  // BUG-025: 削除ハンドラ
  // ──────────────────────────
  describe("handleDelete (BUG-025)", () => {
    it("編集時、deleteVaccinationFn が呼ばれ成功時に onSuccess コールバックが実行される", () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: { id: "10", petId: "5", vaccineId: "1", date: "2026-07-01", nextDate: "", nextScheduleType: "1year" },
      } as ReturnType<typeof useGetVaccination>);
      const mockMutate = vi.fn((_id: string, opts?: { onSuccess?: () => void }) => opts?.onSuccess?.());
      vi.mocked(useDeleteVaccination).mockReturnValue({
        mutate: mockMutate,
        isPending: false,
      } as ReturnType<typeof useDeleteVaccination>);

      const { result } = renderHook(() => useVaccinationForm("10"));
      const onSuccess = vi.fn();
      act(() => { result.current.handleDelete(onSuccess); });

      expect(mockMutate).toHaveBeenCalledWith("10", expect.objectContaining({
        onSuccess: expect.any(Function),
        onError: expect.any(Function),
      }));
      expect(onSuccess).toHaveBeenCalled();
    });

    it("新規作成時（id なし）は何もしない", () => {
      const mockMutate = vi.fn();
      vi.mocked(useDeleteVaccination).mockReturnValue({
        mutate: mockMutate,
        isPending: false,
      } as ReturnType<typeof useDeleteVaccination>);

      const { result } = renderHook(() => useVaccinationForm());
      act(() => { result.current.handleDelete(); });

      expect(mockMutate).not.toHaveBeenCalled();
    });
  });
});
