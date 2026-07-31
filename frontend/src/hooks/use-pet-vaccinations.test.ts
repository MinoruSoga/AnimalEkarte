import { describe, it, expect } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";
import { useGetPetVaccinations } from "./use-pet-vaccinations";
import type { Vaccination } from "@/types/generated/models";

function makeVaccination(overrides: Partial<Vaccination> = {}): Vaccination {
  return {
    id: 1,
    clinic_id: 1,
    vaccine_id: 1,
    date: "2026-03-25T16:30:00Z",
    supplemental: "",
    lot1: "",
    lot2: "",
    lot3: "",
    lot4: "",
    remarks: "",
    created_at: "2026-03-25T16:30:00Z",
    updated_at: "2026-03-25T16:30:00Z",
    ...overrides,
  };
}

describe("useGetPetVaccinations (SD-19: JST 壁日付整形)", () => {
  it("UTC instant を JST 壁日付で整形する（ローカル TZ の getter に依存しない）", async () => {
    // 2026-03-25T16:30:00Z は JST(+9h) では 2026-03-26 01:30 — UTC のまま
    // getFullYear/getMonth/getDate すると 03/25 になり 1 日ずれる（SD-19 の再現条件）。
    server.use(
      http.get("/api/v1/vaccinations", () =>
        HttpResponse.json({
          data: [
            makeVaccination({
              date: "2026-03-25T16:30:00Z",
              next_date: "2026-04-25T16:30:00Z",
            }),
          ],
        }),
      ),
    );

    const { result } = renderHook(() => useGetPetVaccinations("7"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const [item] = result.current.data ?? [];
    expect(item.date).toBe("26/3/26");
    expect(item.next).toBe("26/4/26");
    // SD-19: 判定用 nextDate も表示 next と同じ JST 暦日（UTC 切り出し禁止）
    expect(item.nextDate).toBe("2026-04-26");
  });

  it("UTC 境界 instant の nextDate は formatJSTDate と同じ JST 壁日付になる", async () => {
    // T15:00:00Z = JST 翌日 00:00、T16:30:00Z = JST 翌日 01:30
    // split("T")[0] だと UTC 暦日のまま 1 日ずれる。
    server.use(
      http.get("/api/v1/vaccinations", () =>
        HttpResponse.json({
          data: [
            makeVaccination({
              id: 1,
              date: "2026-03-25T15:00:00Z",
              next_date: "2026-03-25T15:00:00Z",
            }),
            makeVaccination({
              id: 2,
              date: "2026-03-25T16:30:00Z",
              next_date: "2026-03-25T16:30:00Z",
            }),
          ],
        }),
      ),
    );

    const { result } = renderHook(() => useGetPetVaccinations("7"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const items = result.current.data ?? [];
    const byId = Object.fromEntries(items.map((item) => [item.id, item]));

    expect(byId[1].date).toBe("26/3/26");
    expect(byId[1].next).toBe("26/3/26");
    expect(byId[1].nextDate).toBe("2026-03-26");

    expect(byId[2].date).toBe("26/3/26");
    expect(byId[2].next).toBe("26/3/26");
    expect(byId[2].nextDate).toBe("2026-03-26");
  });

  it("未設定の日付は '-' を表示する", async () => {
    server.use(
      http.get("/api/v1/vaccinations", () =>
        HttpResponse.json({ data: [makeVaccination({ next_date: undefined })] }),
      ),
    );

    const { result } = renderHook(() => useGetPetVaccinations("7"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[0].next).toBe("-");
    expect(result.current.data?.[0].nextDate).toBe("");
  });

  it("不正な next_date は判定用 nextDate を空にし期限判定に載せない", async () => {
    server.use(
      http.get("/api/v1/vaccinations", () =>
        HttpResponse.json({
          data: [makeVaccination({ next_date: "not-a-date" as unknown as string })],
        }),
      ),
    );

    const { result } = renderHook(() => useGetPetVaccinations("7"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[0].next).toBe("-");
    expect(result.current.data?.[0].nextDate).toBe("");
  });

  it("petId 未指定ならクエリは無効でフェッチしない", () => {
    const { result } = renderHook(() => useGetPetVaccinations(undefined), {
      wrapper: createTestWrapper(),
    });
    expect(result.current.fetchStatus).toBe("idle");
  });

  it("BIGINT の petId を数値変換せずクエリへ渡す", async () => {
    const largePetId = "9007199254740993";
    let receivedPetId = "";
    server.use(
      http.get("/api/v1/vaccinations", ({ request }) => {
        receivedPetId = new URL(request.url).searchParams.get("pet_id") ?? "";
        return HttpResponse.json({ data: [] });
      }),
    );

    const { result } = renderHook(() => useGetPetVaccinations(largePetId), {
      wrapper: createTestWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(receivedPetId).toBe(largePetId);
  });
});
