import { describe, it, expect } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { useMedicalRecordsList } from "./use-medical-records";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";

const createWrapper = createTestWrapper;

function mockList() {
  let capturedUrl: URL | undefined;
  server.use(
    http.get("/api/v1/medical-records", ({ request }) => {
      capturedUrl = new URL(request.url);
      return HttpResponse.json({ data: [], total: 0, page: 1, limit: 20 });
    }),
  );
  return () => capturedUrl;
}

// BUG-B1: PropertyFilter の ActiveFilter → server-side query 変換の回帰防止
describe("useMedicalRecordsList", () => {
  it("診療日フィルタを start_date/end_date に変換する", async () => {
    const getUrl = mockList();
    const activeFilters: ActiveFilter[] = [
      { key: "date", condition: "is_between", value: { from: "2026-01-01", to: "2026-01-31" }, displayValue: "" },
    ];

    const { result } = renderHook(
      () => useMedicalRecordsList({ searchTerm: "", activeFilters, page: 1 }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("start_date")).toBe("2026-01-01");
    expect(getUrl()?.searchParams.get("end_date")).toBe("2026-01-31");
  });

  it("ステータスフィルタ（is条件・日本語ラベル）を BE enum に変換して送信する", async () => {
    const getUrl = mockList();
    const activeFilters: ActiveFilter[] = [
      { key: "status", condition: "is", value: "確定済", displayValue: "確定済" },
    ];

    const { result } = renderHook(
      () => useMedicalRecordsList({ searchTerm: "", activeFilters, page: 1 }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("status")).toBe("finalized");
  });

  it("ステータスフィルタが is 以外の条件のときは送信しない（server は単一値一致のみ対応）", async () => {
    const getUrl = mockList();
    const activeFilters: ActiveFilter[] = [
      { key: "status", condition: "is_not", value: "確定済", displayValue: "確定済ではない" },
    ];

    const { result } = renderHook(
      () => useMedicalRecordsList({ searchTerm: "", activeFilters, page: 1 }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("status")).toBeNull();
  });

  it("担当医・種フィルタ（IDベース）を doctor_id/animal_species_id に変換する", async () => {
    const getUrl = mockList();
    const activeFilters: ActiveFilter[] = [
      { key: "doctor", condition: "is", value: "5", displayValue: "山田医師" },
      { key: "species", condition: "is", value: "1", displayValue: "犬" },
    ];

    const { result } = renderHook(
      () => useMedicalRecordsList({ searchTerm: "", activeFilters, page: 1 }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("doctor_id")).toBe("5");
    expect(getUrl()?.searchParams.get("animal_species_id")).toBe("1");
  });

  it("検索語・ページ番号をそのまま渡す", async () => {
    const getUrl = mockList();

    const { result } = renderHook(
      () => useMedicalRecordsList({ searchTerm: "田中太郎", activeFilters: [], page: 3, limit: 20 }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("search")).toBe("田中太郎");
    expect(getUrl()?.searchParams.get("page")).toBe("3");
  });

  // B-1 follow-up: 列ソート server 化
  it("sort/order をそのまま BE query に渡す", async () => {
    const getUrl = mockList();

    const { result } = renderHook(
      () => useMedicalRecordsList({ searchTerm: "", activeFilters: [], page: 1, sort: "owner_name", order: "asc" }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("sort")).toBe("owner_name");
    expect(getUrl()?.searchParams.get("order")).toBe("asc");
  });

  it("sort 未指定のときは sort/order を送信しない（BE 既定順に委譲）", async () => {
    const getUrl = mockList();

    const { result } = renderHook(
      () => useMedicalRecordsList({ searchTerm: "", activeFilters: [], page: 1 }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(getUrl()?.searchParams.get("sort")).toBeNull();
    expect(getUrl()?.searchParams.get("order")).toBeNull();
  });
});
