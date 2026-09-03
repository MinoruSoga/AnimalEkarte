import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { CategorySidePanel } from "./ReservationTypeSidePanel";
import type { ReservationType } from "../api/reservation-types";

function makeType(overrides: Partial<ReservationType>): ReservationType {
  return {
    id: "1",
    clinicId: "1",
    name: "テスト区分",
    color: "#3B82F6",
    isActive: true,
    description: "",
    sortOrder: 1,
    groupId: undefined,
    createdAt: "2026-05-29T00:00:00Z",
    updatedAt: "2026-05-29T00:00:00Z",
    reservationDisplayName: "",
    durationMinutes: 15,
    shortName: "",
    showShortName: false,
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    reservationDayOption: "none",
    isInternal: false,
    category: "general",
    parentId: undefined,
    parentName: undefined,
    isLeaf: true,
    depth: 0,
    childIds: [],
    ...overrides,
  };
}

function createWrapper() {
  return createTestWrapper({ router: true });
}

const noop = vi.fn();

describe("CategorySidePanel", () => {
  it("leaf type: ReservationTypeAvailableSlotsSection が render される", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/1/available-slots", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/masters/reservation-types/1/unavailable-times", () =>
        HttpResponse.json([]),
      ),
      http.get("/api/v1/masters/reservation-types/1/occupations", () =>
        HttpResponse.json([]),
      ),
    );

    const leafType = makeType({ id: "1", isLeaf: true });

    render(
      <CategorySidePanel
        item={leafType}
        onClose={noop}
        onSave={noop}
        groups={[]}
      />,
      { wrapper: createWrapper() },
    );

    // AvailableSlotsSection のヘッダーが表示される
    expect(await screen.findByText("予約可能枠")).toBeInTheDocument();
    // 案内文は表示されない
    expect(
      screen.queryByText("子予約区分ごとに予約枠を設定してください"),
    ).not.toBeInTheDocument();
  });

  it("parent type: セクションが render されず、案内文が表示される", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/1/occupations", () =>
        HttpResponse.json([]),
      ),
    );

    const parentType = makeType({ id: "1", isLeaf: false, childIds: ["2", "3"] });

    render(
      <CategorySidePanel
        item={parentType}
        onClose={noop}
        onSave={noop}
        groups={[]}
      />,
      { wrapper: createWrapper() },
    );

    expect(
      await screen.findByText("子予約区分ごとに予約枠を設定してください"),
    ).toBeInTheDocument();
    // AvailableSlotsSection のヘッダーは表示されない
    expect(screen.queryByText("予約可能枠")).not.toBeInTheDocument();
  });
});
