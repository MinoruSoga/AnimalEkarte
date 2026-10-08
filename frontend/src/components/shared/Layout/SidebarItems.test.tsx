import { act, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { paths } from "@/config/paths";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import type { MenuItem } from "@/types";

import { sidebarItemDescriptions, sidebarMenuSections } from "./SidebarMenu";
import { SidebarItemWithPermission } from "./SidebarItems";

// String constants avoid generated/models imports in this layout test (TASK-444-S1 allowlist).
const ResourceHospitalSettings = "hospital-settings" as NonNullable<MenuItem["resource"]>;
const ResourceLstepAnalytics = "lstep-analytics" as NonNullable<MenuItem["resource"]>;

const defaultHasPermission = vi.fn(() => true);

function authValue(
  hasPermission: AuthContextValue["hasPermission"] = defaultHasPermission,
): AuthContextValue {
  return {
    user: null,
    currentClinicId: "clinic-test-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission,
    refreshPermissions: async () => {},
  };
}

function renderSidebarItem(
  item: MenuItem,
  collapsed = false,
  initialEntry = "/",
  hasPermission: AuthContextValue["hasPermission"] = defaultHasPermission,
) {
  return render(
    <AuthContext.Provider value={authValue(hasPermission)}>
      <MemoryRouter initialEntries={[initialEntry]}>
        <SidebarItemWithPermission item={item} collapsed={collapsed} />
      </MemoryRouter>
    </AuthContext.Provider>,
  );
}

describe("SidebarItemWithPermission", () => {
  it("サブメニュー付き項目で button の入れ子を作らない", () => {
    const { container } = renderSidebarItem({
      label: "LINE予約管理",
      path: "/line-reservation",
      subItems: [{ label: "基本設定", path: "/line-reservation/settings" }],
    });

    expect(container.querySelector("button button")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "LINE予約管理" })).toHaveAttribute(
      "aria-expanded",
      "false",
    );
    expect(screen.getByRole("button", { name: "LINE予約管理を展開" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });

  it("折りたたみ時も親メニューbuttonにaccessible nameと44px以上の操作領域を保つ", () => {
    renderSidebarItem(
      {
        label: "マスタ設定",
        path: "/settings/master",
        subItems: [{ label: "スタッフ", path: "/settings/master/staffs" }],
      },
      true,
      "/settings/master/staffs",
    );

    const parentButton = screen.getByRole("button", { name: "マスタ設定" });
    expect(parentButton).toHaveClass("min-h-11", "min-w-11");
    expect(parentButton).not.toHaveAttribute("aria-expanded");
  });

  it("展開ボタンからサブメニューを開閉できる", () => {
    renderSidebarItem({
      label: "LINE予約管理",
      path: "/line-reservation",
      subItems: [{ label: "基本設定", path: "/line-reservation/settings" }],
    });

    fireEvent.click(screen.getByRole("button", { name: "LINE予約管理を展開" }));

    expect(screen.getByRole("button", { name: "LINE予約管理" })).toHaveAttribute(
      "aria-expanded",
      "true",
    );
    expect(screen.getByRole("button", { name: "LINE予約管理を折りたたむ" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
    expect(screen.getByRole("link", { name: "基本設定" })).toBeInTheDocument();
  });

  it("親が HospitalSettings でも Analytics 子が view 可ならグループを表示する (R-06/R-07)", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) => resource === ResourceLstepAnalytics && action === "view",
    );
    renderSidebarItem(
      {
        label: "Lステップ連携",
        path: "/settings/integrations/lstep",
        resource: ResourceHospitalSettings,
        subItems: [
          {
            label: "連携設定",
            path: "/settings/integrations/lstep",
            resource: ResourceHospitalSettings,
          },
          {
            label: "配信監視",
            path: "/lstep/delivery-monitor",
            resource: ResourceLstepAnalytics,
          },
        ],
      },
      false,
      "/",
      hasPermission,
    );

    fireEvent.click(screen.getByRole("button", { name: "Lステップ連携を展開" }));
    expect(screen.getByRole("link", { name: "配信監視" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "連携設定" })).not.toBeInTheDocument();
  });

  it("Analytics-only では親 path へ navigate しない (expand-only shell)", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) => resource === ResourceLstepAnalytics && action === "view",
    );
    renderSidebarItem(
      {
        label: "Lステップ連携",
        path: "/settings/integrations/lstep",
        resource: ResourceHospitalSettings,
        subItems: [
          {
            label: "配信監視",
            path: "/lstep/delivery-monitor",
            resource: ResourceLstepAnalytics,
          },
        ],
      },
      false,
      "/",
      hasPermission,
    );

    // Parent click expands only; unauthorized settings path must not become current route label via Link.
    fireEvent.click(screen.getByRole("button", { name: "Lステップ連携" }));
    expect(screen.getByRole("link", { name: "配信監視" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Lステップ連携" })).not.toBeInTheDocument();
  });

  it("neither 権限では Lステップ group を出さない", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(() => false);
    renderSidebarItem(
      {
        label: "Lステップ連携",
        path: "/settings/integrations/lstep",
        resource: ResourceHospitalSettings,
        subItems: [
          {
            label: "配信監視",
            path: "/lstep/delivery-monitor",
            resource: ResourceLstepAnalytics,
          },
        ],
      },
      false,
      "/",
      hasPermission,
    );
    expect(screen.queryByRole("button", { name: "Lステップ連携" })).not.toBeInTheDocument();
  });
});

const CLOSING_PURPOSE = "日次の売上集計、レジ内の現金の突合、および過去の締め記録の管理。";
const LINE_PURPOSE =
  "飼主向け LINE 予約システムの稼働状態、受付ルール、および表示内容を一元管理する。";
const LSTEP_PURPOSE =
  "Lステップ（LINE公式アカウント拡張ツール）との高度な連携設定、自動配信タグの管理、およびマーケティング施策の効果分析。";

const DESCRIBED_MENU: ReadonlyArray<{ label: string; path: string; purpose: string }> = [
  {
    label: "当日の受付",
    path: paths.home.getHref(),
    purpose:
      "院内全体の稼働状況を30秒間隔の定期同期で把握・管理するためのカンバンボード型ホーム画面。",
  },
  {
    label: "飼主・ペット",
    path: paths.owners.getHref(),
    purpose: "登録されている飼主および患者（ペット）の横断検索、および詳細ページへのアクセス。",
  },
  {
    label: "予約管理",
    path: paths.reservations.getHref(),
    purpose: "診療、トリミング、手術等の全予約をカレンダー形式で統合管理する。",
  },
  {
    label: "カルテ",
    path: paths.medicalRecords.getHref(),
    purpose: "院内で行われた全カルテ（診療記録）の時系列検索、および診察内容の参照。",
  },
  {
    label: "検査管理",
    path: paths.examinations.getHref(),
    purpose:
      "院内および院外（外部委託）で実施される全検査のオーダー状況および、測定結果ステータスの一元管理。",
  },
  {
    label: "トリミング",
    path: paths.trimming.getHref(),
    purpose: "院内で実施されるトリミング施術の予約・実績・進捗管理。",
  },
  {
    label: "予防接種",
    path: paths.vaccinations.getHref(),
    purpose:
      "院内で行われた全予防接種（混合ワクチン、狂犬病ワクチン、フィラリア・ノミダニ予防等）の実績管理。",
  },
  {
    label: "定期健診",
    path: paths.checkups.getHref(),
    purpose: "ペットの身体測定や一般健診の履歴を管理し、次回予定日に基づく受診漏れを防止する。",
  },
  {
    label: "会計管理",
    path: paths.accounting.getHref(),
    purpose:
      "院内で行われた全診療・サービスの請求レコードを一覧表示・管理し、未精算の把握や過去の会計参照を行う。",
  },
  { label: "レジ締め", path: paths.accounting.close.getHref(), purpose: CLOSING_PURPOSE },
  { label: "締め履歴", path: paths.accounting.closeHistory.getHref(), purpose: CLOSING_PURPOSE },
  {
    label: "月次集計レポート",
    path: paths.accounting.reports.getHref(),
    purpose:
      "月単位または任意期間の売上実績、支払方法内訳、消費税明細を把握し、経営分析や確定申告・会計処理の基礎データを提供する。",
  },
  {
    label: "入院・ホテル",
    path: paths.hospitalization.getHref(),
    purpose: "入院・ホテル預かり中の患者（ペット）を管理し、ケージの稼働状況を可視化する。",
  },
  {
    label: "在庫管理",
    path: paths.inventory.getHref(),
    purpose:
      "院内で使用・販売する医薬品、フード、備品等の在庫状況の監視と、適切な発注タイミングの把握。",
  },
  {
    label: "シフト管理",
    path: paths.shifts.getHref(),
    purpose:
      "スタッフの勤務シフトをカレンダー形式で視覚的に把握・管理し、診察予約やLINE予約の「空き枠計算」の基礎データとして提供する。",
  },
  {
    label: "集計・分析",
    path: paths.aggregation.getHref(),
    purpose:
      "医院の全顧客を対象とした売上貢献度（LTV）および来院頻度（CPM）の多角的な分析、および優良顧客・離脱懸念顧客の可視化。",
  },
  { label: "LINE予約管理", path: paths.lineReservation.getHref(), purpose: LINE_PURPOSE },
  { label: "基本設定", path: paths.lineReservation.settings.getHref(), purpose: LINE_PURPOSE },
  { label: "ページ編集", path: paths.lineReservation.pageEditor.getHref(), purpose: LINE_PURPOSE },
  { label: "予約枠", path: paths.lineReservation.slots.getHref(), purpose: LINE_PURPOSE },
  { label: "Lステップ連携", path: paths.lstep.settings.getHref(), purpose: LSTEP_PURPOSE },
  { label: "連携設定", path: paths.lstep.settings.getHref(), purpose: LSTEP_PURPOSE },
  { label: "タグ管理", path: paths.lstep.tags.getHref(), purpose: LSTEP_PURPOSE },
  { label: "健診対象者抽出", path: paths.lstep.checkupSync.getHref(), purpose: LSTEP_PURPOSE },
  { label: "分析レポート", path: paths.lstep.analytics.getHref(), purpose: LSTEP_PURPOSE },
  {
    label: "配信監視",
    path: paths.lstep.deliveryMonitor.getHref(),
    purpose:
      "システムが自動生成した Lステップ配信トリガーの実行状況、除外判定、および API 通信の成否を取得時点のログで監視する。",
  },
  {
    label: "マスタ設定",
    path: paths.settings.getHref(),
    purpose:
      "診療、トリミング、入院、スタッフ、権限、および外部連携など、システム全体で共通利用される定義データ（マスタ）を一元的に管理するポータル。",
  },
  {
    label: "医院",
    path: paths.settings.clinic.getHref(),
    purpose:
      "システムを利用する各医院（本院・分院）の基本情報、連絡先、およびインボイス制度に対応した各種パラメータの一元管理。",
  },
  {
    label: "動物種類",
    path: paths.settings.animalSpecies.getHref(),
    purpose: "システムで取り扱う動物種（犬、猫、うさぎ等）の定義。",
  },
  {
    label: "診療項目",
    path: paths.settings.treatmentItems.getHref(),
    purpose: "診察、処置、手術、検査等の臨床行為の定義、および標準価格・税率の設定。",
  },
  {
    label: "診断マスタ",
    path: paths.settings.diagnosis.getHref(),
    purpose: "臨床現場で使用される標準的な診断名、および疾患カテゴリの体系的管理。",
  },
  {
    label: "問診設定",
    path: paths.settings.inquiryTemplates.getHref(),
    purpose: "カルテの主訴、所見、指導内容の入力を効率化するための定型文テンプレートの管理。",
  },
  {
    label: "主訴カテゴリ",
    path: paths.settings.interview.chiefComplaint.getHref(),
    purpose:
      "カルテ作成時に飼主が訴える「主訴」を分類し、統計や入力補助に使用するための大分類（消化器、皮膚、眼科等）の管理。",
  },
  {
    label: "薬剤マスタ",
    path: paths.settings.medicine.getHref(),
    purpose: "院内で使用・処方される医薬品の定義、単価設定、および在庫システムとの紐付け。",
  },
  {
    label: "検査マスタ",
    path: paths.settings.examination.getHref(),
    purpose:
      "血液検査や生化学検査等、検査プラン（exam_types）の名称・価格の定義、および測定項目（exam_type_fields）と基準値の管理。",
  },
  {
    label: "予約区分",
    path: paths.settings.reservationType.getHref(),
    purpose:
      "診察、ワクチン、手術、トリミング等の予約枠（スロット）の定義と、LINE 予約への公開設定。",
  },
  {
    label: "入院マスタ",
    path: paths.settings.hospitalization.getHref(),
    purpose: "入院治療やペットホテルの基本料金体系の管理。",
  },
  {
    label: "ケージマスタ",
    path: paths.settings.cage.getHref(),
    purpose: "院内のケージ、入院室、ペットホテルの宿泊スペースの定義。",
  },
  {
    label: "トリミングマスタ",
    path: paths.settings.trimming.getHref(),
    purpose: "トリミングにおける「基本コース」と「追加オプション」の定義。",
  },
  {
    label: "コース種別マスタ",
    path: paths.settings.trimmingCourseType.getHref(),
    purpose:
      "トリミングコースの大分類（シャンプー、シャンプー+カットなど）を管理し、トリミングコース管理画面の分類基盤を整える。",
  },
  {
    label: "権限グループ",
    path: paths.settings.permissionGroups.getHref(),
    purpose: "スタッフの役割（院長、獣医師、看護師、事務等）に応じたシステム操作権限の体系的定義。",
  },
  {
    label: "職種マスタ",
    path: paths.settings.occupations.getHref(),
    purpose: "医院内で勤務するスタッフの職種（獣医師、看護師、トリマー、受付等）の定義。",
  },
  {
    label: "スタッフ管理",
    path: paths.settings.staff.getHref(),
    purpose:
      "システムを利用する全スタッフの基本情報、職種、および LINE 予約用の公開プロフィールの統合管理。",
  },
  {
    label: "保険マスタ",
    path: paths.settings.insurance.getHref(),
    purpose:
      "ペット保険（アニコム、アイペット等）の定義、および保険が補償する側の割合（「補償率」）の設定管理。",
  },
  {
    label: "物販・フード",
    path: paths.settings.merchandiseItems.getHref(),
    purpose: "療法食、サプリメント、ヘアケア用品等の販売商品の定義および価格管理。",
  },
  {
    label: "支払方法",
    path: paths.settings.paymentMethods.getHref(),
    purpose:
      "決済手段（現金、クレジットカード、電子マネー等）マスタの名称・有効/無効の管理。レジ締めや月次レポートの決済手段別集計の軸となる。",
  },
  {
    label: "割引キャンペーン",
    path: paths.settings.campaigns.getHref(),
    purpose:
      "会計施策としてのキャンペーンを期間・割引条件付きで定義し、明細計算時に適用可能にする。",
  },
  {
    label: "締め時間設定",
    path: paths.settings.closingTime.getHref(),
    purpose: "日次および月次の売上集計に使用する境界時刻、休診日、および特別な集計期間の管理。",
  },
  {
    label: "シフトテンプレート",
    path: paths.settings.shiftTemplates.getHref(),
    purpose:
      "スタッフの標準的な勤務時間（例：フルタイム、早番、遅番等）をテンプレートとして定義し、シフト管理の入力を大幅に効率化する。",
  },
  {
    label: "同一飼主・ペット連携",
    path: paths.identityLinks.getHref(),
    purpose:
      "所属医院内の飼主・ペットを手動で identity group にリンク／unlink し、連携先ペットの最小治療履歴を参照する（Phase 1 workbench）。",
  },
  {
    label: "取扱説明書",
    path: paths.manual.getHref(),
    purpose:
      "システムの操作方法、運用ルール、およびトラブルシューティングをスタッフが院内でいつでも閲覧・検索できる統合ヘルプセンター。",
  },
];

const UNDESCRIBED_MENU: ReadonlyArray<{ label: string; path?: string }> = [
  { label: "検査受信", path: paths.labDevice.getHref() },
  { label: "検査機器マスタ", path: paths.settings.labDeviceItemMasters.getHref() },
  { label: "カルテ関連" },
  { label: "入院・ケージ" },
  { label: "トリミング" },
];

function flattenMenu(items: MenuItem[]): MenuItem[] {
  return items.flatMap((item) => [item, ...(item.subItems ? flattenMenu(item.subItems) : [])]);
}

const flatMenu = sidebarMenuSections.flatMap((section) => flattenMenu(section.items));

function requireMenuItem(label: string, path: string | undefined): MenuItem {
  const found = flatMenu.find((item) => item.label === label && item.path === path);
  if (!found) {
    throw new Error(`menu item not found: ${label}`);
  }
  return found;
}

function focusVisible(element: HTMLElement) {
  act(() => {
    element.ownerDocument.defaultView?.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Tab", bubbles: true }),
    );
    element.focus();
  });
}

function expectPurposeExposed(
  element: HTMLElement,
  purpose: string,
  exposure: "hover" | "focus-visible",
) {
  if (exposure === "hover") {
    fireEvent.mouseEnter(element);
  } else {
    focusVisible(element);
  }
  const tooltip = screen.getByRole("tooltip");
  expect(tooltip).toHaveTextContent(purpose);
  expect(tooltip).toHaveAttribute("data-exposure", expect.stringContaining(exposure));
  expect(element).not.toHaveAttribute("title", purpose);
}

describe("sidebar item purpose descriptions", () => {
  it("展開中の項目はホバーと focus-visible で役割説明を出す", () => {
    const item = requireMenuItem("当日の受付", paths.home.getHref());
    renderSidebarItem(item, false);

    const link = screen.getByRole("link", { name: "当日の受付" });
    expect(link).not.toHaveAttribute("title");
    expectPurposeExposed(link, DESCRIBED_MENU[0].purpose, "hover");
    expectPurposeExposed(link, DESCRIBED_MENU[0].purpose, "focus-visible");
  });

  it("折りたたみ中もラベルが隠れてもホバーと focus-visible で役割説明を出す", () => {
    const item = requireMenuItem("当日の受付", paths.home.getHref());
    renderSidebarItem(item, true);

    expect(screen.queryByText("当日の受付")).not.toBeInTheDocument();
    const link = screen.getByRole("link", { name: "当日の受付" });
    expect(link).toHaveAttribute("title", "当日の受付");
    expectPurposeExposed(link, DESCRIBED_MENU[0].purpose, "hover");
    expectPurposeExposed(link, DESCRIBED_MENU[0].purpose, "focus-visible");
  });

  it("展開した子の focus-visible とホバーではその子の役割説明だけを出す", () => {
    const item = requireMenuItem("Lステップ連携", paths.lstep.settings.getHref());
    renderSidebarItem(item);
    fireEvent.click(screen.getByRole("button", { name: "Lステップ連携を展開" }));

    const child = screen.getByRole("link", { name: "配信監視" });
    focusVisible(child);

    const focused = screen.getAllByRole("tooltip");
    expect(focused).toHaveLength(1);
    expect(focused[0]).toHaveTextContent("配信トリガー");
    expect(focused[0]).not.toHaveTextContent(LSTEP_PURPOSE);

    fireEvent.blur(child);
    fireEvent.mouseEnter(child);

    const hovered = screen.getAllByRole("tooltip");
    expect(hovered).toHaveLength(1);
    expect(hovered[0]).toHaveTextContent("配信トリガー");
    expect(hovered[0]).not.toHaveTextContent(LSTEP_PURPOSE);
    expect(hovered[0]).not.toHaveTextContent("**");
  });

  it("役割説明は Markdown の強調と識別子引用を含まない", () => {
    for (const text of Object.values(sidebarItemDescriptions)) {
      expect(text).not.toMatch(/\*\*|`/);
    }
  });

  it("権限で遷移先を外した親でも元の役割説明を出す", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) => resource === ResourceLstepAnalytics && action === "view",
    );
    renderSidebarItem(
      {
        label: "Lステップ連携",
        path: paths.lstep.settings.getHref(),
        resource: ResourceHospitalSettings,
        subItems: [
          {
            label: "配信監視",
            path: paths.lstep.deliveryMonitor.getHref(),
            resource: ResourceLstepAnalytics,
          },
        ],
      },
      false,
      "/",
      hasPermission,
    );

    focusVisible(screen.getByRole("button", { name: "Lステップ連携" }));
    expect(screen.getByRole("tooltip")).toHaveTextContent(LSTEP_PURPOSE);
  });

  it("折りたたみ時の説明はアンカーの右に残った幅へ収める", () => {
    const original = window.innerWidth;
    Object.defineProperty(window, "innerWidth", { configurable: true, value: 375 });
    try {
      const item = requireMenuItem("当日の受付", paths.home.getHref());
      renderSidebarItem(item, true);
      const link = screen.getByRole("link", { name: "当日の受付" });
      const row = link.closest("[data-sidebar-row]");
      if (!(row instanceof HTMLElement)) throw new Error("sidebar row missing");
      row.getBoundingClientRect = () =>
        ({
          x: 0,
          y: 0,
          top: 10,
          left: 0,
          right: 64,
          bottom: 58,
          width: 64,
          height: 48,
          toJSON: () => ({}),
        }) as DOMRect;
      focusVisible(link);
      const tooltip = screen.getByRole("tooltip");
      expect(tooltip.style.left).toBe("72px");
      expect(tooltip.style.maxWidth).toBe("min(28rem, 295px)");
    } finally {
      Object.defineProperty(window, "innerWidth", { configurable: true, value: original });
    }
  });

  it("画面の目的がある項目だけその文を受け取り、無い5件は説明を持たない", () => {
    for (const row of DESCRIBED_MENU) {
      const item = requireMenuItem(row.label, row.path);
      expect(item.path && sidebarItemDescriptions[item.path]).toBe(row.purpose);
      expect(row.purpose).not.toMatch(/\/manual|help-article|記事/);
    }

    for (const row of UNDESCRIBED_MENU) {
      const item = requireMenuItem(row.label, row.path);
      expect(item.path ? sidebarItemDescriptions[item.path] : undefined).toBeUndefined();
    }

    const describedKeys = new Set(DESCRIBED_MENU.map((row) => `${row.label}\t${row.path}`));
    const undescribedKeys = new Set(
      UNDESCRIBED_MENU.map((row) => `${row.label}\t${row.path ?? ""}`),
    );
    expect(flatMenu.map((item) => `${item.label}\t${item.path ?? ""}`).sort()).toEqual(
      [...describedKeys, ...undescribedKeys].sort(),
    );
  });
});
