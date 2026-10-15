import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { DEFAULT_OWNER_DATA } from "../hooks/use-owner-form-model";
import { OwnerBasicFields } from "./OwnerBasicFields";

describe("OwnerBasicFields required accessibility", () => {
  it("エラー表示前から飼主名・よみ・電話番号を必須と通知し任意欄は必須にしない", () => {
    render(
      <OwnerBasicFields
        ownerData={{ ...DEFAULT_OWNER_DATA }}
        fieldErrors={{}}
        isEdit={false}
        onChange={vi.fn()}
        onClearError={vi.fn()}
      />,
    );

    for (const name of ["飼主名", "飼主名よみ", "電話番号"]) {
      const input = screen.getByRole("textbox", { name });
      expect(input).toHaveAttribute("aria-required", "true");
      expect(input).not.toHaveAttribute("required");
      expect(input).not.toHaveAttribute("aria-invalid", "true");
    }
    for (const name of ["会社名", "メールアドレス", "会社 電話番号"]) {
      expect(screen.getByRole("textbox", { name })).not.toHaveAttribute("aria-required");
    }
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
