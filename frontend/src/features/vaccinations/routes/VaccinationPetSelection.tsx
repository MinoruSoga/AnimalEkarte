import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { Label } from "../../../components/ui/label";
import { Badge } from "../../../components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../../../components/ui/table";
import { Search, Check } from "lucide-react";
import { Pet } from "../../../types";
import { MOCK_PETS } from "../../../lib/constants";
import { PageLayout } from "../../../components/shared/PageLayout";

export const VaccinationPetSelection = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useState({
    ownerId: "",
    ownerName: "",
    ownerNameKana: "",
    phone: "",
    petName: "",
    petNameKana: "",
    species: "",
    address: "",
  });

  const filteredPets = MOCK_PETS.filter((pet) => {
    if (searchParams.ownerId && !pet.ownerId.includes(searchParams.ownerId)) return false;
    if (searchParams.ownerName && !pet.ownerName.includes(searchParams.ownerName)) return false;
    if (searchParams.phone && (!pet.phone || !pet.phone.includes(searchParams.phone))) return false;
    if (searchParams.petName && !pet.name.includes(searchParams.petName)) return false;
    if (searchParams.species && !pet.species.includes(searchParams.species)) return false;
    return true;
  });

  const handleSearch = () => {
    // Search is already handled by filteredPets reactive filter
  };

  const handleSelect = (pet: Pet) => {
    navigate(`/vaccinations/new?petId=${pet.id}`);
  };

  const handleBack = () => {
    navigate("/vaccinations");
  };

  return (
    <PageLayout
      title="予防接種登録 - ペット選択"
      onBack={handleBack}
      maxWidth="max-w-full"
    >
        {/* Search Form */}
        <div className="mb-4 rounded-lg bg-white p-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)] border border-[rgba(55,53,47,0.16)]">
          <h2 className="mb-2 text-sm font-medium text-[#37352F]">検索条件</h2>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
            <div className="space-y-1.5">
              <Label htmlFor="ownerId" className="text-sm text-[#37352F]/60">
                飼主No
              </Label>
              <Input
                id="ownerId"
                placeholder="例: 30042"
                value={searchParams.ownerId}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, ownerId: e.target.value })
                }
                className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="ownerName" className="text-sm text-[#37352F]/60">
                飼主名
              </Label>
              <Input
                id="ownerName"
                placeholder="例: 林 文明"
                value={searchParams.ownerName}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, ownerName: e.target.value })
                }
                className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="ownerNameKana" className="text-sm text-[#37352F]/60">
                飼主名(カナ)
              </Label>
              <Input
                id="ownerNameKana"
                placeholder="例: ハヤシ フミアキ"
                value={searchParams.ownerNameKana}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, ownerNameKana: e.target.value })
                }
                className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="phone" className="text-sm text-[#37352F]/60">
                電話番号
              </Label>
              <Input
                id="phone"
                placeholder="例: 090-1234-5678"
                value={searchParams.phone}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, phone: e.target.value })
                }
                className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="petName" className="text-sm text-[#37352F]/60">
                ペット名
              </Label>
              <Input
                id="petName"
                placeholder="例: Iris"
                value={searchParams.petName}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, petName: e.target.value })
                }
                className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="petNameKana" className="text-sm text-[#37352F]/60">
                ペット名(カナ)
              </Label>
              <Input
                id="petNameKana"
                placeholder="例: イリス"
                value={searchParams.petNameKana}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, petNameKana: e.target.value })
                }
                className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="species" className="text-sm text-[#37352F]/60">
                種別
              </Label>
              <Input
                id="species"
                placeholder="例: 犬"
                value={searchParams.species}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, species: e.target.value })
                }
                className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="address" className="text-sm text-[#37352F]/60">
                住所
              </Label>
              <Input
                id="address"
                placeholder="例: 東京都"
                value={searchParams.address}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, address: e.target.value })
                }
                className="text-sm h-10 bg-white text-[#37352F] focus-visible:ring-[#2EAADC] focus-visible:ring-1 border-[rgba(55,53,47,0.16)]"
              />
            </div>
          </div>
          <div className="flex justify-end">
            <Button 
              size="sm"
              onClick={handleSearch} 
              className="gap-2 bg-[#37352F] hover:bg-[#37352F]/90 text-white focus-visible:ring-[#2EAADC] h-10 text-sm"
            >
              <Search className="size-4" />
              検索
            </Button>
          </div>
        </div>

        {/* Results */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-medium text-[#37352F]">検索結果</h2>
            <span className="text-sm text-[#37352F]/60">
              {filteredPets.length}件
            </span>
          </div>

          <div className="rounded-lg bg-white overflow-hidden shadow-sm border border-[rgba(55,53,47,0.16)]">
            <Table>
              <TableHeader>
                <TableRow 
                  className="hover:bg-transparent bg-[#F7F6F3] border-b border-[rgba(55,53,47,0.16)] h-12"
                >
                  <TableHead className="min-w-[80px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">飼主No</TableHead>
                  <TableHead className="min-w-[140px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">飼主名</TableHead>
                  <TableHead className="min-w-[80px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">ペット番号</TableHead>
                  <TableHead className="min-w-[100px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">ペット名</TableHead>
                  <TableHead className="min-w-[50px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">生死</TableHead>
                  <TableHead className="min-w-[50px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">種</TableHead>
                  <TableHead className="min-w-[90px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">生年月日</TableHead>
                  <TableHead className="min-w-[60px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">体重</TableHead>
                  <TableHead className="min-w-[100px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">環境</TableHead>
                  <TableHead className="min-w-[90px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">前回来院</TableHead>
                  <TableHead className="min-w-[80px] text-sm text-[#37352F]/60 whitespace-nowrap h-12">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPets.map((pet, index) => (
                  <TableRow 
                    key={pet.id}
                    className={`transition-colors hover:bg-[rgba(55,53,47,0.06)] cursor-pointer h-12 ${
                      index < filteredPets.length - 1 ? "border-b border-[rgba(55,53,47,0.09)]" : "border-none"
                    }`}
                    onClick={() => handleSelect(pet)}
                  >
                    <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.ownerId}</TableCell>
                    <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.ownerName}</TableCell>
                    <TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">{pet.petNumber || "-"}</TableCell>
                    <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.name}</TableCell>
                    <TableCell className="whitespace-nowrap py-2">
                      {pet.status && (
                          <Badge
                            variant="secondary"
                            className={
                            pet.status === "生存"
                                ? "bg-[#DDEDEA] text-[#0F7B6C] border-[#DDEDEA] hover:bg-[#DDEDEA] text-sm px-2 py-0 h-7"
                                : "bg-[#EBECED] text-[#9B9A97] border-[#EBECED] hover:bg-[#EBECED] text-sm px-2 py-0 h-7"
                            }
                          >
                            {pet.status}
                          </Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.species}</TableCell>
                    <TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">{pet.birthDate || "-"}</TableCell>
                    <TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">{pet.weight || "-"}</TableCell>
                    <TableCell className="text-sm text-[#37352F] whitespace-nowrap py-2">{pet.environment || "-"}</TableCell>
                    <TableCell className="font-mono text-sm text-[#37352F] whitespace-nowrap py-2">{pet.lastVisit || "-"}</TableCell>
                    <TableCell className="whitespace-nowrap py-2" onClick={(e) => e.stopPropagation()}>
                      <Button 
                        size="sm" 
                        className="h-10 gap-1 bg-[#37352F] hover:bg-[#37352F]/90 text-white text-sm px-4" 
                        onClick={() => handleSelect(pet)}
                      >
                        <Check className="size-4" />
                        選択
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
                {filteredPets.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={11} className="h-24 text-center text-sm text-[#37352F]/60">
                      該当するペットが見つかりません
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </div>
    </PageLayout>
  );
}
