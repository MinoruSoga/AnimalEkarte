import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { ResourceIdentityLinks } from "@/types/generated/models";

const hasPermission = vi.fn();

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    hasPermission,
  }),
}));

vi.mock("../api/identity-links-api", () => ({
  searchOwnersForLink: vi.fn().mockResolvedValue([]),
  searchPetsForLink: vi.fn().mockResolvedValue([]),
  createOwnerIdentityGroup: vi.fn(),
  unlinkOwnerIdentityMember: vi.fn(),
  createPetIdentityGroup: vi.fn(),
  unlinkPetIdentityMember: vi.fn(),
  getLinkedTreatmentHistory: vi.fn(),
}));

import { IdentityLinksPage } from "./IdentityLinksPage";

describe("IdentityLinksPage permission gates", () => {
  beforeEach(() => {
    hasPermission.mockReset();
  });

  it("view が無い場合はホームへリダイレクトする", () => {
    hasPermission.mockReturnValue(false);
    render(
      <MemoryRouter>
        <IdentityLinksPage />
      </MemoryRouter>,
    );
    // Navigate replaces content; page title must not appear
    expect(screen.queryByText("同一飼主・ペット連携")).not.toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(ResourceIdentityLinks, "view");
  });

  it("view のみの場合は閲覧UIを出し link ボタンを出さない", () => {
    hasPermission.mockImplementation((resource: string, action: string) => {
      return resource === ResourceIdentityLinks && action === "view";
    });
    render(
      <MemoryRouter>
        <IdentityLinksPage />
      </MemoryRouter>,
    );
    expect(screen.getByText("同一飼主・ペット連携")).toBeInTheDocument();
    expect(screen.getByText(/閲覧のみ/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "飼主をリンク" })).not.toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(ResourceIdentityLinks, "edit");
  });

  it("edit がある場合は link ボタンを表示する", () => {
    hasPermission.mockImplementation((resource: string, action: string) => {
      return resource === ResourceIdentityLinks && (action === "view" || action === "edit");
    });
    render(
      <MemoryRouter>
        <IdentityLinksPage />
      </MemoryRouter>,
    );
    expect(screen.getByRole("button", { name: "飼主をリンク" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ペットをリンク" })).toBeInTheDocument();
  });
});
