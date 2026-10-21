import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import {
  useDeleteTreatmentPackage,
  useReorderTreatmentPackages,
  useTreatmentPackageEditor,
  useUpdateTreatmentPackage,
} from "./use-treatment-packages";
import { draftFromPackage } from "../lib/treatment-package-form";
import type { TreatmentPackage, TreatmentPackagePatch } from "../api/treatment-packages";

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
    vi.mocked(axios.patch).mockRejectedValue(conflict());
    const { result } = renderHook(() => useTreatmentPackageEditor(pkg), { wrapper: wrapper() });
    const draft = result.current.draft!;
    draft.name = "変更";
    act(() => result.current.updateDraft({ name: "変更" }));
    const outcome = await result.current.save({
      create: vi.fn(),
      update: (id, body) => useUpdateImpl(id, body),
    });
    async function useUpdateImpl(id: string, body: TreatmentPackagePatch) {
      await axios.patch(`/v1/masters/treatment-packages/${id}`, body);
      throw conflict();
    }
    expect(outcome.kind === "error" || outcome.kind === "conflict").toBe(true);
    expect(result.current.draft!.name).toBe("変更");
  });

  it("create 応答不明は create-unknown で盲目再送しない", async () => {
    const { result } = renderHook(() => useTreatmentPackageEditor("new"), { wrapper: wrapper() });
    act(() => {
      result.current.updateDraft({ name: "n" });
      result.current.draft!.components[0]!.masterId = "1";
    });
    const create = vi.fn().mockRejectedValue(new AxiosError("network"));
    const outcome = await result.current.save({ create, update: vi.fn() });
    expect(outcome.kind).toBe("create-unknown");
    expect(create).toHaveBeenCalledTimes(1);
    expect(result.current.createUnknown).toBe(true);
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
