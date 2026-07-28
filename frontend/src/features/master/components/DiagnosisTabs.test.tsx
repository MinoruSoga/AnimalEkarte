import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { DiagnosisNameTab, DiagnosisTypeTab } from "./DiagnosisTabs";

const mocks = vi.hoisted(() => ({
  reorderCallbacks: [] as Array<(ids: string[]) => void>,
  reorderNames: vi.fn(),
  reorderTypes: vi.fn(),
}));

vi.mock("@/hooks/use-sortable-list", () => ({
  useSortableList: ({
    items,
    onReorder,
  }: {
    items: Array<{ id: string }>;
    onReorder: (ids: string[]) => void;
  }) => {
    mocks.reorderCallbacks.push(onReorder);
    return {
      orderedItems: items,
      sensors: [],
      handleDragEnd: vi.fn(),
    };
  },
}));

vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: () => null,
}));

vi.mock("./DiagnosisSortableTable", () => ({
  DiagnosisSortableTable: () => null,
}));

vi.mock("../api/diagnosis", () => ({
  useGetDiagnosisNames: () => ({ data: [] }),
  useGetDiagnosisTypes: () => ({ data: [] }),
  useReorderDiagnosisNames: () => ({ mutate: mocks.reorderNames }),
  useReorderDiagnosisTypes: () => ({ mutate: mocks.reorderTypes }),
}));

function callLatestReorder() {
  const callback = mocks.reorderCallbacks.at(-1);
  expect(callback).toBeDefined();
  callback?.(["2", "1"]);
}

describe("diagnosis reorder permission guards", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.reorderCallbacks.length = 0;
  });

  it("blocks diagnosis type reorder mutation when edit permission is absent", () => {
    render(<DiagnosisTypeTab canEdit={false} onEditTargetChange={vi.fn()} />);
    callLatestReorder();
    expect(mocks.reorderTypes).not.toHaveBeenCalled();
  });

  it("blocks diagnosis name reorder mutation when edit permission is absent", () => {
    render(<DiagnosisNameTab canEdit={false} onEditTargetChange={vi.fn()} />);
    callLatestReorder();
    expect(mocks.reorderNames).not.toHaveBeenCalled();
  });
});
