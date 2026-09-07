import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import { useGetExaminationPrintSnapshot } from "./get-examination-print-snapshot";

const createWrapper = createTestWrapper;

describe("BUG-20260906-004 useGetExaminationPrintSnapshot", () => {
  it("enabled が false のとき print-snapshot を取得しない", async () => {
    let fetched = false;
    server.use(
      http.get("/api/v1/examinations/:id/print-snapshot", () => {
        fetched = true;
        return HttpResponse.json({ error: "not found" }, { status: 404 });
      }),
    );

    const { result } = renderHook(
      () => useGetExaminationPrintSnapshot("11012293", { enabled: false }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => {
      expect(result.current.fetchStatus).toBe("idle");
    });
    expect(result.current.isFetching).toBe(false);
    expect(fetched).toBe(false);
  });

  it("enabled 省略時は id があれば取得する", async () => {
    server.use(
      http.get("/api/v1/examinations/:id/print-snapshot", () =>
        HttpResponse.json({
          examination_id: 1,
          clinic_id: 1,
          version: 1,
          kind: "working",
          status: "completed",
          print_boundary: "draft",
          date: "2026-09-06T00:00:00+09:00",
          result_summary: "",
          machine: "",
          exam_type_id: 1,
          display: {
            medical_record_no: "",
            pet_name: "",
            medical_record_owner_name: "",
            pet_owner_name: "",
            species_name: "",
            exam_type_name: "",
            doctor_name: "",
          },
          items: [],
        }),
      ),
    );

    const { result } = renderHook(() => useGetExaminationPrintSnapshot("1"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    expect(result.current.data?.examinationId).toBe("1");
  });
});
