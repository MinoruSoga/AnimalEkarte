import { act, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { useClearNextVisitIntent } from "./use-next-visit-intent";

vi.mock("@/lib/axios", () => ({ axios: { delete: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

describe("guarded next visit intent deletion", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(axios.delete).mockResolvedValue({});
  });

  it("sends identity, full revision and the displayed row's clinic", async () => {
    const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const { result } = renderHook(() => useClearNextVisitIntent("1"), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({
        petId: "22",
        clinicId: "3",
        expectedReservationId: "42",
        expectedUpdatedAt: "2026-10-11T00:00:00.123456789Z",
      });
    });
    expect(axios.delete).toHaveBeenCalledWith("/v1/reservations/pets/22/next-visit-intent", {
      headers: { "X-Clinic-ID": "3" },
      params: {
        expected_reservation_id: "42",
        expected_updated_at: "2026-10-11T00:00:00.123456789Z",
      },
    });
    client.clear();
  });
});
