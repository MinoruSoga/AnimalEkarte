// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { useState, useCallback, useDeferredValue, useTransition, memo } from "react";

// External
import { Search, Users } from "lucide-react";

// Internal
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { TableCell, TableHead } from "@/components/ui/table";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { EmptyState } from "@/components/shared/DataStates";
import { handleApiError } from "@/lib/handle-api-error";
import { axios } from "@/lib/axios";

// Types
import { transformOwner, type OwnerApiResponse } from "@/lib/transforms/owner";

interface OwnerSummary {
  id: string;
  name: string;
  phone: string;
  address: string;
  discountRate: number;
  membershipType: string;
}

interface OwnerSearchModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (owner: { id: string; name: string; discountRate: number; membershipType: string }) => void;
  currentOwnerName?: string;
}

interface OwnerSearchResponse {
  data: OwnerApiResponse[];
  total?: number;
}

const OWNER_SEARCH_LIMIT = 100;

function toOwnerSummary(o: OwnerApiResponse): OwnerSummary {
  const owner = transformOwner(o);
  return {
    id: owner.id,
    name: owner.ownerName,
    phone: owner.phone,
    address: [owner.address1, owner.address2].filter(Boolean).join(" "),
    discountRate: owner.discountRate,
    membershipType: owner.membershipType,
  };
}

export const OwnerSearchModal = memo(function OwnerSearchModal({
  open,
  onOpenChange,
  onSelect,
  currentOwnerName,
}: OwnerSearchModalProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [owners, setOwners] = useState<OwnerSummary[]>([]);
  const [isSearching, startSearchTransition] = useTransition();
  const [hasSearched, setHasSearched] = useState(false);
  const [isTruncated, setIsTruncated] = useState(false);
  const [confirmTarget, setConfirmTarget] = useState<OwnerSummary | null>(null);

  const handleSearch = useCallback(() => {
    if (!searchTerm.trim()) return;
    setHasSearched(true);
    startSearchTransition(async () => {
      try {
        const { data } = await axios.get<OwnerSearchResponse>("/v1/owners", {
          params: { search: searchTerm.trim(), page: 1, limit: OWNER_SEARCH_LIMIT },
        });
        const nextOwners = (data.data ?? []).map(toOwnerSummary);
        setOwners(nextOwners);
        setIsTruncated(typeof data.total === "number" && data.total > nextOwners.length);
      } catch (error) {
        handleApiError(error, "飼主検索");
        setOwners([]);
        setIsTruncated(false);
      }
    });
  }, [searchTerm]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") handleSearch();
    },
    [handleSearch],
  );

  const handleRowClick = useCallback((owner: OwnerSummary) => {
    setConfirmTarget(owner);
  }, []);

  const handleConfirm = useCallback(() => {
    if (!confirmTarget) return;
    onSelect({
      id: confirmTarget.id,
      name: confirmTarget.name,
      discountRate: confirmTarget.discountRate,
      membershipType: confirmTarget.membershipType,
    });
    setConfirmTarget(null);
    onOpenChange(false);
    // reset
    setSearchTerm("");
    setOwners([]);
    setHasSearched(false);
    setIsTruncated(false);
  }, [confirmTarget, onSelect, onOpenChange]);

  const handleCancel = useCallback(() => {
    setConfirmTarget(null);
  }, []);

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      if (!nextOpen) {
        setSearchTerm("");
        setOwners([]);
        setHasSearched(false);
        setIsTruncated(false);
        setConfirmTarget(null);
      }
      onOpenChange(nextOpen);
    },
    [onOpenChange],
  );

  const isFiltering = searchTerm !== deferredSearch;

  const filteredOwners = owners;

  return (
    <>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="sm:max-w-2xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className={`flex items-center gap-2 text-base font-semibold ${C.text}`}>
              <Users className={ICON.action} />
              飼主検索
            </DialogTitle>
            <DialogDescription className={`text-sm ${C.text50}`}>
              飼主名、飼主No、電話番号で検索できます
            </DialogDescription>
          </DialogHeader>

          {/* Search */}
          <div className="flex items-center gap-2 px-1">
            <div className="relative flex-1">
              <Search className={`absolute left-3 top-1/2 -translate-y-1/2 ${ICON.action} ${C.text40}`} />
              <Input
                autoFocus
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="飼主名 / 飼主No / 電話番号"
                className={`pl-9 h-10 text-base ${C.bgPage} ${C.borderMedium} ${C.focusBorderAccent} rounded-xs`}
              />
            </div>
            <Button
              onClick={handleSearch}
              disabled={!searchTerm.trim() || isSearching}
              className="h-10 px-4 text-base"
            >
              検索
            </Button>
          </div>

          {/* Results */}
          <div className={`flex-1 overflow-auto min-h-[200px] ${isFiltering ? "opacity-60" : ""}`}>
            {isSearching ? (
              <div className={`flex items-center justify-center h-full text-sm ${C.text40}`}>
                検索中...
              </div>
            ) : filteredOwners.length > 0 ? (
              <>
                {isTruncated ? (
                  <p className={`px-2 py-2 text-xs ${C.text50}`} role="status">
                    検索結果の先頭100件までを表示しています
                  </p>
                ) : null}
                <table className="w-full">
                  <thead>
                    {/* DESIGN.md ex-data-table-cell: header は canvas-soft 背景 + eyebrow 相当タイポグラフィ */}
                    <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
                      <TableHead className={C.text55}>飼主No</TableHead>
                      <TableHead className={C.text55}>飼主名</TableHead>
                      <TableHead className={C.text55}>電話番号</TableHead>
                      <TableHead className={C.text55}>住所</TableHead>
                      <TableHead className={C.text55}>操作</TableHead>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredOwners.map((owner) => (
                      <tr key={owner.id} className={`border-b ${C.borderDivider} ${C.hoverBgLight} transition-colors`}>
                        <TableCell className={`${C.text60} font-mono`}>{owner.id}</TableCell>
                        <TableCell className={`font-medium ${C.text}`}>{owner.name}</TableCell>
                        <TableCell className={C.text}>{owner.phone || "-"}</TableCell>
                        <TableCell className={`${C.text60} truncate max-w-[200px]`}>{owner.address || "-"}</TableCell>
                        <TableCell>
                          <DataTableRowButton
                            aria-label={`選択: 飼主 ${owner.name} (ID ${owner.id})`}
                            onClick={() => handleRowClick(owner)}
                          >
                            選択
                          </DataTableRowButton>
                        </TableCell>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </>
            ) : hasSearched ? (
              <div className="flex items-center justify-center h-full">
                <EmptyState message="該当する飼主が見つかりません" />
              </div>
            ) : (
              <div className={`flex items-center justify-center h-full text-sm ${C.text30}`}>
                検索してください
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* Confirm Dialog */}
      <ConfirmDialog
        open={confirmTarget !== null}
        onClose={handleCancel}
        title="飼主変更の確認"
        description={
          currentOwnerName
            ? `飼主を「${currentOwnerName}」→「${confirmTarget?.name ?? ""}」に変更します。よろしいですか？`
            : `飼主を「${confirmTarget?.name ?? ""}」に変更します。よろしいですか？`
        }
        confirmLabel="変更する"
        onConfirm={handleConfirm}
      />
    </>
  );
});
