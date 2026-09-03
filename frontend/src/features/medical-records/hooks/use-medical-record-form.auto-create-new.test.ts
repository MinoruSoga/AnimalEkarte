import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor, render, screen } from "@testing-library/react";
import { createElement } from "react";
import { useMedicalRecordForm } from "./use-medical-record-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation } from "@/hooks/use-create-reservation";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { MedicalRecordAutoCreateFailure } from "../components/MedicalRecordAutoCreateFailure";

// ──────────────────────────────────────────────────────────
// モック定義
// ──────────────────────────────────────────────────────────
//
// FE4-18: use-medical-record-form.test.ts（821 行）から describe 境界で分割。
// 新規作成 auto-create effect / formAction（useActionState）の回帰をカバーする。
// vi.mock はファイルスコープで hoist されるため、この定義ブロックは
// use-medical-record-form.test.ts と同一の内容を逐語複製している
// （値・ロジックは 1 文字も変えていない）。基本的な state/handler の回帰は
// use-medical-record-form.test.ts を参照。

const mockNavigate = vi.fn();
let mockSearchParams = new URLSearchParams();
let mockLocationState: Record<string, unknown> | null = null;

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams, vi.fn()],
  useLocation: () => ({ pathname: "/medical-records/10", search: "", state: mockLocationState }),
}));

vi.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}));

vi.mock("sonner", () => ({ toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
const mockUsePermission = vi.hoisted(() => vi.fn());
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mockUsePermission(),
}));

// API フック群をすべてスタブ化（デフォルト: データなし・ローディングなし）
const noData = { data: undefined, isLoading: false, isError: false };
const noMutation = { mutateAsync: vi.fn().mockResolvedValue({}), isPending: false };

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({ data: undefined, isLoading: false, isError: false })),
  useGetPets: vi.fn(() => ({ data: [], isLoading: false, isError: false })),
}));
vi.mock("@/hooks/use-owner", () => ({ useGetOwner: vi.fn(() => noData) }));
const mockUseGetMedicalRecord = vi.fn(() => noData);
vi.mock("../api/get-medical-record", () => ({ useGetMedicalRecord: (...args: unknown[]) => mockUseGetMedicalRecord(...args) }));
vi.mock("../api/create-medical-record", () => ({ useCreateMedicalRecord: vi.fn(() => noMutation) }));
vi.mock("@/hooks/use-create-reservation", () => ({
  useCreateReservation: vi.fn(() => noMutation),
}));
vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: vi.fn(() => ({ data: [], isLoading: false })),
}));
vi.mock("../api/update-medical-record", () => ({ useUpdateMedicalRecord: () => noMutation }));
vi.mock("../api/inquiries", () => ({ useUpdateInquiry: () => noMutation }));
// BUG-416③: useGetClinicalPlan は clinical_plan の version（楽観ロック用）取得元。
// per-test で version を差し替えられるよう mockUseGetMedicalRecord と同じ間接呼び出しパターンにする。
const mockUseGetClinicalPlan = vi.fn(() => noData);
vi.mock("../api/clinical-plan", () => ({
  useUpdateClinicalPlan: () => noMutation,
  useGetClinicalPlan: (...args: unknown[]) => mockUseGetClinicalPlan(...args),
}));
vi.mock("@/hooks/use-reservation-types", () => ({
  useGetReservationTypesGrouped: vi.fn(() => ({
    data: [
      {
        label: "一般診療",
        types: [
          {
            id: 1,
            name: "一般診察",
            color: "#000000",
            is_active: true,
            duration_minutes: 15,
            sort_order: 1,
            is_internal: false,
            category: "general",
            group_id: 1,
            group: { id: 1, name: "一般診療", color: "#000000" },
          },
        ],
      },
    ],
  })),
}));

// ──────────────────────────────────────────────────────────
// テスト
// ──────────────────────────────────────────────────────────

// FE-RC-045: use-medical-record-form.auto-create.test.ts (1219行) を分割した1ファイル目。
// 親 describe("useMedicalRecordForm") の beforeEach/afterEach を複製し、独立 describe として実行する（振る舞いは維持）。

describe("useMedicalRecordForm — 新規作成 auto-create effect (新規予約/カルテ作成)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUsePermission.mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    });
    mockSearchParams = new URLSearchParams();
    mockLocationState = null;
    mockUseGetMedicalRecord.mockReturnValue(noData);
    mockUseGetClinicalPlan.mockReturnValue(noData);
    // デフォルト: pet データなし
    vi.mocked(useGetPet).mockReturnValue({ data: undefined, isLoading: false, isError: false });
    // デフォルト: owner データなし
    vi.mocked(useGetOwner).mockReturnValue(noData as never);
    vi.mocked(useGetReservationTypesGrouped).mockReturnValue({
      data: [
        {
          label: "一般診療",
          types: [
            {
              id: 1,
              name: "一般診察",
              color: "#000000",
              is_active: true,
              duration_minutes: 15,
              sort_order: 1,
              is_internal: false,
              category: "general",
              group_id: 1,
              group: { id: 1, name: "一般診療", color: "#000000" },
            },
          ],
        },
      ],
    } as ReturnType<typeof useGetReservationTypesGrouped>);
    vi.mocked(useCreateReservation).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({ id: "appointment-1" }),
      isPending: false,
    } as ReturnType<typeof useCreateReservation>);
    vi.mocked(useGetReservations).mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useGetReservations>);
  });

  afterEach(() => {
    vi.clearAllTimers();
    vi.useRealTimers();
  });

    const mockPet = {
      id: "5",
      name: "ポチ",
      ownerId: "2",
      ownerName: "田中",
      species: "犬",
      breed: "",
      birthday: "",
      gender: "男" as const,
      weight: null,
      imageUrl: null,
      status: "生存" as const,
      microchipNumber: null,
      insuranceNumber: null,
      insuranceExpiry: null,
      memo: null,
    };

    it("isNewRecord && selectedPet あり → createMutation.mutateAsync を呼ぶ", async () => {
      // petId あり → isNewRecord=true かつ selectedPet を設定
      mockSearchParams = new URLSearchParams({ petId: "5" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockMutateAsync = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm()); // recordId なし → isNewRecord=true
      });

      await waitFor(() => {
        expect(mockCreateReservation).toHaveBeenCalledWith(
          expect.objectContaining({
            pet_id: Number(mockPet.id),
            owner_id: Number(mockPet.ownerId),
            reservation_type_id: 1,
            status: "in_consultation",
            reservation_route: "record_shortcut",
          })
        );
        expect(mockMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            pet_id: mockPet.id,
            owner_id: mockPet.ownerId,
            appointment_id: "appointment-1",
            status: "draft",
          })
        );
      });
    });

    it.each([
      {
        currentTime: "2026-06-01T14:59:00Z",
        jstDate: "2026-06-01",
        jstTime: "23:59",
        appointmentStartTime: "2026-06-01T23:59:00+09:00",
        appointmentEndTime: "2026-06-02T00:14:00+09:00",
      },
      {
        currentTime: "2026-06-01T15:01:00Z",
        jstDate: "2026-06-02",
        jstTime: "00:01",
        appointmentStartTime: "2026-06-02T00:01:00+09:00",
        appointmentEndTime: "2026-06-02T00:16:00+09:00",
      },
    ])(
      "visitDate 未指定: $currentTime (JST $jstDate $jstTime) を appointment 検索・予約作成・カルテ作成で共有する",
      async ({ currentTime, jstDate, appointmentStartTime, appointmentEndTime }) => {
        vi.useFakeTimers({ toFake: ["Date"] });
        vi.setSystemTime(new Date(currentTime));
        mockSearchParams = new URLSearchParams({ petId: "5" });
        vi.mocked(useGetPet).mockReturnValue({
          data: mockPet,
          isLoading: false,
          isError: false,
        });

        const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
        vi.mocked(useCreateMedicalRecord).mockReturnValue({
          mutateAsync: mockCreateRecord,
          isPending: false,
        } as ReturnType<typeof useCreateMedicalRecord>);
        const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
        vi.mocked(useCreateReservation).mockReturnValue({
          mutateAsync: mockCreateReservation,
          isPending: false,
        } as ReturnType<typeof useCreateReservation>);

        await act(async () => {
          renderHook(() => useMedicalRecordForm());
        });

        await waitFor(() => {
          expect(useGetReservations).toHaveBeenCalledWith({
            date: jstDate,
            petId: mockPet.id,
            enabled: true,
          });
          expect(mockCreateReservation).toHaveBeenCalledWith(
            expect.objectContaining({
              start_time: appointmentStartTime,
              end_time: appointmentEndTime,
            }),
          );
          expect(mockCreateRecord).toHaveBeenCalledWith(
            expect.objectContaining({
              appointment_id: "appointment-1",
              visit_date: jstDate,
            }),
          );
        });
      },
    );

    it("受付から遷移した新規作成では既存 appointment_id をカルテに紐付ける", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      mockLocationState = { appointmentId: "77", visitDate: "2026-05-29" };
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      await waitFor(() => {
        expect(mockCreateReservation).not.toHaveBeenCalled();
        expect(mockCreateRecord).toHaveBeenCalledWith(
          expect.objectContaining({
            appointment_id: "77",
            visit_date: "2026-05-29",
          })
        );
      });
    });

    it("visitDate 指定の一覧新規作成では appointment も同じ日付で作成する", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5", visitDate: "2026-06-01" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      await waitFor(() => {
        expect(mockCreateReservation).toHaveBeenCalledWith(
          expect.objectContaining({
            start_time: expect.stringMatching(/^2026-06-01T/),
            end_time: expect.stringMatching(/^2026-06-01T/),
          })
        );
        expect(mockCreateRecord).toHaveBeenCalledWith(
          expect.objectContaining({
            appointment_id: "appointment-1",
            visit_date: "2026-06-01",
          })
        );
      });
    });

    it("一覧新規作成では同日同ペットの未完了通常 appointment を再利用する", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5", visitDate: "2026-06-01" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });
      vi.mocked(useGetReservations).mockReturnValue({
        data: [
          {
            id: "appointment-existing",
            category: "general",
            status: "checked_in",
          },
        ],
        isLoading: false,
      } as ReturnType<typeof useGetReservations>);

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      await waitFor(() => {
        expect(mockCreateReservation).not.toHaveBeenCalled();
        expect(mockCreateRecord).toHaveBeenCalledWith(
          expect.objectContaining({
            appointment_id: "appointment-existing",
            visit_date: "2026-06-01",
          })
        );
      });
    });

    it("一覧新規作成では当日 appointment 検索中に自動作成しない", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5", visitDate: "2026-06-01" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });
      vi.mocked(useGetReservations).mockReturnValue({
        data: [],
        isLoading: true,
      } as ReturnType<typeof useGetReservations>);

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      expect(mockCreateReservation).not.toHaveBeenCalled();
      expect(mockCreateRecord).not.toHaveBeenCalled();
    });

    it("isNewRecord && selectedPet あり → 作成後に detail ページへナビゲート", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockMutateAsync = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      // navigate が呼ばれ、detail ページパスに new-record-1 が含まれる
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringContaining("new-record-1"),
          expect.objectContaining({ replace: true })
        );
      });
    });

    // BUG-503: clinic に general 予約区分が無い（trimming のみ）とき silent no-op しない
    it("一般予約区分が無い（trimming のみ）場合は silent no-op せず appointment failure を出し再試行できる", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });
      vi.mocked(useGetReservationTypesGrouped).mockReturnValue({
        data: [
          {
            label: "トリミング",
            types: [
              {
                id: 9,
                name: "トリミング",
                color: "#000000",
                is_active: true,
                duration_minutes: 30,
                sort_order: 1,
                is_internal: false,
                category: "trimming",
                group_id: 2,
                group: { id: 2, name: "トリミング", color: "#000000" },
              },
            ],
          },
        ],
        isLoading: false,
        isError: false,
        isPending: false,
        isFetched: true,
      } as ReturnType<typeof useGetReservationTypesGrouped>);

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      const { result } = renderHook(() => useMedicalRecordForm());

      await waitFor(() => {
        expect(result.current.autoCreateFailurePhase).toBe("appointment");
      });
      expect(mockCreateReservation).not.toHaveBeenCalled();
      expect(mockCreateRecord).not.toHaveBeenCalled();
      expect(mockNavigate).not.toHaveBeenCalled();

      render(createElement(MedicalRecordAutoCreateFailure, {
        failurePhase: "appointment",
        isRetrying: false,
        onRetry: () => result.current.retryAutoCreate(),
      }));
      expect(screen.getByRole("button", { name: "カルテ作成を再試行する" })).toBeEnabled();
      expect(screen.getByRole("alert")).toBeInTheDocument();

      // 再試行時も general が無ければ再び failure（silent に戻らない）
      act(() => {
        result.current.retryAutoCreate();
      });
      await waitFor(() => {
        expect(result.current.autoCreateFailurePhase).toBe("appointment");
      });
      expect(mockCreateReservation).not.toHaveBeenCalled();
      expect(mockCreateRecord).not.toHaveBeenCalled();
    });

    // BUG-503 / S11#3: 同一 auto-draft root。appointment 再利用があれば general 無しでも promote する
    it("trimming のみでも同日 general appointment 再利用で detail へ promote する", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5", visitDate: "2026-06-01" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });
      vi.mocked(useGetReservationTypesGrouped).mockReturnValue({
        data: [
          {
            label: "トリミング",
            types: [
              {
                id: 9,
                name: "トリミング",
                color: "#000000",
                is_active: true,
                duration_minutes: 30,
                sort_order: 1,
                is_internal: false,
                category: "trimming",
                group_id: 2,
                group: { id: 2, name: "トリミング", color: "#000000" },
              },
            ],
          },
        ],
        isLoading: false,
        isError: false,
        isPending: false,
        isFetched: true,
      } as ReturnType<typeof useGetReservationTypesGrouped>);
      vi.mocked(useGetReservations).mockReturnValue({
        data: [
          {
            id: "appointment-existing",
            category: "general",
            status: "checked_in",
          },
        ],
        isLoading: false,
      } as ReturnType<typeof useGetReservations>);

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      await waitFor(() => {
        expect(mockCreateReservation).not.toHaveBeenCalled();
        expect(mockCreateRecord).toHaveBeenCalledWith(
          expect.objectContaining({
            appointment_id: "appointment-existing",
            visit_date: "2026-06-01",
          }),
        );
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringContaining("new-record-1"),
          expect.objectContaining({ replace: true }),
        );
      });
    });

    it("?tab= があるとき作成後の detail でもタブを残す", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5", tab: "検査" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockMutateAsync = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith(
          "/medical-records/new-record-1?tab=%E6%A4%9C%E6%9F%BB",
          expect.objectContaining({ replace: true }),
        );
      });
    });
});
