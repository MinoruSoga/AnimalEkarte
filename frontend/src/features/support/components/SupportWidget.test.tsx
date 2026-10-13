import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SupportWidget } from "./SupportWidget";
import globalsCss from "@/styles/globals.css?raw";

// BugReportTab が動的 import する html2canvas-pro への依存を切る
vi.mock("../lib/capture-screenshot", () => ({
  capturePageScreenshot: vi.fn().mockResolvedValue(null),
}));

// EMR-387 の open-state 属性/Escape は SupportWidget 自身の責務。
// パネル内部（履歴 fetch 等）は別ユニットのためここでは差し替える。
vi.mock("./SupportPanel", () => ({
  SupportPanel: ({ onClose }: { onClose: () => void }) => (
    <div data-testid="support-panel">
      <button type="button" onClick={onClose}>
        閉じる
      </button>
    </div>
  ),
}));

vi.mock("./HelpChatHistoryProvider", () => ({
  HelpChatHistoryProvider: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

describe("SupportWidget", () => {
  it("起動ボタンを表示する", () => {
    render(<SupportWidget />);
    expect(screen.getByRole("button", { name: "サポート・ヘルプを開く" })).toBeInTheDocument();
  });

  it("コンテナは pointer-events-auto を持ち、モーダル Dialog の body ロック中も操作できる", () => {
    render(<SupportWidget />);
    const button = screen.getByRole("button", { name: "サポート・ヘルプを開く" });
    // Radix modal Dialog は body を pointer-events:none にする。ダイアログ外の
    // ウィジェットは見えているのにクリックが背面へ透過するため、明示的に復帰させる。
    const container = button.parentElement;
    expect(container).toHaveClass("pointer-events-auto");
    expect(container).toHaveClass("fixed");
  });

  it("ウィジェット内の pointerdown が document に伝播しない（モーダルの外側クリック判定を防ぐ）", () => {
    const docSpy = vi.fn();
    document.addEventListener("pointerdown", docSpy);
    render(<SupportWidget />);
    fireEvent(
      screen.getByRole("button", { name: "サポート・ヘルプを開く" }),
      new MouseEvent("pointerdown", { bubbles: true }),
    );
    expect(docSpy).not.toHaveBeenCalled();
    document.removeEventListener("pointerdown", docSpy);
  });

  it("ウィジェットへの focusin/focusout が document に伝播しない（FocusScope のフォーカス回収を防ぐ）", () => {
    const focusInSpy = vi.fn();
    const focusOutSpy = vi.fn();
    document.addEventListener("focusin", focusInSpy);
    document.addEventListener("focusout", focusOutSpy);
    render(<SupportWidget />);
    const button = screen.getByRole("button", { name: "サポート・ヘルプを開く" });
    fireEvent.focusIn(button);
    fireEvent(document.body, new FocusEvent("focusout", { bubbles: true, relatedTarget: button }));
    expect(focusInSpy).not.toHaveBeenCalled();
    expect(focusOutSpy).not.toHaveBeenCalled();
    document.removeEventListener("focusin", focusInSpy);
    document.removeEventListener("focusout", focusOutSpy);
  });
});

describe("SupportWidget — EMR-387 モーダル/SidePeek 中の閉じたランチャー非表示", () => {
  it("閉じている間は open-state 属性を付けず、開くと data-support-widget-open を付ける", async () => {
    const user = userEvent.setup();
    const { container } = render(<SupportWidget />);

    const widget = container.querySelector("[data-support-widget]");
    expect(widget).not.toBeNull();
    expect(widget).not.toHaveAttribute("data-support-widget-open");

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    expect(widget).toHaveAttribute("data-support-widget-open", "");
    expect(screen.getByTestId("support-panel")).toBeInTheDocument();
  });

  it("開いたパネルは Escape で閉じられ、open-state 属性も外れる", async () => {
    const user = userEvent.setup();
    const { container } = render(<SupportWidget />);
    const widget = container.querySelector("[data-support-widget]");

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    await user.keyboard("{Escape}");

    expect(widget).not.toHaveAttribute("data-support-widget-open");
    expect(screen.queryByTestId("support-panel")).not.toBeInTheDocument();
  });

  it("globals.css は無関係なモーダル/SidePeek 中に閉じたランチャーを display:none で隠す", () => {
    // 閉じたウィジェットのみを対象にする :not([data-support-widget-open]) と、
    // 起動しているモーダル dialog/alertdialog・SidePeek の検出子を両方含むこと。
    // display:none は pointer/キーボード操作からも除去する（opacity 偽装ではない）。
    expect(globalsCss).toMatch(
      /body:has\(\s*\[data-side-peek\][^)]*\[role="dialog"\]\[data-state="open"\][^)]*\[role="alertdialog"\]\[data-state="open"\][^)]*\)\s*\[data-support-widget\]:not\(\[data-support-widget-open\]\)\s*\{[^}]*display:\s*none/,
    );
    // パネル表示中の既存 SidePeek 退避ルールは open-state 付きウィジェットに限定される
    expect(globalsCss).toMatch(
      /body:has\(\[data-side-peek\]\)\s*\[data-support-widget\]\[data-support-widget-open\]\s*\{[^}]*right:\s*calc\(520px \+ 1rem\)/,
    );
  });
});

// FORMS-A11Y-01: 閉じた起動ボタンは外部 SidePeek/dialog が開いている間 display:none
// になりフォーカスを保持できない。閉じる操作はオーバーレイ内の適切な要素へ
// フォーカスを移す必要がある（jsdom では CSS 非適用のため対象選択ロジックを検証）。
describe("SupportWidget — FORMS-A11Y-01 オーバーレイ中のフォーカス復帰", () => {
  const mountedFixtures: HTMLElement[] = [];

  afterEach(() => {
    for (const el of mountedFixtures.splice(0)) el.remove();
  });

  const mountSidePeek = () => {
    const peek = document.createElement("div");
    peek.setAttribute("data-side-peek", "");
    const titleInput = document.createElement("input");
    titleInput.type = "text";
    titleInput.setAttribute("aria-label", "名称");
    const saveButton = document.createElement("button");
    saveButton.type = "button";
    saveButton.textContent = "保存";
    peek.append(titleInput, saveButton);
    document.body.appendChild(peek);
    mountedFixtures.push(peek);
    return {
      titleInput,
      saveButton,
    };
  };

  // 非表示祖先・type=hidden・inert・disabled の「フォーカス不能」候補が先に
  // 並ぶ SidePeek。末尾の有効な 保存 ボタンが唯一の適格復帰先。
  const mountSidePeekWithUnfocusable = () => {
    const peek = document.createElement("div");
    peek.setAttribute("data-side-peek", "");
    const wrap = document.createElement("div");
    wrap.setAttribute("data-wrap", "");
    const wrappedInput = document.createElement("input");
    wrappedInput.type = "text";
    wrappedInput.setAttribute("aria-label", "ラップ内入力");
    wrap.append(wrappedInput);
    const hiddenInput = document.createElement("input");
    hiddenInput.type = "hidden";
    const inertWrap = document.createElement("div");
    inertWrap.setAttribute("inert", "");
    const inertButton = document.createElement("button");
    inertButton.type = "button";
    inertButton.textContent = "inertボタン";
    inertWrap.append(inertButton);
    const disabledButton = document.createElement("button");
    disabledButton.type = "button";
    disabledButton.disabled = true;
    disabledButton.textContent = "無効ボタン";
    const saveButton = document.createElement("button");
    saveButton.type = "button";
    saveButton.textContent = "保存";
    peek.append(wrap, hiddenInput, inertWrap, disabledButton, saveButton);
    document.body.appendChild(peek);
    mountedFixtures.push(peek);
    return {
      wrap,
      wrappedInput,
      saveButton,
    };
  };

  it("外部オーバーレイが無い通常の close は起動ボタンへフォーカスを戻す（既存動作）", async () => {
    const user = userEvent.setup();
    render(<SupportWidget />);
    const launcher = screen.getByRole("button", { name: "サポート・ヘルプを開く" });

    await user.click(launcher);
    await user.click(screen.getByRole("button", { name: "閉じる" }));

    expect(document.activeElement).toBe(launcher);
  });

  it("SidePeek 表示中にパネルを閉じると直前にフォーカスしていた SidePeek 内要素へ戻す", async () => {
    const user = userEvent.setup();
    const { titleInput } = mountSidePeek();
    render(<SupportWidget />);
    titleInput.focus();

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    await user.click(screen.getByRole("button", { name: "閉じる" }));

    expect(document.activeElement).toBe(titleInput);
  });

  it("直前要素が無効な場合はオーバーレイ内の最初のフォーカス可能要素へ戻す", async () => {
    const user = userEvent.setup();
    const { titleInput } = mountSidePeek();
    render(<SupportWidget />);

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    await user.click(screen.getByRole("button", { name: "閉じる" }));

    expect(document.activeElement).toBe(titleInput);
  });

  it("直前要素が DOM から消えていれば次のフォーカス可能要素へフォールバックする", async () => {
    const user = userEvent.setup();
    const { titleInput, saveButton } = mountSidePeek();
    render(<SupportWidget />);
    titleInput.focus();
    titleInput.remove();

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    await user.click(screen.getByRole("button", { name: "閉じる" }));

    expect(document.activeElement).toBe(saveButton);
  });

  it("Escape で閉じた場合も SidePeek 内へフォーカスを戻す", async () => {
    const user = userEvent.setup();
    const { titleInput } = mountSidePeek();
    render(<SupportWidget />);
    titleInput.focus();

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    await user.keyboard("{Escape}");

    expect(document.activeElement).toBe(titleInput);
  });

  it("直前要素が display:none の祖先配下なら次の有効な要素へフォールバックする", async () => {
    const user = userEvent.setup();
    const { wrap, wrappedInput, saveButton } = mountSidePeekWithUnfocusable();
    render(<SupportWidget />);
    wrappedInput.focus();
    // フォーカス追跡後に祖先ごと非表示になったケース
    wrap.style.display = "none";

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    await user.click(screen.getByRole("button", { name: "閉じる" }));

    expect(document.activeElement).toBe(saveButton);
  });

  it("非表示祖先・hidden input・inert・disabled の候補をスキップして有効な要素へ戻す", async () => {
    const user = userEvent.setup();
    const { wrap, saveButton } = mountSidePeekWithUnfocusable();
    wrap.style.display = "none";
    render(<SupportWidget />);

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    await user.click(screen.getByRole("button", { name: "閉じる" }));

    expect(document.activeElement).toBe(saveButton);
  });

  it("focus() が受理されなかった候補はスキップして次の候補へ進む", async () => {
    const user = userEvent.setup();
    const { titleInput, saveButton } = mountSidePeek();
    render(<SupportWidget />);
    titleInput.focus();
    titleInput.blur();
    // フォーカス拒否（モーダルの奪回等）を模す
    titleInput.focus = () => {};

    await user.click(screen.getByRole("button", { name: "サポート・ヘルプを開く" }));
    await user.click(screen.getByRole("button", { name: "閉じる" }));

    expect(document.activeElement).toBe(saveButton);
  });
});
