import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { useTreatmentPackageApply } from "./use-treatment-package-apply";
import type { TreatmentPackagePreview } from "../api/treatment-package-applications";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn(), post: vi.fn(), patch: vi.fn(), delete: vi.fn() },
}));
beforeEach(() => vi.clearAllMocks());

const wrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
};

const preview = {
  record_id: "91",
  clinic_id: "7",
  pet_id: "801",
  package_id: "201",
  package_version: "3",
  plan_collection_version: "17",
  version_state: "ready",
  preview_token: `pkg-preview-v1:${"a".repeat(64)}`,
  applicable: true,
  total: { status: "unresolved", amount: null },
  items: [
    {
      component_key: "a0000000-0000-4000-8000-000000000011",
      item: { kind: "consultation", consultation_id: "101" },
      quantity: "1",
      name: "診察",
      unit_label: null,
      source_stamp: { native_version: "2", dependency_token: `pkg-source-v1:${"0".repeat(64)}` },
      price: {
        unit_price: "110",
        tax_type: "included",
        tax_rate: "0.1",
        is_non_insurance: true,
        insurance_setting_state: "explicit",
        price_source: "master",
        tax_source: "master",
        insurance_source: "master",
        amount_basis: "included",
        uncertainty: [],
      },
      classification: "clinical",
      issues: [],
    },
  ],
} as TreatmentPackagePreview;

const ctx = {
  editSessionRef: "a0000000-0000-4000-8000-000000000002",
  editEpoch: "9",
  ownerExpectedVersion: "17",
};

const ok = (key: string) => ({
  status: "succeeded",
  operation_key: key,
  result_refs: { record_id: "91", application_id: "302" },
  committed_versions: { application: "1", plan_items: "18" },
  recorded_at: "t",
});

describe("useTreatmentPackageApply（同key照会・成功後の再取得・scope破棄）", () => {
  it("明示 apply → succeeded → application id を保持し plan/applications を再取得", async () => {
    vi.mocked(axios.post).mockImplementation(async (_url, body) => ({
      data: ok((body as { operation_key: string }).operation_key),
    }));
    const { result } = renderHook(
      () =>
        useTreatmentPackageApply({
          scope: { recordId: "91", clinicId: "7", petId: "801" },
          editContext: ctx,
          runtimeAvailable: true,
        }),
      { wrapper: wrapper() },
    );
    await result.current.submit(preview);
    expect(result.current.state.kind).toBe("succeeded");
    expect(axios.post).toHaveBeenCalledTimes(1);
    const body = vi.mocked(axios.post).mock.calls[0]![1] as Record<string, string>;
    expect(body.package_id).toBe("201");
    expect(body.edit_epoch).toBe("9");
    // 2回目の明示 apply は別 application（別key）
    await result.current.submit(preview);
    expect(vi.mocked(axios.post).mock.calls[1]![1]).not.toEqual(body);
  });

  it("応答断は unknown → 元key照会のみ（盲目再送なし）", async () => {
    vi.mocked(axios.post).mockRejectedValue(new Error("network"));
    vi.mocked(axios.get).mockImplementation(async (url) => {
      expect(String(url)).toContain("/operations/");
      return { data: ok("a0000000-0000-4000-8000-000000000001") };
    });
    const { result } = renderHook(
      () =>
        useTreatmentPackageApply({
          scope: { recordId: "91" },
          editContext: ctx,
          runtimeAvailable: true,
        }),
      { wrapper: wrapper() },
    );
    await result.current.submit(preview);
    expect(result.current.state.kind).toBe("unknown");
    await result.current.resolveUnknown();
    expect(result.current.state.kind).toBe("succeeded");
    // POST は1回だけ（照会は GET のみ）
    expect(axios.post).toHaveBeenCalledTimes(1);
    expect(String(vi.mocked(axios.get).mock.calls[0]![0])).toContain(
      "/v1/medical-records/91/operations/",
    );
  });

  it("runtime 未提供/editContext なしでは送信しない", async () => {
    const { result } = renderHook(
      () => useTreatmentPackageApply({ scope: { recordId: "91" }, editContext: null }),
      { wrapper: wrapper() },
    );
    await result.current.submit(preview);
    expect(result.current.state.kind).toBe("idle");
    expect(axios.post).not.toHaveBeenCalled();
  });

  it("in-progress は unknown 維持で成功にしない", async () => {
    vi.mocked(axios.post).mockResolvedValue({
      data: { status: "operation-in-progress", operation_key: uuid() },
    });
    const { result } = renderHook(
      () =>
        useTreatmentPackageApply({
          scope: { recordId: "91" },
          editContext: ctx,
          runtimeAvailable: true,
        }),
      { wrapper: wrapper() },
    );
    await result.current.submit(preview);
    expect(result.current.state.kind).toBe("unknown");
  });
});

function uuid() {
  return "a0000000-0000-4000-8000-000000000001";
}
