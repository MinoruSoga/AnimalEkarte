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
// FE-RC-045: use-vaccination-form.test.ts (1317行) をトピック別に分割した1ファイル。
// 元は describe("useVaccinationForm") 配下のネスト describe だったが、親の beforeEach/afterEach を
// 各ファイルの独立 describe に複製して分割している（振る舞いは維持）。

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

describe("useVaccinationForm — 新規登録時の接種日デフォルト（BUG-004）", () => {
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

describe("useVaccinationForm — 新規登録時のバリデーション", () => {
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

describe("useVaccinationForm — 編集時のバリデーション", () => {
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

describe("useVaccinationForm — 新規登録: SD-1 回帰", () => {
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

describe("useVaccinationForm — 新規登録: SD-6 回帰", () => {
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

describe("useVaccinationForm — 次回予定日の自動計算（BUG-026）", () => {
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
