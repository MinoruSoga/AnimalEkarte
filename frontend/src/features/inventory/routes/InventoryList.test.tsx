import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";
import type { InventoryItem as BackendInventoryItem } from "@/types/generated/models";
import { InventoryList } from "./InventoryList";

// PageLayout の resource prop 経由で PermissionBadges → usePermission → useAuth が呼ばれるため、
// AuthProvider 非依存でレンダーできるよう最小限のモックを用意する（CashRegisterHistoryPage.test.tsx と同パターン）。
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

// BUG-414: category/status の is_not はクライアント側除外が必要なため、PropertyFilter の内部実装
// (ポップオーバー操作) は別レイヤーの責務としてモックし、InventoryList 側の onFilterChange 配線のみに
// 焦点を当てる（OwnersList.filter-wiring.test.tsx と同パターン）。
vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: ({ onFilterChange }: { onFilterChange: (f: unknown[]) => void }) => (
    <div>
      <button
        type="button"
        onClick={() =>
          onFilterChange([
            { key: "category", condition: "is_not", value: "medicine", displayValue: "医薬品以外" },
          ])
        }
      >
        カテゴリ「医薬品以外」で絞り込む
      </button>
      <button
        type="button"
        onClick={() =>
          onFilterChange([
            { key: "status", condition: "is_not", value: "sufficient", displayValue: "十分以外" },
          ])
        }
      >
        ステータス「十分以外」で絞り込む
      </button>
    </div>
  ),
}));

const makeItem = (overrides: Partial<BackendInventoryItem>): BackendInventoryItem => ({
  id: 1,
  clinic_id: 1,
  name: "アイテム",
  category: "medicine",
  quantity: 10,
  unit: "個",
  min_stock_level: 5,
  location: "棚A",
  supplier: "仕入先A",
  status: "sufficient",
  created_at: "2026-07-01T00:00:00Z",
  updated_at: "2026-07-01T00:00:00Z",
  ...overrides,
});

const mockItems: BackendInventoryItem[] = [
  makeItem({ id: 1, name: "ワクチンA", category: "medicine", status: "sufficient" }),
  makeItem({ id: 2, name: "フードB", category: "food", status: "low" }),
];

const renderList = () =>
  render(<InventoryList />, { wrapper: createTestWrapper({ initialEntries: ["/inventory"] }) });

describe("InventoryList — BUG-414 category/status の is_not フィルタ", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("BUG-414: category に is_not 条件を指定すると、一致する item が除外され一致しない item は表示される", async () => {
    server.use(
      http.get("*/v1/inventory", () =>
        HttpResponse.json({ data: mockItems, total: mockItems.length, page: 1, limit: 20 })
      )
    );

    renderList();

    await screen.findByText("ワクチンA");
    expect(screen.getByText("フードB")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "カテゴリ「医薬品以外」で絞り込む" }));

    await waitFor(() => {
      expect(screen.queryByText("ワクチンA")).not.toBeInTheDocument();
    });
    expect(screen.getByText("フードB")).toBeInTheDocument();
  });

  it("BUG-414: status に is_not 条件を指定すると、一致する item が除外され一致しない item は表示される", async () => {
    server.use(
      http.get("*/v1/inventory", () =>
        HttpResponse.json({ data: mockItems, total: mockItems.length, page: 1, limit: 20 })
      )
    );

    renderList();

    await screen.findByText("ワクチンA");
    expect(screen.getByText("フードB")).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "ステータス「十分以外」で絞り込む" }));

    await waitFor(() => {
      expect(screen.queryByText("ワクチンA")).not.toBeInTheDocument();
    });
    expect(screen.getByText("フードB")).toBeInTheDocument();
  });
});
