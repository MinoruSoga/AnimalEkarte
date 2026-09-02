import type { ReactNode } from "react";
import { act, render } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { CourseFormData, OptionFormData } from "../components/trimming-side-panel-model";
import { TrimmingSettings } from "./TrimmingSettings";

const mocks = vi.hoisted(() => {
  const dirty = {
    isDirty: false,
    markDirty: vi.fn(() => {
      dirty.isDirty = true;
    }),
    markClean: vi.fn(() => {
      dirty.isDirty = false;
    }),
    confirmDiscard: vi.fn(() => true),
    runWithDiscardCheck: (fn: () => void) => {
      fn();
    },
  };
  return {
    dirty,
    permissions: {
      canCreate: true,
      canEdit: true,
      canDelete: true,
    },
    editTarget: "new" as "new" | { id: string } | null,
    setEditTarget: vi.fn(),
    createCourse: vi.fn(),
    updateCourse: vi.fn(),
    createOption: vi.fn(),
    updateOption: vi.fn(),
    deleteCourse: vi.fn(),
    deleteOption: vi.fn(),
  };
});

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mocks.permissions,
}));

vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => mocks.dirty,
}));

vi.mock("../hooks/use-master-crud", () => ({
  useMasterCRUD: () => ({
    editTarget: mocks.editTarget,
    setEditTarget: mocks.setEditTarget,
    startSaveTransition: (callback: () => void) => callback(),
    panelItem: null,
    pendingDelete: null,
    handleNew: vi.fn(),
    handleClose: vi.fn(),
    setPendingDelete: vi.fn(),
    handleDeleteCancel: vi.fn(),
    handleDeleteConfirm: vi.fn(),
  }),
}));

vi.mock("../components/MasterTabPage", () => ({
  MasterTabPage: ({ sidePanel }: { sidePanel: ReactNode }) => <>{sidePanel}</>,
}));

type SidePanelSaveProps = {
  onCourseSave: (data: CourseFormData) => void | Promise<boolean>;
  onOptionSave: (data: OptionFormData) => void | Promise<boolean>;
};

let latestSidePanel: SidePanelSaveProps | null = null;

vi.mock("../components/TrimmingSettingsSidePanels", () => ({
  TrimmingSettingsSidePanels: (props: SidePanelSaveProps) => {
    latestSidePanel = props;
    return null;
  },
}));

vi.mock("../components/TrimmingTabs", () => ({
  TrimmingCourseTab: () => null,
  TrimmingOptionTab: () => null,
}));

vi.mock("../components/TrimmingDeleteDialogs", () => ({
  TrimmingDeleteDialogs: () => null,
}));

function mutationStub(mutate: ReturnType<typeof vi.fn>) {
  return { mutate, mutateAsync: mutate };
}

vi.mock("../api/trimming", () => ({
  useCreateTrimmingCourse: () => mutationStub(mocks.createCourse),
  useUpdateTrimmingCourse: () => mutationStub(mocks.updateCourse),
  useDeleteTrimmingCourse: () => mutationStub(mocks.deleteCourse),
  useCreateTrimmingOption: () => mutationStub(mocks.createOption),
  useUpdateTrimmingOption: () => mutationStub(mocks.updateOption),
  useDeleteTrimmingOption: () => mutationStub(mocks.deleteOption),
}));

function savedCourse() {
  return {
    id: "1",
    clinicId: "1",
    name: "カット",
    price: 3000,
    isActive: true,
    description: "",
    targetSize: "small" as const,
    courseTypeId: "1",
    duration: 90,
    sortOrder: 1,
    createdAt: "",
    updatedAt: "",
  };
}

function savedOption() {
  return {
    id: "2",
    clinicId: "1",
    name: "シャンプー",
    price: 1000,
    isActive: true,
    description: "",
    duration: 30,
    combinable: true,
    sortOrder: 1,
    createdAt: "",
    updatedAt: "",
  };
}

function courseForm(overrides: Partial<CourseFormData> = {}): CourseFormData {
  return {
    name: "カット",
    price: "3000",
    targetSize: "small",
    courseTypeId: "1",
    duration: "90",
    description: "",
    isActive: true,
    ...overrides,
  };
}

function optionForm(overrides: Partial<OptionFormData> = {}): OptionFormData {
  return {
    name: "シャンプー",
    price: "1000",
    duration: "30",
    combinable: true,
    description: "",
    isActive: true,
    ...overrides,
  };
}

function renderPage(path = "/") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <TrimmingSettings />
    </MemoryRouter>,
  );
}

describe("TrimmingSettings dirty flag after save", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    latestSidePanel = null;
    mocks.editTarget = "new";
    mocks.permissions.canCreate = true;
    mocks.permissions.canEdit = true;
    mocks.permissions.canDelete = true;
    mocks.dirty.isDirty = true;
    mocks.createCourse.mockResolvedValue(savedCourse());
    mocks.updateCourse.mockResolvedValue(savedCourse());
    mocks.createOption.mockResolvedValue(savedOption());
    mocks.updateOption.mockResolvedValue(savedOption());
  });

  it("コース保存成功後に isDirty が false になる", async () => {
    renderPage();

    await act(async () => {
      await latestSidePanel?.onCourseSave(courseForm());
    });

    expect(mocks.createCourse).toHaveBeenCalledOnce();
    expect(mocks.dirty.markClean).toHaveBeenCalled();
    expect(mocks.dirty.isDirty).toBe(false);
  });

  it("オプション保存成功後に isDirty が false になる", async () => {
    renderPage("/?tab=option");

    await act(async () => {
      await latestSidePanel?.onOptionSave(optionForm());
    });

    expect(mocks.createOption).toHaveBeenCalledOnce();
    expect(mocks.dirty.markClean).toHaveBeenCalled();
    expect(mocks.dirty.isDirty).toBe(false);
  });

  it("コース更新成功後に isDirty が false になる", async () => {
    mocks.editTarget = { id: "1" };
    renderPage();

    await act(async () => {
      await latestSidePanel?.onCourseSave(courseForm({ name: "カット（更新）" }));
    });

    expect(mocks.updateCourse).toHaveBeenCalledOnce();
    expect(mocks.createCourse).not.toHaveBeenCalled();
    expect(mocks.dirty.markClean).toHaveBeenCalled();
    expect(mocks.dirty.isDirty).toBe(false);
  });

  it("保存失敗後は isDirty が残る", async () => {
    mocks.createCourse.mockRejectedValue(new Error("network"));
    renderPage();

    await act(async () => {
      await latestSidePanel?.onCourseSave(courseForm());
    });

    expect(mocks.dirty.markClean).not.toHaveBeenCalled();
    expect(mocks.dirty.isDirty).toBe(true);
  });

  it("名称が空の場合は保存せず isDirty を維持する", async () => {
    renderPage();

    await act(async () => {
      await latestSidePanel?.onCourseSave(courseForm({ name: "   " }));
    });

    expect(mocks.createCourse).not.toHaveBeenCalled();
    expect(mocks.dirty.markClean).not.toHaveBeenCalled();
    expect(mocks.dirty.isDirty).toBe(true);
  });

  it("create権限がない場合は保存せず isDirty を維持する", async () => {
    mocks.permissions.canCreate = false;
    renderPage();

    await act(async () => {
      await latestSidePanel?.onCourseSave(courseForm());
    });

    expect(mocks.createCourse).not.toHaveBeenCalled();
    expect(mocks.dirty.markClean).not.toHaveBeenCalled();
    expect(mocks.dirty.isDirty).toBe(true);
  });
});
