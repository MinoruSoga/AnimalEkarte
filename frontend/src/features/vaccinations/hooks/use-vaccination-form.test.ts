import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition, useLayoutEffect, useRef } from "react";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { calculateNextDate, useVaccinationForm } from "./use-vaccination-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetAllVaccinesMaster } from "@/hooks/use-treatment-master";
import { useGetVaccination } from "../api/get-vaccination";
import { useCreateVaccination } from "../api/create-vaccination";
import { useUpdateVaccination } from "../api/update-vaccination";
import { useDeleteVaccination } from "../api/delete-vaccination";
import { jstDateStartISOString } from "@/lib/jst-date";

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
// BUG-401: 実マスタ参照化に伴い useGetAllVaccinesMaster をモックする。id="1"/"2" は既存の BUG-026
// 回帰テスト期待値（両方とも 1year）を保つ interval="1年" の固定値。BUG-401 固有の interval マッピング
// テストは個別に mockReturnValueOnce で上書きする。
vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllVaccinesMaster: vi.fn(() => ({
    data: [
      { id: "1", name: "混合ワクチン", isActive: true, interval: "1年" },
      { id: "2", name: "狂犬病ワクチン", isActive: true, interval: "1年" },
    ],
  })),
}));
vi.mock("../api/get-vaccination", () => ({
  useGetVaccination: vi.fn(() => ({
    data: undefined,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  })),
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

const ALLOWED_MUTATION_PERMISSIONS = {
  canCreate: true,
  canEdit: true,
  canDelete: true,
} as const;

const DECEASED_PET = {
  id: "5",
  ownerId: "1",
  name: "ポチ",
  status: "死亡",
} as NonNullable<ReturnType<typeof useGetPet>["data"]>;

function renderVaccinationForm(id?: string) {
  return renderHook(() => useVaccinationForm(id, ALLOWED_MUTATION_PERMISSIONS));
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
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);
    // BUG-401: renderHook 内の act() で複数回再レンダーが起きるため、mockReturnValueOnce だと
    // 2 回目以降の呼び出しでモック実装がデフォルトへ巻き戻ってしまう（1 回限りの upvalue が枯渇する）。
    // beforeEach でテストごとに明示的にデフォルトへ戻し、個別テストは mockReturnValue（永続）で上書きする。
    vi.mocked(useGetAllVaccinesMaster).mockReturnValue({
      data: [
        { id: "1", name: "混合ワクチン", isActive: true, interval: "1年" },
        { id: "2", name: "狂犬病ワクチン", isActive: true, interval: "1年" },
      ],
    } as ReturnType<typeof useGetAllVaccinesMaster>);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  // ──────────────────────────
  // BUG-004: 新規オープン時の接種日デフォルト＝当日（JST）
  // ──────────────────────────
  describe("新規登録時の接種日デフォルト（BUG-004）", () => {
    it("新規オープンで form.date が JST 当日になる", () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      const { result } = renderVaccinationForm();

      expect(result.current.form.date).toBe("2026-07-10");
    });

    it("接種日は手動変更できる（当日デフォルト後に上書き可能）", () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      const { result } = renderVaccinationForm();

      expect(result.current.form.date).toBe("2026-07-10");
      act(() => {
        result.current.form.setDate("2026-07-01");
      });
      expect(result.current.form.date).toBe("2026-07-01");
    });

    it("編集モードでは既存レコードの接種日を使い当日で上書きしない", () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: {
          id: "10",
          petId: "5",
          vaccineId: "1",
          date: "2026-01-15",
          nextDate: "2027-01-15",
          nextScheduleType: "1year",
        },
        isLoading: false,
        isError: false,
        error: null,
        refetch: vi.fn(),
      } as unknown as ReturnType<typeof useGetVaccination>);

      const { result } = renderVaccinationForm("10");

      expect(result.current.form.date).toBe("2026-01-15");
    });
  });

  describe("新規登録時のバリデーション", () => {
    it("BUG-074: vaccineId 未選択 → fieldErrors.vaccineId がセットされる", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      const { result } = renderVaccinationForm();

      act(() => { result.current.form.setDate("2026-07-01"); });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.vaccineId).toBe("ワクチン種別を選択してください");
      });
    });

    it("vaccineId が '0' の場合も未選択扱いでエラーになる", async () => {
      const { result } = renderVaccinationForm();
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
      const { result } = renderVaccinationForm();
      // BUG-004: 新規は当日デフォルトのため、未入力検証は明示クリアが必要
      act(() => {
        result.current.form.setVaccineId("1");
        result.current.form.setDate("");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.fieldErrors.date).toBe("接種日を入力してください");
      });
    });

    it("BUG-024: 接種日が未来日 → fieldErrors.date がセットされる", async () => {
      const { result } = renderVaccinationForm();
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
      const { result } = renderVaccinationForm();
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
      const { result } = renderVaccinationForm();
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
      const { result } = renderVaccinationForm("10");

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
      const { result } = renderVaccinationForm("10");

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

      const { result } = renderVaccinationForm("10");
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
      const { result } = renderVaccinationForm("10");

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

      const { result } = renderVaccinationForm("10");
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

    it("SD-1 回帰: supplemental が updateMutation の payload に含まれる（サイレント消失防止）", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: { id: "10", petId: "5", vaccineId: "1", date: "2026-07-01", nextDate: "", nextScheduleType: "1year" },
      } as ReturnType<typeof useGetVaccination>);
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);

      const { result } = renderVaccinationForm("10");
      act(() => {
        result.current.form.setDate("2026-07-05");
        result.current.form.setNextDate("2026-07-20");
        result.current.form.setSupplemental("補助説明テキスト");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ req: expect.objectContaining({ supplemental: "補助説明テキスト" }) })
      );
    });

    it("SD-6 回帰: next_schedule_type が updateMutation の payload に含まれる（サイレント消失防止）", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: { id: "10", petId: "5", vaccineId: "1", date: "2026-07-01", nextDate: "", nextScheduleType: "1year" },
      } as ReturnType<typeof useGetVaccination>);
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);

      const { result } = renderVaccinationForm("10");
      act(() => {
        result.current.form.setDate("2026-07-05");
        result.current.form.setNextDate("2026-07-20");
        result.current.form.setNextScheduleType("4weeks");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ req: expect.objectContaining({ next_schedule_type: "4weeks" }) })
      );
    });
  });

  describe("新規登録: SD-1 回帰", () => {
    it("第3象限: 作成権限あり・編集権限なし・IDなしでも従来の create payload を維持する", async () => {
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      const mockUpdateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useCreateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateVaccination>);
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockUpdateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);

      const { result } = renderHook(() =>
        useVaccinationForm(undefined, {
          canCreate: true,
          canEdit: false,
          canDelete: false,
        }),
      );
      act(() => {
        result.current.petSelection.setSelectedPets([
          { id: "5", ownerId: "1", name: "ポチ" } as Parameters<typeof result.current.petSelection.setSelectedPets>[0][number],
        ]);
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-01");
        result.current.form.setSupplemental("補助説明テキスト");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });
      expect(mockMutateAsync).toHaveBeenCalledWith({
        medical_record_id: null,
        pet_id: 5,
        vaccine_id: 1,
        date: jstDateStartISOString("2026-07-01"),
        next_date: jstDateStartISOString("2027-07-01"),
        lot1: undefined,
        lot2: undefined,
        lot3: undefined,
        lot4: undefined,
        remarks: undefined,
        supplemental: "補助説明テキスト",
        next_schedule_type: "1year",
      });
      expect(mockUpdateAsync).not.toHaveBeenCalled();
    });
  });

  describe("新規登録: SD-6 回帰", () => {
    it("next_schedule_type が createMutation の payload に含まれる（サイレント消失防止）", async () => {
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useCreateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateVaccination>);

      const { result } = renderVaccinationForm();
      act(() => {
        result.current.petSelection.setSelectedPets([
          { id: "5", ownerId: "1", name: "ポチ" } as Parameters<typeof result.current.petSelection.setSelectedPets>[0][number],
        ]);
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-01");
        result.current.form.setNextScheduleType("3weeks");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ next_schedule_type: "3weeks" })
      );
    });
  });

  // ──────────────────────────
  // BUG-026: ワクチン種別/接種日変更時の次回予定日自動計算
  // ──────────────────────────
  describe("次回予定日の自動計算（BUG-026）", () => {
    it("setVaccineId で狂犬病ワクチン(id=2)を選択 → nextScheduleType=1year かつ nextDate が自動計算される", async () => {
      const { result } = renderVaccinationForm();
      act(() => { result.current.form.setDate("2026-07-01"); });
      await waitFor(() => expect(result.current.form.date).toBe("2026-07-01"));

      act(() => { result.current.form.setVaccineId("2"); });

      await waitFor(() => {
        expect(result.current.form.nextScheduleType).toBe("1year");
        expect(result.current.form.nextDate).toBe("2027-07-01");
      });
    });

    it("setDate で接種日を変更 → 現在の nextScheduleType に基づき nextDate が再計算される", async () => {
      const { result } = renderVaccinationForm();
      act(() => { result.current.form.setNextScheduleType("3weeks"); });
      await waitFor(() => expect(result.current.form.nextScheduleType).toBe("3weeks"));

      act(() => { result.current.form.setDate("2026-07-01"); });

      await waitFor(() => {
        expect(result.current.form.nextDate).toBe("2026-07-22");
      });
    });

    it("setNextScheduleType で種別を変更 → 既存の接種日を基準に nextDate が再計算される", async () => {
      const { result } = renderVaccinationForm();
      act(() => { result.current.form.setDate("2026-07-01"); });
      await waitFor(() => expect(result.current.form.date).toBe("2026-07-01"));

      act(() => { result.current.form.setNextScheduleType("4weeks"); });

      await waitFor(() => {
        expect(result.current.form.nextDate).toBe("2026-07-29");
      });
    });
  });

  // ──────────────────────────
  // BUG-401: 実マスタ参照化後も vaccine_id → schedule 自動計算が退行しないこと
  // ──────────────────────────
  describe("ワクチンマスタ interval に基づく次回予定自動計算（BUG-401）", () => {
    it("interval='1ヶ月' のマスタを選択 → nextScheduleType=4weeks で自動計算される（旧ハードコード '1'/'2' 依存では発生しなかった分岐）", async () => {
      vi.mocked(useGetAllVaccinesMaster).mockReturnValue({
        data: [{ id: "14", name: "狂犬病ワクチン", isActive: true, interval: "1ヶ月" }],
      } as ReturnType<typeof useGetAllVaccinesMaster>);

      const { result } = renderVaccinationForm();
      act(() => { result.current.form.setDate("2026-07-01"); });
      await waitFor(() => expect(result.current.form.date).toBe("2026-07-01"));

      act(() => { result.current.form.setVaccineId("14"); });

      await waitFor(() => {
        expect(result.current.form.nextScheduleType).toBe("4weeks");
        expect(result.current.form.nextDate).toBe("2026-07-29");
      });
    });

    it("マスタに存在しない/interval 未設定の vaccineId を選択 → デフォルト(1year)にフォールバックする（サイレント誤スケジュールにしない）", async () => {
      vi.mocked(useGetAllVaccinesMaster).mockReturnValue({
        data: [{ id: "3", name: "ワクチンエキゾ", isActive: true, interval: "" }],
      } as ReturnType<typeof useGetAllVaccinesMaster>);

      const { result } = renderVaccinationForm();
      act(() => { result.current.form.setDate("2026-07-01"); });
      await waitFor(() => expect(result.current.form.date).toBe("2026-07-01"));

      act(() => { result.current.form.setVaccineId("3"); });

      await waitFor(() => {
        expect(result.current.form.nextScheduleType).toBe("1year");
        expect(result.current.form.nextDate).toBe("2027-07-01");
      });
    });

    it("vaccineOptions は isActive なマスタ項目のみを {value, label} で公開する", () => {
      vi.mocked(useGetAllVaccinesMaster).mockReturnValue({
        data: [
          { id: "11", name: "混合ワクチン5種（犬）", isActive: true, interval: "1年" },
          { id: "99", name: "廃盤ワクチン", isActive: false, interval: "1年" },
        ],
      } as ReturnType<typeof useGetAllVaccinesMaster>);

      const { result } = renderVaccinationForm();

      expect(result.current.form.vaccineOptions).toEqual([
        { value: "11", label: "混合ワクチン5種（犬）" },
      ]);
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

      const { result } = renderVaccinationForm("10");
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

      const { result } = renderVaccinationForm();
      act(() => { result.current.handleDelete(); });

      expect(mockMutate).not.toHaveBeenCalled();
    });
  });

  describe("mutation permission boundary (FE12-02 U8)", () => {
    it("作成権限なしでは有効な新規入力でも create mutation を発行しない", async () => {
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useCreateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateVaccination>);

      const { result } = renderHook(() =>
        useVaccinationForm(undefined, {
          canCreate: false,
          canEdit: true,
          canDelete: true,
        }),
      );
      act(() => {
        result.current.petSelection.setSelectedPets([
          { id: "5", ownerId: "1", name: "ポチ" } as Parameters<
            typeof result.current.petSelection.setSelectedPets
          >[0][number],
        ]);
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-01");
      });

      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(false);
      });
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it("編集権限なしでは有効な編集入力でも update mutation を発行しない", async () => {
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
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);

      const { result } = renderHook(() =>
        useVaccinationForm("10", {
          canCreate: true,
          canEdit: false,
          canDelete: true,
        }),
      );

      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(false);
      });
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it("削除権限なしでは編集IDがあっても delete mutation を発行しない", () => {
      const mockMutate = vi.fn();
      vi.mocked(useDeleteVaccination).mockReturnValue({
        mutate: mockMutate,
        isPending: false,
      } as ReturnType<typeof useDeleteVaccination>);

      const { result } = renderHook(() =>
        useVaccinationForm("10", {
          canCreate: true,
          canEdit: true,
          canDelete: false,
        }),
      );

      act(() => {
        result.current.handleDelete();
      });

      expect(mockMutate).not.toHaveBeenCalled();
    });

    it("作成権限が剥奪された後は取得済み formAction でも create mutation を発行しない", async () => {
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useCreateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateVaccination>);
      const { result, rerender } = renderHook(
        ({ canCreate }: { canCreate: boolean }) =>
          useVaccinationForm(undefined, {
            canCreate,
            canEdit: true,
            canDelete: true,
          }),
        { initialProps: { canCreate: true } },
      );
      act(() => {
        result.current.petSelection.setSelectedPets([
          { id: "5", ownerId: "1", name: "ポチ" } as Parameters<
            typeof result.current.petSelection.setSelectedPets
          >[0][number],
        ]);
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-01");
      });
      const capturedAction = result.current.formAction;

      rerender({ canCreate: false });
      runFormAction(capturedAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(false);
      });
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it("作成権限剥奪をcommitした直後のlayout phaseで取得済みformActionが発火してもmutationを発行しない", async () => {
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useCreateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateVaccination>);
      const { result, rerender } = renderHook(
        ({ canCreate }: { canCreate: boolean }) => {
          const form = useVaccinationForm(undefined, {
            canCreate,
            canEdit: true,
            canDelete: true,
          });
          const capturedActionRef = useRef(form.formAction);
          useLayoutEffect(() => {
            if (!canCreate) {
              startTransition(() => capturedActionRef.current(new FormData()));
            }
          }, [canCreate]);
          return form;
        },
        { initialProps: { canCreate: true } },
      );
      act(() => {
        result.current.petSelection.setSelectedPets([
          { id: "5", ownerId: "1", name: "ポチ" } as Parameters<
            typeof result.current.petSelection.setSelectedPets
          >[0][number],
        ]);
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-01");
      });
      const initialTimestamp = result.current.formState.timestamp;

      await act(async () => {
        rerender({ canCreate: false });
      });

      await waitFor(() => {
        expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
      });
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });
  });

  describe("deceased pet mutation boundary (FE12-02 C6a)", () => {
    it("direct petIdのpetが死亡へ変わったcommit直後のlayout phaseでも取得済みformActionはcreate mutationを発行しない", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      const livingPet = { ...DECEASED_PET, status: "生存" as const };
      const petSnapshot = { current: livingPet };
      vi.mocked(useGetPet).mockImplementation((requestedPetId) => ({
        data: requestedPetId === "5" ? petSnapshot.current : undefined,
        isLoading: false,
      } as ReturnType<typeof useGetPet>));
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useCreateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateVaccination>);
      const { result, rerender } = renderHook(
        ({ status }: { status: "生存" | "死亡" }) => {
          petSnapshot.current = status === "死亡" ? DECEASED_PET : livingPet;
          const form = useVaccinationForm(undefined, ALLOWED_MUTATION_PERMISSIONS);
          const capturedActionRef = useRef(form.formAction);
          useLayoutEffect(() => {
            if (status === "死亡") {
              startTransition(() => capturedActionRef.current(new FormData()));
            }
          }, [status]);
          return form;
        },
        { initialProps: { status: "生存" as const } },
      );
      act(() => {
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-01");
      });
      const initialTimestamp = result.current.formState.timestamp;

      await act(async () => {
        rerender({ status: "死亡" });
      });

      await waitFor(() => {
        expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
      });
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it("編集petが死亡へ変わったcommit直後のlayout phaseでも取得済みformActionはupdate mutationを発行しない", async () => {
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
      const livingPet = { ...DECEASED_PET, status: "生存" as const };
      const petSnapshot = { current: livingPet };
      vi.mocked(useGetPet).mockImplementation((requestedPetId) => ({
        data: requestedPetId === "5" ? petSnapshot.current : undefined,
        isLoading: false,
      } as ReturnType<typeof useGetPet>));
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);
      const { result, rerender } = renderHook(
        ({ status }: { status: "生存" | "死亡" }) => {
          petSnapshot.current = status === "死亡" ? DECEASED_PET : livingPet;
          const form = useVaccinationForm("10", ALLOWED_MUTATION_PERMISSIONS);
          const capturedActionRef = useRef(form.formAction);
          useLayoutEffect(() => {
            if (status === "死亡") {
              startTransition(() => capturedActionRef.current(new FormData()));
            }
          }, [status]);
          return form;
        },
        { initialProps: { status: "生存" as const } },
      );
      const initialTimestamp = result.current.formState.timestamp;

      await act(async () => {
        rerender({ status: "死亡" });
      });

      await waitFor(() => {
        expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
      });
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it("編集対象が明示的な死亡ペットならdelete mutationを発行しない", () => {
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
      vi.mocked(useGetPet).mockImplementation((requestedPetId) => ({
        data: requestedPetId === "5" ? DECEASED_PET : undefined,
        isLoading: false,
      } as ReturnType<typeof useGetPet>));
      const mockMutate = vi.fn();
      vi.mocked(useDeleteVaccination).mockReturnValue({
        mutate: mockMutate,
        isPending: false,
      } as ReturnType<typeof useDeleteVaccination>);

      const { result } = renderVaccinationForm("10");

      act(() => {
        result.current.handleDelete();
      });

      expect(mockMutate).not.toHaveBeenCalled();
    });

    it("direct petIdから明示的な死亡ペットをhydrateしてもcreate mutationを発行しない", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      vi.mocked(useGetPet).mockImplementation((requestedPetId) => ({
        data: requestedPetId === "5" ? DECEASED_PET : undefined,
        isLoading: false,
      } as ReturnType<typeof useGetPet>));
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useCreateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateVaccination>);

      const { result } = renderVaccinationForm();
      await waitFor(() => {
        expect(result.current.petSelection.selectedPets[0]?.status).toBe("死亡");
      });
      act(() => {
        result.current.form.setVaccineId("1");
        result.current.form.setDate("2026-07-01");
      });
      const initialTimestamp = result.current.formState.timestamp;

      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
      });
      expect(result.current.formState.success).toBe(false);
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it("編集対象から明示的な死亡ペットをhydrateしてもupdate mutationを発行しない", async () => {
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
      vi.mocked(useGetPet).mockImplementation((requestedPetId) => ({
        data: requestedPetId === "5" ? DECEASED_PET : undefined,
        isLoading: false,
      } as ReturnType<typeof useGetPet>));
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);

      const { result } = renderVaccinationForm("10");
      await waitFor(() => {
        expect(result.current.petSelection.selectedPets[0]?.status).toBe("死亡");
      });
      const initialTimestamp = result.current.formState.timestamp;

      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
      });
      expect(result.current.formState.success).toBe(false);
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });
  });
});

// ──────────────────────────────────────────────────────────
// BUG-016: 不存在 ID / 別 clinic / network error を空 edit に潰さない
// ──────────────────────────────────────────────────────────

describe("useVaccinationForm BUG-016 entity read", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-07-10T01:00:00.000Z"));
    vi.mocked(useGetPet).mockReturnValue({ data: undefined, isLoading: false } as ReturnType<typeof useGetPet>);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  function axiosError(status: number | undefined) {
    const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
    if (status === undefined) {
      return new AxiosError("Network Error", AxiosError.ERR_NETWORK, config, undefined, undefined);
    }
    return new AxiosError(
      "request failed",
      AxiosError.ERR_BAD_RESPONSE,
      config,
      undefined,
      {
        config,
        data: { error: "not found" },
        headers: new AxiosHeaders(),
        status,
        statusText: "Error",
      },
    );
  }

  it("404 → isReadNotFound、formAction で update/create が 0 回", async () => {
    const updateMutate = vi.fn().mockResolvedValue({});
    const createMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: updateMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);
    vi.mocked(useCreateVaccination).mockReturnValue({
      mutateAsync: createMutate,
      isPending: false,
    } as ReturnType<typeof useCreateVaccination>);
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(404),
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);

    const { result } = renderVaccinationForm("999999999");
    expect(result.current.isReadNotFound).toBe(true);
    expect(result.current.isReadError).toBe(false);
    expect(result.current.entityRead.status).toBe("notFound");

    await act(async () => {
      startTransition(() => {
        result.current.formAction(new FormData());
      });
    });
    await waitFor(() => {
      expect(result.current.formState.success).toBe(false);
    });
    expect(updateMutate).not.toHaveBeenCalled();
    expect(createMutate).not.toHaveBeenCalled();
  });

  it("403（別 clinic 相当）→ isReadNotFound と同一非開示、mutation 0 回", async () => {
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: updateMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(403),
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);

    const { result } = renderVaccinationForm("42");
    expect(result.current.isReadNotFound).toBe(true);
    expect(result.current.entityRead.status).toBe("forbiddenOrHidden");

    await act(async () => {
      startTransition(() => {
        result.current.formAction(new FormData());
      });
    });
    await waitFor(() => {
      expect(result.current.formState.success).toBe(false);
    });
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("network error → isReadError（notFound と区別）かつ retry あり、mutation 0 回", async () => {
    const refetch = vi.fn();
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateVaccination).mockReturnValue({
      mutateAsync: updateMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateVaccination>);
    vi.mocked(useGetVaccination).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(undefined),
      refetch,
    } as unknown as ReturnType<typeof useGetVaccination>);

    const { result } = renderVaccinationForm("999999999");
    expect(result.current.isReadError).toBe(true);
    expect(result.current.isReadNotFound).toBe(false);
    expect(result.current.retryRead).toBeTypeOf("function");
    result.current.retryRead?.();
    expect(refetch).toHaveBeenCalledTimes(1);

    await act(async () => {
      startTransition(() => {
        result.current.formAction(new FormData());
      });
    });
    await waitFor(() => {
      expect(result.current.formState.success).toBe(false);
    });
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("正常 edit: found レコードを form に反映する", () => {
    vi.mocked(useGetVaccination).mockReturnValue({
      data: {
        id: "10",
        petId: "5",
        vaccineId: "1",
        date: "2026-01-15T00:00:00+09:00",
        nextScheduleType: "1year",
        nextDate: "2027-01-15T00:00:00+09:00",
        doctor: "Dr.A",
      },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetVaccination>);

    const { result } = renderVaccinationForm("10");
    expect(result.current.entityRead.status).toBe("found");
    expect(result.current.isEdit).toBe(true);
    expect(result.current.form.vaccineId).toBe("1");
    expect(result.current.form.date).toBe("2026-01-15");
  });

  it("create route (id なし): idle かつ default form", () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
    const { result } = renderVaccinationForm();
    expect(result.current.entityRead.status).toBe("idle");
    expect(result.current.isEdit).toBe(false);
    expect(result.current.form.vaccineId).toBe("");
  });
});
