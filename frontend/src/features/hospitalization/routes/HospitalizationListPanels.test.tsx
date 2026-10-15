import type { ReactNode } from "react";
import type { useDraggable, useDroppable } from "@dnd-kit/core";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi, afterEach } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { CURRENT_CLINIC_STORAGE_KEY, setStoredClinicId } from "@/lib/current-clinic";

import type { MasterItem } from "@/types";
import { HospitalizationListPageView } from "./HospitalizationListPanels";

// ---- dnd-kit mock（HospitalizationBoard.test.tsx と同型） ----
const dndHarness = vi.hoisted(() => ({
  onDragEnd: undefined as ((event: unknown) => void) | undefined,
  active: null as unknown,
  useDraggable: vi.fn((_options: Parameters<typeof useDraggable>[0]) => ({
    attributes: { "data-drag-attributes": "enabled" },
    listeners: { "data-drag-listener": "enabled" },
    setNodeRef: vi.fn(),
    isDragging: false,
  })),
  useDroppable: vi.fn((_options: Parameters<typeof useDroppable>[0]) => ({
    setNodeRef: vi.fn(),
    isOver: false,
  })),
}));

vi.mock("@dnd-kit/core", () => ({
  DndContext: ({
    children,
    onDragEnd,
  }: {
    children: ReactNode;
    onDragEnd: (event: unknown) => void;
  }) => {
    dndHarness.onDragEnd = onDragEnd;
    return children;
  },
  PointerSensor: function PointerSensor() {},
  KeyboardSensor: function KeyboardSensor() {},
  closestCorners: vi.fn(),
  useSensor: vi.fn(() => ({})),
  useSensors: vi.fn(() => []),
  useDraggable: dndHarness.useDraggable,
  useDroppable: dndHarness.useDroppable,
  useDndContext: vi.fn(() => ({ active: dndHarness.active })),
}));

// ---- permission mock: master-hospitalization:edit だけ可変にする ----
const perm = vi.hoisted(() => ({ masterCanEdit: true }));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: (resource: string) => ({
    canView: true,
    canCreate: true,
    canEdit: resource === "master-hospitalization" ? perm.masterCanEdit : true,
    canDelete: true,
  }),
}));

const mergeCages: MasterItem[] = [
  { id: "10", name: "ケージ10", category: "犬舎", price: 0, status: "active" },
  { id: "11", name: "ケージ11", category: "犬舎", price: 0, status: "active" },
  {
    id: "12",
    name: "ケージ12",
    category: "犬舎",
    price: 0,
    status: "active",
    mergedIntoCageId: "10",
  },
  { id: "13", name: "ケージ13", category: "犬舎", price: 0, status: "active" },
];

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  );
}

function renderPageView(
  overrides: Partial<Parameters<typeof HospitalizationListPageView>[0]> = {},
) {
  return render(
    <HospitalizationListPageView
      canCreateHeader={false}
      onCreate={vi.fn()}
      statusFilter="active"
      onStatusTabChange={vi.fn()}
      filterProperties={[]}
      activeFilters={[]}
      onFilterChange={vi.fn()}
      searchTerm=""
      onSearchChange={vi.fn()}
      serverTotal={0}
      activeSorts={[]}
      onSortChange={vi.fn()}
      viewMode="board"
      onViewModeChange={vi.fn()}
      cages={mergeCages}
      boardHospitalizations={[]}
      boardLoading={false}
      boardError={false}
      cagesLoading={false}
      cagesError={false}
      onNavigateToForm={vi.fn()}
      onMovePet={vi.fn()}
      canCreate
      canEdit
      pagination={{
        paginatedData: [],
        totalPages: 1,
        totalCount: 0,
        startIndex: 0,
        endIndex: 0,
        currentPage: 1,
      }}
      onPageChange={vi.fn()}
      {...overrides}
    />,
    { wrapper: createWrapper() },
  );
}

beforeEach(() => {
  expect(setStoredClinicId("1")).toBe(true);
  dndHarness.useDraggable.mockClear();
  dndHarness.useDroppable.mockClear();
  dndHarness.active = null;
  perm.masterCanEdit = true;
  // EMR-306: merge-mode ON で走る占有クエリ（GET /masters/cages/occupancy）の既定
  // ハンドラ。個別テストの server.use が上書きする。
  server.use(
    http.get("*/v1/masters/cages/occupancy", () => HttpResponse.json({ occupied_cage_ids: [] })),
  );
});

describe("HospitalizationListPageView empty board", () => {
  it("患者0件でも空状態とケージ操作を残す", () => {
    renderPageView({ boardHospitalizations: [], cages: mergeCages });

    expect(screen.getByText("入院データが見つかりません")).toBeInTheDocument();
    expect(screen.getByText("ケージ10")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ケージ10.*空き枠に入院/ })).toBeInTheDocument();
  });
});

afterEach(() => {
  server.resetHandlers();
  localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
});

describe("HospitalizationListPanels EMR-306 merge mode toggle", () => {
  it("master-hospitalization:edit 権限があると ケージ連結モード toggle が表示され、ON で merge-cage- drag に切り替わる", async () => {
    const user = userEvent.setup();
    renderPageView();

    const toggle = screen.getByRole("button", { name: "ケージ連結モード" });
    expect(toggle).toHaveAttribute("aria-pressed", "false");

    // OFF: pet-move 用の drag id（occupant なし → empty- 系）が登録される
    expect(dndHarness.useDraggable).toHaveBeenCalledWith(
      expect.objectContaining({ id: expect.stringMatching(/^empty-/) }),
    );
    expect(dndHarness.useDraggable).not.toHaveBeenCalledWith(
      expect.objectContaining({ id: expect.stringMatching(/^merge-cage-/) }),
    );

    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-pressed", "true");

    // ON: merge-cage- id で standalone cage(11) が drag 可能になる
    await waitFor(() =>
      expect(dndHarness.useDraggable).toHaveBeenCalledWith(
        expect.objectContaining({ id: "merge-cage-11", disabled: false }),
      ),
    );
  });

  it("master-hospitalization:edit 権限がないと toggle は表示されず、board は通常モードのまま", () => {
    perm.masterCanEdit = false;
    renderPageView();

    expect(screen.queryByRole("button", { name: "ケージ連結モード" })).not.toBeInTheDocument();
    expect(dndHarness.useDraggable).not.toHaveBeenCalledWith(
      expect.objectContaining({ id: expect.stringMatching(/^merge-cage-/) }),
    );
    // member card の 連結解除 action も権限ゲートされる
    expect(screen.queryByRole("button", { name: /連結を解除/ })).not.toBeInTheDocument();
  });

  it("merge mode で eligible drop → 確認ダイアログ → 連結する で POST /v1/masters/cages/:target/merge（dropped=member）", async () => {
    let capturedId: string | undefined;
    let capturedBody: unknown;
    server.use(
      http.post("*/v1/masters/cages/:id/merge", async ({ request, params }) => {
        capturedId = String(params.id);
        capturedBody = await request.json();
        return HttpResponse.json({ id: Number(params.id) });
      }),
    );

    const user = userEvent.setup();
    renderPageView();
    await user.click(screen.getByRole("button", { name: "ケージ連結モード" }));

    await waitFor(() =>
      expect(dndHarness.useDraggable).toHaveBeenCalledWith(
        expect.objectContaining({ id: "merge-cage-11", disabled: false }),
      ),
    );
    // cage 11 を drag → cage 13（standalone）へ drop
    // （代表 cage への追加連結は BE が単一階層制約で 409 拒否するため target 不適格）
    act(() => {
      dndHarness.onDragEnd?.({
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-13" },
      });
    });

    // 確認ダイアログが開く（dropped=11 を target=13 へ連結）
    expect(await screen.findByText("ケージを連結しますか？")).toBeInTheDocument();
    expect(screen.getByText(/ケージ11.*ケージ13.*連結/)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "連結する" }));

    await waitFor(() => expect(capturedId).toBe("13"));
    expect(capturedBody).toEqual({ member_cage_ids: [11] });
    await waitFor(() =>
      expect(screen.queryByText("ケージを連結しますか？")).not.toBeInTheDocument(),
    );
  });

  it("merge mode で ineligible drop（連結メンバーへの drop）はダイアログを開かない", async () => {
    const user = userEvent.setup();
    renderPageView();
    await user.click(screen.getByRole("button", { name: "ケージ連結モード" }));

    act(() => {
      dndHarness.onDragEnd?.({
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-12" },
      });
    });

    expect(screen.queryByText("ケージを連結しますか？")).not.toBeInTheDocument();
  });

  it("占有クエリが cage 占有を返すと merge-mode で該当ケージは merge source 不可", async () => {
    server.use(
      http.get("*/v1/masters/cages/occupancy", () =>
        HttpResponse.json({ occupied_cage_ids: [11] }),
      ),
    );

    const user = userEvent.setup();
    renderPageView();
    await user.click(screen.getByRole("button", { name: "ケージ連結モード" }));

    // 占有ルックアップで cage 11 は occupied → disabled
    await waitFor(() => {
      const calls = dndHarness.useDraggable.mock.calls.filter((c) => c[0].id === "merge-cage-11");
      expect(calls.at(-1)?.[0].disabled).toBe(true);
    });
    // 空きの 13 は引き続き source 適格
    await waitFor(() =>
      expect(dndHarness.useDraggable).toHaveBeenCalledWith(
        expect.objectContaining({ id: "merge-cage-13", disabled: false }),
      ),
    );
  });

  it("占有情報の取得中は連結を拒否し、全件取得後に許可する", async () => {
    let release = () => {};
    const ready = new Promise<void>((resolve) => {
      release = resolve;
    });
    server.use(
      http.get("*/v1/masters/cages/occupancy", async () => {
        await ready;
        return HttpResponse.json({ occupied_cage_ids: [] });
      }),
    );
    const user = userEvent.setup();
    renderPageView();
    await user.click(screen.getByRole("button", { name: "ケージ連結モード" }));
    expect(screen.getByRole("status")).toHaveTextContent("連結できません");
    act(() => {
      dndHarness.onDragEnd?.({
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-13" },
      });
    });
    expect(screen.queryByText("ケージを連結しますか？")).not.toBeInTheDocument();
    release();
    await waitFor(() =>
      expect(dndHarness.useDraggable).toHaveBeenCalledWith(
        expect.objectContaining({ id: "merge-cage-11", disabled: false }),
      ),
    );
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it.each(["malformed", "error"])(
    "占有クエリ %s 時は表示中の患者がゼロでも連結を許可しない",
    async (response) => {
      let requested = false;
      server.use(
        http.get("*/v1/masters/cages/occupancy", () => {
          requested = true;
          return response === "error"
            ? new HttpResponse(null, { status: 500 })
            : HttpResponse.json({});
        }),
      );
      const user = userEvent.setup();
      renderPageView({ boardHospitalizations: [] });
      await user.click(screen.getByRole("button", { name: "ケージ連結モード" }));
      await waitFor(() => expect(requested).toBe(true));
      expect(screen.getByRole("status")).toHaveTextContent("連結できません");
      for (const id of ["11", "13"]) {
        const calls = dndHarness.useDraggable.mock.calls.filter(
          (c) => c[0].id === `merge-cage-${id}`,
        );
        expect(calls.at(-1)?.[0].disabled).toBe(true);
      }
      act(() => {
        dndHarness.onDragEnd?.({
          active: { data: { current: { mergeCageId: "11" } } },
          over: { id: "cage-13" },
        });
      });
      expect(screen.queryByText("ケージを連結しますか？")).not.toBeInTheDocument();
    },
  );

  it("member card の 連結解除 → 確認ダイアログ → 解除 で POST /v1/masters/cages/:rep/unmerge", async () => {
    let capturedId: string | undefined;
    server.use(
      http.post("*/v1/masters/cages/:id/unmerge", ({ params }) => {
        capturedId = String(params.id);
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const user = userEvent.setup();
    renderPageView();

    // 連結解除 action は権限があるので通常モードでも見える（代表 id 解決: 12→10）
    await user.click(screen.getByRole("button", { name: "ケージ10への連結を解除" }));

    expect(await screen.findByText("ケージの連結を解除しますか？")).toBeInTheDocument();
    expect(screen.getByText(/ケージ10.*連結を解除/)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "解除" }));

    await waitFor(() => expect(capturedId).toBe("10"));
    await waitFor(() =>
      expect(screen.queryByText("ケージの連結を解除しますか？")).not.toBeInTheDocument(),
    );
  });

  it.each([
    { label: "clinic 未選択", storedClinicId: null },
    { label: "clinic が空白文字", storedClinicId: "   " },
  ])(
    "$label では連結モード ON でも占有クエリも merge POST も発行せず連結できない",
    async ({ storedClinicId }) => {
      if (storedClinicId === null) {
        localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
      } else {
        localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, storedClinicId);
      }
      let occupancyRequests = 0;
      let mergeRequests = 0;
      server.use(
        http.get("*/v1/masters/cages/occupancy", () => {
          occupancyRequests += 1;
          return HttpResponse.json({ occupied_cage_ids: [] });
        }),
        http.post("*/v1/masters/cages/:id/merge", () => {
          mergeRequests += 1;
          return HttpResponse.json({ id: 13 });
        }),
      );

      const user = userEvent.setup();
      renderPageView({ boardHospitalizations: [] });
      const toggle = screen.getByRole("button", { name: "ケージ連結モード" });
      await user.click(toggle);
      expect(toggle).toHaveAttribute("aria-pressed", "true");
      expect(screen.getByRole("status")).toHaveTextContent("連結できません");

      act(() => {
        dndHarness.onDragEnd?.({
          active: { data: { current: { mergeCageId: "11" } } },
          over: { id: "cage-13" },
        });
      });

      expect(screen.queryByText("ケージを連結しますか？")).not.toBeInTheDocument();
      expect(occupancyRequests).toBe(0);
      expect(mergeRequests).toBe(0);
    },
  );
});

// EMR-412 R4: board の表示可否は board 全件クエリと cage master クエリの両方に依存する。
// cage 未確定（loading/error）では「ケージが登録されていません」と言い切れない。
describe("HospitalizationListPageView cage master loading/error (EMR-412 R4)", () => {
  it("cages 取得中（board 完了）は部分描画せず loading を出し、ケージ未登録を出さない", () => {
    renderPageView({
      boardHospitalizations: [],
      cages: [],
      boardLoading: false,
      cagesLoading: true,
      cagesError: false,
    });

    expect(screen.queryByText("ケージが登録されていません")).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toBeInTheDocument();
  });

  it("cages 取得失敗（board 完了）は error 表示に倒し、ケージ未登録を出さない", () => {
    renderPageView({
      boardHospitalizations: [],
      cages: [],
      boardLoading: false,
      cagesLoading: false,
      cagesError: true,
    });

    expect(screen.getByRole("alert")).toHaveTextContent("データの取得に失敗しました");
    expect(screen.queryByText("ケージが登録されていません")).not.toBeInTheDocument();
  });

  it("cages 取得成功で0件ならケージ未登録メッセージを出す", () => {
    renderPageView({
      boardHospitalizations: [],
      cages: [],
      boardLoading: false,
      cagesLoading: false,
      cagesError: false,
    });

    expect(screen.getByText("ケージが登録されていません")).toBeInTheDocument();
  });

  it("cages 取得成功（件数あり）で入院0件なら board を描画する", () => {
    renderPageView({
      boardHospitalizations: [],
      cages: mergeCages,
      boardLoading: false,
      cagesLoading: false,
      cagesError: false,
    });

    expect(screen.getByText("ケージ10")).toBeInTheDocument();
    expect(screen.queryByText("ケージが登録されていません")).not.toBeInTheDocument();
  });

  it("list 表示は cages の loading/error に影響されない", () => {
    renderPageView({
      viewMode: "list",
      cages: [],
      cagesLoading: true,
      cagesError: true,
    });

    expect(screen.getByText("入院No")).toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
