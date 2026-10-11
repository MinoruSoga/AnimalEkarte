import { cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { useGetReservations } from "./use-get-reservations";

vi.mock("@/lib/axios", () => ({ axios: { get: vi.fn() } }));
vi.mock("@/hooks/use-current-clinic-id", () => ({ useCurrentClinicId: () => "1" }));

const row = (id: number) => ({
  id,
  clinic_id: 1,
  pet_id: id,
  owner_id: 1,
  start_time: "2026-10-11T10:00:00+09:00",
  end_time: "2026-10-11T11:00:00+09:00",
  status: "confirmed",
});
const clients: QueryClient[] = [];
function mount() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, throwOnError: false } },
  });
  clients.push(client);
  const hook = renderHook(
    () => useGetReservations({ startDate: "2026-10-11", endDate: "2026-10-12", clinicId: "1" }),
    {
      wrapper: ({ children }) => (
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      ),
    },
  );
  return { ...hook, client };
}
beforeEach(() => {
  vi.mocked(axios.get).mockReset();
});
afterEach(() => {
  cleanup();
  clients.splice(0).forEach((client) => client.clear());
});

describe("reservation snapshot", () => {
  it.each(["insert", "delete", "same-count reschedule"])(
    "uses one snapshot despite a concurrent %s",
    async (change) => {
      const rows = Array.from({ length: 1002 }, (_, index) => row(index + 1));
      vi.mocked(axios.get).mockResolvedValueOnce({
        data: { data: rows, total: rows.length, snapshot: true, page: 1, limit: 10000 },
      });
      // A later offset request would observe this changed ordering/membership.
      const changedRows =
        change === "insert"
          ? [row(2000), ...rows]
          : change === "delete"
            ? rows.slice(1)
            : [...rows.slice(1), rows[0]];
      vi.mocked(axios.get).mockResolvedValue({
        data: { data: changedRows.slice(1000), total: changedRows.length, page: 2, limit: 1000 },
      });
      const { result } = mount();
      await waitFor(() => expect(result.current.isSuccess).toBe(true));
      expect(result.current.data).toHaveLength(1002);
      expect(new Set(result.current.data?.map((reservation) => reservation.id)).size).toBe(1002);
      expect(axios.get).toHaveBeenCalledTimes(1);
      expect(axios.get).toHaveBeenCalledWith("/v1/reservations", {
        params: { snapshot: true, start_date: "2026-10-11", end_date: "2026-10-12" },
        headers: { "X-Clinic-ID": "1" },
      });
    },
  );
  it("rejects a legacy paginated response rather than silently combining pages", async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: { data: [row(1)], total: 1, page: 1, limit: 1000 },
    });
    const { client } = mount();
    await waitFor(() => expect(client.getQueryCache().getAll()[0]?.state.status).toBe("error"));
    expect(client.getQueryCache().getAll()[0]?.state.data).toBeUndefined();
    expect(axios.get).toHaveBeenCalledTimes(1);
  });
  it("rejects a truncated snapshot", async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { data: [row(1)], total: 2, snapshot: true } });
    const { client } = mount();
    await waitFor(() => expect(client.getQueryCache().getAll()[0]?.state.status).toBe("error"));
    expect(client.getQueryCache().getAll()[0]?.state.data).toBeUndefined();
    expect(axios.get).toHaveBeenCalledTimes(1);
  });
  it("propagates an overflow or database failure without returning partial rows", async () => {
    vi.mocked(axios.get).mockImplementation(async () => {
      throw new Error("期間を絞り込んでください");
    });
    const { client } = mount();
    await waitFor(() => expect(client.getQueryCache().getAll()[0]?.state.status).toBe("error"));
    expect(client.getQueryCache().getAll()[0]?.state.data).toBeUndefined();
  });
});
