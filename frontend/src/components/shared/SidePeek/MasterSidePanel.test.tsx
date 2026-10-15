import { type ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

import { MasterSidePanel } from "./MasterSidePanel";

function renderPanel({
  isNew = true,
  isPending,
  onClose = vi.fn(),
  children,
}: {
  isNew?: boolean;
  isPending?: boolean;
  onClose?: () => void;
  children?: ReactNode;
} = {}) {
  render(
    <MasterSidePanel
      isNew={isNew}
      title=""
      onTitleChange={vi.fn()}
      onClose={onClose}
      icon={<svg data-testid="panel-icon" />}
      isPending={isPending}
    >
      {children}
    </MasterSidePanel>,
  );
  return { onClose };
}

describe("MasterSidePanel 非モーダル dialog の支援技術契約 (EMR-401 C5-4)", () => {
  it("isNew=true では role=dialog と accessible name 新規登録 を公開し、aria-modal を付けない", () => {
    renderPanel({ isNew: true });

    const dialog = screen.getByRole("dialog", { name: "新規登録" });
    expect(dialog).not.toHaveAttribute("aria-modal", "true");
  });

  it("isNew=false では accessible name が 編集 になる", () => {
    renderPanel({ isNew: false });

    expect(screen.getByRole("dialog", { name: "編集" })).toBeInTheDocument();
  });

  it("マウント時にタイトル入力へ自動フォーカスする（既存契約の維持）", () => {
    renderPanel();

    expect(document.getElementById("master-title")).toHaveFocus();
  });

  it("Escape で onClose を呼ぶ", () => {
    const { onClose } = renderPanel();

    fireEvent.keyDown(document.getElementById("master-title")!, { key: "Escape" });
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("isPending 中は Escape で閉じない", () => {
    const { onClose } = renderPanel({ isPending: true });

    fireEvent.keyDown(document.getElementById("master-title")!, { key: "Escape" });
    expect(onClose).not.toHaveBeenCalled();
  });

  it("ネストしたポップオーバー等が Escape を defaultPrevented した場合は閉じない", () => {
    const { onClose } = renderPanel({
      children: (
        <input
          aria-label="nested"
          onKeyDown={(e) => {
            if (e.key === "Escape") e.preventDefault();
          }}
        />
      ),
    });

    fireEvent.keyDown(screen.getByLabelText("nested"), { key: "Escape" });
    expect(onClose).not.toBeCalled();
  });

  it("アンマウント時に開いたトリガーへフォーカスを戻す", () => {
    render(<button type="button">open-panel</button>);
    const trigger = screen.getByRole("button", { name: "open-panel" });
    trigger.focus();

    const panel = render(
      <MasterSidePanel isNew title="" onTitleChange={vi.fn()} onClose={vi.fn()} icon={<svg />}>
        {null}
      </MasterSidePanel>,
    );

    // パネル mount 時点では trigger が activeElement として捕捉されている
    expect(document.getElementById("master-title")).toHaveFocus();

    panel.unmount();
    expect(trigger).toHaveFocus();
  });
});

// EMR-411 C-3: tablet overlay は workspace コンテキストでのみ成立する
// opt-in。MasterSidePanel は既定で relative（セクション内埋め込み等を壊さない）、
// tabletOverlay=true で SidePeekPanel へ overlay 契約を委譲する。
describe("MasterSidePanel tabletOverlay forwarding (EMR-411)", () => {
  it("既定ではパネルは relative のまま（max-xl:absolute 無し）", () => {
    renderPanel();

    const panel = document.querySelector("[data-side-peek]");
    expect(panel).toHaveClass("relative");
    expect(panel).not.toHaveClass("max-xl:absolute");
  });

  it("tabletOverlay=true で max-xl overlay クラスがパネルに付く", () => {
    render(
      <MasterSidePanel
        isNew
        title=""
        onTitleChange={vi.fn()}
        onClose={vi.fn()}
        icon={<svg />}
        tabletOverlay
      >
        <div>body</div>
      </MasterSidePanel>,
    );

    const panel = document.querySelector("[data-side-peek]");
    expect(panel).toHaveClass(
      "max-xl:absolute",
      "max-xl:inset-y-0",
      "max-xl:right-0",
      "max-xl:z-20",
    );
  });
});
