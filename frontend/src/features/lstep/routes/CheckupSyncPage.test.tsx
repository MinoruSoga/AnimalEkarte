import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";
import { C } from "@/lib/design-tokens";
import type { CheckupSyncPreviewResponse } from "../api/get-checkup-sync-preview";
import { CheckupSyncPage } from "./CheckupSyncPage";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const mocks = vi.hoisted(() => ({
  permission: { canCreate: true },
  preview: {
    data: undefined as CheckupSyncPreviewResponse | undefined,
    isFetching: false,
  },
  createMutate: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mocks.permission,
}));

vi.mock("../api/get-checkup-sync-preview", () => ({
  useGetCheckupSyncPreview: () => mocks.preview,
}));

vi.mock("../api/create-checkup-sync", () => ({
  useCreateCheckupSync: () => ({ mutate: mocks.createMutate, isPending: false }),
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const previewData: CheckupSyncPreviewResponse = {
  owners: [
    {
      owner_id: "owner-1",
      owner_name: "山田 太郎",
      pet_names: ["ポチ"],
      last_visit_date: "2026-07-01",
      has_line: true,
      is_opt_out: false,
      has_living_pet: true,
      exclusion_reason: null,
      current_tags: [],
      min_pet_age_years: 3,
      max_pet_age_years: 3,
      has_chronic_condition: false,
      cpm_stage: "",
      total_amount: 0,
      annual_visit_count: 1,
      last_checkup_date: null,
    },
  ],
  eligible_count: 1,
  line_linked_count: 1,
  opt_out_count: 0,
  no_living_pet_count: 0,
  total_count: 1,
};

function hasClassInAncestry(element: Element | null, className: string): boolean {
  let current = element;
  while (current !== null) {
    if (current.classList.contains(className)) return true;
    current = current.parentElement;
  }
  return false;
}

async function searchAndOpenConfirm(user: ReturnType<typeof userEvent.setup>) {
  await user.selectOptions(screen.getByLabelText(/検診種別/), "annual");
  await user.click(screen.getByRole("button", { name: "対象者を検索" }));
  await user.click(await screen.findByRole("checkbox", { name: "山田 太郎を選択" }));
  await user.click(screen.getByRole("button", { name: "タグを一括付与する" }));
  expect(await screen.findByRole("dialog")).toBeInTheDocument();
}

describe("CheckupSyncPage", () => {
  beforeEach(() => {
    mocks.permission.canCreate = true;
    mocks.preview.data = undefined;
    mocks.preview.isFetching = false;
    mocks.createMutate.mockReset();
    vi.mocked(toast.error).mockClear();
  });

  it("DESIGN.md の canvas-soft shell 上に表示する", () => {
    render(<CheckupSyncPage />);

    const heading = screen.getByRole("heading", { name: "健診リマインダー抽出" });
    expect(hasClassInAncestry(heading, C.bgPage)).toBe(true);
  });

  it("canCreate=false では createCheckupSync せず toast.error する", async () => {
    mocks.preview.data = previewData;
    const user = userEvent.setup();
    const { rerender } = render(<CheckupSyncPage />);

    await searchAndOpenConfirm(user);

    mocks.permission.canCreate = false;
    rerender(<CheckupSyncPage />);

    await user.click(screen.getByRole("button", { name: "タグを一括付与する" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(mocks.createMutate).not.toHaveBeenCalled();
  });
});
