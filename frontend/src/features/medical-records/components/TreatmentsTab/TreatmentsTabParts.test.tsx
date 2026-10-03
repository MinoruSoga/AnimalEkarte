import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { TreatmentAddControls } from "./TreatmentsTabParts";
import type { TreatmentItemType } from "../../types";

// ─────────────────────────────────────────────────────────────
// EMR-226: raw <select> → ui/select 移行後の保持動作
// TreatmentAddControls は純粋な表示コンポーネントのため props 駆動で直接検証する。
// ─────────────────────────────────────────────────────────────

interface RenderOptions {
  canCreate?: boolean;
  isAdding?: boolean;
  addItemType?: TreatmentItemType;
  addAdminRoute?: string;
}

function renderAddControls({
  canCreate = true,
  isAdding = true,
  addItemType = "consultation",
  addAdminRoute = "",
}: RenderOptions = {}) {
  const props = {
    canCreate,
    isAdding,
    isPending: false,
    addItemType,
    addContent: "",
    addAdminRoute,
    onItemTypeChange: vi.fn(),
    onContentChange: vi.fn(),
    onAdminRouteChange: vi.fn(),
    onSubmit: vi.fn(),
    onCancel: vi.fn(),
    onOpenSearch: vi.fn(),
    onStartAdding: vi.fn(),
  };
  render(<TreatmentAddControls {...props} />);
  return props;
}

describe("TreatmentAddControls — shared ui/select (EMR-226)", () => {
  it("種別セレクトは既存値を trigger に表示する", () => {
    renderAddControls({ addItemType: "procedure" });
    expect(screen.getByRole("combobox", { name: "種別" })).toHaveTextContent("処置");
  });

  it("種別を変更すると onItemTypeChange が選択値で呼ばれる", async () => {
    const user = userEvent.setup();
    const props = renderAddControls({ addItemType: "consultation" });

    await user.click(screen.getByRole("combobox", { name: "種別" }));
    await user.click(screen.getByRole("option", { name: "処置" }));

    expect(props.onItemTypeChange).toHaveBeenCalledWith("procedure");
  });

  it("addItemType=medicine のときだけ投与方法セレクトを表示する（薬-only conditional）", () => {
    const props = {
      canCreate: true,
      isAdding: true,
      isPending: false,
      addItemType: "medicine" as TreatmentItemType,
      addContent: "",
      addAdminRoute: "",
      onItemTypeChange: vi.fn(),
      onContentChange: vi.fn(),
      onAdminRouteChange: vi.fn(),
      onSubmit: vi.fn(),
      onCancel: vi.fn(),
      onOpenSearch: vi.fn(),
      onStartAdding: vi.fn(),
    };
    const { rerender } = render(<TreatmentAddControls {...props} />);
    expect(screen.getByRole("combobox", { name: "投与方法" })).toBeInTheDocument();

    rerender(<TreatmentAddControls {...props} addItemType="procedure" />);
    expect(screen.queryByRole("combobox", { name: "投与方法" })).not.toBeInTheDocument();
  });

  it("投与方法は未選択時にプレースホルダラベルを表示し、選択で onAdminRouteChange が呼ばれる", async () => {
    const user = userEvent.setup();
    const props = renderAddControls({ addItemType: "medicine", addAdminRoute: "" });

    const routeTrigger = screen.getByRole("combobox", { name: "投与方法" });
    expect(routeTrigger).toHaveTextContent("投与方法を選択");

    await user.click(routeTrigger);
    await user.click(screen.getByRole("option", { name: "経口" }));

    expect(props.onAdminRouteChange).toHaveBeenCalledWith("経口");
  });

  it("投与方法は選択後に未選択肢へ戻せる（空値コールバック）", async () => {
    const user = userEvent.setup();
    const props = renderAddControls({ addItemType: "medicine", addAdminRoute: "経口" });

    const routeTrigger = screen.getByRole("combobox", { name: "投与方法" });
    expect(routeTrigger).toHaveTextContent("経口");

    await user.click(routeTrigger);
    await user.click(screen.getByRole("option", { name: "投与方法を選択" }));

    expect(props.onAdminRouteChange).toHaveBeenCalledWith("");
  });

  it("種別セレクトはキーボードだけで操作できる（Enter で開き矢印と Enter で確定）", async () => {
    const user = userEvent.setup();
    const props = renderAddControls({ addItemType: "consultation" });

    const trigger = screen.getByRole("combobox", { name: "種別" });
    trigger.focus();
    await user.keyboard("{Enter}");
    // consultation（先頭・選択中）→ procedure へ移動して確定
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{Enter}");

    expect(props.onItemTypeChange).toHaveBeenCalledWith("procedure");
  });

  it("canCreate=false では追加導線ごと何も描画しない（locked-state denial）", () => {
    renderAddControls({ canCreate: false, isAdding: false });

    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
