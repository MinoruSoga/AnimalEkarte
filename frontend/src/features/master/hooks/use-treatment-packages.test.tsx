import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import {
  useDeleteTreatmentPackage,
  useReorderTreatmentPackages,
  useTreatmentPackageEditor,
  useUpdateTreatmentPackage,
} from "./use-treatment-packages";
import type { TreatmentPackage } from "../api/treatment-packages";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn(), post: vi.fn(), patch: vi.fn(), delete: vi.fn() },
}));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
beforeEach(() => vi.clearAllMocks());

const wrapper = () => {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
};

const pkg: TreatmentPackage = {
  id: "201",
  clinic_id: "7",
  version: "3",
  name: "セット",
  description: "",
  is_active: true,
  sort_order: 0,
  created_at: "t",
  updated_at: "t",
  components: [
    {
      component_key: "a0000000-0000-4000-8000-000000000011",
      item: { kind: "consultation", consultation_id: "101" },
      quantity: "1",
    },
  ],
};

const conflict = () => {
  const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
  return new AxiosError("conflict", "409", config, undefined, {
    status: 409,
    statusText: "",
    headers: {},
    config,
    data: { error: "conflict", code: "version-conflict" },
  });
};

describe("editor session（CAS・競合・入力保持）", () => {
  it("409 は conflict 化し draft を保持。無断上書きしない", async () => {
    const { result } = renderHook(() => useTreatmentPackageEditor(pkg), { wrapper: wrapper() });
    act(() => result.current.updateDraft({ name: "変更" }));
    const update = vi.fn().mockRejectedValue(conflict());
    let outcome!: Awaited<ReturnType<typeof result.current.save>>;
    await act(async () => {
      outcome = await result.current.save({ create: vi.fn(), update });
    });
    expect(update).toHaveBeenCalledWith("201", { version: "3", name: "変更" });
    expect(outcome.kind).toBe("conflict");
    expect(result.current.conflict).toBe(true);
    expect(result.current.draft!.name).toBe("変更");
  });

  it("create 応答不明は create-unknown で盲目再送しない", async () => {
    const { result } = renderHook(() => useTreatmentPackageEditor("new"), { wrapper: wrapper() });
    // draft を直接 mutate せず、editor の更新 API（イミュータブル）で入力する。
    act(() => {
      result.current.updateDraft({ name: "n" });
      result.current.updateComponents(
        result.current.draft!.components.map((c) => ({ ...c, masterId: "1" })),
      );
    });
    const create = vi.fn().mockRejectedValue(new AxiosError("network"));
    let outcome!: Awaited<ReturnType<typeof result.current.save>>;
    await act(async () => {
      outcome = await result.current.save({ create, update: vi.fn() });
    });
    expect(outcome.kind).toBe("create-unknown");
    expect(create).toHaveBeenCalledTimes(1);
    expect(result.current.createUnknown).toBe(true);
  });

  it("A1-01: 保存中に編集 session が変わったら discarded（conflict 化しない）", async () => {
    const { result } = renderHook(() => useTreatmentPackageEditor(pkg), { wrapper: wrapper() });
    act(() => result.current.updateDraft({ name: "変更" }));
    let resolvePatch!: (v: unknown) => void;
    const update = vi.fn().mockImplementation(
      () =>
        new Promise((r) => {
          resolvePatch = r;
        }),
    );
    let promise!: Promise<Awaited<ReturnType<typeof result.current.save>>>;
    await act(async () => {
      promise = result.current.save({ create: vi.fn(), update }, () => true);
    });
    await act(async () => {
      resolvePatch(pkg);
    });
    expect((await promise).kind).toBe("discarded");
    expect(result.current.conflict).toBe(false);
    expect(result.current.draft!.name).toBe("変更");
  });

  it("A1-01: 遅延 409 も discarded で conflict 化しない", async () => {
    const { result } = renderHook(() => useTreatmentPackageEditor(pkg), { wrapper: wrapper() });
    act(() => result.current.updateDraft({ name: "変更" }));
    let rejectPatch!: (e: unknown) => void;
    const update = vi.fn().mockImplementation(
      () =>
        new Promise((_r, rej) => {
          rejectPatch = rej;
        }),
    );
    let promise!: Promise<Awaited<ReturnType<typeof result.current.save>>>;
    await act(async () => {
      promise = result.current.save({ create: vi.fn(), update }, () => true);
    });
    await act(async () => {
      rejectPatch(conflict());
    });
    expect((await promise).kind).toBe("discarded");
    expect(result.current.conflict).toBe(false);
  });
});

describe("mutation の版規則", () => {
  it("update は非canonical版を送らない", async () => {
    const { result } = renderHook(() => useUpdateTreatmentPackage(), { wrapper: wrapper() });
    await expect(result.current.mutateAsync({ id: "201", req: { version: "0" } })).rejects.toThrow(
      "編集開始時",
    );
    expect(axios.patch).not.toHaveBeenCalled();
  });

  it("delete は version query 必須で送る", async () => {
    vi.mocked(axios.delete).mockResolvedValue({ data: undefined });
    const { result } = renderHook(() => useDeleteTreatmentPackage(), { wrapper: wrapper() });
    await result.current.mutateAsync({ id: "201", version: "3" });
    expect(axios.delete).toHaveBeenCalledWith("/v1/masters/treatment-packages/201", {
      params: { version: "3" },
    });
  });

  it("reorder は全件 ids + order_version を送る", async () => {
    vi.mocked(axios.patch).mockResolvedValue({ data: undefined });
    const { result } = renderHook(() => useReorderTreatmentPackages(), { wrapper: wrapper() });
    await result.current.mutateAsync({ order_version: "17", ids: ["201", "202"] });
    expect(axios.patch).toHaveBeenCalledWith("/v1/masters/treatment-packages/reorder", {
      order_version: "17",
      ids: ["201", "202"],
    });
  });
});

describe("S07/S08: HTTP status helpers", () => {
  const withStatus = (status: number) => {
    const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
    return new AxiosError("e", String(status), config, undefined, {
      status,
      statusText: "",
      headers: {},
      config,
      data: {},
    });
  };

  it("isForbiddenError は 403 のみ、isNotFoundError は 404 のみ", async () => {
    const { isForbiddenError, isNotFoundError } = await import("./use-treatment-packages");
    expect(isForbiddenError(withStatus(403))).toBe(true);
    expect(isForbiddenError(withStatus(500))).toBe(false);
    expect(isForbiddenError(new Error("network"))).toBe(false);
    expect(isNotFoundError(withStatus(404))).toBe(true);
    expect(isNotFoundError(withStatus(403))).toBe(false);
  });

  it("R4: update 409/403 は handleApiError を呼ばない", async () => {
    const { result } = renderHook(() => useUpdateTreatmentPackage(), { wrapper: wrapper() });
    const send = () => result.current.mutateAsync({ id: "201", req: { version: "3" } });
    vi.mocked(axios.patch).mockRejectedValue(withStatus(409));
    await expect(send()).rejects.toThrow();
    vi.mocked(axios.patch).mockRejectedValue(withStatus(403));
    await expect(send()).rejects.toThrow();
    expect(handleApiError).not.toHaveBeenCalled();
  });

  it("R4: update 500 は handleApiError を呼ぶ", async () => {
    const { result } = renderHook(() => useUpdateTreatmentPackage(), { wrapper: wrapper() });
    const send = () => result.current.mutateAsync({ id: "201", req: { version: "3" } });
    vi.mocked(axios.patch).mockRejectedValue(withStatus(500));
    await expect(send()).rejects.toThrow();
    expect(handleApiError).toHaveBeenCalledTimes(1);
  });
});
