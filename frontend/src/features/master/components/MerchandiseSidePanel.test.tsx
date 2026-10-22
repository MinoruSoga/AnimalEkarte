import { createMemoryRouter, RouterProvider } from "react-router";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { MerchandiseSidePanel } from "./MerchandiseSidePanel";
import type { FrontendMerchandiseItem } from "../api/merchandise-items";

// EMR-537 E-1: 税率selectorが isMerchandiseTaxRate を呼ぶため、import 欠落があると
// パネル render 時に ReferenceError で落ちる。新規・既存パネルの render 回帰を固定する。
function makeItem(taxRate: string): FrontendMerchandiseItem {
  return {
    id: "merch-1",
    name: "既存品目",
    category: "goods",
    unitPrice: 1200,
    taxType: "excluded",
    taxRate,
    isActive: true,
    sortOrder: 1,
  };
}

function renderPanel(
  item: FrontendMerchandiseItem | null,
  onSave = vi.fn(() => true),
  opts: { onClose?: () => void; readOnly?: boolean } = {},
) {
  // NavigationBlocker (dirty 検知) が useBlocker を呼ぶため data router 配下で描画する
  const router = createMemoryRouter(
    [
      {
        path: "*",
        element: (
          <MerchandiseSidePanel
            item={item}
            onClose={opts.onClose ?? vi.fn()}
            onSave={onSave}
            readOnly={opts.readOnly}
          />
        ),
      },
    ],
    { initialEntries: ["/"] },
  );
  render(<RouterProvider router={router} />);
  return onSave;
}

describe("MerchandiseSidePanel render (EMR-537 E-1)", () => {
  it("新規パネル（item=null）が未定義helper参照なしでrenderされる", () => {
    renderPanel(null);
    expect(document.getElementById("master-title")).not.toBeNull();
    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
  });

  it.each(["0.10", "0.08", "0", "0.08333", "0.10000000000000001"])(
    "既存パネル taxRate=%s も未定義helper参照なしでrenderされる",
    (taxRate) => {
      renderPanel(makeItem(taxRate));
      expect(document.getElementById("master-title")).not.toBeNull();
      expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    },
  );
});

describe("MerchandiseSidePanel exact taxRate payload (EMR-537 AC1)", () => {
  // EMR-530: form 値・保存 payload は入力字句の exact decimal 文字列を保持する
  // （selector の number 契約への合わせ込みで丸めない）。
  it.each(["0.10", "0.08", "0", "0.08333", "0.10000000000000001"])(
    "taxRate=%s を編集せず保存すると payload が同じ文字列を保つ",
    async (taxRate) => {
      const user = userEvent.setup();
      const onSave = renderPanel(makeItem(taxRate));

      const title = document.getElementById("master-title")!;
      await user.type(title, "改");
      await user.click(screen.getByRole("button", { name: "保存" }));

      expect(onSave).toHaveBeenCalledTimes(1);
      expect(onSave).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "既存品目改",
          taxRate,
        }),
      );
    },
  );

  it("新規パネルの既定 taxRate は 0.1 の文字列のまま保存される", async () => {
    const user = userEvent.setup();
    const onSave = renderPanel(null);

    const title = document.getElementById("master-title")!;
    await user.type(title, "新規品目");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ name: "新規品目", taxRate: "0.1" }),
    );
  });
});

describe("MerchandiseSidePanel readOnly / cancel (EMR-537 AC1 挙動不変)", () => {
  it("readOnly では保存・削除ボタンを出さず閉じるだけを表示する", () => {
    renderPanel(
      makeItem("0.1"),
      vi.fn(() => true),
      { readOnly: true },
    );

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "閉じる" }).length).toBeGreaterThan(0);
  });

  it("名称未入力で保存すると onSave を呼ばず名称エラーを出す", async () => {
    const user = userEvent.setup();
    const onSave = renderPanel(null);

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("名称を入力してください")).toBeInTheDocument();
    expect(document.getElementById("master-title-error")).toHaveTextContent(
      "名称を入力してください",
    );
    expect(onSave).not.toHaveBeenCalled();
  });

  it("キャンセルで onClose が呼ばれる", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    renderPanel(
      null,
      vi.fn(() => true),
      { onClose },
    );

    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
