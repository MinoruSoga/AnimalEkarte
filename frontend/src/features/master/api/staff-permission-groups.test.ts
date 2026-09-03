import { describe, it, expect, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { useGetAllStaffPermissionGroupMap } from "./staff-permission-groups";

afterEach(() => {
  server.resetHandlers();
});

describe("useGetAllStaffPermissionGroupMap（FE5-17: 一括取得の握りつぶし是正）", () => {
  it("一括取得: 404 のスタッフは空配列として継続する", async () => {
    server.use(
      http.get("/api/v1/masters/staffs/:id/permission-groups", ({ params }) => {
        if (params.id === "1") {
          return HttpResponse.json({ group_ids: [10, 20] });
        }
        return HttpResponse.json(null, { status: 404 });
      }),
    );

    const { result } = renderHook(() => useGetAllStaffPermissionGroupMap(["1", "2"]), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.get("1")).toEqual(["10", "20"]);
    expect(result.current.data?.get("2")).toEqual([]);
  });

  it("一括取得: 500 が 1 件でもあれば全体が reject する", async () => {
    server.use(
      http.get("/api/v1/masters/staffs/:id/permission-groups", ({ params }) => {
        if (params.id === "1") {
          return HttpResponse.json({ group_ids: [10] });
        }
        return HttpResponse.json(null, { status: 500 });
      }),
    );

    const { result } = renderHook(() => useGetAllStaffPermissionGroupMap(["1", "2"]), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
