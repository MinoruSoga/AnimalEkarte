// React/Framework
import { useState, useCallback } from "react";
import { useNavigate } from "react-router";

// External
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import ChevronDown from "lucide-react/dist/esm/icons/chevron-down";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import Plus from "lucide-react/dist/esm/icons/plus";
import Search from "lucide-react/dist/esm/icons/search";

// Internal
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageLayout } from "@/components/shared/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { C, PALETTE } from "@/lib/design-tokens";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface TreatmentItem {
  id: string;
  type: string;
  name: string;
  price: number | null;
  status: "active" | "inactive";
  children?: TreatmentItem[];
}

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const TABS = [
  { value: "consultation", label: "診察" },
  { value: "examination", label: "検査" },
  { value: "procedure", label: "処置" },
  { value: "vaccine", label: "予防接種" },
  { value: "checkup", label: "定期健診" },
] as const;

const MOCK_CONSULTATION_ITEMS: TreatmentItem[] = [
  { id: "1", type: "診察", name: "再診料(再診)", price: 800, status: "active" },
  { id: "2", type: "診察", name: "初診料", price: 1500, status: "active" },
  { id: "3", type: "診察", name: "時間外診察料", price: 2000, status: "active" },
  {
    id: "4",
    type: "診察",
    name: "再診A",
    price: null,
    status: "active",
    children: [
      { id: "4-1", type: "診察", name: "再診A：500", price: 1000, status: "active" },
      { id: "4-2", type: "診察", name: "再診A：100", price: 200, status: "active" },
    ],
  },
];

// ─────────────────────────────────────────────────
// StatusDot
// ─────────────────────────────────────────────────

function StatusDot({ status }: { status: "active" | "inactive" }) {
  return (
    <div className="flex items-center gap-1.5">
      <span
        className="size-[8.8px] rounded-full shrink-0"
        style={{
          background: status === "active" ? PALETTE.brand : PALETTE.borderMedium,
        }}
      />
      <span className={`text-xs ${C.text50}`}>
        {status === "active" ? "有効" : "無効"}
      </span>
    </div>
  );
}

// ─────────────────────────────────────────────────
// TreatmentRow
// ─────────────────────────────────────────────────

interface TreatmentRowProps {
  item: TreatmentItem;
  depth?: number;
  isExpanded?: boolean;
  onToggle?: (id: string) => void;
}

function TreatmentRow({
  item,
  depth = 0,
  isExpanded,
  onToggle,
}: TreatmentRowProps) {
  const hasChildren = !!item.children?.length;
  const indentPx = depth * 24;

  return (
    <TableRow
      className={`border-b ${C.borderDivider} h-[45px] hover:bg-[rgba(55,53,47,0.02)] group`}
    >
      {/* 種別 */}
      <TableCell className={`w-[120px] py-0 text-sm ${C.text70} shrink-0`}>
        {item.type}
      </TableCell>

      {/* 名称 */}
      <TableCell className="py-0">
        <div
          className="flex items-center gap-1"
          style={{ paddingLeft: `${indentPx}px` }}
        >
          {hasChildren ? (
            <button
              type="button"
              onClick={() => onToggle?.(item.id)}
              className={`size-[22px] flex items-center justify-center rounded-[3px] ${C.text40} ${C.hoverBgMedium} transition-colors shrink-0`}
            >
              {isExpanded ? (
                <ChevronDown className="size-3.5" />
              ) : (
                <ChevronRight className="size-3.5" />
              )}
            </button>
          ) : (
            <span className="size-[22px] shrink-0" />
          )}
          <span className={`text-sm ${C.text}`}>{item.name}</span>
          {hasChildren && (
            <span className={`text-xs ${C.text25} ml-0.5`}>
              {item.children!.length}
            </span>
          )}
        </div>
      </TableCell>

      {/* 単価(税込) */}
      <TableCell className="w-[100px] py-0 text-right">
        {item.price !== null ? (
          <span className={`text-sm ${C.text70} font-mono`}>
            ¥{item.price.toLocaleString()}
          </span>
        ) : (
          <span className={`text-sm ${C.text30}`}>-</span>
        )}
      </TableCell>

      {/* ステータス */}
      <TableCell className="w-[80px] py-0">
        <StatusDot status={item.status} />
      </TableCell>
    </TableRow>
  );
}

// ─────────────────────────────────────────────────
// TreatmentTreeTable
// ─────────────────────────────────────────────────

interface TreatmentTreeTableProps {
  items: TreatmentItem[];
  searchTerm: string;
}

function TreatmentTreeTable({ items, searchTerm }: TreatmentTreeTableProps) {
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  const toggle = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const filtered = searchTerm
    ? items.filter(
        (item) =>
          item.name.includes(searchTerm) ||
          item.children?.some((c) => c.name.includes(searchTerm)),
      )
    : items;

  const renderRows = (list: TreatmentItem[], depth = 0): React.ReactNode[] => {
    return list.flatMap((item) => {
      const isExpanded = expandedIds.has(item.id);
      const rows: React.ReactNode[] = [
        <TreatmentRow
          key={item.id}
          item={item}
          depth={depth}
          isExpanded={isExpanded}
          onToggle={toggle}
        />,
      ];
      if (isExpanded && item.children?.length) {
        rows.push(...renderRows(item.children, depth + 1));
      }
      return rows;
    });
  };

  return (
    <div className={`border ${C.borderLight} rounded-[3px] overflow-hidden`}>
      <Table>
        <TableHeader>
          <TableRow
            className={`border-b ${C.borderLight} h-[40px] ${C.bgPage}`}
          >
            <TableHead
              className={`w-[120px] text-xs font-medium ${C.text50}`}
            >
              種別
            </TableHead>
            <TableHead className={`text-xs font-medium ${C.text50}`}>
              名称
            </TableHead>
            <TableHead
              className={`w-[100px] text-xs font-medium ${C.text50} text-right`}
            >
              単価(税込)
            </TableHead>
            <TableHead className={`w-[80px] text-xs font-medium ${C.text50}`}>
              ステータス
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filtered.length === 0 ? (
            <TableRow>
              <TableCell
                colSpan={4}
                className={`text-center py-12 text-sm ${C.text40}`}
              >
                データが見つかりません
              </TableCell>
            </TableRow>
          ) : (
            renderRows(filtered)
          )}
        </TableBody>
      </Table>
    </div>
  );
}

// ─────────────────────────────────────────────────
// TreatmentPlanMaster (main)
// ─────────────────────────────────────────────────

export function TreatmentPlanMaster() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<string>("consultation");
  const [searchTerm, setSearchTerm] = useState("");

  const handleTabChange = useCallback((value: string) => {
    setActiveTab(value);
    setSearchTerm("");
  }, []);

  const totalCount =
    activeTab === "consultation"
      ? MOCK_CONSULTATION_ITEMS.reduce(
          (sum, item) => sum + 1 + (item.children?.length ?? 0),
          0,
        )
      : 0;

  const activeItems =
    activeTab === "consultation" ? MOCK_CONSULTATION_ITEMS : [];

  return (
    <PageLayout
      title="治療プランマスタ"
      icon={<Stethoscope className="size-5 text-[#37352F]" />}
      onBack={() => navigate("/settings")}
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-0">
        <Tabs value={activeTab} onValueChange={handleTabChange}>
          <TabsList
            className={`h-9 bg-transparent border-b ${C.borderLight} rounded-none w-full justify-start gap-0 p-0`}
          >
            {TABS.map((tab) => (
              <TabsTrigger
                key={tab.value}
                value={tab.value}
                className={`h-9 rounded-none border-b-2 border-transparent px-4 text-sm ${C.text60}
                  data-[state=active]:border-[#37352F] data-[state=active]:${C.text}
                  data-[state=active]:shadow-none data-[state=active]:bg-transparent`}
              >
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>

          {/* ツールバー */}
          <div className="flex items-center gap-3 py-3 px-1">
            <span className={`text-sm ${C.text50} shrink-0`}>
              {totalCount}件
            </span>
            <div className="relative flex-1 max-w-[280px]">
              <Search
                className={`absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 ${C.text30}`}
              />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="名称で検索..."
                className={`w-full h-8 pl-8 pr-3 text-sm rounded-[3px] outline-none ${C.bgPage} border ${C.borderLight} ${C.textPlaceholder} ${C.text} transition-colors focus:border-[#038B94]`}
              />
            </div>
            <div className="flex-1" />
            <PrimaryButton onClick={() => {}}>
              <Plus className="size-4 mr-1.5" />
              新規登録
            </PrimaryButton>
          </div>

          {/* 診察タブのコンテンツ */}
          <TabsContent value="consultation" className="mt-0">
            <TreatmentTreeTable
              items={activeItems}
              searchTerm={searchTerm}
            />
          </TabsContent>

          {/* その他のタブ */}
          {(["examination", "procedure", "vaccine", "checkup"] as const).map(
            (tab) => (
              <TabsContent key={tab} value={tab} className="mt-0">
                <TreatmentTreeTable items={[]} searchTerm={searchTerm} />
              </TabsContent>
            ),
          )}
        </Tabs>
      </div>
    </PageLayout>
  );
}
