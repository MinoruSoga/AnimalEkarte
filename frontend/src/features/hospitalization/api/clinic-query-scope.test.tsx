import { cleanup, renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { useGetHospitalizations, useGetAllHospitalizations } from "./get-hospitalizations";

vi.mock("@/lib/axios", () => ({ axios: { get: vi.fn() } }));
// AuthProvider remains at A when the storage update to B is followed by a
// cancelled beforeunload reload. Re-rendering/refetching must still use A.
vi.mock("@/hooks/use-current-clinic-id", () => ({ useCurrentClinicId: () => "1" }));
const clients: QueryClient[] = [];
beforeEach(() => {
  vi.mocked(axios.get).mockReset();
  localStorage.setItem("auth_current_clinic:v1", "2");
});
afterEach(() => {
  cleanup();
  clients.splice(0).forEach((client) => client.clear());
  localStorage.clear();
});
const hooks = [
  {
    name: "hospitalization list",
    use: (enabled: boolean) => useGetHospitalizations(undefined, { enabled }),
  },
  {
    name: "hospitalization snapshot",
    use: (enabled: boolean) => useGetAllHospitalizations(undefined, { enabled }),
  },
  { name: "reservations", use: (enabled: boolean) => useGetReservations({ enabled }) },
];
function mount(
  useHook: (
    enabled: boolean,
  ) =>
    | ReturnType<typeof useGetReservations>
    | ReturnType<typeof useGetHospitalizations>
    | ReturnType<typeof useGetAllHospitalizations>,
  enabled = true,
) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: 1, retryDelay: 0 } } });
  clients.push(client);
  const hook = renderHook(({ enabled }) => useHook(enabled), {
    initialProps: { enabled },
    wrapper: ({ children }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    ),
  });
  return { ...hook, client };
}
function response() {
  return { data: { data: [], total: 0, snapshot: true, page: 1, limit: 20 } };
}
function expectPinned(client: QueryClient, clinicId = "1") {
  expect(axios.get).toHaveBeenCalled();
  for (const [, config] of vi.mocked(axios.get).mock.calls) {
    expect(config?.headers).toEqual({ "X-Clinic-ID": clinicId });
  }
  const queries = client.getQueryCache().getAll();
  for (const query of queries) {
    expect(query.queryKey).toContainEqual(expect.objectContaining({ scopeClinicId: clinicId }));
  }
  expect(queries.filter((query) => query.state.status === "success")).toHaveLength(1);
}
describe.each(hooks)("$name query clinic scope", ({ use: useHook }) => {
  it("pins Context A despite storage B and refetch after cancelled reload", async () => {
    vi.mocked(axios.get).mockResolvedValue(response());
    const { result, client } = mount(useHook);
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    localStorage.setItem("auth_current_clinic:v1", "3");
    await act(async () => {
      await result.current.refetch();
    });
    expect(axios.get).toHaveBeenCalledTimes(2);
    expectPinned(client);
  });
  it("keeps the captured clinic on retry", async () => {
    vi.mocked(axios.get)
      .mockImplementationOnce(async () => {
        localStorage.setItem("auth_current_clinic:v1", "3");
        throw new Error("retry fixture");
      })
      .mockResolvedValue(response());
    const { result, client } = mount(useHook);
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(axios.get).toHaveBeenCalledTimes(2);
    expectPinned(client);
  });
  it("pins the context when disabled becomes enabled", async () => {
    vi.mocked(axios.get).mockResolvedValue(response());
    const { result, rerender, client } = mount(useHook, false);
    expect(axios.get).not.toHaveBeenCalled();
    localStorage.setItem("auth_current_clinic:v1", "3");
    rerender({ enabled: true });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expectPinned(client);
  });
});
it("preserves an explicit reservation clinic override", async () => {
  vi.mocked(axios.get).mockResolvedValue(response());
  const { result, client } = mount((enabled) => useGetReservations({ enabled, clinicId: "4" }));
  await waitFor(() => expect(result.current.isSuccess).toBe(true));
  expectPinned(client, "4");
});
