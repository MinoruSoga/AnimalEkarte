import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { MasterSettingsIndex } from "./MasterSettingsIndex";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ hasPermission: () => true }),
}));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true }),
}));

describe("MasterSettingsIndex", () => {
  it("ページタイトルとセクション内容を表示し、装飾の Settings アイコンは出さない", () => {
    const { container } = render(
      <MemoryRouter>
        <MasterSettingsIndex />
      </MemoryRouter>,
    );

    expect(screen.getAllByRole("heading", { name: "マスタ設定" }).length).toBeGreaterThan(0);
    expect(screen.getByText("動物病院の各種マスタデータを管理します")).toBeInTheDocument();
    expect(container.querySelector(".lucide-settings")).not.toBeInTheDocument();
  });
});
