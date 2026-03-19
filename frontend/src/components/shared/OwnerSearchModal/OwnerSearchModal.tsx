// React/Framework
import { useState, useCallback, useMemo, useDeferredValue, memo } from "react";

// External
import { Search, Users } from "lucide-react";

// Internal
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { axios } from "@/lib/axios";

// Types
import type { Owner as BackendOwner } from "@/types/generated/models";

interface OwnerSummary {
  id: string;
  name: string;
  phone: string;
  address: string;
}

interface OwnerSearchModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (owner: { id: string; name: string }) => void;
  currentOwnerName?: string;
}

function transformOwner(o: BackendOwner): OwnerSummary {
  return {
    id: String(o.id ?? 0),
    name: o.owner_name ?? "",
    phone: o.phone ?? "",
    address: [o.address1, o.address2].filter(Boolean).join(" "),
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
  const [isSearching, setIsSearching] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [confirmTarget, setConfirmTarget] = useState<OwnerSummary | null>(null);

  const handleSearch = useCallback(async () => {
    if (!searchTerm.trim()) return;
    setIsSearching(true);
    setHasSearched(true);
    try {
      const { data } = await axios.get<{ data: BackendOwner[] }>("/v1/owners", {
        params: { search: searchTerm.trim() },
      });
      setOwners((data.data ?? []).map(transformOwner));
    } catch {
      setOwners([]);
    } finally {
      setIsSearching(false);
    }
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
    onSelect({ id: confirmTarget.id, name: confirmTarget.name });
    setConfirmTarget(null);
    onOpenChange(false);
    // reset
    setSearchTerm("");
    setOwners([]);
    setHasSearched(false);
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
        setConfirmTarget(null);
      }
      onOpenChange(nextOpen);
    },
    [onOpenChange],
  );

  const isFiltering = searchTerm !== deferredSearch;

  const filteredOwners = useMemo(() => owners, [owners]);

  return (
    <>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="sm:max-w-2xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-base font-semibold text-[#37352F]">
              <Users className="size-4" />
              飼主検索
            </DialogTitle>
            <DialogDescription className="text-sm text-[#37352F]/50">
              飼主名、飼主No、電話番号で検索できます
            </DialogDescription>
          </DialogHeader>

          {/* Search */}
          <div className="flex items-center gap-2 px-1">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-[#37352F]/40" />
              <Input
                autoFocus
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="飼主名 / 飼主No / 電話番号"
                className="pl-9 h-10 text-base bg-[#F7F6F3] border-[rgba(55,53,47,0.16)] focus:border-[#2383E2] rounded-[4px]"
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
              <div className="flex items-center justify-center h-full text-sm text-[#37352F]/40">
                検索中...
              </div>
            ) : filteredOwners.length > 0 ? (
              <table className="w-full">
                <thead>
                  <tr className="border-b border-[rgba(55,53,47,0.09)] bg-[#F7F6F3]">
                    <th className="px-3 py-2 text-left text-xs font-medium text-[#37352F]/60">飼主No</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-[#37352F]/60">飼主名</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-[#37352F]/60">電話番号</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-[#37352F]/60">住所</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredOwners.map((owner) => (
                    <tr
                      key={owner.id}
                      onClick={() => handleRowClick(owner)}
                      className="border-b border-[rgba(55,53,47,0.06)] cursor-pointer hover:bg-[rgba(55,53,47,0.04)] transition-colors"
                    >
                      <td className="px-3 py-2.5 text-sm text-[#37352F]/60 font-mono">{owner.id}</td>
                      <td className="px-3 py-2.5 text-sm font-medium text-[#37352F]">{owner.name}</td>
                      <td className="px-3 py-2.5 text-sm text-[#37352F]">{owner.phone || "-"}</td>
                      <td className="px-3 py-2.5 text-sm text-[#37352F]/60 truncate max-w-[200px]">{owner.address || "-"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : hasSearched ? (
              <div className="flex items-center justify-center h-full text-sm text-[#37352F]/40">
                該当する飼主が見つかりません
              </div>
            ) : (
              <div className="flex items-center justify-center h-full text-sm text-[#37352F]/30">
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
