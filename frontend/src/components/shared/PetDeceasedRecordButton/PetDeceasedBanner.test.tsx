import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { PetDeceasedBanner } from "./PetDeceasedBanner";

vi.mock("@/hooks/use-revoke-pet-death", () => ({
  useRevokePetDeath: () => ({ mutate: vi.fn(), isPending: false }),
}));

// FE4-9 fix 回帰テスト: `new Date(deceasedAt)` + ブラウザローカル TZ getter だった旧実装は
// 非 JST ブラウザで前日を表示するバグがあった。toJSTWallDate 経由に修正後、
// JST 相当の入力（このコンテナの既定 TZ は Asia/Tokyo）で現行出力と同一であることを固定する。
// TZ=America/New_York 等への vitest プロセス TZ 切替による分岐検証は計画上スコープ外
// （FE-refactor.md FE4-9 参照）— toJSTWallDate 自体の TZ 非依存動作は jst-date.test.ts が担保する。
describe("PetDeceasedBanner (FE4-9)", () => {
  it("JST 深夜 instant の deceasedAt を正しい日付（曜日表示なし）で表示する", () => {
    render(
      <PetDeceasedBanner deceasedAt="2026-07-11T00:00:00+09:00" petId="1" canEdit={false} />,
    );
    expect(screen.getByText("2026年7月11日 永眠")).toBeInTheDocument();
  });

  it("時刻付きの deceasedAt でも JST の日付で表示する", () => {
    render(
      <PetDeceasedBanner deceasedAt="2026-07-11T23:30:00+09:00" petId="1" canEdit={false} />,
    );
    expect(screen.getByText("2026年7月11日 永眠")).toBeInTheDocument();
  });
});
