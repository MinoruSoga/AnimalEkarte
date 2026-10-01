import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";
import { DangerBadge } from "./DangerBadge";

const REASON = "診察時に咬傷歴あり";

// EMR-231: 画面・支援技術経路のどこにも出してはいけない直接文言。
const FORBIDDEN_TEXT = [/危険/, /注意/, /噛/];

describe("DangerBadge variant=pet", () => {
  it("高は赤いアイコンのみのバッジを出し、click で補足メモを開閉できる", async () => {
    const user = userEvent.setup();
    render(<DangerBadge variant="pet" level="高" subjectName="ポチ" reason={REASON} />);

    const trigger = screen.getByRole("button", { name: "ポチの詳細を表示" });
    expect(trigger.tagName).toBe("BUTTON");
    expect(trigger).toHaveAttribute("type", "button");
    // アイコンのみ: trigger に直接文言を含めない
    expect(trigger.textContent).toBe("");
    expect(trigger.querySelector("svg")).not.toBeNull();
    expect(trigger).toHaveClass(C.bgDanger10, C.danger, C.borderDanger20);
    expect(trigger).toHaveAttribute("aria-expanded", "false");

    await user.click(trigger);
    expect(await screen.findByText(REASON)).toBeInTheDocument();
    expect(screen.getByText("特記レベル: 高")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(trigger).toHaveAttribute("aria-controls");

    await user.click(trigger);
    await waitFor(() => {
      expect(screen.queryByText(REASON)).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it("中は黄いアイコンのみのバッジを出し、同じ中立名で Popover を開く", async () => {
    const user = userEvent.setup();
    render(<DangerBadge variant="pet" level="中" subjectName="ミドル" reason={REASON} />);

    const trigger = screen.getByRole("button", { name: "ミドルの詳細を表示" });
    expect(trigger.textContent).toBe("");
    expect(trigger).toHaveClass(C.bgNotice, C.textBadgeYellow, C.borderNotice);

    await user.click(trigger);
    expect(await screen.findByText(REASON)).toBeInTheDocument();
    expect(screen.getByText("特記レベル: 中")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
  });

  it("高と中は色だけでなくアイコン形状でも区別できる", () => {
    render(
      <>
        <DangerBadge variant="pet" level="高" subjectName="ポチ" reason={REASON} />
        <DangerBadge variant="pet" level="中" subjectName="ミケ" reason={REASON} />
      </>,
    );

    const highIcon = screen.getByRole("button", { name: "ポチの詳細を表示" }).querySelector("svg");
    const mediumIcon = screen
      .getByRole("button", { name: "ミケの詳細を表示" })
      .querySelector("svg");
    // lucide はアイコン名を class に含める（lucide-octagon-alert / lucide-triangle-alert）
    expect(highIcon).toHaveClass("lucide-octagon-alert");
    expect(mediumIcon).toHaveClass("lucide-triangle-alert");
  });

  it.each([
    ["wire high", "high"],
    ["wire medium", "medium"],
  ])("wire 値 %s も同じアイコンバッジを描画する", (_caseName, level) => {
    render(<DangerBadge variant="pet" level={level} subjectName="ポチ" reason={REASON} />);

    const trigger = screen.getByRole("button", { name: "ポチの詳細を表示" });
    expect(trigger.textContent).toBe("");
    expect(trigger.querySelector("svg")).not.toBeNull();
  });

  it.each([
    ["低", "低"],
    ["wire low", "low"],
    ["null", null],
    ["空文字", ""],
    ["未設定", undefined],
    ["未知値", "extreme"],
  ])("特記レベルが %s ならバッジを描画しない", (_caseName, level) => {
    const { container } = render(
      <DangerBadge variant="pet" level={level} subjectName="ポチ" reason={REASON} />,
    );

    expect(container).toBeEmptyDOMElement();
  });

  it.each([
    ["undefined", undefined],
    ["空文字", ""],
    ["空白のみ", "   "],
  ])("補足メモが %s なら内容未登録を表示する", async (_caseName, reason) => {
    const user = userEvent.setup();
    render(<DangerBadge variant="pet" level="高" subjectName="ポチ" reason={reason} />);

    await user.click(screen.getByRole("button", { name: "ポチの詳細を表示" }));

    expect(await screen.findByText("内容未登録")).toBeInTheDocument();
  });

  it.each([
    ["Enter", "{Enter}"],
    ["Space", " "],
  ])("%s で Popover を開閉できる", async (_keyName, key) => {
    const user = userEvent.setup();
    render(<DangerBadge variant="pet" level="高" subjectName="ポチ" reason={REASON} />);
    const trigger = screen.getByRole("button", { name: "ポチの詳細を表示" });

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

    await user.click(screen.getByRole("button", { name: "ポチの詳細を表示" }));

    expect(onParentClick).not.toHaveBeenCalled();
    expect(await screen.findByText(REASON)).toBeInTheDocument();
  });
});

describe("DangerBadge variant=owner", () => {
  it("アイコンのみの静的マークを出し、interactive な trigger を持たない", () => {
    render(<DangerBadge variant="owner" />);

    const mark = screen.getByRole("img", { name: "特記" });
    expect(mark.tagName).toBe("SPAN");
    expect(mark.textContent).toBe("");
    expect(mark.querySelector("svg")).not.toBeNull();
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

    await user.click(screen.getByRole("img", { name: "特記" }));

    expect(onParentClick).not.toHaveBeenCalled();
  });
});

describe("DangerBadge プライバシー (EMR-231)", () => {
  it("バッジ・aria-label・Popover 見出しに直接文言を出さない", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <DangerBadge variant="pet" level="高" subjectName="ポチ" reason={REASON} />,
    );

    for (const pattern of FORBIDDEN_TEXT) {
      expect(container).not.toHaveTextContent(pattern);
    }
    const trigger = screen.getByRole("button");
    expect(trigger.getAttribute("aria-label")).toBe("ポチの詳細を表示");
    expect(trigger.getAttribute("title")).toBeNull();

    await user.click(trigger);
    await screen.findByText(REASON);
    // Popover 見出しも中立文言のみ（ユーザー入力のメモ本文はそのまま出す）
    expect(screen.getByText("特記レベル: 高")).toBeInTheDocument();
    expect(screen.queryByText(/危険|注意/)).not.toBeInTheDocument();
  });
});
