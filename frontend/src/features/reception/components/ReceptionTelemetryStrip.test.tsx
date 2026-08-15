import { readFileSync } from "node:fs";
import { join } from "node:path";
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { C } from "@/lib/design-tokens";
import { ReceptionTelemetryStrip } from "./ReceptionTelemetryStrip";

describe("ReceptionTelemetryStrip", () => {
  it("Phase 1: waitStats 未指定では本日受付件数のみ表示し、平均待ち/最長待ちは表示しない", () => {
    render(<ReceptionTelemetryStrip totalCount={32} />);

    expect(screen.getByText("本日受付")).toBeInTheDocument();
    expect(screen.getByText("32")).toBeInTheDocument();
    expect(screen.queryByText("平均待ち")).not.toBeInTheDocument();
    expect(screen.queryByText("最長待ち")).not.toBeInTheDocument();
  });

  it("Phase 2: 受付済 0 件では平均待ち/最長待ちともに「—」を表示する", () => {
    render(
      <ReceptionTelemetryStrip
        totalCount={0}
        waitStats={{ averageMinutes: null, longest: null }}
      />,
    );

    expect(screen.getByText("平均待ち")).toBeInTheDocument();
    expect(screen.getByText("最長待ち")).toBeInTheDocument();
    const dashes = screen.getAllByText("—");
    expect(dashes).toHaveLength(2);
  });

  it("Phase 2: 受付済ありでは平均待ち(分)と最長待ち(患者名つき)を表示する", () => {
    render(
      <ReceptionTelemetryStrip
        totalCount={32}
        waitStats={{ averageMinutes: 14, longest: { minutes: 32, petName: "ミルク" } }}
      />,
    );

    expect(screen.getByText("14分")).toBeInTheDocument();
    expect(screen.getByText("32分 — ミルク")).toBeInTheDocument();
  });

  it("最長待ちの強調はテキスト色のみで、primary/brand identity の構造クラスを使わない", () => {
    render(
      <ReceptionTelemetryStrip
        totalCount={32}
        waitStats={{ averageMinutes: 14, longest: { minutes: 32, petName: "ミルク" } }}
      />,
    );

    const longest = screen.getByText("32分 — ミルク");
    expect(longest).not.toHaveClass(C.textActionPrimary);
    expect(longest).not.toHaveClass(C.bgActionPrimary);
    expect(longest).not.toHaveClass(C.textBrandIdentity);
    expect(longest).not.toHaveClass(C.bgBrandIdentity);
  });

  it("コンポーネントソースに raw hex カラーリテラルを直書きしない(design-tokens 経由のみ)", () => {
    const sourcePath = join(process.cwd(), "src/features/reception/components/ReceptionTelemetryStrip.tsx");
    const source = readFileSync(sourcePath, "utf-8");

    // C.xxx / STYLE.xxx トークン定義側の #hex は design-tokens.ts に閉じ込め、
    // コンポーネント側では `#` リテラルを一切書かない
    expect(source).not.toMatch(/#[0-9a-fA-F]{3,6}/);
  });
});
