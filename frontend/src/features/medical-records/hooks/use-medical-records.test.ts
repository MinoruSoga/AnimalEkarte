import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import { useMedicalRecordsList } from "./use-medical-records";

function captureListUrl() {
  let capturedUrl: URL | undefined;
  server.use(
    http.get("/api/v1/medical-records", ({ request }) => {
      capturedUrl = new URL(request.url);
      return HttpResponse.json({ data: [], total: 0, page: 1, limit: 20 });
    }),
  );
  return () => capturedUrl;
}

const activeFilters: ActiveFilter[] = [];

describe("useMedicalRecordsList medicine filter", () => {
  it("薬剤フィルタを medicine_id クエリへ伝播する", async () => {
    const getUrl = captureListUrl();
    const { result } = renderHook(
      () =>
        useMedicalRecordsList({
          searchTerm: "",
          activeFilters: [
            {
              key: "medicine",
              condition: "is",
              value: "7",
              displayValue: "アモキシシリン",
            },
          ],
          page: 1,
        }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("medicine_id")).toBe("7");
  });

  it("処置・診察フィルタを procedure_id / consultation_id クエリへ伝播する", async () => {
    const getUrl = captureListUrl();
    const { result } = renderHook(
      () =>
        useMedicalRecordsList({
          searchTerm: "",
          activeFilters: [
            { key: "procedure", condition: "is", value: "11", displayValue: "切開" },
            { key: "consultation", condition: "is", value: "13", displayValue: "再診" },
            { key: "inventory", condition: "is", value: "71", displayValue: "シリンジ" },
          ],
          page: 1,
        }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("procedure_id")).toBe("11");
    expect(getUrl()?.searchParams.get("consultation_id")).toBe("13");
    expect(getUrl()?.searchParams.get("inventory_id")).toBe("71");
  });

  it("薬剤フィルタ未指定なら medicine_id を送信しない", async () => {
    const getUrl = captureListUrl();
    const { result } = renderHook(
      () =>
        useMedicalRecordsList({
          searchTerm: "",
          activeFilters,
          page: 1,
        }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.has("medicine_id")).toBe(false);
  });
});

describe("useMedicalRecordsList pet_id filter", () => {
  it("petId を MedicalRecordFilters.petId へ伝播する", async () => {
    const getUrl = captureListUrl();
    const { result } = renderHook(
      () =>
        useMedicalRecordsList({
          searchTerm: "",
          activeFilters,
          petId: "22",
          page: 1,
        }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("pet_id")).toBe("22");
  });

  it("petId 未指定なら pet_id を送信しない", async () => {
    const getUrl = captureListUrl();
    const { result } = renderHook(
      () =>
        useMedicalRecordsList({
          searchTerm: "",
          activeFilters,
          page: 1,
        }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.has("pet_id")).toBe(false);
  });
});
