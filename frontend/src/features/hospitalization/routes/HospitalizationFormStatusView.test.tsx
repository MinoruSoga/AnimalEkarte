import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { createTestWrapper } from "@/testing/TestUtils";

import { HospitalizationFormStatusView } from "./HospitalizationFormStatusView";

vi.mock("@/hooks/use-auth", () => ({ useAuth: () => ({ hasPermission: () => true }) }));

// EMR-412 R4/N2: 編集系ステータス画面の見出しは「保存済みの区分」からのみ作る。
// ロード中/不在では推測せず汎用「入院・ホテル編集」。新規系は「入院・ホテル登録」。
describe("HospitalizationFormStatusView 見出し (EMR-412 R4)", () => {
  it("保存済み区分が未取得の edit-loading は汎用見出しにする", () => {
    render(<HospitalizationFormStatusView gate={{ kind: "edit-loading" }} onBack={vi.fn()} />, {
      wrapper: createTestWrapper(),
    });

    expect(screen.getByRole("heading", { name: "入院・ホテル編集" })).toBeInTheDocument();
  });

  it("保存済み区分がある edit-loading はその区分名を使う", () => {
    render(
      <HospitalizationFormStatusView
        gate={{ kind: "edit-loading" }}
        onBack={vi.fn()}
        savedHospitalizationType="ホテル"
      />,
      { wrapper: createTestWrapper() },
    );

    expect(screen.getByRole("heading", { name: "ホテル編集" })).toBeInTheDocument();
  });

  it("edit-not-found は区分を推測せず汎用見出しにする", () => {
    render(<HospitalizationFormStatusView gate={{ kind: "edit-not-found" }} onBack={vi.fn()} />, {
      wrapper: createTestWrapper(),
    });

    expect(screen.getByRole("heading", { name: "入院・ホテル編集" })).toBeInTheDocument();
    expect(screen.getByText("入院情報が見つかりません")).toBeInTheDocument();
  });

  it("new-deceased は「入院・ホテル登録」と区別できる文言を出す", () => {
    render(<HospitalizationFormStatusView gate={{ kind: "new-deceased" }} onBack={vi.fn()} />, {
      wrapper: createTestWrapper(),
    });

    expect(screen.getByRole("heading", { name: "入院・ホテル登録" })).toBeInTheDocument();
    expect(screen.getByText("死亡したペットは入院・ホテル登録できません")).toBeInTheDocument();
  });
});
