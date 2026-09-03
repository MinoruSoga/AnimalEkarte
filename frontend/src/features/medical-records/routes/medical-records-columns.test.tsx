import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";

import { LIST_TABLE_COL } from "@/components/shared/DataTable/list-table-col";

import { useMedicalRecordsColumns } from "./MedicalRecordsColumns";

describe("useMedicalRecordsColumns status column (BUG-020)", () => {
  it("ステータス列は共有 LIST_TABLE_COL.status（min-width + nowrap）を使う", () => {
    const { result } = renderHook(() =>
      useMedicalRecordsColumns({
        showClinicColumn: false,
        directionForSort: () => "none",
        onSortToggle: vi.fn(),
      }),
    );

    const statusColumn = result.current.find(
      (col) => col.className === LIST_TABLE_COL.status,
    );

    expect(statusColumn).toBeDefined();
    expect(statusColumn?.className).toBe(LIST_TABLE_COL.status);
    expect(statusColumn?.className).toContain("w-[100px]");
    expect(statusColumn?.className).toContain("min-w-[100px]");
    expect(statusColumn?.className).toContain("whitespace-nowrap");
  });
});
