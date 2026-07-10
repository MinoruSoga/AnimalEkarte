import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { PropertyRow } from "./PropertyRow";

describe("PropertyRow", () => {
  it("input 子要素をラベルと関連付ける（getByLabelText で取得できる）", () => {
    render(
      <PropertyRow label="カテゴリ">
        <input type="text" defaultValue="" />
      </PropertyRow>,
    );
    expect(screen.getByLabelText("カテゴリ")).toBeInTheDocument();
  });

  it("textarea 子要素をラベルと関連付ける", () => {
    render(
      <PropertyRow label="説明">
        <textarea defaultValue="" />
      </PropertyRow>,
    );
    expect(screen.getByLabelText("説明")).toBeInTheDocument();
  });

  it("button 子要素（トグル）もラベルと関連付ける", () => {
    render(
      <PropertyRow label="ステータス">
        <button type="button">切替</button>
      </PropertyRow>,
    );
    expect(screen.getByLabelText("ステータス")).toBeInTheDocument();
  });

  it("ラベルテキストは表示される", () => {
    render(
      <PropertyRow label="備考">
        <input type="text" defaultValue="" />
      </PropertyRow>,
    );
    expect(screen.getByText("備考")).toBeInTheDocument();
  });
});
