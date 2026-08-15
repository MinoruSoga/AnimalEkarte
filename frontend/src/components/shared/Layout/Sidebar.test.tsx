import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AuthContext } from "@/hooks/auth-context";
import { C } from "@/lib/design-tokens";
import type { AuthUser, ResourceAction } from "@/types/auth";
import { Sidebar } from "./Sidebar";

vi.mock("./SidebarItems", () => ({
  SidebarItemWithPermission: () => null,
}));

vi.mock("@/components/shared/ChangePasswordDialog/ChangePasswordDialog", () => ({
  ChangePasswordDialog: () => <div data-testid="change-password-dialog" />,
}));

function makeAuthContext(user: AuthUser | null = null) {
  return {
    user,
    currentClinicId: "clinic-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (_resource: string, _action: ResourceAction): boolean => true,
    refreshPermissions: async () => {},
  };
}

function renderSidebar(width: number, user: AuthUser | null = null) {
  Object.defineProperty(window, "innerWidth", { configurable: true, value: width });

  return render(
    <AuthContext.Provider value={makeAuthContext(user)}>
      <MemoryRouter>
        <Sidebar />
      </MemoryRouter>
    </AuthContext.Provider>,
  );
}

describe("Sidebar touch targets", () => {
  beforeEach(() => {
    Object.defineProperty(window, "matchMedia", {
      configurable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: query === "(max-width: 1279px)" && window.innerWidth < 1280,
        media: query,
        onchange: null,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        addListener: vi.fn(),
        removeListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    });
  });

  it("展開時の折りたたみ・ログアウトbuttonを44x44px以上に保つ", async () => {
    renderSidebar(1440);

    await screen.findByTestId("change-password-dialog");

    expect(screen.getByRole("button", { name: "サイドバーを折りたたむ" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
    expect(screen.getByRole("button", { name: "ログアウト" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });

  it("折りたたみ時の展開・ログアウトbuttonを44x44px以上に保つ", () => {
    renderSidebar(500);

    expect(screen.getByRole("button", { name: "サイドバーを展開" })).toHaveClass(
      "h-11",
      "min-w-11",
    );
    expect(screen.getByRole("button", { name: "ログアウト" })).toHaveClass(
      "h-11",
      "min-w-11",
    );
  });

  it("複数医院のtriggerと候補buttonを44px以上にしfocus ringを表示する", () => {
    const user: AuthUser = {
      id: "user-1",
      email: "staff@example.com",
      displayName: "スタッフ",
      isSystemAdmin: false,
      mainClinicId: "clinic-1",
      clinic: null,
      clinics: [
        { clinicId: "clinic-1", clinicName: "本院", isMain: true },
        { clinicId: "clinic-2", clinicName: "分院", isMain: false },
      ],
      permissions: {},
    };
    renderSidebar(1440, user);

    const trigger = screen.getByRole("button", { name: "本院" });
    expect(trigger).toHaveClass(
      "min-h-11",
      "min-w-11",
      "focus-visible:ring-2",
      C.focusRingAccent40,
    );
    fireEvent.click(trigger);

    for (const name of ["本院", "分院"]) {
      expect(screen.getAllByRole("button", { name }).at(-1)).toHaveClass(
        "min-h-11",
        "min-w-11",
        "focus-visible:ring-2",
        C.focusRingAccent40,
      );
    }
  });

  it("現在医院をpressed状態として公開し、再選択では確認dialogを開かない", () => {
    const user: AuthUser = {
      id: "user-1",
      email: "staff@example.com",
      displayName: "スタッフ",
      isSystemAdmin: false,
      mainClinicId: "clinic-1",
      clinic: null,
      clinics: [
        { clinicId: "clinic-1", clinicName: "本院", isMain: true },
        { clinicId: "clinic-2", clinicName: "分院", isMain: false },
      ],
      permissions: {},
    };
    renderSidebar(1440, user);

    fireEvent.click(screen.getByRole("button", { name: "本院" }));

    const currentClinicButton = screen.getAllByRole("button", { name: "本院" }).at(-1);
    const otherClinicButton = screen.getByRole("button", { name: "分院" });
    expect(currentClinicButton).toHaveAttribute("aria-pressed", "true");
    expect(currentClinicButton).toBeDisabled();
    expect(otherClinicButton).toHaveAttribute("aria-pressed", "false");
    expect(otherClinicButton).not.toBeDisabled();

    fireEvent.click(currentClinicButton!);
    expect(
      screen.queryByRole("alertdialog", { name: "医院を切り替えますか？" }),
    ).not.toBeInTheDocument();

    fireEvent.click(otherClinicButton);
    expect(
      screen.getByRole("alertdialog", { name: "医院を切り替えますか？" }),
    ).toBeInTheDocument();
  });
});
