import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import type { ManualArticle } from "../lib/manual-index";
import { ManualEditor } from "./ManualEditor";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const mutate = vi.fn();

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock("../api/upsert-manual-article", () => ({
  useUpsertManualArticle: () => ({ mutate, isPending: false }),
}));

vi.mock("./ManualContent", () => ({
  ManualContent: () => <div>preview</div>,
}));

const article: ManualArticle = {
  category: "screens",
  slug: "00-overview",
  title: "概要",
  order: 0,
  section: "基本",
  content: "初期本文",
  searchText: "概要 初期本文",
};

describe("ManualEditor permission re-check", () => {
  beforeEach(() => {
    mutate.mockReset();
    vi.mocked(toast.error).mockClear();
  });

  it("canEdit=true なら dirty 保存で mutate を呼ぶ", async () => {
    const user = userEvent.setup();
    render(<ManualEditor article={article} onClose={vi.fn()} canEdit={true} />);

    await user.clear(screen.getByLabelText("マニュアル本文編集"));
    await user.type(screen.getByLabelText("マニュアル本文編集"), "更新本文");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(mutate).toHaveBeenCalledWith(
        expect.objectContaining({
          category: "screens",
          slug: "00-overview",
          body_markdown: "更新本文",
        }),
        expect.any(Object),
      );
    });
    expect(toast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false では mutate せず toast.error する", async () => {
    const user = userEvent.setup();
    render(<ManualEditor article={article} onClose={vi.fn()} canEdit={false} />);

    await user.clear(screen.getByLabelText("マニュアル本文編集"));
    await user.type(screen.getByLabelText("マニュアル本文編集"), "権限失効後の本文");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(mutate).not.toHaveBeenCalled();
  });

  it("canEdit 省略時は fail-closed で mutate せず toast.error する", async () => {
    const user = userEvent.setup();
    render(<ManualEditor article={article} onClose={vi.fn()} />);

    await user.clear(screen.getByLabelText("マニュアル本文編集"));
    await user.type(screen.getByLabelText("マニュアル本文編集"), "省略時本文");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(mutate).not.toHaveBeenCalled();
  });
});
