import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";

import { LIST_TABLE_COL } from "@/components/shared/DataTable/list-table-col";

import { useMedicalRecordsColumns } from "../hooks/use-medical-records-columns";

describe("useMedicalRecordsColumns status column (BUG-020)", () => {
  it("ステータス列は共有 LIST_TABLE_COL.status（min-width + nowrap）を使う", () => {
    const { result } = renderHook(() =>
      useMedicalRecordsColumns({
        showClinicColumn: false,
        directionForSort: () => "none",
        onSortToggle: vi.fn(),
      }),
    );

    const statusColumn = result.current.find((col) => col.className === LIST_TABLE_COL.status);

    expect(statusColumn).toBeDefined();
    expect(statusColumn?.className).toBe(LIST_TABLE_COL.status);
    expect(statusColumn?.className).toContain("w-[100px]");
    expect(statusColumn?.className).toContain("min-w-[100px]");
    expect(statusColumn?.className).toContain("whitespace-nowrap");
  });
});

describe("useMedicalRecordsColumns identifier columns (EMR-379)", () => {
  it("飼主名・ペット名ヘッダは読める最小幅を持ち、氏名が1字縦積みにならない", () => {
    const { result } = renderHook(() =>
      useMedicalRecordsColumns({
        showClinicColumn: false,
        directionForSort: () => "none",
        onSortToggle: vi.fn(),
      }),
    );

    // 列順は date→owner→pet→species→... の固定契約
    expect(result.current[1].className).toContain("min-w-[140px]");
    expect(result.current[1].className).toContain("whitespace-nowrap");
    expect(result.current[2].className).toContain("min-w-[140px]");
    expect(result.current[2].className).toContain("whitespace-nowrap");
  });

  it("担当医ヘッダは氏名が2字縦積みにならない最小幅を持つ (T003)", () => {
    const { result } = renderHook(() =>
      useMedicalRecordsColumns({
        showClinicColumn: false,
        directionForSort: () => "none",
        onSortToggle: vi.fn(),
      }),
    );

    // 列順は date→owner→pet→species→主訴→関連→担当医 の固定契約
    expect(result.current[6].className).toContain("min-w-[160px]");
    expect(result.current[6].className).toContain("whitespace-nowrap");
  });

  it("診療日ヘッダは折り返さない", () => {
    const { result } = renderHook(() =>
      useMedicalRecordsColumns({
        showClinicColumn: false,
        directionForSort: () => "none",
        onSortToggle: vi.fn(),
      }),
    );

    expect(result.current[0].className).toContain("whitespace-nowrap");
  });
});
