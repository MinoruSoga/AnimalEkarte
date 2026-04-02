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

  return (
    <div className="flex items-center justify-between py-3 px-1">
      <div className={STYLE.paginationInfo}>
        {totalCount.toLocaleString()}件中 {startIndex.toLocaleString()}-{endIndex.toLocaleString()}件
      </div>

      <div className="flex items-center gap-1">
        {/* First page */}
        <Button
          variant="ghost"
          size="icon"
          className={STYLE.paginationBtn}
          onClick={() => onPageChange(1)}
          disabled={currentPage === 1}
        >
          <ChevronsLeft className={ICON.action} />
        </Button>

        {/* Previous */}
        <Button
          variant="ghost"
          size="icon"
          className={STYLE.paginationBtn}
          onClick={onPrev}
          disabled={currentPage === 1}
        >
          <ChevronLeft className={ICON.action} />
        </Button>

        {/* Page numbers */}
        {pageNumbers.map((page, idx) =>
          page === "ellipsis" ? (
            <span
              key={`ellipsis-${idx}`}
              className={`px-1 text-base ${C.text40}`}
            >
              ...
            </span>
          ) : (
            <Button
              key={page}
              variant={currentPage === page ? "default" : "ghost"}
              size="icon"
              className={
                currentPage === page
                  ? STYLE.paginationBtnActive
                  : STYLE.paginationBtn
              }
              onClick={() => onPageChange(page)}
            >
              {page}
            </Button>
          )
        )}

        {/* Next */}
        <Button
          variant="ghost"
          size="icon"
          className={STYLE.paginationBtn}
          onClick={onNext}
          disabled={currentPage === totalPages}
        >
          <ChevronRight className={ICON.action} />
        </Button>

        {/* Last page */}
        <Button
          variant="ghost"
          size="icon"
          className={STYLE.paginationBtn}
          onClick={() => onPageChange(totalPages)}
          disabled={currentPage === totalPages}
        >
          <ChevronsRight className={ICON.action} />
        </Button>
      </div>
    </div>
  );
});
