import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { CampaignSidePanel } from "./CampaignSidePanel";

vi.mock("@/components/shared/SidePeek", () => ({
  MasterSidePanel: ({
    children,
    onTitleChange,
    onSave,
  }: {
    children: ReactNode;
    onTitleChange?: (value: string) => void;
    onSave?: () => void;
  }) => (
    <div>
      <input aria-label="名称" onChange={(e) => onTitleChange?.(e.target.value)} />
      <button type="button" onClick={() => onSave?.()}>
        保存
      </button>
      {children}
    </div>
  ),
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
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={() => {}} />);

    expect(screen.getByLabelText("開始日")).toBeInTheDocument();
    expect(screen.getByLabelText("終了日")).toBeInTheDocument();
    expect(screen.getByLabelText("割引種別")).toBeInTheDocument();

    const categoryGrid = screen.getByText("フード").closest(".grid");
    expect(categoryGrid).toHaveClass("w-full", "grid-cols-1");
  });

  // EMR-319: 開始日・終了日の必須エラーで両 date Input に aria-invalid+describedby。
  it("期間未入力で保存すると開始日・終了日が aria-invalid で共通エラー文を指す", async () => {
    const user = userEvent.setup();
    render(<CampaignSidePanel item={null} onClose={() => {}} onSave={() => {}} />);

    await user.type(screen.getByLabelText("名称"), "テストキャンペーン");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("開始日・終了日を入力してください");
    expect(alert).toHaveAttribute("id", "campaign-period-error");

    for (const label of ["開始日", "終了日"]) {
      const input = screen.getByLabelText(label);
      expect(input).toHaveAttribute("aria-invalid", "true");
      expect(input).toHaveAttribute("aria-describedby", "campaign-period-error");
    }
  });
});
