import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { ClinicMasterList } from "./ClinicMasterList";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ hasPermission: () => true }),
}));

describe("ClinicMasterList touch targets", () => {
  it("新しい医院を追加buttonを44px以上に保つ", () => {
    render(
      <MemoryRouter>
        <ClinicMasterList
          canCreate
          canEdit
          items={[]}
          searchTerm=""
          onSearchChange={vi.fn()}
          onBack={vi.fn()}
          onCreate={vi.fn()}
          onEdit={vi.fn()}
        />
      </MemoryRouter>,
    );

    expect(screen.getByRole("button", { name: "新しい医院を追加..." })).toHaveClass("min-h-11");
  });
});
