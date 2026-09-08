import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { StaffBasicInfoSection } from "./StaffBasicInfoSection";
import type { Staff } from "../api/staffs";
import type { StaffFormData } from "../lib/staff-side-panel-model";

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

function formData(): StaffFormData {
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
  };
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
    render(
      <StaffBasicInfoSection
        item={staffWithoutEmail()}
        isNew={false}
        formData={formData()}
        setFormDataDirty={vi.fn()}
        allOccupations={[]}
      />,
    );

    expect(screen.queryByLabelText("パスワード")).not.toBeInTheDocument();
    await user.type(screen.getByLabelText("追加するメールアドレス"), "staff@example.test");
    await user.click(screen.getByRole("button", { name: "ログインアカウントを追加" }));
    expect(mutateAsync).toHaveBeenCalledWith({ id: "10", email: "staff@example.test" });
  });
});
