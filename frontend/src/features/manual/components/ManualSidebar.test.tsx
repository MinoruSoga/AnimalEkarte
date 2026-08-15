import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { ManualSidebar } from "./ManualSidebar";

describe("ManualSidebar touch targets", () => {
  it("tabs・検索・目次linksを44px以上に保つ", () => {
    render(
      <MemoryRouter initialEntries={["/manual/screens/00-overview"]}>
        <ManualSidebar
          viewMode="screens"
          onChangeViewMode={vi.fn()}
          query=""
          onChangeQuery={vi.fn()}
          filteredArticles={[]}
          isSearching={false}
        />
      </MemoryRouter>,
    );

    expect(screen.getByRole("tab", { name: "画面別" })).toHaveClass("min-h-11");
    expect(screen.getByRole("tab", { name: "業務フロー" })).toHaveClass("min-h-11");
    expect(screen.getByRole("searchbox", { name: "マニュアル内検索" })).toHaveClass("min-h-11");
    screen.getAllByRole("link").forEach((link) => expect(link).toHaveClass("min-h-11"));
  });
});
