import { memo } from "react";
import { C, ICON } from "@/lib/design-tokens";
import { Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/components/ui/utils";

export interface PetSelectionSearchParams {
  ownerId: string;
  ownerName: string;
  ownerNameKana: string;
  phone: string;
  petName: string;
  petNameKana: string;
  species: string;
  address: string;
}

interface PetSelectionSearchFormProps {
  searchParams: PetSelectionSearchParams;
  setSearchParams: (params: PetSelectionSearchParams) => void;
  onSearch: () => void;
  onClear: () => void;
}

const FIELD_DEFS = [
  { id: "ownerId", label: "飼主No", placeholder: "例: 30042" },
  { id: "ownerName", label: "飼主名", placeholder: "例: 林 文明" },
  { id: "ownerNameKana", label: "飼主名よみ", placeholder: "例: はやし ふみあき" },
  { id: "phone", label: "電話番号", placeholder: "例: 090-1234-5678" },
  { id: "petName", label: "ペット名", placeholder: "例: Iris" },
  { id: "petNameKana", label: "ペット名よみ", placeholder: "例: いりす" },
  { id: "species", label: "種別", placeholder: "例: 犬" },
  { id: "address", label: "住所", placeholder: "例: 東京都" },
] as const;

export const PetSelectionSearchForm = memo(function PetSelectionSearchForm({ searchParams, setSearchParams, onSearch, onClear }: PetSelectionSearchFormProps) {
  return (
    <div className={cn("mb-4 rounded-lg bg-white p-4 border", C.borderMedium)}>
      <h2 className={cn("mb-2 text-sm font-medium", C.text)}>検索条件</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-4">
        {FIELD_DEFS.map(({ id, label, placeholder }) => (
          <div key={id} className="space-y-1.5">
            <Label htmlFor={id} className={cn("text-sm", C.text60)}>
              {label}
            </Label>
            <Input
              id={id}
              placeholder={placeholder}
              value={searchParams[id as keyof PetSelectionSearchParams]}
              onChange={(e) =>
                setSearchParams({ ...searchParams, [id]: e.target.value })
              }
              className={cn("text-sm h-11 bg-white focus-visible:ring-1", C.text, C.borderMedium)}
            />
          </div>
        ))}
      </div>
      <div className="flex justify-end gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={onClear}
          className={cn("h-11 text-sm", C.borderMedium, C.text60, C.hoverBgPage, C.hoverText)}
        >
          クリア
        </Button>
        {/* docs/spec/design-system.md button-primary: brand teal #038B94 + pill。cn()/tailwind-merge が
            クラス競合を解決するため、後続の brand トークンで安全に上書きできる。 */}
        <Button
          size="sm"
          onClick={onSearch}
          className={cn("gap-2 h-11 text-sm rounded-full", C.textWhite, C.bgBrand, C.hoverBgBrand)}
        >
          <Search className={ICON.action} />
          検索
        </Button>
      </div>
    </div>
  );
});
