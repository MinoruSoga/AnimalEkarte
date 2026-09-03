import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { toast } from "sonner";
import { createTestWrapper } from "@/testing/TestUtils";
import { LinkedLineCustomers } from "./LinkedLineCustomers";
import type { LineCustomer } from "../api/types";

const permission = vi.hoisted(() => ({
  current: { canView: true, canCreate: true, canEdit: true, canDelete: false },
}));

const mutate = vi.hoisted(() => vi.fn());

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => permission.current,
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const linkedCustomer: LineCustomer = {
  id: 1,
  clinic_id: 1,
  line_user_id: "U1",
  display_name: "LINE太郎",
  real_name: "山田太郎",
  additional_fields: "",
  owner_id: 10,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const unlinkedCustomer: LineCustomer = {
  id: 2,
  clinic_id: 1,
  line_user_id: "U2",
  display_name: "未紐付け花子",
  real_name: "佐藤花子",
  additional_fields: "",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

vi.mock("../api/get-line-customers", () => ({
  useGetLineCustomers: () => ({ data: [linkedCustomer, unlinkedCustomer] }),
}));

vi.mock("../api/update-owner-link", () => ({
  useUpdateOwnerLink: () => ({ mutate, isPending: false }),
}));

function renderSection() {
  return render(<LinkedLineCustomers clinicId="clinic-1" ownerId={10} />, {
    wrapper: createTestWrapper(),
  });
}

describe("LinkedLineCustomers — FE-RC-212 mutation permission re-check", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: false };
    mutate.mockReset();
    vi.mocked(toast.error).mockClear();
  });

  it("canEdit=false では解除 mutate を呼ばず toast.error する", async () => {
    permission.current = { canView: true, canCreate: true, canEdit: false, canDelete: false };
    renderSection();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /解除/ }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(mutate).not.toHaveBeenCalled();
  });

  it("canEdit=false では紐付け mutate を呼ばず toast.error する", async () => {
    permission.current = { canView: true, canCreate: true, canEdit: false, canDelete: false };
    renderSection();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "LINEアカウントを紐付け" }));
    await user.click(screen.getByRole("button", { name: /未紐付け花子/ }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(mutate).not.toHaveBeenCalled();
  });

  it("canEdit=true では解除 mutate を呼ぶ", async () => {
    renderSection();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /解除/ }));

    expect(mutate).toHaveBeenCalledWith({ customerId: 1, ownerID: null });
    expect(toast.error).not.toHaveBeenCalled();
  });
});
