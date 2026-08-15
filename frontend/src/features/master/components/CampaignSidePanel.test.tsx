import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { CampaignSidePanel } from "./CampaignSidePanel";

vi.mock("@/components/shared/SidePeek", () => ({
  MasterSidePanel: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  StatusToggleButton: () => null,
}));

vi.mock("../api/merchandise-items", () => ({
  useGetAllMerchandiseItems: () => ({ data: [] }),
}));

describe("CampaignSidePanel", () => {
  it("mobileではフォームの複数列を単一列にし、sm以上で既存の2列へ戻す", () => {
    render(
      <CampaignSidePanel
        item={null}
        onClose={() => {}}
        onSave={() => {}}
      />,
    );

    const periodGrid = screen.getByText("開始日").closest(".grid");
    const discountGrid = screen.getByText("割引種別").closest(".grid");
    const categoryGrid = screen.getByText("フード").closest(".grid");

    for (const grid of [periodGrid, discountGrid, categoryGrid]) {
      expect(grid).toHaveClass("w-full", "grid-cols-1", "sm:grid-cols-2");
    }
  });
});
