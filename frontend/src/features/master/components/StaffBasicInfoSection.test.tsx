import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ComponentProps } from "react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { paths } from "@/config/paths";

import type { Occupation } from "../api/occupations";
import type { Staff } from "../api/staffs";
import type { StaffFormData } from "../lib/staff-side-panel-model";
import { StaffBasicInfoSection } from "./StaffBasicInfoSection";

const { mutateAsync } = vi.hoisted(() => ({
  mutateAsync: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { isSystemAdmin: true } }),
}));

vi.mock("../api/staffs", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/staffs")>();
  return {
    ...actual,
    useAttachStaffAccount: () => ({ mutateAsync, isPending: false }),
  };
});

function staffWithoutEmail(overrides: Partial<Staff> = {}): Staff {
  return {
    id: "10",
    clinicId: "1",
    name: "既存 太郎",
    isActive: true,
    occupationId: null,
    occupationName: null,
    licenseNumber: "",
    sortOrder: 1,
    email: "",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    staffType: "doctor",
    reservationDisplayName: "",
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    ...overrides,
  };
}

function formData(overrides: Partial<StaffFormData> = {}): StaffFormData {
  return {
    name: "既存 太郎",
    jobTitleId: null,
    licenseNumber: "",
    isActive: true,
    email: "",
    password: "",
    staffType: "doctor",
    reservationDisplayName: "",
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    ...overrides,
  };
}

function occupation(overrides: Partial<Occupation> = {}): Occupation {
  return {
    id: "occupation-1",
    name: "獣医師",
    description: "",
    isActive: true,
    sortOrder: 1,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function renderSection(props: Partial<ComponentProps<typeof StaffBasicInfoSection>> = {}) {
  return render(
    <MemoryRouter>
      <StaffBasicInfoSection
        item={staffWithoutEmail()}
        isNew={false}
        formData={formData()}
        setFormDataDirty={vi.fn()}
        allOccupations={[]}
        {...props}
      />
    </MemoryRouter>,
  );
}

describe("StaffBasicInfoSection attach account", () => {
  it("lets a system admin attach a login account without showing a password", async () => {
    mutateAsync.mockResolvedValue({
      staffId: "10",
      accountId: "88",
      email: "staff@example.test",
      message: "アカウントを追加しました。本人がログイン画面のパスワード再設定から設定してください",
    });
    const user = userEvent.setup({ delay: null });
    renderSection();

    expect(screen.queryByLabelText("パスワード")).not.toBeInTheDocument();
    await user.type(screen.getByLabelText("追加するメールアドレス"), "staff@example.test");
    await user.click(screen.getByRole("button", { name: "ログインアカウントを追加" }));
    expect(mutateAsync).toHaveBeenCalledWith({ id: "10", email: "staff@example.test" });
  });
});

describe("StaffBasicInfoSection occupation empty master", () => {
  it("shows unregistered guidance and a path to the occupation master when occupations are empty", () => {
    renderSection({ allOccupations: [] });

    expect(screen.getByText("職種が未登録です")).toBeInTheDocument();
    const link = screen.getByRole("link", { name: "職種マスタを開く" });
    expect(link).toHaveAttribute("href", paths.settings.occupations.getHref());
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
    expect(screen.queryByText("獣医師")).not.toBeInTheDocument();
    expect(screen.queryByText("動物看護師")).not.toBeInTheDocument();
  });

  it("lists active occupations only and keeps occupation optional when masters exist", async () => {
    const user = userEvent.setup({ delay: null });
    const setFormDataDirty = vi.fn();
    renderSection({
      formData: formData({ jobTitleId: null }),
      setFormDataDirty,
      allOccupations: [
        occupation({ id: "occupation-1", name: "獣医師", isActive: true }),
        occupation({ id: "occupation-2", name: "停止職種", isActive: false }),
      ],
    });

    expect(screen.queryByText("職種が未登録です")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "職種マスタを開く" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("combobox"));
    expect(screen.getByRole("option", { name: "獣医師" })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "停止職種" })).not.toBeInTheDocument();
    expect(setFormDataDirty).not.toHaveBeenCalled();
  });

  it("retains an inactive selected occupation for display without inventing options", async () => {
    const user = userEvent.setup({ delay: null });
    renderSection({
      formData: formData({ jobTitleId: "occupation-2" }),
      allOccupations: [
        occupation({ id: "occupation-1", name: "獣医師", isActive: true }),
        occupation({ id: "occupation-2", name: "停止職種", isActive: false }),
      ],
    });

    expect(screen.getByRole("combobox")).toHaveTextContent("停止職種");
    await user.click(screen.getByRole("combobox"));
    expect(screen.getByRole("option", { name: "獣医師" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "停止職種" })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "動物看護師" })).not.toBeInTheDocument();
  });
});
