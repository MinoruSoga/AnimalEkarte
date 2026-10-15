import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryCache, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { describe, expect, it, vi } from "vitest";
import { server } from "@/testing/mocks/node";
import { useStaffSettingsLookups } from "./use-staff-settings-lookups";

const { permission } = vi.hoisted(() => ({ permission: { groupsView: false } }));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: (resource: string) => ({
    canView: resource === "master-permission" ? permission.groupsView : true,
  }),
}));
vi.mock("../api/staffs", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/staffs")>();
  return {
    ...actual,
    useGetStaffs: () => ({ data: [{ id: "1", name: "合成スタッフ" }] }),
    useGetClinicsList: () => ({ data: [] }),
  };
});
vi.mock("../api/occupations", () => ({ useGetAllOccupations: () => ({ data: [] }) }));
vi.mock("../api/reservation-types", () => ({ useGetReservationTypes: () => ({ data: [] }) }));

describe("staff permission-group lookup boundary", () => {
  it.each([false, true])(
    "staff view allowed / group view %s gates both requests",
    async (groupsView) => {
      permission.groupsView = groupsView;
      const requests: string[] = [];
      const permissionToast = vi.fn();
      server.use(
        http.get("/api/v1/masters/permission-groups", () => {
          requests.push("groups");
          return groupsView
            ? HttpResponse.json([
                { id: 7, clinic_id: 1, name: "合成権限", is_active: true, rules: [] },
              ])
            : new HttpResponse(null, { status: 403 });
        }),
        http.get("/api/v1/masters/staffs/1/permission-groups", () => {
          requests.push("staff-groups");
          return groupsView
            ? HttpResponse.json({ group_ids: [7] })
            : new HttpResponse(null, { status: 403 });
        }),
      );
      const client = new QueryClient({
        queryCache: new QueryCache({ onError: permissionToast }),
        defaultOptions: { queries: { retry: false } },
      });
      function Wrapper({ children }: { children: ReactNode }) {
        return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
      }
      const { result } = renderHook(() => useStaffSettingsLookups(), { wrapper: Wrapper });
      if (groupsView) {
        await waitFor(() => expect(result.current.allGroups[0]?.name).toBe("合成権限"));
        await waitFor(() => expect(requests).toContain("staff-groups"));
        await waitFor(() =>
          expect(result.current.groupsByStaffId.get("1")?.[0]?.name).toBe("合成権限"),
        );
      } else {
        await act(async () => {});
        expect(requests).toEqual([]);
        expect(result.current.allGroups).toEqual([]);
        expect(result.current.groupsByStaffId.has("1")).toBe(false);
      }
      expect(permissionToast).not.toHaveBeenCalled();
    },
  );
});
