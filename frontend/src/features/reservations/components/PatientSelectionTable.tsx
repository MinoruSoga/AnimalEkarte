import { useState } from "react";
import { Check, Filter } from "lucide-react";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { Label } from "../../../components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../../../components/ui/table";
import { Pet } from "../../../types";
import { MOCK_PETS } from "../../../lib/constants";

interface PatientSelectionTableProps {
  onSelect: (pet: Pet) => void;
  selectedPets: Pet[];
}

export function PatientSelectionTable({ onSelect, selectedPets }: PatientSelectionTableProps) {
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

  const hasSearchConditions = Object.values(searchParams).some(value => value.trim() !== "");

  const filteredPets = hasSearchConditions 
    ? MOCK_PETS.filter((pet) => {
        if (searchParams.ownerId && !pet.ownerId.includes(searchParams.ownerId)) return false;
        if (searchParams.ownerName && !pet.ownerName.includes(searchParams.ownerName)) return false;
        if (searchParams.phone && (!pet.phone || !pet.phone.includes(searchParams.phone))) return false; 
        if (searchParams.petName && !pet.name.includes(searchParams.petName)) return false;
        if (searchParams.species && !pet.species.includes(searchParams.species)) return false;
        return true;
      }).slice(0, 20) // Limit to 20 results for performance
    : [];

  const isSelected = (pet: Pet) => selectedPets.some((p) => p.id === pet.id);

  return (
    <div className="flex flex-col gap-4 h-full">
      {/* Search Criteria */}
      <div className="rounded-lg bg-white p-3 shadow-sm border border-[rgba(55,53,47,0.16)] shrink-0">
        <div className="flex items-center gap-2 mb-2">
            <Filter className="size-3 text-muted-foreground" />
            <h2 className="text-sm font-medium text-[#37352F]">検索条件</h2>
        </div>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-2 mb-2">
          <div className="space-y-0.5">
            <Label htmlFor="ownerId" className="text-sm text-[#37352F]/60">飼主No</Label>
            <Input
              id="ownerId"
              placeholder="例: 30042"
              value={searchParams.ownerId}
              onChange={(e) => setSearchParams({ ...searchParams, ownerId: e.target.value })}
              className="text-sm h-10 bg-white"
            />
          </div>
          <div className="space-y-0.5">
            <Label htmlFor="ownerName" className="text-sm text-[#37352F]/60">飼主名</Label>
            <Input
              id="ownerName"
              placeholder="例: 林 文明"
              value={searchParams.ownerName}
              onChange={(e) => setSearchParams({ ...searchParams, ownerName: e.target.value })}
              className="text-sm h-10 bg-white"
            />
          </div>
          <div className="space-y-0.5">
            <Label htmlFor="ownerNameKana" className="text-sm text-[#37352F]/60">飼主名(カナ)</Label>
            <Input
              id="ownerNameKana"
              placeholder="例: ハヤシ"
              value={searchParams.ownerNameKana}
              onChange={(e) => setSearchParams({ ...searchParams, ownerNameKana: e.target.value })}
              className="text-sm h-10 bg-white"
            />
          </div>
          <div className="space-y-0.5">
            <Label htmlFor="phone" className="text-sm text-[#37352F]/60">電話番号</Label>
            <Input
              id="phone"
              placeholder="例: 090..."
              value={searchParams.phone}
              onChange={(e) => setSearchParams({ ...searchParams, phone: e.target.value })}
              className="text-sm h-10 bg-white"
            />
          </div>
          <div className="space-y-0.5">
            <Label htmlFor="petName" className="text-sm text-[#37352F]/60">ペット名</Label>
            <Input
              id="petName"
              placeholder="例: Iris"
              value={searchParams.petName}
              onChange={(e) => setSearchParams({ ...searchParams, petName: e.target.value })}
              className="text-sm h-10 bg-white"
            />
          </div>
          <div className="space-y-0.5">
            <Label htmlFor="petNameKana" className="text-sm text-[#37352F]/60">ペット名(カナ)</Label>
            <Input
              id="petNameKana"
              placeholder="例: イリス"
              value={searchParams.petNameKana}
              onChange={(e) => setSearchParams({ ...searchParams, petNameKana: e.target.value })}
              className="text-sm h-10 bg-white"
            />
          </div>
          <div className="space-y-0.5">
            <Label htmlFor="species" className="text-sm text-[#37352F]/60">種別</Label>
            <Input
              id="species"
              placeholder="例: 犬"
              value={searchParams.species}
              onChange={(e) => setSearchParams({ ...searchParams, species: e.target.value })}
              className="text-sm h-10 bg-white"
            />
          </div>
           {/* Address omitted for brevity/space, kept key fields */}
        </div>
      </div>

      {/* Results Table */}
      <div className="flex-1 rounded-lg bg-white overflow-hidden shadow-sm border border-[rgba(55,53,47,0.16)] min-h-0">
        <div className="overflow-auto h-full">
            <Table>
            <TableHeader className="bg-[#F7F6F3] sticky top-0 z-10">
                <TableRow className="border-b border-[rgba(55,53,47,0.16)] h-10 hover:bg-[#F7F6F3]">
                <TableHead className="min-w-[80px] text-sm text-[#37352F]/60 h-10">飼主No</TableHead>
                <TableHead className="min-w-[120px] text-sm text-[#37352F]/60 h-10">飼主名</TableHead>
                <TableHead className="min-w-[100px] text-sm text-[#37352F]/60 h-10">ペット名</TableHead>
                <TableHead className="min-w-[60px] text-sm text-[#37352F]/60 h-10">種別</TableHead>
                <TableHead className="min-w-[60px] text-sm text-[#37352F]/60 h-10">性別</TableHead>
                <TableHead className="min-w-[80px] text-sm text-[#37352F]/60 h-10">生年月日</TableHead>
                <TableHead className="min-w-[60px] text-sm text-[#37352F]/60 h-10">体重</TableHead>
                <TableHead className="min-w-[60px] text-sm text-[#37352F]/60 h-10">操作</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {filteredPets.length === 0 ? (
                    <TableRow>
                        <TableCell colSpan={8} className="h-32 text-center text-sm text-muted-foreground">
                            {hasSearchConditions ? "条件に一致するペットが見つかりません" : "検索条件を入力して患者を検索してください"}
                        </TableCell>
                    </TableRow>
                ) : (
                    filteredPets.map((pet) => (
                    <TableRow
                        key={pet.id}
                        className={`transition-colors hover:bg-[rgba(55,53,47,0.06)] cursor-pointer h-10 ${
                        isSelected(pet) ? "bg-blue-50 hover:bg-blue-100" : ""
                        }`}
                        onClick={() => onSelect(pet)}
                    >
                        <TableCell className="text-sm py-1 font-mono">{pet.ownerId}</TableCell>
                        <TableCell className="text-sm py-1 font-medium">{pet.ownerName}</TableCell>
                        <TableCell className="text-sm py-1 font-bold">{pet.name}</TableCell>
                        <TableCell className="text-sm py-1">{pet.species}</TableCell>
                        <TableCell className="text-sm py-1">-</TableCell> {/* Gender not in mock */}
                        <TableCell className="text-sm py-1 font-mono">{pet.birthDate || "-"}</TableCell>
                        <TableCell className="text-sm py-1 font-mono">{pet.weight || "-"}</TableCell>
                        <TableCell className="py-1">
                        <Button
                            size="sm"
                            className={`h-10 gap-1 text-sm px-2 ${
                                isSelected(pet) 
                                ? "bg-blue-600 hover:bg-blue-700 text-white" 
                                : "bg-white border border-gray-300 text-gray-700 hover:bg-gray-50"
                            }`}
                            onClick={(e) => {
                                e.stopPropagation();
                                onSelect(pet);
                            }}
                        >
                            <Check className={`size-3 ${isSelected(pet) ? "" : "opacity-0"}`} />
                            {isSelected(pet) ? "選択中" : "選択"}
                        </Button>
                        </TableCell>
                    </TableRow>
                    ))
                )}
            </TableBody>
            </Table>
        </div>
      </div>
    </div>
  );
}
