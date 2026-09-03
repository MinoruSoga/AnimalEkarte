import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { useVaccinationForm } from "./use-vaccination-form";
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

function renderVaccinationForm(id?: string) {
  return renderHook(() => useVaccinationForm(id, ALLOWED_MUTATION_PERMISSIONS));
}
// FE-RC-045: use-vaccination-form.test.ts (1317行) をトピック別に分割した1ファイル。
// 元は describe("useVaccinationForm") 配下のネスト describe だったが、親の beforeEach/afterEach を
// 各ファイルの独立 describe に複製して分割している（振る舞いは維持）。

describe("useVaccinationForm — 次回予定 type/日付の整合（BUG-005）", () => {
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

    it("標準間隔選択後に nextDate を手動上書き → nextScheduleType が other になる", async () => {
      const { result } = renderVaccinationForm();
      act(() => {
        result.current.form.setDate("2026-07-01");
        result.current.form.setNextScheduleType("1year");
      });
      await waitFor(() => {
        expect(result.current.form.nextDate).toBe("2027-07-01");
        expect(result.current.form.nextScheduleType).toBe("1year");
      });

      act(() => {
        result.current.form.setNextDate("2027-07-20");
      });

      await waitFor(() => {
        expect(result.current.form.nextDate).toBe("2027-07-20");
        expect(result.current.form.nextScheduleType).toBe("other");
      });
    });

    it("手動上書き後の保存 payload は next_schedule_type=other と手動 next_date を送る", async () => {
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
        // systemTime = 2026-07-10 のため接種日は過去、次回は本日以降
        result.current.form.setDate("2026-07-01");
        result.current.form.setNextScheduleType("1year");
        result.current.form.setNextDate("2027-07-20");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          next_schedule_type: "other",
          next_date: jstDateStartISOString("2027-07-20"),
        }),
      );
    });

    it("手動 nextDate が標準計算と一致するなら type は維持する", async () => {
      const { result } = renderVaccinationForm();
      act(() => {
        result.current.form.setDate("2026-07-01");
        result.current.form.setNextScheduleType("1year");
      });
      await waitFor(() => expect(result.current.form.nextDate).toBe("2027-07-01"));

      act(() => {
        result.current.form.setNextDate("2027-07-01");
      });

      await waitFor(() => {
        expect(result.current.form.nextScheduleType).toBe("1year");
        expect(result.current.form.nextDate).toBe("2027-07-01");
      });
    });

    it("編集時: localOverrides に date がなくても setNextScheduleType は server date から再計算する", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: {
          id: "10",
          petId: "5",
          vaccineId: "1",
          date: "2026-07-01",
          nextDate: "2027-07-01",
          nextScheduleType: "1year",
        },
      } as ReturnType<typeof useGetVaccination>);

      const { result } = renderVaccinationForm("10");
      await waitFor(() => expect(result.current.form.date).toBe("2026-07-01"));

      act(() => {
        result.current.form.setNextScheduleType("4weeks");
      });

      await waitFor(() => {
        expect(result.current.form.nextScheduleType).toBe("4weeks");
        expect(result.current.form.nextDate).toBe("2026-07-29");
      });
    });

    it("編集時: 手動 nextDate 上書きで type が other になり update payload に載る", async () => {
      vi.mocked(useGetVaccination).mockReturnValue({
        data: {
          id: "10",
          petId: "5",
          vaccineId: "1",
          date: "2026-07-01",
          nextDate: "2027-07-01",
          nextScheduleType: "1year",
        },
      } as ReturnType<typeof useGetVaccination>);
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.mocked(useUpdateVaccination).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useUpdateVaccination>);

      const { result } = renderVaccinationForm("10");
      await waitFor(() => expect(result.current.form.nextScheduleType).toBe("1year"));

      act(() => {
        result.current.form.setNextDate("2027-07-20");
      });
      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          req: expect.objectContaining({
            next_schedule_type: "other",
            next_date: jstDateStartISOString("2027-07-20"),
          }),
        }),
      );
    });
});

describe("useVaccinationForm — ワクチンマスタ interval に基づく次回予定自動計算（BUG-401）", () => {
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

describe("useVaccinationForm — handleDelete (BUG-025)", () => {
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
