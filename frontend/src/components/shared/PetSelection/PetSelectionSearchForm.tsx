import { Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

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
}

const FIELD_DEFS = [
  { id: "ownerId", label: "飼主No", placeholder: "例: 30042" },
  { id: "ownerName", label: "飼主名", placeholder: "例: 林 文明" },
  { id: "ownerNameKana", label: "飼主名(カナ)", placeholder: "例: ハヤシ フミアキ" },
  { id: "phone", label: "電話番号", placeholder: "例: 090-1234-5678" },
  { id: "petName", label: "ペット名", placeholder: "例: Iris" },
  { id: "petNameKana", label: "ペット名(カナ)", placeholder: "例: イリス" },
  { id: "species", label: "種別", placeholder: "例: 犬" },
  { id: "address", label: "住所", placeholder: "例: 東京都" },
] as const;

export const PetSelectionSearchForm = ({ searchParams, setSearchParams, onSearch }: PetSelectionSearchFormProps) => {
  return (
    <div className="mb-4 rounded-lg bg-white p-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)] border border-[rgba(55,53,47,0.16)]">
      <h2 className="mb-2 text-sm font-medium text-[#37352F]">検索条件</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-4">
        {FIELD_DEFS.map(({ id, label, placeholder }) => (
          <div key={id} className="space-y-1.5">
            <Label htmlFor={id} className="text-sm text-[#37352F]/60">
              {label}
            </Label>
            <Input
              id={id}
              placeholder={placeholder}
              value={searchParams[id as keyof PetSelectionSearchParams]}
              onChange={(e) =>
                setSearchParams({ ...searchParams, [id]: e.target.value })
              }
              className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
            />
          </div>
        ))}
      </div>
      <div className="flex justify-end">
        <Button
          size="sm"
          onClick={onSearch}
          className="gap-2 bg-[#37352F] hover:bg-[#37352F]/90 text-white focus-visible:ring-[#2EAADC] h-10 text-sm"
        >
          <Search className="size-4" />
          検索
        </Button>
      </div>
    </div>
  );
};
