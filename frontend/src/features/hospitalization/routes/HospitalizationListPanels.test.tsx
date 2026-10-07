import type { ReactNode } from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi, afterEach } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import type { MasterItem } from "@/types";
import { HospitalizationListPageView } from "./HospitalizationListPanels";

// ---- dnd-kit mock（HospitalizationBoard.test.tsx と同型） ----
const dndHarness = vi.hoisted(() => ({
  onDragEnd: undefined as ((event: unknown) => void) | undefined,
  active: null as unknown,
  useDraggable: vi.fn(() => ({
    attributes: { "data-drag-attributes": "enabled" },
    listeners: { "data-drag-listener": "enabled" },
    setNodeRef: vi.fn(),
    isDragging: false,
  })),
  useDroppable: vi.fn(() => ({
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

function renderPageView() {
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
    />,
    { wrapper: createWrapper() },
  );
}

beforeEach(() => {
  dndHarness.useDraggable.mockClear();
  dndHarness.useDroppable.mockClear();
  dndHarness.active = null;
  perm.masterCanEdit = true;
  // EMR-306: merge-mode ON で走る占有クエリ（admitted/reserved 全件取得）の既定
  // ハンドラ。個別テストの server.use が上書きする。
  server.use(
    http.get("*/v1/hospitalizations", () =>
      HttpResponse.json({ data: [], total: 0, page: 1, limit: 100 }),
    ),
  );
});

afterEach(() => {
  server.resetHandlers();
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
});
