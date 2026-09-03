import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { UnlinkedLineIdForm } from "./LineIntegrationCardParts";

const LINE_ID_STATE = { error: null, success: false };

describe("UnlinkedLineIdForm", () => {
  it("OwnerForm配下でもformを入れ子にせずconsole errorを発生させない", () => {
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => undefined);

    try {
      const { container } = render(
        <form>
          <UnlinkedLineIdForm canEdit lineIdFormAction={vi.fn()} lineIdState={LINE_ID_STATE} />
        </form>,
      );

      expect(container.querySelectorAll("form")).toHaveLength(1);
      expect(consoleError.mock.calls.flat().join(" ")).not.toContain(
        "cannot be a descendant of <form>",
      );
    } finally {
      consoleError.mockRestore();
    }
  });

  it("設定ボタンからLINE User IDをFormDataでactionへ渡す", async () => {
    const user = userEvent.setup();
    const lineIdFormAction = vi.fn();
    render(
      <form>
        <UnlinkedLineIdForm
          canEdit
          lineIdFormAction={lineIdFormAction}
          lineIdState={LINE_ID_STATE}
        />
      </form>,
    );

    await user.type(screen.getByLabelText("LINE User ID"), "U1234567890");
    await user.click(screen.getByRole("button", { name: "設定" }));

    await waitFor(() => expect(lineIdFormAction).toHaveBeenCalledTimes(1));
    const payload = lineIdFormAction.mock.calls[0]?.[0];
    expect(payload).toBeInstanceOf(FormData);
    expect(payload?.get("line_user_id")).toBe("U1234567890");
  });
});
