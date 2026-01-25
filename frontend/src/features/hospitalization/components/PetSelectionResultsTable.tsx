import { Check } from "lucide-react";
import { Button } from "../../../components/ui/button";
import { Badge } from "../../../components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../../../components/ui/table";
import { Pet } from "../../../types";
import { H_STYLES } from "../styles";

interface PetSelectionResultsTableProps {
  pets: Pet[];
  onSelect: (pet: Pet) => void;
}

export const PetSelectionResultsTable = ({ pets, onSelect }: PetSelectionResultsTableProps) => {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className={`${H_STYLES.text.base} font-medium text-[#37352F]`}>検索結果</h2>
        <span className={`${H_STYLES.text.base} text-[#37352F]/60`}>
          {pets.length}件
        </span>
      </div>

      <div className="rounded-lg bg-white overflow-hidden shadow-sm border border-[rgba(55,53,47,0.16)]">
        <Table>
          <TableHeader>
            <TableRow 
              className="hover:bg-transparent bg-[#F7F6F3] border-b border-[rgba(55,53,47,0.16)] h-12"
            >
              <TableHead className={`min-w-[80px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>飼主No</TableHead>
              <TableHead className={`min-w-[140px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>飼主名</TableHead>
              <TableHead className={`min-w-[80px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>ペット番号</TableHead>
              <TableHead className={`min-w-[100px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>ペット名</TableHead>
              <TableHead className={`min-w-[50px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>生死</TableHead>
              <TableHead className={`min-w-[50px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>種</TableHead>
              <TableHead className={`min-w-[90px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>生年月日</TableHead>
              <TableHead className={`min-w-[60px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>体重</TableHead>
              <TableHead className={`min-w-[100px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>環境</TableHead>
              <TableHead className={`min-w-[90px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>前回来院</TableHead>
              <TableHead className={`min-w-[80px] ${H_STYLES.text.sm} text-[#37352F]/60 whitespace-nowrap h-12`}>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {pets.map((pet, index) => (
              <TableRow 
                key={pet.id}
                className={`transition-colors hover:bg-[rgba(55,53,47,0.06)] cursor-pointer h-12 ${
                  index < pets.length - 1 ? "border-b border-[rgba(55,53,47,0.09)]" : "border-none"
                }`}
                onClick={() => onSelect(pet)}
              >
                <TableCell className={`${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.ownerId}</TableCell>
                <TableCell className={`${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.ownerName}</TableCell>
                <TableCell className={`font-mono ${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.petNumber || "-"}</TableCell>
                <TableCell className={`${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.name}</TableCell>
                <TableCell className="whitespace-nowrap py-2">
                  {pet.status && (
                      <Badge
                        variant="secondary"
                        className={
                        pet.status === "生存"
                            ? `bg-[#DDEDEA] text-[#0F7B6C] border-[#DDEDEA] hover:bg-[#DDEDEA] ${H_STYLES.text.xs} px-2 py-0 h-7`
                            : `bg-[#EBECED] text-[#9B9A97] border-[#EBECED] hover:bg-[#EBECED] ${H_STYLES.text.xs} px-2 py-0 h-7`
                        }
                      >
                        {pet.status}
                      </Badge>
                  )}
                </TableCell>
                <TableCell className={`${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.species}</TableCell>
                <TableCell className={`font-mono ${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.birthDate || "-"}</TableCell>
                <TableCell className={`font-mono ${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.weight || "-"}</TableCell>
                <TableCell className={`${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.environment || "-"}</TableCell>
                <TableCell className={`font-mono ${H_STYLES.text.base} text-[#37352F] whitespace-nowrap py-2`}>{pet.lastVisit || "-"}</TableCell>
                <TableCell className="whitespace-nowrap py-2" onClick={(e) => e.stopPropagation()}>
                  <Button 
                    size="sm" 
                    className={`h-10 gap-1 bg-[#37352F] hover:bg-[#37352F]/90 text-white ${H_STYLES.button.action} px-4`} 
                    onClick={() => onSelect(pet)}
                  >
                    <Check className={H_STYLES.button.icon} />
                    選択
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {pets.length === 0 && (
              <TableRow>
                <TableCell colSpan={11} className={`h-24 text-center ${H_STYLES.text.base} text-[#37352F]/60`}>
                  該当するペットが見つかりません
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
};
