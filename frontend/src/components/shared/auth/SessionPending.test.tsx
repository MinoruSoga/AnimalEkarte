import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { SessionPending } from "./SessionPending";

describe("SessionPending", () => {
  it("announces the default pending message with role=status", () => {
    render(<SessionPending />);
    expect(screen.getByRole("status")).toHaveTextContent("ログイン状態を確認しています");
    expect(screen.getByText("ノア動物病院")).toBeInTheDocument();
  });

  it("accepts a custom non-sensitive message", () => {
    render(<SessionPending message="画面を読み込んでいます" />);
    expect(screen.getByRole("status")).toHaveTextContent("画面を読み込んでいます");
  });
});
