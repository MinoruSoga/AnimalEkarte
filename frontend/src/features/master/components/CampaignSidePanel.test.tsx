import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { CampaignSidePanel } from "./CampaignSidePanel";

vi.mock("@/components/shared/SidePeek", () => ({
  MasterSidePanel: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  StatusToggleButton: () => null,
  PropertyRow: ({ label, children }: { label?: ReactNode; children: ReactNode }) => (
    <div>
      {label}
      {children}
    </div>
  ),
}));

vi.mock("../api/merchandise-items", () => ({
  useGetAllMerchandiseItems: () => ({ data: [] }),
}));

describe("CampaignSidePanel", () => {
  // e93ac185b で responsive grid から PropertyRow 行構成へ移行済み。
  // 行入力の存在と、対象カテゴリの単一列 grid を現行仕様として固定する。
  it("PropertyRow 行構成で開始日・割引種別・対象カテゴリを描画する", () => {
    render(
      <CampaignSidePanel
        item={null}
        onClose={() => {}}
        onSave={() => {}}
      />,
    );

    expect(screen.getByLabelText("開始日")).toBeInTheDocument();
    expect(screen.getByLabelText("終了日")).toBeInTheDocument();
    expect(screen.getByLabelText("割引種別")).toBeInTheDocument();

    const categoryGrid = screen.getByText("フード").closest(".grid");
    expect(categoryGrid).toHaveClass("w-full", "grid-cols-1");
  });
});
