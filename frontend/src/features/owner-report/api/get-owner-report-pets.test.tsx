import type { ReactNode } from "react";
import {
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";

import { queryKeys } from "@/lib/query-keys";
import { server } from "@/testing/mocks/node";

import { useGetOwnerReportPets } from "./get-owner-report-pets";

function createWrapper(queryClient: QueryClient) {
  return function TestQueryClientProvider({
    children,
  }: {
    children: ReactNode;
  }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  };
}

describe("useGetOwnerReportPets", () => {
  it("専用 endpoint と cache key を使い、描画項目だけを owner-report 用へ変換する", async () => {
    let requestedPath = "";
    server.use(
      http.get("/api/v1/owners/42/report/pets", ({ request }) => {
        requestedPath = new URL(request.url).pathname;
        return HttpResponse.json({
          data: [
            {
              id: 7,
              name: "ポチ",
              pet_name_kana: "ぽち",
              gender: "male",
              status: "alive",
              birth_date: "2015-04-14T00:00:00+09:00",
              breed: "柴犬",
              color: "赤",
              blood_type: "DEA1.1陽性",
              microchip_number: "392140000123456",
              weight: 12.3,
              neutered_date: "2016-05-20T00:00:00+09:00",
              acquisition_type: "purchased",
              food: "療法食",
              environment: "室内",
              last_visit: "2024-08-25T09:30:00+09:00",
              remarks: "咬傷注意",
              animal_species: { name: "犬" },
              insurance: { name: "アニコム", coverage_rate: 70 },
              danger_level: "high",
              danger_reason: "staff-only",
            },
          ],
        });
      }),
    );
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    const { result } = renderHook(() => useGetOwnerReportPets("42"), {
      wrapper: createWrapper(queryClient),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const expectedPet = {
      id: "7",
      name: "ポチ",
      petNameKana: "ぽち",
      gender: "雄",
      status: "生存",
      birthDate: "2015-04-14",
      breed: "柴犬",
      color: "赤",
      bloodType: "DEA1.1陽性",
      microchipNumber: "392140000123456",
      weight: "12.3",
      neuteredDate: "2016-05-20",
      acquisitionType: "購入",
      food: "療法食",
      environment: "室内",
      lastVisit: "2024-08-25",
      remarks: "咬傷注意",
      species: "犬",
      insuranceName: "アニコム",
      insuranceDetails: "70%補償",
    };
    expect(requestedPath).toBe("/api/v1/owners/42/report/pets");
    expect(result.current.data).toEqual([expectedPet]);
    expect(queryKeys.ownerReportPets("42")).toEqual([
      "owner-report-pets",
      "42",
    ]);
    expect(
      queryClient.getQueryData(queryKeys.ownerReportPets("42")),
    ).toEqual([expectedPet]);
    expect(
      queryClient.getQueryData(
        queryKeys.pets.list("42", { includeDeceased: true }),
      ),
    ).toBeUndefined();
    expect(result.current.data?.[0]).not.toHaveProperty("dangerLevel");
    expect(result.current.data?.[0]).not.toHaveProperty("dangerReason");
  });

  it("ownerId が空なら query を無効化して fetch しない", () => {
    let requestCount = 0;
    server.use(
      http.get("/api/v1/owners/:ownerId/report/pets", () => {
        requestCount += 1;
        return HttpResponse.json({ data: [] });
      }),
    );
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    const { result } = renderHook(() => useGetOwnerReportPets(""), {
      wrapper: createWrapper(queryClient),
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(requestCount).toBe(0);
  });
});
