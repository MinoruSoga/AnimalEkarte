import { describe, expect, it, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";
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
  return render(<LineReservationSlotsSettings />, {
    wrapper: createTestWrapper({ initialEntries: [initialEntry] }),
  });
}

/** BE の tree response 形式 (root は children 付き) */
function reservationType(
  id: number,
  name: string,
  isActive: boolean,
  children: ReturnType<typeof reservationType>[] = [],
) {
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
    reservation_display_name: "",
    duration_minutes: 15,
    short_name: "",
    show_short_name: false,
    reservation_visible: true,
    reservation_comment: "",
    reservation_image_url: "",
    reservation_day_option: "none",
    is_internal: false,
    category: "general",
    children,
  };
}

afterEach(() => {
  server.resetHandlers();
});

describe("LineReservationSlotsSettings", () => {
  it("tree/calendar は mobile で縦積み・全幅、md 以上で横並びにする", () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () => HttpResponse.json([])),
    );
    const { container } = renderPage("/line-reservation/slots");

    const pageShell = container.firstElementChild;
    const splitLayout = pageShell?.children.item(2);
    const treePanel = splitLayout?.firstElementChild;

    expect(splitLayout).toHaveClass("flex-col", "md:flex-row");
    expect(treePanel).toHaveClass("w-full", "md:w-[260px]");
  });

  it("typeId 未指定なら最初の有効な leaf のカレンダーを表示する", async () => {
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

  it("typeId が親ノード ID の場合 → 最初の leaf ID に URL 正規化される", async () => {
    let requestedTypeId: string | null = null;
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          reservationType(1, "LINEコース", true, [
            reservationType(2, "初診コース", true),
            reservationType(3, "再診コース", true),
          ]),
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

    // 親ノード ID=1 を指定
    renderPage("/line-reservation/slots?typeId=1");

    // 最初の子 leaf (ID=2) に正規化される
    await waitFor(() => {
      expect(requestedTypeId).toBe("2");
    });
  });

  it("breadcrumb に「親名 / 子名」が表示される", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          reservationType(1, "LINEコース", true, [
            reservationType(2, "初診コース", true),
          ]),
        ]),
      ),
      http.get("/api/v1/masters/reservation-types/:id/available-slots", () =>
        HttpResponse.json([]),
      ),
    );

    renderPage("/line-reservation/slots?typeId=2");

    expect(await screen.findByText("LINEコース / 初診コース")).toBeInTheDocument();
  });

  it("root-only leaf は breadcrumb が区分名のみ", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([reservationType(6, "一般診療", true)]),
      ),
      http.get("/api/v1/masters/reservation-types/:id/available-slots", () =>
        HttpResponse.json([]),
      ),
    );

    renderPage("/line-reservation/slots?typeId=6");

    // ツリーとパンくずの両方に「一般診療」が出るため findAllByText でまとめて待機
    const elements = await screen.findAllByText("一般診療");
    expect(elements.length).toBeGreaterThanOrEqual(1);
    // breadcrumb にスラッシュは含まない
    expect(screen.queryByText(/\/.*一般診療|一般診療.*\//)).not.toBeInTheDocument();
  });

  it("全 leaf が inactive でも最初の inactive leaf を選択する", async () => {
    let requestedTypeId: string | null = null;
    server.use(
      http.get("/api/v1/masters/reservation-types", () =>
        HttpResponse.json([
          reservationType(5, "停止中A", false),
          reservationType(6, "停止中B", false),
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

    // active leaf がないので最初の leaf (ID=5) を選択
    await waitFor(() => {
      expect(requestedTypeId).toBe("5");
    });
  });
});
