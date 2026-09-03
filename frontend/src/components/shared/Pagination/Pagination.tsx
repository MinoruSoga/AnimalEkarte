import { memo } from "react";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export interface PaginationProps {
  currentPage: number;
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  onPageChange: (page: number) => void;
  onPrev: () => void;
  onNext: () => void;
}

export const Pagination = memo(function Pagination({
  currentPage,
  totalPages,
  totalCount,
  startIndex,
  endIndex,
  onPageChange,
  onPrev,
  onNext,
}: PaginationProps) {
  if (totalCount === 0) return null;

  // Generate page numbers to display
  const getPageNumbers = () => {
    const pages: (number | "ellipsis")[] = [];
    const maxVisible = 5;

    if (totalPages <= maxVisible + 2) {
      // Show all pages
      for (let i = 1; i <= totalPages; i++) pages.push(i);
    } else {
      // Always show first page
      pages.push(1);

      if (currentPage > 3) {
        pages.push("ellipsis");
      }

      // Pages around current
      const start = Math.max(2, currentPage - 1);
      const end = Math.min(totalPages - 1, currentPage + 1);
      for (let i = start; i <= end; i++) {
        pages.push(i);
      }

      if (currentPage < totalPages - 2) {
        pages.push("ellipsis");
      }

      // Always show last page
      pages.push(totalPages);
    }

    return pages;
  };

  const pageNumbers = getPageNumbers();

  // FE-RC-044: nav ランドマークはこのコンポーネント側で付けない。少なくとも
  // OwnerAccountingHistory.tsx が既に `<nav aria-label="ページネーション">` で外側から
  // ラップする規約を持っており、ここで同名の nav を追加すると入れ子で aria-label が重複し
  // 「複数要素が見つかる」a11y クエリの破壊的衝突になる（実測: pagination.test.tsx で検出）。
  return (
    <div className="flex items-center justify-between py-3 px-1">
      <div className={STYLE.paginationInfo}>
        {totalCount.toLocaleString()}件中 {startIndex.toLocaleString()}-{endIndex.toLocaleString()}
        件
      </div>

      <div className="flex items-center gap-1">
        {/* First page */}
        <Button
          variant="ghost"
          size="icon"
          className={STYLE.paginationBtn}
          onClick={() => onPageChange(1)}
          disabled={currentPage === 1}
          aria-label="最初のページ"
          data-testid="pagination-first"
        >
          <ChevronsLeft aria-hidden="true" className={ICON.action} />
        </Button>

        {/* Previous */}
        <Button
          variant="ghost"
          size="icon"
          className={STYLE.paginationBtn}
          onClick={onPrev}
          disabled={currentPage === 1}
          aria-label="前のページ"
          data-testid="pagination-prev"
        >
          <ChevronLeft aria-hidden="true" className={ICON.action} />
        </Button>

        {/* Page numbers */}
        {pageNumbers.map((page, idx) =>
          page === "ellipsis" ? (
            <span
              key={`ellipsis-${idx}`}
              aria-hidden="true"
              className={`px-1 text-base ${C.text40}`}
            >
              ...
            </span>
          ) : (
            <Button
              key={page}
              variant={currentPage === page ? "default" : "ghost"}
              size="icon"
              className={currentPage === page ? STYLE.paginationBtnActive : STYLE.paginationBtn}
              onClick={() => onPageChange(page)}
              // FE-RC-044: aria-label は付けない — 付けると accessible name が数字テキストから
              // 上書きされ、他 feature の既存テスト（getByRole("button", { name: "2" }) 等）が
              // 割れる。aria-current="page" のみで現在ページを SR に伝える（WAI-ARIA pagination
              // pattern としても十分）。
              aria-current={currentPage === page ? "page" : undefined}
            >
              {page}
            </Button>
          ),
        )}

        {/* Next */}
        <Button
          variant="ghost"
          size="icon"
          className={STYLE.paginationBtn}
          onClick={onNext}
          disabled={currentPage === totalPages}
          aria-label="次のページ"
          data-testid="pagination-next"
        >
          <ChevronRight aria-hidden="true" className={ICON.action} />
        </Button>

        {/* Last page */}
        <Button
          variant="ghost"
          size="icon"
          className={STYLE.paginationBtn}
          onClick={() => onPageChange(totalPages)}
          disabled={currentPage === totalPages}
          aria-label="最後のページ"
          data-testid="pagination-last"
        >
          <ChevronsRight aria-hidden="true" className={ICON.action} />
        </Button>
      </div>
    </div>
  );
});
