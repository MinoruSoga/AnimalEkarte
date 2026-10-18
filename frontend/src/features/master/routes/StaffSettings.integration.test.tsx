import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";
import { PALETTE } from "@/lib/design-tokens";
import { StaffSettings } from "./StaffSettings";
import type { ClinicSummary, Staff } from "../api/staffs";
import type { PermissionGroup } from "../api/permission-groups";
import type { ReservationType } from "../api/reservation-types";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

const m = vi.hoisted(() => ({
  createMutateAsync: vi.fn(),
  updateMutateAsync: vi.fn(),
  deleteMutate: vi.fn(),
  groupsMutate: vi.fn(),
  groupsMutateAsync: vi.fn(),
  clinicsMutate: vi.fn(),
  clinicsMutateAsync: vi.fn(),
  capableMutate: vi.fn(),
  capableMutateAsync: vi.fn(),
}));

const serverAssoc = vi.hoisted(() => ({
  groups: [] as string[] | undefined,
  clinics: [] as string[] | undefined,
  capable: [] as string[] | undefined,
}));

const permState = vi.hoisted(() => ({
  check: ((_resource: string, _action: string) => true) as (
    resource: string,
    action: string,
  ) => boolean,
}));

vi.mock("../api/staffs", () => ({
  useGetStaffs: () => ({ isPending: false, isError: false }),
  useCreateStaff: () => ({ mutateAsync: m.createMutateAsync }),
  useUpdateStaff: () => ({ mutateAsync: m.updateMutateAsync }),
  useAttachStaffAccount: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteStaff: () => ({ mutate: m.deleteMutate }),
  useGetClinicsList: () => ({ data: [] }),
  useGetAllStaffPermissionGroupMap: () => ({ data: new Map() }),
  useGetStaffPermissionGroups: () => ({ data: serverAssoc.groups }),
  useGetStaffClinics: () => ({ data: serverAssoc.clinics }),
  useGetStaffCapableReservationTypes: () => ({ data: serverAssoc.capable }),
  useUpdateStaffPermissionGroups: () => ({
    mutate: m.groupsMutate,
    mutateAsync: m.groupsMutateAsync,
  }),
  useUpdateStaffClinics: () => ({
    mutate: m.clinicsMutate,
    mutateAsync: m.clinicsMutateAsync,
  }),
  useUpdateStaffCapableReservationTypes: () => ({
    mutate: m.capableMutate,
    mutateAsync: m.capableMutateAsync,
  }),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { isSystemAdmin: false },
    currentClinicId: "1",
    isAuthenticated: true,
    isLoading: false,
    hasPermission: (resource: string, action: string) => permState.check(resource, action),
    login: vi.fn(),
    logout: vi.fn(),
    switchClinic: vi.fn(),
    refreshPermissions: vi.fn(),
  }),
}));

function makeStaff(overrides: Partial<Staff> = {}): Staff {
  return {
    id: "1",
    clinicId: "1",
    name: "Synthetic Staff One",
    isActive: true,
    occupationId: null,
    occupationName: null,
    licenseNumber: "",
    sortOrder: 1,
    email: "",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    staffType: "doctor",
    reservationDisplayName: "",
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    ...overrides,
  };
}

const STAFFS: Staff[] = [
  makeStaff({ id: "1", name: "Synthetic Staff One" }),
  makeStaff({ id: "2", name: "Synthetic Staff Two" }),
];

const GROUPS: PermissionGroup[] = [
  {
    id: "10",
    clinicId: "1",
    name: "Synthetic Group A",
    description: "",
    color: null,
    isActive: true,
    sortOrder: 1,
    rules: [],
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  },
  {
    id: "11",
    clinicId: "1",
    name: "Synthetic Group B",
    description: "",
    color: null,
    isActive: true,
    sortOrder: 1,
    rules: [],
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  },
];

const CLINICS: ClinicSummary[] = [{ id: "1", name: "Synthetic Clinic A", isActive: true }];

const RESERVATION_TYPES: ReservationType[] = [
  {
    id: "100",
    clinicId: "1",
    name: "Synthetic RT A",
    color: PALETTE.pickerDefaultBlue,
    isActive: true,
    description: "",
    sortOrder: 1,
    groupId: undefined,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    reservationDisplayName: "",
    durationMinutes: 15,
    shortName: "",
    showShortName: false,
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    reservationDayOption: "none",
    isInternal: false,
    category: "general",
    parentId: undefined,
    parentName: undefined,
    isLeaf: true,
    depth: 0,
    childIds: [],
  },
];

vi.mock("../hooks/use-staff-settings-lookups", () => ({
  useStaffSettingsLookups: () => ({
    data: STAFFS,
    allOccupations: [],
    allGroups: GROUPS,
    allReservationTypes: RESERVATION_TYPES,
    allClinics: CLINICS,
    groupsByStaffId: new Map<string, PermissionGroup[]>(),
    staffFilterProperties: [],
  }),
}));

interface Deferred<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
  reject: (reason?: unknown) => void;
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

type AssociationKey = "groups" | "clinics" | "capable";

const ASSOCIATION_MOCKS: Record<
  AssociationKey,
  { mutate: () => ReturnType<typeof vi.fn>; mutateAsync: () => ReturnType<typeof vi.fn> }
> = {
  groups: { mutate: () => m.groupsMutate, mutateAsync: () => m.groupsMutateAsync },
  clinics: { mutate: () => m.clinicsMutate, mutateAsync: () => m.clinicsMutateAsync },
  capable: { mutate: () => m.capableMutate, mutateAsync: () => m.capableMutateAsync },
};

const ASSOCIATION_LABEL: Record<AssociationKey, string> = {
  groups: "Synthetic Group A",
  clinics: "Synthetic Clinic A",
  capable: "Synthetic RT A",
};

const ASSOCIATION_PAYLOAD_KEY: Record<AssociationKey, string> = {
  groups: "groupIds",
  clinics: "clinicIds",
  capable: "reservationTypeIds",
};

const ASSOCIATION_EXPECTED_IDS: Record<AssociationKey, string[]> = {
  groups: ["10"],
  clinics: ["1"],
  capable: ["100"],
};

function associationCalls(key: AssociationKey): unknown[][] {
  const mocks = ASSOCIATION_MOCKS[key];
  return [...mocks.mutate().mock.calls, ...mocks.mutateAsync().mock.calls];
}

let pendingAssoc: Deferred<unknown> | null = null;
let basicUpdate: Deferred<Staff> | null = null;

function installAssociationMutate(target: AssociationKey | null) {
  for (const key of Object.keys(ASSOCIATION_MOCKS) as AssociationKey[]) {
    const mocks = ASSOCIATION_MOCKS[key];
    mocks.mutate().mockImplementation((vars, opts) => {
      const d = deferred<unknown>();
      if (key === target) {
        pendingAssoc = d;
      } else {
        d.resolve(undefined);
      }
      void d.promise.catch(() => {});
      void d.promise.then(
        (v) => opts?.onSuccess?.(v, vars),
        (e) => opts?.onError?.(e, vars),
      );
    });
    mocks.mutateAsync().mockImplementation(() => {
      const d = deferred<unknown>();
      if (key === target) {
        pendingAssoc = d;
      } else {
        d.resolve(undefined);
      }
      return d.promise;
    });
  }
}

function renderPage() {
  const router = createMemoryRouter([{ path: "*", element: <StaffSettings /> }], {
    initialEntries: ["/"],
  });
  const utils = render(<RouterProvider router={router} />);
  return { router, ...utils };
}

const rowButtonName = (staff: Staff) => `詳細: スタッフ ${staff.name} (ID ${staff.id})`;

async function openRow(user: ReturnType<typeof userEvent.setup>, staff: Staff) {
  await user.click(await screen.findByRole("button", { name: rowButtonName(staff) }));
  await waitFor(() => expect(document.getElementById("master-title")).not.toBeNull());
}

async function editTitle(user: ReturnType<typeof userEvent.setup>, value: string) {
  const title = document.getElementById("master-title") as HTMLInputElement;
  await user.clear(title);
  await user.type(title, value);
}

async function clickSave(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("button", { name: "保存" }));
}

const panelOpen = () => document.getElementById("master-title") !== null;

async function waitPanelClosed() {
  await waitFor(() => expect(document.getElementById("master-title")).toBeNull());
}

const DISCARD_TITLE = "未保存の変更があります";

beforeEach(() => {
  vi.clearAllMocks();
  pendingAssoc = null;
  basicUpdate = null;
  serverAssoc.groups = [];
  serverAssoc.clinics = [];
  serverAssoc.capable = [];
  permState.check = () => true;
  installAssociationMutate(null);
  m.createMutateAsync.mockResolvedValue(makeStaff({ id: "3" }));
  m.updateMutateAsync.mockImplementation(() => {
    basicUpdate = deferred<Staff>();
    return basicUpdate.promise;
  });
  m.deleteMutate.mockReturnValue(undefined);
});

describe("StaffSettings integration — dirty guard control", () => {
  it("別行クリックは未保存変更の確認ダイアログを出し、キャンセルで編集が保持される", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");

    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));

    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(DISCARD_TITLE)).toBeInTheDocument();

    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(panelOpen()).toBe(true);
    expect((document.getElementById("master-title") as HTMLInputElement).value).toBe("Edited Name");
  });
});

describe("StaffSettings integration — basic save lifecycle", () => {
  it("基本 PATCH が pending の間は編集パネルを保持し、成功後は未保存ガードなく別行へ移れる", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    expect(panelOpen()).toBe(true);

    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitPanelClosed();
    for (const key of Object.keys(ASSOCIATION_MOCKS) as AssociationKey[]) {
      expect(associationCalls(key)).toHaveLength(0);
    }

    await user.click(await screen.findByRole("button", { name: rowButtonName(STAFFS[0]) }));
    expect(screen.queryByText(DISCARD_TITLE)).toBeNull();
    await waitFor(() => expect(panelOpen()).toBe(true));

    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));
    expect(screen.queryByText(DISCARD_TITLE)).toBeNull();
    expect((document.getElementById("master-title") as HTMLInputElement).value).toBe(
      "Synthetic Staff Two",
    );
  });

  it("基本 PATCH が失敗した場合は編集が保持され、別行移動は破棄確認を要求する（control）", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.reject(new Error("synthetic update failure"));
    });

    expect(panelOpen()).toBe(true);

    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));
    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(DISCARD_TITLE)).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(panelOpen()).toBe(true);
  });
});

describe.each(["groups", "clinics", "capable"] as AssociationKey[])(
  "StaffSettings integration — association %s pending",
  (target) => {
    it("関連付け PUT が pending の間も編集パネルを保持し成功完了扱いしない", async () => {
      installAssociationMutate(target);
      const user = userEvent.setup();
      renderPage();
      await openRow(user, STAFFS[0]);

      await user.click(screen.getByRole("checkbox", { name: ASSOCIATION_LABEL[target] }));
      await clickSave(user);

      await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
      await act(async () => {
        basicUpdate!.resolve(makeStaff({ id: "1" }));
      });

      await waitFor(() => expect(associationCalls(target)).toHaveLength(1));
      expect(pendingAssoc).not.toBeNull();
      expect(associationCalls(target)[0][0]).toMatchObject({
        staffId: "1",
        [ASSOCIATION_PAYLOAD_KEY[target]]: ASSOCIATION_EXPECTED_IDS[target],
      });
      expect(toast.success).not.toHaveBeenCalled();

      await act(async () => {});
      expect(panelOpen()).toBe(true);

      await act(async () => {
        pendingAssoc!.resolve(undefined);
      });
      await waitPanelClosed();
      expect(toast.success).toHaveBeenCalledTimes(1);

      await openRow(user, STAFFS[0]);
      expect(screen.queryByText(DISCARD_TITLE)).toBeNull();
      expect(panelOpen()).toBe(true);
      await user.click(screen.getByRole("button", { name: "キャンセル" }));
      await waitPanelClosed();
      await openRow(user, STAFFS[1]);
      expect(screen.queryByText(DISCARD_TITLE)).toBeNull();
      expect(panelOpen()).toBe(true);
    });
  },
);

describe.each(["groups", "clinics", "capable"] as AssociationKey[])(
  "StaffSettings integration — association %s failure",
  (target) => {
    it("関連付け PUT が失敗した場合も編集・dirty 保護を維持する", async () => {
      installAssociationMutate(target);
      const user = userEvent.setup();
      renderPage();
      await openRow(user, STAFFS[0]);

      await user.click(screen.getByRole("checkbox", { name: ASSOCIATION_LABEL[target] }));
      await clickSave(user);

      await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
      await act(async () => {
        basicUpdate!.resolve(makeStaff({ id: "1" }));
      });

      await waitFor(() => expect(associationCalls(target)).toHaveLength(1));
      expect(pendingAssoc).not.toBeNull();
      expect(associationCalls(target)[0][0]).toMatchObject({
        staffId: "1",
        [ASSOCIATION_PAYLOAD_KEY[target]]: ASSOCIATION_EXPECTED_IDS[target],
      });

      await act(async () => {
        pendingAssoc!.reject(new Error(`synthetic ${target} failure`));
      });

      expect(toast.success).not.toHaveBeenCalled();
      await act(async () => {});
      expect(panelOpen()).toBe(true);
      expect(screen.getByRole("checkbox", { name: ASSOCIATION_LABEL[target] })).toBeChecked();
      expect((document.getElementById("master-title") as HTMLInputElement).value).toBe(
        "Synthetic Staff One",
      );

      await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));
      const dialog = await screen.findByRole("alertdialog");
      expect(within(dialog).getByText(DISCARD_TITLE)).toBeInTheDocument();
      await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));
      await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
      expect(panelOpen()).toBe(true);
      expect(screen.getByRole("checkbox", { name: ASSOCIATION_LABEL[target] })).toBeChecked();
    });
  },
);

const DENY_CASES: Array<{
  target: AssociationKey;
  resource: string;
  action: string;
  message: string;
}> = [
  {
    target: "groups",
    resource: "master-permission",
    action: "edit",
    message: "権限グループを変更する権限がありません",
  },
  {
    target: "groups",
    resource: "master-staff",
    action: "edit",
    message: "権限グループを変更する権限がありません",
  },
  {
    target: "clinics",
    resource: "master-staff",
    action: "edit",
    message: "所属医院を変更する権限がありません",
  },
  {
    target: "capable",
    resource: "master-staff",
    action: "edit",
    message: "対応区分を変更する権限がありません",
  },
];

describe("StaffSettings integration — contract cases", () => {
  it("基本情報のみの保存は権限グループ編集権限が無くても成功し、関連付けPUTを送らない", async () => {
    permState.check = (resource, action) =>
      !(resource === "master-permission" && action === "edit");
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitPanelClosed();
    for (const key of Object.keys(ASSOCIATION_MOCKS) as AssociationKey[]) {
      expect(associationCalls(key)).toHaveLength(0);
    }
    expect(toast.success).toHaveBeenCalledTimes(1);
  });

  it.each(DENY_CASES)(
    "基本PATCH pending中に権限剥奪された %s の変更PUTは発行されず編集を保持する ($resource:$action)",
    async ({ target, resource, action, message }) => {
      installAssociationMutate(target);
      const user = userEvent.setup();
      renderPage();
      await openRow(user, STAFFS[0]);

      await user.click(screen.getByRole("checkbox", { name: ASSOCIATION_LABEL[target] }));
      await clickSave(user);

      await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
      permState.check = (r, a) => !(r === resource && a === action);
      await user.type(screen.getByPlaceholderText("氏名、職種で検索..."), "x{Enter}");
      await waitFor(() =>
        expect(screen.queryByRole("button", { name: rowButtonName(STAFFS[0]) })).toBeNull(),
      );

      await act(async () => {
        basicUpdate!.resolve(makeStaff({ id: "1" }));
      });
      await act(async () => {});

      expect(associationCalls(target)).toHaveLength(0);
      expect(toast.error).toHaveBeenCalledWith(message);
      expect(toast.success).not.toHaveBeenCalled();
      expect(panelOpen()).toBe(true);

      const searchInput = screen.getByPlaceholderText("氏名、職種で検索...");
      await user.clear(searchInput);
      await user.type(searchInput, "{Enter}");
      await user.click(await screen.findByRole("button", { name: rowButtonName(STAFFS[1]) }));
      const dialog = await screen.findByRole("alertdialog");
      expect(within(dialog).getByText(DISCARD_TITLE)).toBeInTheDocument();
      await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));
      await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
      expect(panelOpen()).toBe(true);
    },
  );

  it("サーバー値を全てクリアした場合は空配列を1回だけ送信する", async () => {
    serverAssoc.groups = ["10"];
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);

    await user.click(screen.getByRole("checkbox", { name: "Synthetic Group A" }));
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });

    await waitFor(() => expect(associationCalls("groups")).toHaveLength(1));
    expect(associationCalls("groups")[0][0]).toMatchObject({ staffId: "1", groupIds: [] });
    expect(associationCalls("clinics")).toHaveLength(0);
    expect(associationCalls("capable")).toHaveLength(0);
    await act(async () => {
      pendingAssoc?.resolve(undefined);
    });
    await waitPanelClosed();
  });

  it("未変更の関連付けはPUTを送らない（サーバー値と同一選択）", async () => {
    serverAssoc.groups = ["10"];
    serverAssoc.clinics = ["1"];
    serverAssoc.capable = ["100"];
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitPanelClosed();
    for (const key of Object.keys(ASSOCIATION_MOCKS) as AssociationKey[]) {
      expect(associationCalls(key)).toHaveLength(0);
    }
  });

  it.each(["groups", "clinics", "capable"] as AssociationKey[])(
    "関連付けクエリ未取得(undefined)の %s セクションは編集不可で、空配列の代わり送りもしない",
    async (target) => {
      serverAssoc[target] = undefined;
      const user = userEvent.setup();
      renderPage();
      await openRow(user, STAFFS[0]);

      const checkbox = screen.getByRole("checkbox", { name: ASSOCIATION_LABEL[target] });
      expect(checkbox.closest("fieldset")).toHaveProperty("disabled", true);
      await user.click(checkbox);
      expect(checkbox).not.toBeChecked();

      await editTitle(user, "Edited Name");
      await clickSave(user);

      await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
      await act(async () => {
        basicUpdate!.resolve(makeStaff({ id: "1" }));
      });
      await waitPanelClosed();
      expect(associationCalls(target)).toHaveLength(0);
    },
  );

  it("空でロード済みの関連付けは編集可能で、追加すればPUTされる", async () => {
    serverAssoc.groups = [];
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);

    const checkbox = screen.getByRole("checkbox", { name: "Synthetic Group A" });
    expect(checkbox.closest("fieldset")).toHaveProperty("disabled", false);
    await user.click(checkbox);
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitFor(() => expect(associationCalls("groups")).toHaveLength(1));
    expect(associationCalls("groups")[0][0]).toMatchObject({ staffId: "1", groupIds: ["10"] });
    await act(async () => {
      pendingAssoc?.resolve(undefined);
    });
    await waitPanelClosed();
  });

  it("選択順序だけが異なる同一集合はPUTを送らない", async () => {
    serverAssoc.groups = ["10", "11"];
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);

    await user.click(screen.getByRole("checkbox", { name: "Synthetic Group A" }));
    await user.click(screen.getByRole("checkbox", { name: "Synthetic Group A" }));
    await editTitle(user, "Edited Name");
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitPanelClosed();
    for (const key of Object.keys(ASSOCIATION_MOCKS) as AssociationKey[]) {
      expect(associationCalls(key)).toHaveLength(0);
    }
  });

  it("新規作成は基本POSTだけ成功し、関連付けPUTなしで閉じて再オープンできる", async () => {
    const user = userEvent.setup();
    renderPage();
    await user.click(await screen.findByRole("button", { name: "新規登録" }));
    await waitFor(() => expect(panelOpen()).toBe(true));
    await user.type(document.getElementById("master-title") as HTMLElement, "New Staff");
    await clickSave(user);

    await waitFor(() => expect(m.createMutateAsync).toHaveBeenCalledTimes(1));
    await waitPanelClosed();
    for (const key of Object.keys(ASSOCIATION_MOCKS) as AssociationKey[]) {
      expect(associationCalls(key)).toHaveLength(0);
    }
    expect(toast.success).toHaveBeenCalledTimes(1);

    await openRow(user, STAFFS[0]);
    expect(screen.queryByText(DISCARD_TITLE)).toBeNull();
    expect(panelOpen()).toBe(true);
  });

  it("保存成功後に再度開いて編集すると未保存ガードが再び働く", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitPanelClosed();

    await openRow(user, STAFFS[0]);
    await editTitle(user, "Second Edit");
    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));
    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(DISCARD_TITLE)).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(panelOpen()).toBe(true);
  });

  it("パネルのキャンセルは破棄確認を出し、キャンセル後の別行移動も再度警告する", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");

    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(DISCARD_TITLE)).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(panelOpen()).toBe(true);

    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));
    const dialog2 = await screen.findByRole("alertdialog");
    expect(within(dialog2).getByText(DISCARD_TITLE)).toBeInTheDocument();
    await user.click(within(dialog2).getByRole("button", { name: "キャンセル" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(panelOpen()).toBe(true);
  });

  it("保存pending中の新規編集は保存完了で閉じずdirtyを保持する", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    const pendingButton = await screen.findByRole("button", { name: "保存中..." });
    expect(pendingButton).toBeDisabled();
    const titleInput = document.getElementById("master-title") as HTMLInputElement;
    await user.clear(titleInput);
    await user.type(titleInput, "Newer Edit");
    expect(titleInput).not.toBeDisabled();
    await user.click(screen.getByRole("checkbox", { name: "Synthetic Group A" }));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await act(async () => {});

    expect(toast.success).not.toHaveBeenCalled();
    expect(panelOpen()).toBe(true);
    expect(associationCalls("groups")).toHaveLength(0);
    expect(screen.getByRole("checkbox", { name: "Synthetic Group A" })).toBeChecked();

    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));
    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).getByText(DISCARD_TITLE)).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(panelOpen()).toBe(true);
    expect((document.getElementById("master-title") as HTMLInputElement).value).toBe("Newer Edit");
  });

  it("基本PATCH pending中に破棄した場合、旧結果は新パネルを閉じない", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "確認" }));
    await waitFor(() => expect(panelOpen()).toBe(true));
    expect((document.getElementById("master-title") as HTMLInputElement).value).toBe(
      "Synthetic Staff Two",
    );

    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await act(async () => {});
    expect(panelOpen()).toBe(true);
    expect(toast.success).not.toHaveBeenCalled();
    expect((document.getElementById("master-title") as HTMLInputElement).value).toBe(
      "Synthetic Staff Two",
    );
  });

  it("関連付けpending中に破棄して別行を開いた場合、旧pending結果は新パネルを閉じない", async () => {
    installAssociationMutate("groups");
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await user.click(screen.getByRole("checkbox", { name: "Synthetic Group A" }));
    await clickSave(user);

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitFor(() => expect(associationCalls("groups")).toHaveLength(1));

    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[1]) }));
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "確認" }));
    await waitFor(() => expect(panelOpen()).toBe(true));
    expect((document.getElementById("master-title") as HTMLInputElement).value).toBe(
      "Synthetic Staff Two",
    );

    await act(async () => {
      pendingAssoc?.resolve(undefined);
    });
    await act(async () => {});
    expect(panelOpen()).toBe(true);
    expect(toast.success).not.toHaveBeenCalled();
    expect((document.getElementById("master-title") as HTMLInputElement).value).toBe(
      "Synthetic Staff Two",
    );

    await editTitle(user, "Second Panel Edit");
    await user.click(screen.getByRole("button", { name: rowButtonName(STAFFS[0]) }));
    const dialog2 = await screen.findByRole("alertdialog");
    expect(within(dialog2).getByText(DISCARD_TITLE)).toBeInTheDocument();
    await user.click(within(dialog2).getByRole("button", { name: "キャンセル" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(panelOpen()).toBe(true);
    expect((document.getElementById("master-title") as HTMLInputElement).value).toBe(
      "Second Panel Edit",
    );
  });

  it("タイトル欄で Enter でも基本PATCHが発行される", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    const titleInput = document.getElementById("master-title") as HTMLInputElement;
    await user.clear(titleInput);
    await user.type(titleInput, "Enter Save{Enter}");

    await waitFor(() => expect(m.updateMutateAsync).toHaveBeenCalledTimes(1));
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitPanelClosed();
  });

  it("保存ボタンの重複クリックは基本PATCHを1回しか発行しない", async () => {
    const user = userEvent.setup();
    renderPage();
    await openRow(user, STAFFS[0]);
    await editTitle(user, "Edited Name");

    const saveButton = screen.getByRole("button", { name: "保存" });
    await user.click(saveButton);
    const pendingButton = await screen.findByRole("button", { name: "保存中..." });
    expect(pendingButton).toBeDisabled();
    await user.click(pendingButton);
    (document.getElementById("master-title") as HTMLInputElement).focus();
    await user.keyboard("{Enter}");

    await act(async () => {});
    expect(m.updateMutateAsync).toHaveBeenCalledTimes(1);
    await act(async () => {
      basicUpdate!.resolve(makeStaff({ id: "1" }));
    });
    await waitPanelClosed();
    expect(m.updateMutateAsync).toHaveBeenCalledTimes(1);
  });
});
