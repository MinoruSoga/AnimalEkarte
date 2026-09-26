import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";
import { DangerBadge } from "./DangerBadge";

const REASON = "診察時に咬傷歴あり";

describe("DangerBadge variant=pet", () => {
  it("高は赤い ⚠ 危険 バッジを出し、click で危険理由を開閉できる", async () => {
    const user = userEvent.setup();
    render(<DangerBadge variant="pet" level="高" subjectName="ポチ" reason={REASON} />);

    const trigger = screen.getByRole("button", { name: "ポチの危険理由を表示" });
    expect(trigger.tagName).toBe("BUTTON");
    expect(trigger).toHaveAttribute("type", "button");
    expect(trigger).toHaveTextContent("⚠ 危険");
    expect(trigger).toHaveClass(C.bgDanger10, C.danger, C.borderDanger20);
    expect(trigger).toHaveAttribute("aria-expanded", "false");

    await user.click(trigger);
    expect(await screen.findByText(REASON)).toBeInTheDocument();
    expect(screen.getByText("危険理由")).toHaveClass(C.danger);
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(trigger).toHaveAttribute("aria-controls");

    await user.click(trigger);
    await waitFor(() => {
      expect(screen.queryByText(REASON)).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it("中は黄の ⚠ 注意 バッジを出し、注意理由として Popover を開く", async () => {
    const user = userEvent.setup();
    render(<DangerBadge variant="pet" level="中" subjectName="ミドル" reason={REASON} />);

    // 段階が文言で区別できること: 高の「危険理由」名前は付かず「注意理由」になる。
    const trigger = screen.getByRole("button", { name: "ミドルの注意理由を表示" });
    expect(trigger).toHaveTextContent("⚠ 注意");
    expect(trigger).toHaveClass(C.bgNotice, C.textNotice, C.borderNotice);
    expect(
      screen.queryByRole("button", { name: "ミドルの危険理由を表示" }),
    ).not.toBeInTheDocument();

    await user.click(trigger);
    expect(await screen.findByText(REASON)).toBeInTheDocument();
    expect(screen.getByText("注意理由")).toHaveClass(C.textNotice);
    expect(trigger).toHaveAttribute("aria-expanded", "true");
  });

  it.each([
    ["wire high", "high", "⚠ 危険", "ポチの危険理由を表示"],
    ["wire medium", "medium", "⚠ 注意", "ポチの注意理由を表示"],
  ])("%s 値 %s も同じバッジを描画する", (_caseName, level, label, ariaName) => {
    render(<DangerBadge variant="pet" level={level} subjectName="ポチ" reason={REASON} />);

    const trigger = screen.getByRole("button", { name: ariaName });
    expect(trigger).toHaveTextContent(label);
  });

  it.each([
    ["低", "低"],
    ["wire low", "low"],
    ["null", null],
    ["空文字", ""],
    ["未設定", undefined],
    ["未知値", "extreme"],
  ])("危険度が %s ならバッジを描画しない", (_caseName, level) => {
    const { container } = render(
      <DangerBadge variant="pet" level={level} subjectName="ポチ" reason={REASON} />,
    );

    expect(container).toBeEmptyDOMElement();
    expect(screen.queryByText(/⚠/)).not.toBeInTheDocument();
  });

  it.each([
    ["undefined", undefined],
    ["空文字", ""],
    ["空白のみ", "   "],
  ])("危険理由が %s なら理由未登録を表示する", async (_caseName, reason) => {
    const user = userEvent.setup();
    render(<DangerBadge variant="pet" level="高" subjectName="ポチ" reason={reason} />);

    await user.click(screen.getByRole("button", { name: "ポチの危険理由を表示" }));

    expect(await screen.findByText("理由未登録")).toBeInTheDocument();
  });

  it.each([
    ["Enter", "{Enter}"],
    ["Space", " "],
  ])("%s で Popover を開閉できる", async (_keyName, key) => {
    const user = userEvent.setup();
    render(<DangerBadge variant="pet" level="高" subjectName="ポチ" reason={REASON} />);
    const trigger = screen.getByRole("button", { name: "ポチの危険理由を表示" });

    trigger.focus();
    await user.keyboard(key);
    expect(await screen.findByText(REASON)).toBeInTheDocument();

    trigger.focus();
    await user.keyboard(key);
    await waitFor(() => {
      expect(screen.queryByText(REASON)).not.toBeInTheDocument();
    });
  });

  it("stopPropagation=true のとき click が親へ伝播しない", async () => {
    const user = userEvent.setup();
    const onParentClick = vi.fn();
    render(
      <div onClick={onParentClick} onPointerDown={onParentClick}>
        <DangerBadge variant="pet" level="高" subjectName="ポチ" reason={REASON} stopPropagation />
      </div>,
    );

    await user.click(screen.getByRole("button", { name: "ポチの危険理由を表示" }));

    expect(onParentClick).not.toHaveBeenCalled();
    expect(await screen.findByText(REASON)).toBeInTheDocument();
  });
});

describe("DangerBadge variant=owner", () => {
  it("⚠ 危険人物 をテキスト表示し、interactive な trigger を持たない", () => {
    render(<DangerBadge variant="owner" />);

    const mark = screen.getByText("⚠ 危険人物");
    expect(mark.tagName).toBe("SPAN");
    expect(mark).toHaveClass(C.bgDanger10, C.danger, C.borderDanger20);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("stopPropagation=true のとき pointerdown/click が親へ伝播しない", async () => {
    const user = userEvent.setup();
    const onParentClick = vi.fn();
    render(
      <div onClick={onParentClick} onPointerDown={onParentClick}>
        <DangerBadge variant="owner" stopPropagation />
      </div>,
    );

    await user.click(screen.getByText("⚠ 危険人物"));

    expect(onParentClick).not.toHaveBeenCalled();
  });
});
