import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";

import { useGetPetCheckupResults } from "./use-pet-checkup-results";

describe("useGetPetCheckupResults", () => {
  it("BIGINT の petId を数値変換せずクエリへ渡す", async () => {
    const largePetId = "9007199254740993";
    let receivedPetId = "";
    server.use(
      http.get("/api/v1/checkups/field-results", ({ request }) => {
        receivedPetId = new URL(request.url).searchParams.get("pet_id") ?? "";
        return HttpResponse.json([]);
      }),
    );

    const { result } = renderHook(() => useGetPetCheckupResults(largePetId), {
      wrapper: createTestWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(receivedPetId).toBe(largePetId);
  });
});
