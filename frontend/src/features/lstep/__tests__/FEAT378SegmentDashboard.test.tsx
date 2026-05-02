import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/contexts/auth-context";

import { CPM_SEGMENT_TAGS, LTV_SEGMENT_TAGS, VISIT_SEGMENT_TAGS } from "../constants/segment-tags";
import { isAutoManagedTag } from "@/constants/lstep-auto-tag-prefixes";
import { SegmentDashboard } from "../components/SegmentDashboard";
import { TagSummaryTable } from "../components/TagSummaryTable";
import { TagOwnerListDrawer } from "../components/TagOwnerListDrawer";
import { LstepTagManagementPage } from "../pages/LstepTagManagementPage";
import type { LstepTagSummaryItem } from "../api/get-lstep-tag-summary";

const CLINIC_ID = "clinic-test-1";

const mockAuthCtx = {
  user: null,
  currentClinicId: CLINIC_ID,
  isAuthenticated: true,
  isLoading: false,
  isSwitchingClinic: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => true as boolean,
  refreshPermissions: async () => {},
};

function makeQC() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem("auth_current_clinic:v1");
  vi.restoreAllMocks();
  document.body.removeAttribute("data-scroll-locked");
  document.body.style.removeProperty("pointer-events");
});

// ─────────────────────────────────────────────────────────────
// A: タグ定数定義（純粋インポートテスト）
// ─────────────────────────────────────────────────────────────

describe("FEAT-378 — A: タグ定数定義", () => {
  it("CPM_SEGMENT_TAGS に 5 タグすべてが含まれる", () => {
    const tagNames = CPM_SEGMENT_TAGS.map((t) => t.tagName);
    expect(tagNames).toContain("CPM_01_出会い");
    expect(tagNames).toContain("CPM_02_これから");
    expect(tagNames).toContain("CPM_03_コア");
    expect(tagNames).toContain("CPM_04_スポット");
    expect(tagNames).toContain("CPM_05_ノア");
  });

  it("LTV_SEGMENT_TAGS に LTV_上位20 が含まれる", () => {
    expect(LTV_SEGMENT_TAGS.map((t) => t.tagName)).toContain("LTV_上位20");
  });

  it("VISIT_SEGMENT_TAGS に 4 タグすべてが含まれる", () => {
    const tagNames = VISIT_SEGMENT_TAGS.map((t) => t.tagName);
    expect(tagNames).toContain("VISIT_120日超");
    expect(tagNames).toContain("VISIT_180日超");
    expect(tagNames).toContain("VISIT_220日超");
    expect(tagNames).toContain("VISIT_240日超");
  });
});

// ─────────────────────────────────────────────────────────────
// B: 自動タグ判定（isAutoManagedTag 関数テスト）
// ─────────────────────────────────────────────────────────────

describe("FEAT-378 — B: 自動タグプレフィックス判定", () => {
  it.each([
    ["CPM_01_出会い", "CPM_"],
    ["LTV_上位20", "LTV_"],
    ["VISIT_120日超", "VISIT_"],
    ["PET_犬", "PET_"],
    ["EXCL_配信停止", "EXCL_"],
    ["cpm_コア", "cpm_（旧仕様後方互換）"],
    ["dormant_120", "dormant_（旧仕様後方互換）"],
  ] as const)("%s は自動タグと判定される（%s）", (tagName) => {
    expect(isAutoManagedTag(tagName)).toBe(true);
  });

  it("任意の手動タグ文字列は自動タグと判定されない", () => {
    expect(isAutoManagedTag("カスタムタグ")).toBe(false);
    expect(isAutoManagedTag("manual_tag")).toBe(false);
  });
});

// ─────────────────────────────────────────────────────────────
// C: SegmentDashboard — countMap によるカウント表示
// ─────────────────────────────────────────────────────────────

describe("FEAT-378 — C: SegmentDashboard カウント表示", () => {
  it("summaryTags 空 → CPM セクションが描画され各カードが 0 表示", () => {
    render(
      <MemoryRouter>
        <SegmentDashboard summaryTags={[]} onViewOwners={vi.fn()} />
      </MemoryRouter>
    );
    expect(screen.getByText("CPM ステージ")).toBeInTheDocument();
    expect(screen.getAllByText("0").length).toBeGreaterThan(0);
  });

  it("summaryTags に CPM_01_出会い:5 → 対応カードに 5 が表示される", () => {
    const summaryTags: LstepTagSummaryItem[] = [
      { tag_name: "CPM_01_出会い", owner_count: 5, category: "auto" },
    ];
    render(
      <MemoryRouter>
        <SegmentDashboard summaryTags={summaryTags} onViewOwners={vi.fn()} />
      </MemoryRouter>
    );
    expect(screen.getByText("5")).toBeInTheDocument();
  });

  it("countMap は tag_name 完全一致でカウントを取得する — 別タグは 0 のまま", () => {
    const summaryTags: LstepTagSummaryItem[] = [
      { tag_name: "CPM_03_コア", owner_count: 42, category: "auto" },
    ];
    render(
      <MemoryRouter>
        <SegmentDashboard summaryTags={summaryTags} onViewOwners={vi.fn()} />
      </MemoryRouter>
    );
    expect(screen.getByText("42")).toBeInTheDocument();
    // CPM_01_出会い 等は summaryTags にないので 0 が残る
    expect(screen.getAllByText("0").length).toBeGreaterThan(0);
  });
});

// ─────────────────────────────────────────────────────────────
// D: SegmentDashboard — 対象者一覧コールバック
// ─────────────────────────────────────────────────────────────

describe("FEAT-378 — D: SegmentDashboard onViewOwners コールバック", () => {
  it("「対象者一覧」クリック → onViewOwners(tagName, count) が呼ばれる", async () => {
    const onViewOwners = vi.fn();
    const summaryTags: LstepTagSummaryItem[] = [
      { tag_name: "CPM_01_出会い", owner_count: 7, category: "auto" },
    ];
    render(
      <MemoryRouter>
        <SegmentDashboard summaryTags={summaryTags} onViewOwners={onViewOwners} />
      </MemoryRouter>
    );
    const user = userEvent.setup();
    // SEGMENT_GROUPS の先頭は CPM — 最初の「対象者一覧」ボタンが CPM_01_出会い に対応する
    const buttons = screen.getAllByRole("button", { name: /対象者一覧/ });
    await user.click(buttons[0]);
    expect(onViewOwners).toHaveBeenCalledWith("CPM_01_出会い", 7);
  });
});

// ─────────────────────────────────────────────────────────────
// E: TagSummaryTable — 自動タグの削除ボタン非表示
// ─────────────────────────────────────────────────────────────

describe("FEAT-378 — E: TagSummaryTable 自動タグ削除ボタン非表示", () => {
  const autoTag: LstepTagSummaryItem = {
    tag_name: "CPM_01_出会い",
    owner_count: 3,
    category: "auto",
  };
  const manualTag: LstepTagSummaryItem = {
    tag_name: "手動タグ",
    owner_count: 2,
    category: "manual",
  };

  it("自動タグ + canDelete=true → 削除ボタンが表示されない", () => {
    render(
      <MemoryRouter>
        <TagSummaryTable
          tags={[autoTag]}
          isLoading={false}
          onViewOwners={vi.fn()}
          onBulkRemove={vi.fn()}
          canDelete={true}
        />
      </MemoryRouter>
    );
    expect(screen.queryByRole("button", { name: /削除/ })).not.toBeInTheDocument();
  });

  it("手動タグ + canDelete=true → 削除ボタンが表示される", () => {
    render(
      <MemoryRouter>
        <TagSummaryTable
          tags={[manualTag]}
          isLoading={false}
          onViewOwners={vi.fn()}
          onBulkRemove={vi.fn()}
          canDelete={true}
        />
      </MemoryRouter>
    );
    expect(screen.getByRole("button", { name: /削除/ })).toBeInTheDocument();
  });

  it("手動タグ + canDelete=false → 削除ボタンが表示されない", () => {
    render(
      <MemoryRouter>
        <TagSummaryTable
          tags={[manualTag]}
          isLoading={false}
          onViewOwners={vi.fn()}
          onBulkRemove={vi.fn()}
          canDelete={false}
        />
      </MemoryRouter>
    );
    expect(screen.queryByRole("button", { name: /削除/ })).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// F: TagOwnerListDrawer — CSV ボタン・一括解除ボタン
// ─────────────────────────────────────────────────────────────

describe("FEAT-378 — F: TagOwnerListDrawer CSV ボタン・一括解除", () => {
  type DrawerProps = {
    tagName: string;
    ownerCount: number;
    canDelete?: boolean;
    canExportCsv?: boolean;
    owners?: Array<{
      owner_id: string;
      owner_name: string;
      line_user_id: string | null;
      last_visit_date: string | null;
    }>;
  };

  function renderDrawer(props: DrawerProps) {
    const owners = props.owners ?? [];
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, () =>
        HttpResponse.json({ owners, total: owners.length })
      )
    );
    render(
      <AuthContext.Provider value={mockAuthCtx}>
        <QueryClientProvider client={makeQC()}>
          <MemoryRouter>
            <TagOwnerListDrawer
              open={true}
              onOpenChange={vi.fn()}
              tagName={props.tagName}
              ownerCount={props.ownerCount}
              canDelete={props.canDelete ?? false}
              canExportCsv={props.canExportCsv ?? false}
            />
          </MemoryRouter>
        </QueryClientProvider>
      </AuthContext.Provider>
    );
  }

  it("canExportCsv=true + ownerCount>0 → CSV ボタンが表示され有効", async () => {
    renderDrawer({ tagName: "手動タグ", ownerCount: 5, canExportCsv: true });
    // isLoading が解消されるまで待ってから enabled を確認する
    await waitFor(() => {
      const btn = screen.getByRole("button", { name: /CSV/ });
      expect(btn).toBeInTheDocument();
      expect(btn).not.toBeDisabled();
    });
  });

  it("canExportCsv=false → CSV ボタンが表示されない", async () => {
    renderDrawer({ tagName: "手動タグ", ownerCount: 5, canExportCsv: false });
    await waitFor(() => {
      expect(screen.getByText("5名")).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: /CSV/ })).not.toBeInTheDocument();
  });

  it("ownerCount=0 → CSV ボタンが disabled", async () => {
    renderDrawer({ tagName: "手動タグ", ownerCount: 0, canExportCsv: true });
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /CSV/ })).toBeDisabled();
    });
  });

  it("自動タグ + canDelete=true → 一括解除ボタンが表示されない", async () => {
    renderDrawer({ tagName: "CPM_01_出会い", ownerCount: 3, canDelete: true });
    await waitFor(() => {
      expect(screen.getByText("3名")).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: /一括解除/ })).not.toBeInTheDocument();
  });

  it("手動タグ + canDelete=true + オーナーあり → 一括解除ボタンが表示される", async () => {
    renderDrawer({
      tagName: "手動タグ",
      ownerCount: 1,
      canDelete: true,
      owners: [
        {
          owner_id: "o1",
          owner_name: "テスト飼い主",
          line_user_id: null,
          last_visit_date: null,
        },
      ],
    });
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /一括解除/ })).toBeInTheDocument();
    });
  });
});

// ─────────────────────────────────────────────────────────────
// G: LstepTagManagementPage — SegmentDashboard 組み込み
// ─────────────────────────────────────────────────────────────

describe("FEAT-378 — G: LstepTagManagementPage SegmentDashboard 組み込み", () => {
  function setupPageHandlers() {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/tag-summary`, () =>
        HttpResponse.json({
          tags: [
            { tag_name: "CPM_01_出会い", owner_count: 10, category: "auto" },
          ],
          total_owners_with_lstep: 100,
          as_of: new Date().toISOString(),
        })
      ),
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, () =>
        HttpResponse.json({ owners: [], total: 0 })
      )
    );
  }

  function renderPage() {
    render(
      <AuthContext.Provider value={mockAuthCtx}>
        <QueryClientProvider client={makeQC()}>
          <MemoryRouter>
            <LstepTagManagementPage />
          </MemoryRouter>
        </QueryClientProvider>
      </AuthContext.Provider>
    );
  }

  it("SegmentDashboard が描画され CPM ステージセクションタイトルが表示される", async () => {
    setupPageHandlers();
    renderPage();
    await waitFor(() => {
      expect(screen.getByText("CPM ステージ")).toBeInTheDocument();
    });
  });

  it("タグサマリー取得後 CPM_01_出会い のカウントがカードに反映される", async () => {
    setupPageHandlers();
    renderPage();
    // SegmentDashboard と TagSummaryTable の両方に "10" が表示される（複数一致が正常）
    await waitFor(() => {
      expect(screen.getAllByText("10").length).toBeGreaterThanOrEqual(1);
    });
  });

  it("ページ見出し「Lステップタグ管理」が表示される", async () => {
    setupPageHandlers();
    renderPage();
    await waitFor(() => {
      expect(screen.getByText("Lステップタグ管理")).toBeInTheDocument();
    });
  });
});
