import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { CheckupAlertBadge } from "../CheckupAlertBadge";

const FIXED_DATE = new Date("2026-05-05T12:00:00.000Z");

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(FIXED_DATE);
});

afterEach(() => {
  vi.useRealTimers();
});

describe("CheckupAlertBadge", () => {
  it("A: nextDate が undefined の場合、何も表示しない", () => {
    const { container } = render(<CheckupAlertBadge nextDate={undefined} />);
    expect(container.firstChild).toBeNull();
  });

  it("B: nextDate が今日より過去の場合、「期限切れ」バッジを表示する", () => {
    render(<CheckupAlertBadge nextDate="2026-04-01" />);
    expect(screen.getByText("期限切れ")).toBeInTheDocument();
  });

  it("C: nextDate が today+15日の場合、「期限間近」バッジを表示する", () => {
    render(<CheckupAlertBadge nextDate="2026-05-20" />);
    expect(screen.getByText("期限間近")).toBeInTheDocument();
  });

  it("D: nextDate が today+31日の場合、バッジを表示しない", () => {
    const { container } = render(<CheckupAlertBadge nextDate="2026-06-05" />);
    expect(container.firstChild).toBeNull();
  });
});
