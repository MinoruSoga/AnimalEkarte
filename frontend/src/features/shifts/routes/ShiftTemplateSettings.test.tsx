import { act, render, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { createTestWrapper } from "@/testing/TestUtils";
import type { TemplateFormData } from "../lib/shift-template-form-model";
import type { ShiftTemplate } from "../types";
import { ShiftTemplateSettings } from "./ShiftTemplateSettings";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";
const SAVE_PERMISSION_DENIED_MESSAGE = "シフトテンプレートを保存する権限がありません";

const permission = vi.hoisted(() => ({
  current: { canView: true, canCreate: true, canEdit: true, canDelete: true },
}));

const mutations = vi.hoisted(() => ({
  create: vi.fn(),
  update: vi.fn(),
  remove: vi.fn(),
  reorder: vi.fn(),
}));

const captured = vi.hoisted(() => ({
  onSave: undefined as ((data: TemplateFormData) => void) | undefined,
  onCreate: undefined as (() => void) | undefined,
  onEdit: undefined as ((item: ShiftTemplate) => void) | undefined,
  onDeleteRequest: undefined as (() => void) | undefined,
  onDeleteConfirm: undefined as (() => void) | undefined,
  onReorder: undefined as ((ids: string[]) => void) | undefined,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => permission.current,
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("../api/get-shift-templates", () => ({
  useGetShiftTemplates: () => ({
    data: [
      {
        id: "1",
        clinic_id: "10",
        name: "午前勤務",
        shift_type: "morning",
        start_time: "09:00",
        end_time: "13:00",
        notes: "",
        sort_order: 1,
        is_active: true,
        breaks: [],
        created_at: "",
        updated_at: "",
      },
      {
        id: "2",
        clinic_id: "10",
        name: "午後勤務",
        shift_type: "afternoon",
        start_time: "14:00",
        end_time: "18:00",
        notes: "",
        sort_order: 2,
        is_active: true,
        breaks: [],
        created_at: "",
        updated_at: "",
      },
    ] satisfies ShiftTemplate[],
  }),
}));

vi.mock("../api/create-shift-template", () => ({
  useCreateShiftTemplate: () => ({ mutate: mutations.create, isPending: false }),
}));

vi.mock("../api/update-shift-template", () => ({
  useUpdateShiftTemplate: () => ({ mutate: mutations.update, isPending: false }),
}));

vi.mock("../api/delete-shift-template", () => ({
  useDeleteShiftTemplate: () => ({ mutate: mutations.remove }),
}));

vi.mock("../api/reorder-shift-templates", () => ({
  useReorderShiftTemplates: () => ({ mutate: mutations.reorder }),
}));

vi.mock("@/hooks/use-sortable-list", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/hooks/use-sortable-list")>();
  return {
    ...actual,
    useSortableList: (opts: { items: ShiftTemplate[]; onReorder: (ids: string[]) => void }) => {
      captured.onReorder = opts.onReorder;
      return actual.useSortableList(opts);
    },
  };
});

vi.mock("../components/ShiftTemplateSettingsWorkspace", () => ({
  ShiftTemplateSettingsWorkspace: (props: {
    onSave: (data: TemplateFormData) => void;
    onCreate: () => void;
    onEdit: (item: ShiftTemplate) => void;
    onDeleteRequest?: () => void;
    onDeleteConfirm: () => void;
  }) => {
    captured.onSave = props.onSave;
    captured.onCreate = props.onCreate;
    captured.onEdit = props.onEdit;
    captured.onDeleteRequest = props.onDeleteRequest;
    captured.onDeleteConfirm = props.onDeleteConfirm;
    return <div data-testid="shift-template-settings-workspace" />;
  },
}));

import { toast } from "sonner";

const FORM: TemplateFormData = {
  name: "テスト",
  shift_type: "full",
  start_time: "09:00",
  end_time: "18:00",
  notes: "",
  is_active: true,
  breaks: [],
};

const EXISTING: ShiftTemplate = {
  id: "1",
  clinic_id: "10",
  name: "午前勤務",
  shift_type: "morning",
  start_time: "09:00",
  end_time: "13:00",
  notes: "",
  sort_order: 1,
  is_active: true,
  breaks: [],
  created_at: "",
  updated_at: "",
};

function renderPage() {
  return render(<ShiftTemplateSettings />, {
    wrapper: createTestWrapper({ router: true }),
  });
}

describe("ShiftTemplateSettings mutation permission re-check", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    mutations.create.mockReset();
    mutations.update.mockReset();
    mutations.remove.mockReset();
    mutations.reorder.mockReset();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.success).mockClear();
  });

  it("canCreate=true なら新規保存で create mutate を呼ぶ", () => {
    renderPage();
    act(() => {
      captured.onCreate?.();
    });
    act(() => {
      captured.onSave?.(FORM);
    });

    expect(mutations.create).toHaveBeenCalledTimes(1);
    expect(mutations.update).not.toHaveBeenCalled();
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("canCreate=false なら新規保存で mutate せず専用 toast する", () => {
    const { rerender } = renderPage();
    act(() => {
      captured.onCreate?.();
    });

    permission.current = { ...permission.current, canCreate: false };
    rerender(<ShiftTemplateSettings />);

    act(() => {
      captured.onSave?.(FORM);
    });

    expect(toast.error).toHaveBeenCalledWith(SAVE_PERMISSION_DENIED_MESSAGE);
    expect(mutations.create).not.toHaveBeenCalled();
    expect(mutations.update).not.toHaveBeenCalled();
  });

  it("canEdit=true なら更新保存で update mutate を呼ぶ", () => {
    renderPage();
    act(() => {
      captured.onEdit?.(EXISTING);
    });
    act(() => {
      captured.onSave?.(FORM);
    });

    expect(mutations.update).toHaveBeenCalledTimes(1);
    expect(mutations.create).not.toHaveBeenCalled();
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("canEdit=false なら更新保存で mutate せず専用 toast する", () => {
    const { rerender } = renderPage();
    act(() => {
      captured.onEdit?.(EXISTING);
    });

    permission.current = { ...permission.current, canEdit: false };
    rerender(<ShiftTemplateSettings />);

    act(() => {
      captured.onSave?.(FORM);
    });

    expect(toast.error).toHaveBeenCalledWith(SAVE_PERMISSION_DENIED_MESSAGE);
    expect(mutations.update).not.toHaveBeenCalled();
    expect(mutations.create).not.toHaveBeenCalled();
  });

  it("canDelete=true なら削除確定で delete mutate を呼ぶ", async () => {
    renderPage();
    act(() => {
      captured.onEdit?.(EXISTING);
    });
    await waitFor(() => {
      expect(captured.onDeleteRequest).toBeTypeOf("function");
    });
    act(() => {
      captured.onDeleteRequest?.();
    });
    act(() => {
      captured.onDeleteConfirm?.();
    });

    expect(mutations.remove).toHaveBeenCalledWith("1", expect.any(Object));
    expect(toast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canDelete=false なら削除確定で mutate せず toast.error する", async () => {
    const { rerender } = renderPage();
    act(() => {
      captured.onEdit?.(EXISTING);
    });
    await waitFor(() => {
      expect(captured.onDeleteRequest).toBeTypeOf("function");
    });
    act(() => {
      captured.onDeleteRequest?.();
    });

    permission.current = { ...permission.current, canDelete: false };
    rerender(<ShiftTemplateSettings />);

    act(() => {
      captured.onDeleteConfirm?.();
    });

    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(mutations.remove).not.toHaveBeenCalled();
  });

  it("canEdit=true なら並べ替えで reorder mutate を呼ぶ", () => {
    renderPage();
    act(() => {
      captured.onReorder?.(["2", "1"]);
    });

    expect(mutations.reorder).toHaveBeenCalledWith([2, 1]);
    expect(toast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false なら並べ替えで mutate せず toast.error する", () => {
    const { rerender } = renderPage();
    permission.current = { ...permission.current, canEdit: false };
    rerender(<ShiftTemplateSettings />);

    act(() => {
      captured.onReorder?.(["2", "1"]);
    });

    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(mutations.reorder).not.toHaveBeenCalled();
  });
});
