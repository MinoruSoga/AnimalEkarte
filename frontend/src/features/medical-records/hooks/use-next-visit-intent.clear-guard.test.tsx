import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { useGetReservations } from "@/hooks/use-get-reservations";
import { useClearNextVisitIntent } from "@/hooks/use-next-visit-intent";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { formatJSTWallDate, toJSTWallDate } from "@/lib/jst-date";
import type { Reservation } from "@/lib/transforms/reservation";

import { useNextVisitIntent } from "./use-next-visit-intent";

vi.mock("sonner", () => ({ toast: { success: vi.fn() } }));
vi.mock("@/lib/axios", () => ({ axios: { delete: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
vi.mock("@/hooks/use-get-reservations", () => ({ useGetReservations: vi.fn() }));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: true, canEdit: true }),
}));

const displayedRevision = "2026-10-11T00:00:00.123456789Z";

function plan(id = "42", updatedAt: string | undefined = displayedRevision): Reservation {
  const start = toJSTWallDate(new Date());
  start.setDate(start.getDate() + 45);
  return {
    id,
    updatedAt,
    start,
    intendedDate: formatJSTWallDate(start),
    status: "tentative",
  } as unknown as Reservation;
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(axios.delete).mockResolvedValue({});
});

describe("medical record clear — displayed reservation compare-and-set", () => {
  it.each(["same reservation updated", "replacement reservation"])(
    "queued clear preserves the displayed token after %s",
    async (change) => {
      let displayed = plan();
      vi.mocked(useGetReservations).mockImplementation(
        () => ({ data: [displayed] }) as ReturnType<typeof useGetReservations>,
      );
      let finishSet!: (value: Reservation) => void;
      const setMutation = {
        mutateAsync: vi.fn(
          () =>
            new Promise<Reservation>((resolve) => {
              finishSet = resolve;
            }),
        ),
      };
      const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
      const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      );
      const { result, rerender, unmount } = renderHook(
        () =>
          useNextVisitIntent({
            petId: "10",
            clinicId: "7",
            isPetDeceased: false,
            setIntentMutation: setMutation,
            clearIntentMutation: useClearNextVisitIntent("7"),
          }),
        { wrapper },
      );

      await act(async () => result.current.handleIntentDateChange(formatJSTWallDate(plan().start)));
      await act(async () => result.current.handleIntentDateChange(""));
      expect(axios.delete).not.toHaveBeenCalled();
      displayed = plan(
        change === "replacement reservation" ? "43" : "42",
        "2026-10-11T00:00:01.987654321Z",
      );
      rerender();
      const conflict = { response: { status: 409 } };
      vi.mocked(axios.delete).mockRejectedValue(conflict);
      // R3: queued clear は先行 PUT のレスポンス guard（id + updatedAt）で DELETE する。
      // refetch で新しい revision が表示されてもその guard は採用しない。
      await act(async () =>
        finishSet({ id: "42", updatedAt: "put-rev-1" } as unknown as Reservation),
      );
      await waitFor(() => expect(handleApiError).toHaveBeenCalledWith(conflict, "次回予定クリア"));
      expect(axios.delete).toHaveBeenCalledTimes(1);
      expect(axios.delete).toHaveBeenCalledWith("/v1/reservations/pets/10/next-visit-intent", {
        headers: { "X-Clinic-ID": "7" },
        params: { expected_reservation_id: "42", expected_updated_at: "put-rev-1" },
      });
      expect(result.current.intentFieldValue).toBe(formatJSTWallDate(displayed.start));
      expect(toast.success).toHaveBeenCalledTimes(1); // The preceding PUT succeeded.
      expect(toast.success).not.toHaveBeenCalledWith("次回予定をクリアしました");
      unmount();
      client.clear();
    },
  );

  it.each(["missing reservation", "missing revision"])(
    "%s never sends an unguarded DELETE",
    async (missing) => {
      const rows = missing === "missing reservation" ? [] : [{ ...plan(), updatedAt: undefined }];
      vi.mocked(useGetReservations).mockReturnValue({ data: rows } as ReturnType<
        typeof useGetReservations
      >);
      const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
      const { result } = renderHook(() =>
        useNextVisitIntent({
          petId: "10",
          isPetDeceased: false,
          setIntentMutation: { mutateAsync: vi.fn().mockResolvedValue(undefined) },
          clearIntentMutation: clearMutation,
        }),
      );
      await act(async () => result.current.handleIntentDateChange(""));
      expect(clearMutation.mutateAsync).not.toHaveBeenCalled();
    },
  );
});
