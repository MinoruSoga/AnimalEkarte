import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { CheckupSyncFilterForm } from "./CheckupSyncFilterForm";

// EMR-385: 「最終来院日（以降・任意）」の括弧内で「任/意」の1文字折返しが
// 発生していた。ラベルの日本語まとまり（主語部 / 括弧部）をそれぞれ
// 分割不能な単位にして、まとまり同士の境界でのみ折返せるようにする。
// jsdom は実レイアウトを評価しないため、ここでは「分割不能単位として
// レンダリングされる構造」と label-for/input-id 接続の維持を固定する。
const PAREN_LABELS: [labelText: string, inputId: string][] = [
  ["最終来院日（以降・任意）", "last_visit_after"],
  ["最終来院日（以前・任意）", "last_visit_before"],
  ["最終健診実施日（以降・任意）", "last_checkup_after"],
  ["最終健診実施日（以前・任意）", "last_checkup_before"],
];

describe("CheckupSyncFilterForm — EMR-385 ラベル折返し防止", () => {
  it.each(PAREN_LABELS)("%s のラベルはまとまり単位でしか折返せない", (labelText, inputId) => {
    render(<CheckupSyncFilterForm onSearch={vi.fn()} isLoading={false} />);

    // label htmlFor ↔ input id の接続が維持されている
    const input = screen.getByLabelText(labelText);
    expect(input).toHaveAttribute("id", inputId);

    const label = document.querySelector(`label[for="${inputId}"]`);
    expect(label).not.toBeNull();
    const nowrapParts = label!.querySelectorAll("span.whitespace-nowrap");
    expect(nowrapParts.length).toBeGreaterThanOrEqual(2);
    // 不分割単位の結合が元のラベル全文と一致する（括弧内での途中切れなし）
    expect([...nowrapParts].map((node) => node.textContent).join("")).toBe(labelText);
  });
});
