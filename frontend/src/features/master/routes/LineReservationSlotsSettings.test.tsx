import { describe, expect, it, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { LineReservationSlotsSettings } from "./LineReservationSlotsSettings";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  }),
}));

function renderPage(initialEntry: string) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <QueryClientProvider client={queryClient}>
        <LineReservationSlotsSettings />
      </QueryClientProvider>
    </MemoryRouter>,
  );
}

function reservationType(id: number, name: string, isActive: boolean) {
  return {
    id,
    clinic_id: 1,
    name,
    is_active: isActive,
    color: "#3B82F6",
    description: "",
    sort_order: id,
    created_at: "2026-05-29T00:00:00Z",
    updated_at: "2026-05-29T00:00:00Z",
  };
}

afterEach(() => {
  server.resetHandlers();
});

describe("LineReservationSlotsSettings", () => {
  it("typeId 未指定なら最初の有効な予約区分のカレンダーを表示する", async () => {
    let requestedTypeId: string | null = null;
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          reservationType(5, "停止中の区分", false),
          reservationType(6, "一般診療", true),
        ]),
      ),
      http.get(
        "/api/v1/masters/reservation-types/:id/available-slots",
        ({ params }) => {
          requestedTypeId = String(params.id);
          return HttpResponse.json([]);
        },
      ),
    );

    renderPage("/line-reservation/slots");

    expect(await screen.findByText("LINE予約枠")).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: "今日" })).toBeInTheDocument();

    // 無効区分(5)は飛ばし、最初の有効区分(6)の枠を取得する
    await waitFor(() => {
      expect(requestedTypeId).toBe("6");
    });
  });

  it("typeId クエリパラメータで予約区分を事前選択する", async () => {
    let requestedTypeId: string | null = null;
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          reservationType(6, "一般診療", true),
          reservationType(7, "トリミング", true),
        ]),
      ),
      http.get(
        "/api/v1/masters/reservation-types/:id/available-slots",
        ({ params }) => {
          requestedTypeId = String(params.id);
          return HttpResponse.json([]);
        },
      ),
    );

    renderPage("/line-reservation/slots?typeId=7");

    await waitFor(() => {
      expect(requestedTypeId).toBe("7");
    });
  });

  it("typeId で無効区分を明示指定した場合は別区分へフォールバックしない", async () => {
    let requestedTypeId: string | null = null;
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          reservationType(5, "停止中の区分", false),
          reservationType(6, "一般診療", true),
        ]),
      ),
      http.get(
        "/api/v1/masters/reservation-types/:id/available-slots",
        ({ params }) => {
          requestedTypeId = String(params.id);
          return HttpResponse.json([]);
        },
      ),
    );

    renderPage("/line-reservation/slots?typeId=5");

    await waitFor(() => {
      expect(requestedTypeId).toBe("5");
    });
  });
});
