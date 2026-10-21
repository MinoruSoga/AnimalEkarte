import { Menu, Printer, Edit2 } from "lucide-react";

import { C } from "@/lib/design-tokens";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";

import { ManualSidebar } from "../components/ManualSidebar";
import type { ManualArticle, ManualCategory } from "@/lib/manual-index";

interface ManualPageChromeProps {
  drawerOpen: boolean;
  onDrawerOpenChange: (open: boolean) => void;
  onDrawerCloseAutoFocus: (event: Event) => void;
  desktopSidebarRef: React.RefObject<HTMLDivElement | null>;
  article: ManualArticle | undefined;
  editMode: boolean;
  canEditManual: boolean;
  onStartEdit: () => void;
  viewMode: ManualCategory;
  onChangeViewMode: (mode: ManualCategory) => void;
  query: string;
  onChangeQuery: (query: string) => void;
  filteredArticles: ManualArticle[];
  isSearching: boolean;
  // overrides 取得 pending/error: 検索 input・結果一覧・印刷を無効化する。
  disabled: boolean;
}

export function ManualPageChrome({
  drawerOpen,
  onDrawerOpenChange,
  onDrawerCloseAutoFocus,
  desktopSidebarRef,
  article,
  editMode,
  canEditManual,
  onStartEdit,
  viewMode,
  onChangeViewMode,
  query,
  onChangeQuery,
  filteredArticles,
  isSearching,
  disabled,
}: ManualPageChromeProps) {
  return (
    <>
      <Sheet open={drawerOpen} onOpenChange={onDrawerOpenChange}>
        <SheetTrigger asChild>
          <button
            type="button"
            aria-label="マニュアル目次を開く"
            className={`md:hidden fixed top-3 left-16 z-30 size-11 flex items-center justify-center rounded-xxs border ${C.borderDivider} ${C.bgWhite} ${C.hoverBgLight} no-print`}
          >
            <Menu className="size-5" />
          </button>
        </SheetTrigger>
        <SheetContent
          side="left"
          aria-modal="true"
          aria-describedby={undefined}
          onCloseAutoFocus={onDrawerCloseAutoFocus}
          className="w-[280px] gap-0 p-0 md:hidden [&>button]:min-h-11 [&>button]:min-w-11"
        >
          <SheetHeader className="min-h-14 shrink-0 justify-center py-2 pr-16">
            <SheetTitle>マニュアル目次</SheetTitle>
          </SheetHeader>
          <div className="min-h-0 flex-1">
            <fieldset disabled={disabled} className="contents">
              <ManualSidebar
                viewMode={viewMode}
                onChangeViewMode={onChangeViewMode}
                query={query}
                onChangeQuery={onChangeQuery}
                filteredArticles={filteredArticles}
                isSearching={isSearching}
              />
            </fieldset>
          </div>
        </SheetContent>
      </Sheet>

      {article && !editMode && canEditManual ? (
        <button
          type="button"
          onClick={onStartEdit}
          disabled={disabled}
          aria-label="このページを編集"
          title="このページを編集（編集内容はダウンロードして管理者に共有）"
          className={`fixed top-3 right-16 z-30 size-11 flex items-center justify-center rounded-xxs border ${C.borderDivider} ${C.bgWhite} ${C.hoverBgLight} no-print`}
        >
          <Edit2 className="size-5" />
        </button>
      ) : null}

      {article && !editMode ? (
        <button
          type="button"
          onClick={() => window.print()}
          disabled={disabled}
          aria-label="このページを印刷"
          title="このページを印刷"
          className={`fixed top-3 right-3 z-30 size-11 flex items-center justify-center rounded-xxs border ${C.borderDivider} ${C.bgWhite} ${C.hoverBgLight} no-print disabled:opacity-50`}
        >
          <Printer className="size-5" />
        </button>
      ) : null}

      <div ref={desktopSidebarRef} className="hidden md:flex no-print">
        {/* fieldset disabled で ManualSidebar 内の検索 input 等を一括無効化
            （sidebar 本体の prop は変えない） */}
        <fieldset disabled={disabled} className="contents">
          <ManualSidebar
            viewMode={viewMode}
            onChangeViewMode={onChangeViewMode}
            query={query}
            onChangeQuery={onChangeQuery}
            filteredArticles={filteredArticles}
            isSearching={isSearching}
          />
        </fieldset>
      </div>
    </>
  );
}
