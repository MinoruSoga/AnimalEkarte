import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";
import { CheckupSyncPage } from "./CheckupSyncPage";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: true }),
}));

vi.mock("../api/get-checkup-sync-preview", () => ({
  useGetCheckupSyncPreview: () => ({ data: undefined, isFetching: false }),
}));

vi.mock("../api/create-checkup-sync", () => ({
  useCreateCheckupSync: () => ({ mutate: vi.fn(), isPending: false }),
}));

function hasClassInAncestry(element: Element | null, className: string): boolean {
  let current = element;
  while (current !== null) {
    if (current.classList.contains(className)) return true;
    current = current.parentElement;
  }
  return false;
}

describe("CheckupSyncPage", () => {
  it("DESIGN.md の canvas-soft shell 上に表示する", () => {
    render(<CheckupSyncPage />);

    const heading = screen.getByRole("heading", { name: "健診リマインダー抽出" });
    expect(hasClassInAncestry(heading, C.bgPage)).toBe(true);
  });
});
