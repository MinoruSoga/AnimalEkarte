import { paths } from "@/config/paths";
import { ICON } from "@/lib/design-tokens";
import type { MenuItem } from "@/types";
import {
  ResourceAccounting,
  ResourceAccountingReports,
  ResourceCashRegisterClose,
  ResourceCheckups,
  ResourceClosingSettings,
  ResourceExaminations,
  ResourceHospitalSettings,
  ResourceHospitalization,
  ResourceIdentityLinks,
  ResourceInventory,
  ResourceLabImport,
  ResourceLstepAnalytics,
  ResourceMasterAnimalSpecies,
  ResourceMasterHospitalization as ResourceMasterHosp,
  ResourceMasterInsurance,
  ResourceMasterMedical,
  ResourceMasterMerchandise,
  ResourceMasterPermission,
  ResourceMasterReservationType,
  ResourceMasterStaff,
  ResourceMasterTrimming as ResourceMasterTrim,
  ResourceMedicalRecords,
  ResourceOwners,
  ResourcePaymentMethod,
  ResourceReception,
  ResourceReservations,
  ResourceShifts,
  ResourceTrimming,
  ResourceVaccinations,
} from "@/types/generated/models";
import {
  Activity,
  BarChart3,
  Bed,
  Briefcase,
  Building2,
  Calendar,
  CalendarDays,
  Clipboard,
  ClipboardCheck,
  ClipboardList,
  ClipboardSignature,
  CreditCard,
  FileText,
  FlaskConical,
  HelpCircle,
  History,
  LayoutDashboard,
  Link2,
  Lock,
  MessageSquare,
  MessageSquareText,
  Package,
  PawPrint,
  Pill,
  Scissors,
  Settings,
  ShieldCheck,
  Smartphone,
  Syringe,
  Tag,
  TestTube,
  Users,
} from "lucide-react";

export interface SidebarMenuSection {
  title: string;
  items: MenuItem[];
}

const chartMastersPurposeKey = "group:chart-masters";
const hospitalizationCagePurposeKey = "group:hospitalization-cage";
const trimmingMastersPurposeKey = "group:trimming-masters";

export const sidebarMenuSections: SidebarMenuSection[] = [
  {
    title: "診療業務",
    items: [
      {
        icon: <LayoutDashboard className={ICON.toolbar} />,
        label: "当日の受付",
        path: paths.home.getHref(),
        resource: ResourceReception,
      },
      {
        icon: <Users className={ICON.toolbar} />,
        label: "飼主・ペット",
        path: paths.owners.getHref(),
        resource: ResourceOwners,
      },
      {
        icon: <Calendar className={ICON.toolbar} />,
        label: "予約管理",
        path: paths.reservations.getHref(),
        resource: ResourceReservations,
      },
      {
        icon: <FileText className={ICON.toolbar} />,
        label: "カルテ",
        path: paths.medicalRecords.getHref(),
        resource: ResourceMedicalRecords,
      },
      {
        icon: <TestTube className={ICON.toolbar} />,
        label: "検査管理",
        path: paths.examinations.getHref(),
        resource: ResourceExaminations,
      },
      {
        icon: <TestTube className={ICON.toolbar} />,
        label: "検査受信",
        path: paths.labDevice.getHref(),
        resource: ResourceLabImport,
      },
      {
        icon: <Scissors className={ICON.toolbar} />,
        label: "トリミング",
        path: paths.trimming.getHref(),
        resource: ResourceTrimming,
      },
      {
        icon: <Syringe className={ICON.toolbar} />,
        label: "予防接種",
        path: paths.vaccinations.getHref(),
        resource: ResourceVaccinations,
      },
      {
        icon: <ClipboardCheck className={ICON.toolbar} />,
        label: "定期健診",
        path: paths.checkups.getHref(),
        resource: ResourceCheckups,
      },
    ],
  },
  {
    title: "運用・管理",
    items: [
      {
        icon: <CreditCard className={ICON.toolbar} />,
        label: "会計管理",
        path: paths.accounting.getHref(),
        resource: ResourceAccounting,
      },
      {
        icon: <CreditCard className={ICON.toolbar} />,
        label: "レジ締め",
        path: paths.accounting.close.getHref(),
        resource: ResourceCashRegisterClose,
      },
      {
        icon: <History className={ICON.toolbar} />,
        label: "締め履歴",
        path: paths.accounting.closeHistory.getHref(),
        resource: ResourceCashRegisterClose,
      },
      {
        icon: <BarChart3 className={ICON.toolbar} />,
        label: "月次集計レポート",
        path: paths.accounting.reports.getHref(),
        resource: ResourceAccountingReports,
      },
      {
        icon: <Bed className={ICON.toolbar} />,
        label: "入院・ホテル",
        path: paths.hospitalization.getHref(),
        resource: ResourceHospitalization,
      },
      {
        icon: <Package className={ICON.toolbar} />,
        label: "在庫管理",
        path: paths.inventory.getHref(),
        resource: ResourceInventory,
      },
      {
        icon: <CalendarDays className={ICON.toolbar} />,
        label: "シフト管理",
        path: paths.shifts.getHref(),
        resource: ResourceShifts,
      },
      {
        icon: <BarChart3 className={ICON.toolbar} />,
        label: "集計・分析",
        path: paths.aggregation.getHref(),
        resource: ResourceOwners,
      },
    ],
  },
  {
    title: "LINE予約管理",
    items: [
      {
        icon: <Smartphone className={ICON.toolbar} />,
        label: "LINE予約管理",
        path: paths.lineReservation.getHref(),
        resource: ResourceReservations,
        subItems: [
          {
            label: "基本設定",
            path: paths.lineReservation.settings.getHref(),
            resource: ResourceReservations,
          },
          {
            label: "ページ編集",
            path: paths.lineReservation.pageEditor.getHref(),
            resource: ResourceReservations,
          },
          {
            label: "予約枠",
            path: paths.lineReservation.slots.getHref(),
            resource: ResourceMasterReservationType,
          },
        ],
      },
    ],
  },
  {
    title: "Lステップ連携",
    items: [
      {
        icon: <MessageSquare className={ICON.toolbar} />,
        label: "Lステップ連携",
        path: paths.lstep.settings.getHref(),
        resource: ResourceHospitalSettings,
        subItems: [
          {
            icon: <MessageSquare className={ICON.toolbar} />,
            label: "連携設定",
            path: paths.lstep.settings.getHref(),
            resource: ResourceHospitalSettings,
          },
          {
            icon: <Tag className={ICON.toolbar} />,
            label: "タグ管理",
            path: paths.lstep.tags.getHref(),
            resource: ResourceLstepAnalytics,
          },
          {
            icon: <ClipboardSignature className={ICON.toolbar} />,
            label: "健診対象者抽出",
            path: paths.lstep.checkupSync.getHref(),
            resource: ResourceHospitalSettings,
          },
          {
            icon: <BarChart3 className={ICON.toolbar} />,
            label: "分析レポート",
            path: paths.lstep.analytics.getHref(),
            resource: ResourceLstepAnalytics,
          },
          {
            icon: <Activity className={ICON.toolbar} />,
            label: "配信監視",
            path: paths.lstep.deliveryMonitor.getHref(),
            resource: ResourceLstepAnalytics,
          },
        ],
      },
    ],
  },
  {
    title: "システム設定",
    items: [
      {
        icon: <Settings className={ICON.toolbar} />,
        label: "マスタ設定",
        path: paths.settings.getHref(),
        subItems: [
          {
            icon: <Building2 className={ICON.toolbar} />,
            label: "医院マスタ",
            path: paths.settings.clinic.getHref(),
            resource: ResourceHospitalSettings,
          },
          {
            icon: <PawPrint className={ICON.toolbar} />,
            label: "動物種類マスタ",
            path: paths.settings.animalSpecies.getHref(),
            resource: ResourceMasterAnimalSpecies,
          },
          {
            icon: <FileText className={ICON.toolbar} />,
            label: "カルテ関連",
            purposeKey: chartMastersPurposeKey,
            resource: ResourceMasterMedical,
            subItems: [
              {
                icon: <ClipboardList className={ICON.toolbar} />,
                label: "診療項目マスタ",
                path: paths.settings.treatmentItems.getHref(),
                resource: ResourceMasterMedical,
              },
              {
                icon: <Clipboard className={ICON.toolbar} />,
                label: "診断マスタ",
                path: paths.settings.diagnosis.getHref(),
                resource: ResourceMasterMedical,
              },
              {
                icon: <ClipboardCheck className={ICON.toolbar} />,
                label: "問診テンプレートマスタ",
                path: paths.settings.inquiryTemplates.getHref(),
                resource: ResourceMasterMedical,
              },
              {
                icon: <MessageSquareText className={ICON.toolbar} />,
                label: "主訴カテゴリマスタ",
                path: paths.settings.interview.chiefComplaint.getHref(),
                resource: ResourceMasterMedical,
              },
              {
                icon: <Pill className={ICON.toolbar} />,
                label: "薬剤マスタ",
                path: paths.settings.medicine.getHref(),
                resource: ResourceMasterMedical,
              },
              {
                icon: <TestTube className={ICON.toolbar} />,
                label: "検査マスタ",
                path: paths.settings.examination.getHref(),
                resource: ResourceMasterMedical,
              },
              {
                icon: <FlaskConical className={ICON.toolbar} />,
                label: "検査機器マスタ",
                path: paths.settings.labDeviceItemMasters.getHref(),
                resource: ResourceLabImport,
              },
            ],
          },
          {
            icon: <Activity className={ICON.toolbar} />,
            label: "予約区分マスタ",
            path: paths.settings.reservationType.getHref(),
            resource: ResourceMasterReservationType,
          },
          {
            icon: <Bed className={ICON.toolbar} />,
            label: "入院・ケージ",
            purposeKey: hospitalizationCagePurposeKey,
            resource: ResourceMasterHosp,
            subItems: [
              {
                icon: <Bed className={ICON.toolbar} />,
                label: "入院マスタ",
                path: paths.settings.hospitalization.getHref(),
                resource: ResourceMasterHosp,
              },
              {
                icon: <Bed className={ICON.toolbar} />,
                label: "ケージマスタ",
                path: paths.settings.cage.getHref(),
                resource: ResourceMasterHosp,
              },
            ],
          },
          {
            icon: <Scissors className={ICON.toolbar} />,
            label: "トリミング",
            purposeKey: trimmingMastersPurposeKey,
            resource: ResourceMasterTrim,
            subItems: [
              {
                icon: <Scissors className={ICON.toolbar} />,
                label: "トリミングマスタ",
                path: paths.settings.trimming.getHref(),
                resource: ResourceMasterTrim,
              },
              {
                icon: <Scissors className={ICON.toolbar} />,
                label: "コース種別マスタ",
                path: paths.settings.trimmingCourseType.getHref(),
                resource: ResourceMasterTrim,
              },
            ],
          },
          {
            icon: <Lock className={ICON.toolbar} />,
            label: "権限グループマスタ",
            path: paths.settings.permissionGroups.getHref(),
            resource: ResourceMasterPermission,
          },
          {
            icon: <Briefcase className={ICON.toolbar} />,
            label: "職種マスタ",
            path: paths.settings.occupations.getHref(),
            resource: ResourceMasterStaff,
          },
          {
            icon: <Users className={ICON.toolbar} />,
            label: "スタッフマスタ",
            path: paths.settings.staff.getHref(),
            resource: ResourceMasterStaff,
          },
          {
            icon: <ShieldCheck className={ICON.toolbar} />,
            label: "保険マスタ",
            path: paths.settings.insurance.getHref(),
            resource: ResourceMasterInsurance,
          },
          {
            icon: <Package className={ICON.toolbar} />,
            label: "物販・フードマスタ",
            path: paths.settings.merchandiseItems.getHref(),
            resource: ResourceMasterMerchandise,
          },
          {
            icon: <CreditCard className={ICON.toolbar} />,
            label: "支払方法マスタ",
            path: paths.settings.paymentMethods.getHref(),
            resource: ResourcePaymentMethod,
          },
          {
            icon: <Tag className={ICON.toolbar} />,
            label: "割引キャンペーンマスタ",
            path: paths.settings.campaigns.getHref(),
            resource: ResourceAccounting,
          },
          {
            icon: <CalendarDays className={ICON.toolbar} />,
            label: "締め時間設定",
            path: paths.settings.closingTime.getHref(),
            resource: ResourceClosingSettings,
          },
          {
            icon: <Calendar className={ICON.toolbar} />,
            label: "シフトテンプレート",
            path: paths.settings.shiftTemplates.getHref(),
            resource: ResourceShifts,
          },
        ],
      },
      {
        icon: <Link2 className={ICON.toolbar} />,
        label: "同一飼主・ペット連携",
        path: paths.identityLinks.getHref(),
        resource: ResourceIdentityLinks,
      },
    ],
  },
  {
    title: "ヘルプ",
    items: [
      {
        icon: <HelpCircle className={ICON.toolbar} />,
        label: "取扱説明書",
        path: paths.manual.getHref(),
      },
    ],
  },
];

const closingPurpose = "日次の売上集計、レジ内の現金の突合、および過去の締め記録の管理。";
const lineReservationPurpose =
  "飼主向け LINE 予約システムの稼働状態、受付ルール、および表示内容を一元管理する。";
const lstepPurpose =
  "Lステップ（LINE公式アカウント拡張ツール）との高度な連携設定、自動配信タグの管理、およびマーケティング施策の効果分析。";

/** Screen-purpose sentences keyed by menu path, or by purposeKey for path-less groups. */
export const sidebarItemDescriptions: Readonly<Record<string, string>> = {
  [paths.home.getHref()]:
    "院内全体の稼働状況を30秒間隔の定期同期で把握・管理するためのカンバンボード型ホーム画面。",
  [paths.owners.getHref()]:
    "登録されている飼主および患者（ペット）の横断検索、および詳細ページへのアクセス。",
  [paths.reservations.getHref()]:
    "診療、トリミング、手術等の全予約をカレンダー形式で統合管理する。",
  [paths.medicalRecords.getHref()]:
    "院内で行われた全カルテ（診療記録）の時系列検索、および診察内容の参照。",
  [paths.examinations.getHref()]:
    "院内および院外（外部委託）で実施される全検査のオーダー状況および、測定結果ステータスの一元管理。",
  [paths.labDevice.getHref()]:
    "医院に設定された検査機器の受信を同一画面で見て、本日診療中のカルテへ紐付ける。",
  [paths.trimming.getHref()]: "院内で実施されるトリミング施術の予約・実績・進捗管理。",
  [paths.vaccinations.getHref()]:
    "院内で行われた全予防接種（混合ワクチン、狂犬病ワクチン、フィラリア・ノミダニ予防等）の実績管理。",
  [paths.checkups.getHref()]:
    "ペットの身体測定や一般健診の履歴を管理し、次回予定日に基づく受診漏れを防止する。",
  [paths.accounting.getHref()]:
    "院内で行われた全診療・サービスの請求レコードを一覧表示・管理し、未精算の把握や過去の会計参照を行う。",
  [paths.accounting.close.getHref()]: closingPurpose,
  [paths.accounting.closeHistory.getHref()]: closingPurpose,
  [paths.accounting.reports.getHref()]:
    "月単位または任意期間の売上実績、支払方法内訳、消費税明細を把握し、経営分析や確定申告・会計処理の基礎データを提供する。",
  [paths.hospitalization.getHref()]:
    "入院・ホテル預かり中の患者（ペット）を管理し、ケージの稼働状況を可視化する。",
  [paths.inventory.getHref()]:
    "院内で使用・販売する医薬品、フード、備品等の在庫状況の監視と、適切な発注タイミングの把握。",
  [paths.shifts.getHref()]:
    "スタッフの勤務シフトをカレンダー形式で視覚的に把握・管理し、診察予約やLINE予約の「空き枠計算」の基礎データとして提供する。",
  [paths.aggregation.getHref()]:
    "医院の全顧客を対象とした売上貢献度（LTV）および来院頻度（CPM）の多角的な分析、および優良顧客・離脱懸念顧客の可視化。",
  [paths.lineReservation.getHref()]: lineReservationPurpose,
  [paths.lineReservation.settings.getHref()]: lineReservationPurpose,
  [paths.lineReservation.pageEditor.getHref()]: lineReservationPurpose,
  [paths.lineReservation.slots.getHref()]: lineReservationPurpose,
  [paths.lstep.settings.getHref()]: lstepPurpose,
  [paths.lstep.tags.getHref()]: lstepPurpose,
  [paths.lstep.checkupSync.getHref()]: lstepPurpose,
  [paths.lstep.analytics.getHref()]: lstepPurpose,
  [paths.lstep.deliveryMonitor.getHref()]:
    "システムが自動生成した Lステップ配信トリガーの実行状況、除外判定、および API 通信の成否を取得時点のログで監視する。",
  [paths.settings.getHref()]:
    "診療、トリミング、入院、スタッフ、権限、および外部連携など、システム全体で共通利用される定義データ（マスタ）を一元的に管理するポータル。",
  [paths.settings.clinic.getHref()]:
    "システムを利用する各医院（本院・分院）の基本情報、連絡先、およびインボイス制度に対応した各種パラメータの一元管理。",
  [paths.settings.animalSpecies.getHref()]: "システムで取り扱う動物種（犬、猫、うさぎ等）の定義。",
  [paths.settings.treatmentItems.getHref()]:
    "診察、処置、手術、検査等の臨床行為の定義、および標準価格・税率の設定。",
  [paths.settings.diagnosis.getHref()]:
    "臨床現場で使用される標準的な診断名、および疾患カテゴリの体系的管理。",
  [paths.settings.inquiryTemplates.getHref()]:
    "カルテの主訴、所見、指導内容の入力を効率化するための定型文テンプレートの管理。",
  [paths.settings.interview.chiefComplaint.getHref()]:
    "カルテ作成時に飼主が訴える「主訴」を分類し、統計や入力補助に使用するための大分類（消化器、皮膚、眼科等）の管理。",
  [paths.settings.medicine.getHref()]:
    "院内で使用・処方される医薬品の定義、単価設定、および在庫システムとの紐付け。",
  [paths.settings.examination.getHref()]:
    "血液検査や生化学検査等、検査プラン（exam_types）の名称・価格の定義、および測定項目（exam_type_fields）と基準値の管理。",
  [paths.settings.reservationType.getHref()]:
    "診察、ワクチン、手術、トリミング等の予約枠（スロット）の定義と、LINE 予約への公開設定。",
  [paths.settings.hospitalization.getHref()]: "入院治療やペットホテルの基本料金体系の管理。",
  [paths.settings.cage.getHref()]: "院内のケージ、入院室、ペットホテルの宿泊スペースの定義。",
  [paths.settings.trimming.getHref()]:
    "トリミングにおける「基本コース」と「追加オプション」の定義。",
  [paths.settings.trimmingCourseType.getHref()]:
    "トリミングコースの大分類（シャンプー、シャンプー+カットなど）を管理し、トリミングコース管理画面の分類基盤を整える。",
  [paths.settings.permissionGroups.getHref()]:
    "スタッフの役割（院長、獣医師、看護師、事務等）に応じたシステム操作権限の体系的定義。",
  [paths.settings.occupations.getHref()]:
    "医院内で勤務するスタッフの職種（獣医師、看護師、トリマー、受付等）の定義。",
  [paths.settings.staff.getHref()]:
    "システムを利用する全スタッフの基本情報、職種、および LINE 予約用の公開プロフィールの統合管理。",
  [paths.settings.insurance.getHref()]:
    "ペット保険（アニコム、アイペット等）の定義、および保険が補償する側の割合（「補償率」）の設定管理。",
  [paths.settings.merchandiseItems.getHref()]:
    "療法食、サプリメント、ヘアケア用品等の販売商品の定義および価格管理。",
  [paths.settings.paymentMethods.getHref()]:
    "決済手段（現金、クレジットカード、電子マネー等）マスタの名称・有効/無効の管理。レジ締めや月次レポートの決済手段別集計の軸となる。",
  [paths.settings.campaigns.getHref()]:
    "会計施策としてのキャンペーンを期間・割引条件付きで定義し、明細計算時に適用可能にする。",
  [paths.settings.closingTime.getHref()]:
    "日次および月次の売上集計に使用する境界時刻、休診日、および特別な集計期間の管理。",
  [paths.settings.shiftTemplates.getHref()]:
    "スタッフの標準的な勤務時間（例：フルタイム、早番、遅番等）をテンプレートとして定義し、シフト管理の入力を大幅に効率化する。",
  [paths.identityLinks.getHref()]:
    "所属医院内の飼主・ペットを手動で identity group にリンク／unlink し、連携先ペットの最小治療履歴を参照する（Phase 1 workbench）。",
  [paths.manual.getHref()]:
    "システムの操作方法、運用ルール、およびトラブルシューティングをスタッフが院内でいつでも閲覧・検索できる統合ヘルプセンター。",
  [paths.settings.labDeviceItemMasters.getHref()]:
    "検査機器の項目コードと検査フィールドの対応を定義する。日常の受信操作はここでは行わない。",
  [chartMastersPurposeKey]:
    "カルテで使う診療項目、診断、問診、主訴、薬剤、検査、検査機器の定義をまとめて開く。",
  [hospitalizationCagePurposeKey]: "入院プランと院内ケージの定義をまとめて開く。",
  [trimmingMastersPurposeKey]:
    "トリミングの基本コース、追加オプション、コース種別の定義をまとめて開く。",
};
