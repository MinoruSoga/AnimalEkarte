import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { BirthdayBadge } from "./BirthdayBadge";

// テスト中の「当月」は JST 2026-10 に固定する
const NOW_UTC = new Date("2026-10-15T01:00:00Z"); // JST 2026-10-15 10:00

describe("BirthdayBadge (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(NOW_UTC);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("誕生月ペットに Cake アイコンのみのボタンを出し、click で生年月日を開閉できる", () => {
    // fake timers 下は userEvent が凍結するため fireEvent で同期駆動する
    render(<BirthdayBadge birthDate="2020-10-05" subjectName="ポチ" />);

    const trigger = screen.getByRole("button", { name: "今月誕生日" });
    expect(trigger.tagName).toBe("BUTTON");
    expect(trigger).toHaveAttribute("type", "button");
    // アイコンのみ: trigger に直接文言を含めない
    expect(trigger.textContent).toBe("");
    expect(trigger.querySelector("svg")).not.toBeNull();
    expect(trigger).toHaveAttribute("aria-expanded", "false");

    fireEvent.click(trigger);
    expect(screen.getByText("生年月日: 2020-10-05")).toBeInTheDocument();
    expect(screen.getByText("ポチは今月誕生日")).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");

    fireEvent.click(trigger);
    expect(trigger).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByText("生年月日: 2020-10-05")).not.toBeInTheDocument();
  });

  it("誕生月でないペットには何も描画しない", () => {
    render(<BirthdayBadge birthDate="2020-09-05" subjectName="ミケ" />);
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it.each([
    ["未設定(undefined)", undefined],
    ["null", null],
    ["空文字", ""],
    ["形式不正", "2020/10/05"],
    ["暦上無効", "2011-02-29"],
  ])("birthDate が %s なら描画しない", (_label, birthDate) => {
    render(<BirthdayBadge birthDate={birthDate} subjectName="タマ" />);
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("閏日生まれは2月に描画される", () => {
    vi.setSystemTime(new Date("2026-02-10T00:00:00Z")); // JST 2026-02-10
    render(<BirthdayBadge birthDate="2012-02-29" subjectName="レオ" />);
    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it("stopPropagation=true で click が親へ伝播しない", () => {
    const onParentClick = vi.fn();
    render(
      <div onClick={onParentClick}>
        <BirthdayBadge birthDate="2020-10-05" subjectName="ポチ" stopPropagation />
      </div>,
    );

    fireEvent.click(screen.getByRole("button", { name: "今月誕生日" }));
    expect(onParentClick).not.toHaveBeenCalled();
    // Popover は開く（伝播は止めるが動作自体は有効）
    expect(screen.getByText("生年月日: 2020-10-05")).toBeInTheDocument();
  });

  it("stopPropagation 未指定なら click が親へ伝播する", () => {
    const onParentClick = vi.fn();
    render(
      <div onClick={onParentClick}>
        <BirthdayBadge birthDate="2020-10-05" subjectName="ポチ" />
      </div>,
    );

    fireEvent.click(screen.getByRole("button", { name: "今月誕生日" }));
    expect(onParentClick).toHaveBeenCalledTimes(1);
  });

  it("variant=static は静的アイコン（role=img + title=生年月日）を描画し、button を含まない", () => {
    render(<BirthdayBadge birthDate="2020-10-05" subjectName="モカ" variant="static" />);

    // 祖先 button 内で使う変種 — interactive 要素を含まない
    const mark = screen.getByRole("img", { name: "今月誕生日" });
    expect(mark.tagName).toBe("SPAN");
    expect(mark.querySelector("svg")).not.toBeNull();
    expect(mark).toHaveAttribute("title", "モカの生年月日: 2020-10-05");
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("variant=static でも誕生月でなければ描画しない", () => {
    render(<BirthdayBadge birthDate="2020-09-05" subjectName="モカ" variant="static" />);
    expect(screen.queryByRole("img", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});
