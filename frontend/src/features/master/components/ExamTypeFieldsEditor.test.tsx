import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type {
  AnimalSpecies,
  useGetAnimalSpecies,
} from "../api/animal-species";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import { ExamTypeFieldsEditor } from "./ExamTypeFieldsEditor";

type AnimalSpeciesQueryState = Pick<
  ReturnType<typeof useGetAnimalSpecies>,
  "data" | "error" | "isError" | "isPending"
>;

const mocks = vi.hoisted(() => ({
  create: vi.fn(),
  update: vi.fn(),
  remove: vi.fn(),
  reorder: vi.fn(),
  replace: vi.fn(),
  resetOrder: vi.fn(),
  getAnimalSpecies: vi.fn<() => AnimalSpeciesQueryState>(),
  reorderCallbacks: [] as Array<(ids: string[]) => void>,
}));

vi.mock("@dnd-kit/core", () => ({
  DndContext: ({ children }: { children: ReactNode }) => <>{children}</>,
  closestCenter: vi.fn(),
}));

vi.mock("@dnd-kit/sortable", () => ({
  SortableContext: ({ children }: { children: ReactNode }) => <>{children}</>,
  verticalListSortingStrategy: vi.fn(),
  useSortable: () => ({
    attributes: {},
    listeners: {},
    setNodeRef: vi.fn(),
    setActivatorNodeRef: vi.fn(),
    transform: null,
    transition: undefined,
    isDragging: false,
  }),
}));

vi.mock("@dnd-kit/utilities", () => ({
  CSS: { Transform: { toString: () => undefined } },
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
      resetOrder: mocks.resetOrder,
    };
  },
}));

vi.mock("../api/animal-species", () => ({
  useGetAnimalSpecies: mocks.getAnimalSpecies,
}));

vi.mock("../api/exam-types-master", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/exam-types-master")>();
  return {
    ...actual,
    useCreateExaminationTypeField: () => ({ mutateAsync: mocks.create }),
    useUpdateExaminationTypeField: () => ({ mutateAsync: mocks.update }),
    useDeleteExaminationTypeField: () => ({ mutate: mocks.remove }),
    useReorderExaminationTypeFields: () => ({ mutate: mocks.reorder }),
    useReplaceExamTypeFieldReferenceRanges: () => ({ mutateAsync: mocks.replace }),
  };
});

const examType: ExaminationTypeMaster = {
  id: "3",
  name: "血液検査",
  price: 1000,
  isActive: true,
  description: "",
  sortOrder: 1,
  isNonInsurance: false,
  createdAt: "",
  updatedAt: "",
  items: [{
    id: "31",
    examTypeId: "3",
    name: "白血球",
    inspectionValue: "",
    normalValue: "",
    unit: "/μL",
    sortOrder: 1,
    createdAt: "",
    updatedAt: "",
    referenceRanges: [{
      id: "41",
      examTypeFieldId: "31",
      animalSpeciesId: "2",
      refMin: 5,
      refMax: 10,
    }],
  }],
};

const animalSpecies: AnimalSpecies[] = [
  {
    id: "2",
    name: "犬",
    isActive: true,
    sortOrder: 1,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  },
  {
    id: "3",
    name: "猫",
    isActive: true,
    sortOrder: 2,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  },
];

describe("ExamTypeFieldsEditor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.reorderCallbacks.length = 0;
    mocks.create.mockResolvedValue(undefined);
    mocks.update.mockResolvedValue(undefined);
    mocks.replace.mockResolvedValue(undefined);
    mocks.getAnimalSpecies.mockReturnValue({
      data: animalSpecies,
      error: null,
      isError: false,
      isPending: false,
    });
  });

  it("exposes accessible field create/edit/delete and reorder controls", async () => {
    const user = userEvent.setup();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
      />,
    );

    expect(screen.getByRole("heading", { name: "検査項目" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "検査項目を追加" })).toBeInTheDocument();
    expect(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    })).toBeInTheDocument();
    await user.click(screen.getByRole("button", {
      name: "削除: 検査項目 白血球 (ID 31)",
    }));
    expect(mocks.remove).toHaveBeenCalledWith({ examTypeId: "3", fieldId: "31" });

    mocks.reorderCallbacks.at(-1)?.(["31"]);
    expect(mocks.reorder).toHaveBeenCalledWith(
      { examTypeId: "3", ids: [31] },
      expect.objectContaining({ onError: expect.any(Function) }),
    );
    mocks.reorder.mock.calls[0]?.[1]?.onError();
    expect(mocks.resetOrder).toHaveBeenCalledTimes(1);
  });

  it("prevents all mutation entry points in read-only mode", () => {
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate={false}
        canEdit={false}
        canDelete={false}
      />,
    );

    expect(screen.queryByRole("button", { name: "検査項目を追加" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /編集: 検査項目/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /削除: 検査項目/ })).not.toBeInTheDocument();
    mocks.reorderCallbacks.at(-1)?.(["31"]);
    expect(mocks.reorder).not.toHaveBeenCalled();
  });

  it("validates reversed numeric ranges before full replacement", async () => {
    const user = userEvent.setup();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    const min = screen.getByRole("spinbutton", { name: "犬の数値下限" });
    const max = screen.getByRole("spinbutton", { name: "犬の数値上限" });
    await user.clear(min);
    await user.type(min, "10");
    await user.clear(max);
    await user.type(max, "5");
    await user.click(screen.getByRole("button", { name: "基準範囲を保存" }));

    expect(screen.getByRole("alert")).toHaveTextContent(
      "数値範囲の下限は上限以下にしてください",
    );
    expect(mocks.replace).not.toHaveBeenCalled();
  });

  it("sends an explicit empty array when all species ranges are cleared", async () => {
    const user = userEvent.setup();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    await user.click(screen.getByRole("checkbox", { name: "犬の基準範囲を使用" }));
    await user.click(screen.getByRole("button", { name: "基準範囲を保存" }));

    expect(mocks.replace).toHaveBeenCalledWith({
      examTypeId: "3",
      fieldId: "31",
      ranges: [],
    });
  });

  it("shows an accessible loading status without empty, error, or range controls", async () => {
    mocks.getAnimalSpecies.mockReturnValue({
      data: undefined,
      error: null,
      isError: false,
      isPending: true,
    });
    const user = userEvent.setup();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));

    const status = screen.getByRole("status");
    expect(status).toHaveTextContent(
      "動物種を読み込み中です。基準範囲はまだ設定できません。",
    );
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText(/動物種マスタが登録されていない/))
      .not.toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "基準範囲を保存" }))
      .not.toBeInTheDocument();
  });

  it("shows an accessible generic fetch error without raw details or range controls", async () => {
    const rawError = "GET /v1/masters/animal-species: database timeout";
    mocks.getAnimalSpecies.mockReturnValue({
      data: undefined,
      error: new Error(rawError),
      isError: true,
      isPending: false,
    });
    const user = userEvent.setup();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent(
      "動物種の取得に失敗したため、基準範囲を設定できません。",
    );
    expect(alert).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByText(rawError)).not.toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.queryByText(/動物種を読み込み中/)).not.toBeInTheDocument();
    expect(screen.queryByText(/動物種マスタが登録されていない/))
      .not.toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "基準範囲を保存" }))
      .not.toBeInTheDocument();
  });

  it("shows a distinct accessible status when the species master is empty", async () => {
    mocks.getAnimalSpecies.mockReturnValue({
      data: [],
      error: null,
      isError: false,
      isPending: false,
    });
    const user = userEvent.setup();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));

    const status = screen.getByRole("status");
    expect(status).toHaveTextContent(
      "動物種マスタが登録されていないため、基準範囲を設定できません。",
    );
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText(/動物種を読み込み中/)).not.toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "基準範囲を保存" }))
      .not.toBeInTheDocument();
  });

  it("still saves edited field metadata with the exact payload after a species fetch error", async () => {
    mocks.getAnimalSpecies.mockReturnValue({
      data: undefined,
      error: new Error("network failure"),
      isError: true,
      isPending: false,
    });
    const user = userEvent.setup();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    const name = screen.getByRole("textbox", { name: "検査項目名" });
    await user.clear(name);
    await user.type(name, "赤血球");
    await user.click(screen.getByRole("button", { name: "検査項目情報を保存" }));

    expect(mocks.update).toHaveBeenCalledWith({
      examTypeId: "3",
      fieldId: "31",
      req: {
        name: "赤血球",
        inspection_value: "",
        normal_value: "",
        unit: "/μL",
      },
    });
  });

  it("blocks nested text input Enter from submitting its parent form", async () => {
    const user = userEvent.setup();
    const parentSubmit = vi.fn((event: React.FormEvent) => event.preventDefault());
    const parentKeyDown = vi.fn();
    render(
      <form onSubmit={parentSubmit} onKeyDown={parentKeyDown}>
        <ExamTypeFieldsEditor
          examType={examType}
          canCreate
          canEdit
          canDelete
        />
      </form>,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    await user.type(screen.getByRole("textbox", { name: "検査項目名" }), "{enter}");

    expect(parentSubmit).not.toHaveBeenCalled();
    expect(parentKeyDown).not.toHaveBeenCalled();
  });

  it("reports nested dirty state and clears it on cancel and permission loss", async () => {
    const user = userEvent.setup();
    const onDirtyChange = vi.fn();
    const { rerender } = render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
        onDirtyChange={onDirtyChange}
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    await user.type(screen.getByRole("textbox", { name: "単位" }), "x");
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);

    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    await user.type(screen.getByRole("textbox", { name: "単位" }), "y");
    await user.click(screen.getByRole("button", { name: "検査項目情報を保存" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    await user.type(screen.getByRole("textbox", { name: "単位" }), "z");
    rerender(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit={false}
        canDelete
        onDirtyChange={onDirtyChange}
      />,
    );
    expect(screen.queryByRole("heading", { name: "検査項目を編集" })).not.toBeInTheDocument();
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);
  });

  it("keeps dirty drafts open when a mutation rejects", async () => {
    const user = userEvent.setup();
    const onDirtyChange = vi.fn();
    mocks.update.mockRejectedValueOnce(new Error("network"));
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
        onDirtyChange={onDirtyChange}
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    await user.type(screen.getByRole("textbox", { name: "単位" }), "z");
    await user.click(screen.getByRole("button", { name: "検査項目情報を保存" }));

    expect(screen.getByRole("heading", { name: "検査項目を編集" })).toBeInTheDocument();
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);
  });

  it("preserves dirty ranges when existing field metadata is saved first", async () => {
    const user = userEvent.setup();
    const onDirtyChange = vi.fn();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
        onDirtyChange={onDirtyChange}
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    await user.type(screen.getByRole("textbox", { name: "単位" }), "x");
    const rangeMax = screen.getByRole("spinbutton", { name: "犬の数値上限" });
    await user.clear(rangeMax);
    await user.type(rangeMax, "12");

    expect(screen.getByRole("button", { name: "検査項目を追加" })).toBeDisabled();
    expect(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    })).toBeDisabled();
    expect(screen.getByRole("button", {
      name: "削除: 検査項目 白血球 (ID 31)",
    })).toBeDisabled();
    mocks.reorderCallbacks.at(-1)?.(["31"]);
    expect(mocks.reorder).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "検査項目情報を保存" }));

    expect(mocks.update).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("spinbutton", { name: "犬の数値上限" }))
      .toHaveValue(12);
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);

    await user.click(screen.getByRole("button", { name: "基準範囲を保存" }));
    expect(mocks.replace).toHaveBeenCalledTimes(1);
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);
  });

  it("preserves dirty field metadata when reference ranges are saved first", async () => {
    const user = userEvent.setup();
    const onDirtyChange = vi.fn();
    render(
      <ExamTypeFieldsEditor
        examType={examType}
        canCreate
        canEdit
        canDelete
        onDirtyChange={onDirtyChange}
      />,
    );

    await user.click(screen.getByRole("button", {
      name: "編集: 検査項目 白血球 (ID 31)",
    }));
    const unit = screen.getByRole("textbox", { name: "単位" });
    await user.type(unit, "y");
    const rangeMax = screen.getByRole("spinbutton", { name: "犬の数値上限" });
    await user.clear(rangeMax);
    await user.type(rangeMax, "13");

    await user.click(screen.getByRole("button", { name: "基準範囲を保存" }));

    expect(mocks.replace).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("textbox", { name: "単位" })).toHaveValue("/μLy");
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);

    await user.click(screen.getByRole("button", { name: "検査項目情報を保存" }));
    expect(mocks.update).toHaveBeenCalledTimes(1);
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);
  });
});
