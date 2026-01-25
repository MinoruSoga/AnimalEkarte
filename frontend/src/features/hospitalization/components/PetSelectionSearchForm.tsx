import { Search } from "lucide-react";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { Label } from "../../../components/ui/label";
import { H_STYLES } from "../styles";

interface PetSelectionSearchFormProps {
  searchParams: {
    ownerId: string;
    ownerName: string;
    ownerNameKana: string;
    phone: string;
    petName: string;
    petNameKana: string;
    species: string;
    address: string;
  };
  setSearchParams: (params: PetSelectionSearchFormProps["searchParams"]) => void;
  onSearch: () => void;
}

export const PetSelectionSearchForm = ({ searchParams, setSearchParams, onSearch }: PetSelectionSearchFormProps) => {
  return (
    <div className={`mb-4 rounded-lg bg-white ${H_STYLES.padding.box} shadow-[0_1px_3px_rgba(0,0,0,0.04)] border border-[rgba(55,53,47,0.16)]`}>
      <h2 className={`mb-2 ${H_STYLES.text.base} font-medium text-[#37352F]`}>検索条件</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-4">
        <div className="space-y-1.5">
          <Label htmlFor="ownerId" className={`${H_STYLES.text.sm} text-[#37352F]/60`}>
            飼主No
          </Label>
          <Input
            id="ownerId"
            placeholder="例: 30042"
            value={searchParams.ownerId}
            onChange={(e) =>
              setSearchParams({ ...searchParams, ownerId: e.target.value })
            }
            className={`${H_STYLES.text.base} h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]`}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="ownerName" className={`${H_STYLES.text.sm} text-[#37352F]/60`}>
            飼主名
          </Label>
          <Input
            id="ownerName"
            placeholder="例: 林 文明"
            value={searchParams.ownerName}
            onChange={(e) =>
              setSearchParams({ ...searchParams, ownerName: e.target.value })
            }
            className={`${H_STYLES.text.base} h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]`}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="ownerNameKana" className={`${H_STYLES.text.sm} text-[#37352F]/60`}>
            飼主名(カナ)
          </Label>
          <Input
            id="ownerNameKana"
            placeholder="例: ハヤシ フミアキ"
            value={searchParams.ownerNameKana}
            onChange={(e) =>
              setSearchParams({ ...searchParams, ownerNameKana: e.target.value })
            }
            className={`${H_STYLES.text.base} h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]`}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="phone" className={`${H_STYLES.text.sm} text-[#37352F]/60`}>
            電話番号
          </Label>
          <Input
            id="phone"
            placeholder="例: 090-1234-5678"
            value={searchParams.phone}
            onChange={(e) =>
              setSearchParams({ ...searchParams, phone: e.target.value })
            }
            className={`${H_STYLES.text.base} h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]`}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="petName" className={`${H_STYLES.text.sm} text-[#37352F]/60`}>
            ペット名
          </Label>
          <Input
            id="petName"
            placeholder="例: Iris"
            value={searchParams.petName}
            onChange={(e) =>
              setSearchParams({ ...searchParams, petName: e.target.value })
            }
            className={`${H_STYLES.text.base} h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]`}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="petNameKana" className={`${H_STYLES.text.sm} text-[#37352F]/60`}>
            ペット名(カナ)
          </Label>
          <Input
            id="petNameKana"
            placeholder="例: イリス"
            value={searchParams.petNameKana}
            onChange={(e) =>
              setSearchParams({ ...searchParams, petNameKana: e.target.value })
            }
            className={`${H_STYLES.text.base} h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]`}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="species" className={`${H_STYLES.text.sm} text-[#37352F]/60`}>
            種別
          </Label>
          <Input
            id="species"
            placeholder="例: 犬"
            value={searchParams.species}
            onChange={(e) =>
              setSearchParams({ ...searchParams, species: e.target.value })
            }
            className={`${H_STYLES.text.base} h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]`}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="address" className={`${H_STYLES.text.sm} text-[#37352F]/60`}>
            住所
          </Label>
          <Input
            id="address"
            placeholder="例: 東京都"
            value={searchParams.address}
            onChange={(e) =>
              setSearchParams({ ...searchParams, address: e.target.value })
            }
            className={`${H_STYLES.text.base} h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]`}
          />
        </div>
      </div>
      <div className="flex justify-end">
        <Button 
          size="sm"
          onClick={onSearch} 
          className={`gap-2 bg-[#37352F] hover:bg-[#37352F]/90 text-white focus-visible:ring-[#2EAADC] ${H_STYLES.button.action}`}
        >
          <Search className={H_STYLES.button.icon} />
          検索
        </Button>
      </div>
    </div>
  );
};
