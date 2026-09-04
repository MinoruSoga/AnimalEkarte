import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { BulkTagRemoveDialog } from "./BulkTagRemoveDialog";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const bulkRemove = vi.fn();

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock("../api/delete-owner-tag-bulk", () => ({
  useBulkRemoveTag: () => ({
    bulkRemove,
    progress: { done: 0, total: 0, isRunning: false },
  }),
}));

const defaultProps = {
  open: true,
  onOpenChange: vi.fn(),
  tagName: "campaign_summer",
  ownerCount: 2,
  ownerIds: ["owner-1", "owner-2"],
};

describe("BulkTagRemoveDialog permission re-check", () => {
  beforeEach(() => {
    bulkRemove.mockReset().mockResolvedValue(undefined);
    defaultProps.onOpenChange.mockReset();
    vi.mocked(toast.error).mockClear();
  });

  it("canDelete=true なら bulkRemove を呼ぶ", async () => {
    const user = userEvent.setup();
    render(<BulkTagRemoveDialog {...defaultProps} canDelete={true} />);

    await user.click(screen.getByRole("button", { name: "一括解除" }));

    await waitFor(() => {
      expect(bulkRemove).toHaveBeenCalledWith("campaign_summer", ["owner-1", "owner-2"]);
    });
    expect(toast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canDelete=false では bulkRemove せず toast.error する", async () => {
    const user = userEvent.setup();
    render(<BulkTagRemoveDialog {...defaultProps} canDelete={false} />);

    await user.click(screen.getByRole("button", { name: "一括解除" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(bulkRemove).not.toHaveBeenCalled();
    expect(defaultProps.onOpenChange).not.toHaveBeenCalled();
  });

  it("canDelete 省略時は fail-closed で bulkRemove せず toast.error する", async () => {
    const user = userEvent.setup();
    render(<BulkTagRemoveDialog {...defaultProps} />);

    await user.click(screen.getByRole("button", { name: "一括解除" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(bulkRemove).not.toHaveBeenCalled();
  });
});
