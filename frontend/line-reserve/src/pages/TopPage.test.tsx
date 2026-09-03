import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { TopPage } from "./TopPage";
import type { LiffSettings } from "../types/models";

const settings: LiffSettings = {
  liff_id: "liff-123",
  header_text: "ノア動物病院 八王子",
  phone_number: "042-000-0000",
  status: "running",
  request_example: "",
  reservation_notice: "",
  cancel_notice: "",
  privacy_policy: "",
  show_no_staff_option: false,
  booking_window: 30,
};

describe("TopPage brand chrome (BUG-026 regression)", () => {
  it("renders clinic name in teal header and blue-gray page background", () => {
    const { container } = render(
      <TopPage settings={settings} onNewReservation={() => {}} onMyReservations={() => {}} />,
    );

    expect(screen.getByRole("heading", { name: "ノア動物病院 八王子" })).toBeInTheDocument();
    const header = screen.getByRole("banner");
    expect(header.className).toContain("bg-noah-teal");
    expect(container.firstElementChild?.className).toContain("bg-noah-teal-light");
  });

  it("falls back to default clinic label when header_text is empty (existing behavior)", () => {
    render(
      <TopPage
        settings={{ ...settings, header_text: "" }}
        onNewReservation={() => {}}
        onMyReservations={() => {}}
      />,
    );

    expect(screen.getByRole("heading", { name: "ノア動物病院" })).toBeInTheDocument();
  });
});
