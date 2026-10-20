import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { useUpdateConsultation } from "./consultations";
import { useUpdateProcedure } from "./procedures";

vi.mock("@/lib/axios", () => ({ axios: { patch: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
beforeEach(() => vi.clearAllMocks());

it.each([useUpdateConsultation, useUpdateProcedure])(
  "blocks PATCH without a positive editing version",
  async (useUpdate) => {
    const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const { result } = renderHook(() => useUpdate(), { wrapper });
    for (const version of [undefined, 0, -1, NaN, 1.5]) {
      await act(async () => {
        // Exercise the runtime boundary for legacy/untyped callers too.
        await expect(
          result.current.mutateAsync({ id: "1", req: { version: version! } }),
        ).rejects.toThrow("編集開始時");
      });
    }
    expect(axios.patch).not.toHaveBeenCalled();
    client.clear();
  },
);

it.each([useUpdateConsultation, useUpdateProcedure])(
  "sends editing-start version and reads returned version for status changes",
  async (useUpdate) => {
    const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    vi.mocked(axios.patch).mockResolvedValue({
      data: { id: 1, name: "item", is_active: false, version: 8 },
    });
    const { result } = renderHook(() => useUpdate(), { wrapper });
    await act(async () => {
      const saved = await result.current.mutateAsync({
        id: "1",
        req: { version: 7, is_active: false },
      });
      expect(saved.version).toBe(8);
    });
    expect(axios.patch).toHaveBeenCalledWith(expect.any(String), { version: 7, is_active: false });
    client.clear();
  },
);
