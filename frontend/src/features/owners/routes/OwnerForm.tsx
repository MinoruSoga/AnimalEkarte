// React/Framework
import { useNavigate, useParams } from "react-router";

// External
import {
  Plus,
  Edit,
  User,
  PawPrint,
  MoreHorizontal,
  Calendar,
  FileText,
  Scissors,
  Bed,
  CreditCard
} from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { PageLayout } from "@/components/shared/PageLayout";

// Relative
import { PetEditModal } from "../components/PetEditModal";
import { useOwnerForm } from "../hooks/useOwnerForm";

export const OwnerForm = () => {
  const navigate = useNavigate();
  const { id: ownerId } = useParams();
  
  const {
      isEdit,
      ownerData,
      setOwnerData,
      pets,
      petModalOpen,
      setPetModalOpen,
      editingPet,
      handleAddPet,
      handleEditPet,
      handleSavePet,
      handleSave
  } = useOwnerForm(ownerId);

  const handleBack = () => {
    navigate("/owners");
  };

  return (
    <PageLayout
      title={isEdit ? "飼主・ペット　編集" : "飼主・ペット　登録"}
      onBack={handleBack}
      maxWidth="max-w-[1400px]"
      headerAction={
          <Button
            size="sm"
            onClick={handleSave}
            className="px-4"
          >
            {isEdit ? "更新" : "登録"}
          </Button>
      }
    >
          {/* Owner Information Form */}
        <div className="mb-4 rounded-lg bg-card p-4 shadow-sm border border-border">
          <h2 className="mb-3 text-sm font-bold flex items-center gap-2">
            <User className="h-4 w-4 text-muted-foreground" />
            飼主情報
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {/* Row 1 */}
            <div className="space-y-1.5">
              <Label htmlFor="ownerId" className="text-muted-foreground">
                飼主No
              </Label>
              <Input
                id="ownerId"
                value={ownerData.ownerId}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, ownerId: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="postalCode" className="text-muted-foreground">
                郵便番号
              </Label>
              <Input
                id="postalCode"
                placeholder="123-4567"
                value={ownerData.postalCode}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, postalCode: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="company" className="text-muted-foreground">
                会社名
              </Label>
              <Input
                id="company"
                value={ownerData.company}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, company: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-muted-foreground">会員区分</Label>
              <div className="flex gap-1.5 flex-wrap">
                {["非会員", "会員", "退亡者", "他診/準"].map((type) => (
                  <Button
                    key={type}
                    type="button"
                    variant={
                      ownerData.membershipType === type ? "default" : "outline"
                    }
                    size="sm"
                    className="h-10 px-3"
                    onClick={() =>
                      setOwnerData({ ...ownerData, membershipType: type })
                    }
                  >
                    {type}
                  </Button>
                ))}
              </div>
            </div>

            {/* Row 2 */}
            <div className="space-y-1.5">
              <Label htmlFor="ownerName" className="text-muted-foreground">
                飼主名 <span className="text-destructive">*</span>
              </Label>
              <Input
                id="ownerName"
                value={ownerData.ownerName}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, ownerName: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="address1" className="text-muted-foreground">
                住所1
              </Label>
              <Input
                id="address1"
                value={ownerData.address1}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, address1: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="postalNumber" className="text-muted-foreground">
                郵便番号
              </Label>
              <Input
                id="postalNumber"
                value={ownerData.postalNumber}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, postalNumber: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5 col-span-2 lg:col-span-1 lg:row-span-3">
              <Label className="text-muted-foreground">危険人物</Label>
              <div className="flex items-center space-x-2 mb-2 h-10">
                <Switch
                  id="dangerous"
                  checked={ownerData.isDangerous}
                  onCheckedChange={(checked) =>
                    setOwnerData({
                      ...ownerData,
                      isDangerous: checked as boolean,
                    })
                  }
                  className="origin-left mr-2"
                />
                <label
                  htmlFor="dangerous"
                  className="text-sm cursor-pointer"
                >
                  該当する
                </label>
              </div>
              <Label htmlFor="remarks" className="text-muted-foreground">
                備考・特記事項
              </Label>
              <Textarea
                id="remarks"
                rows={6}
                value={ownerData.remarks}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, remarks: e.target.value })
                }
                className="min-h-[140px] resize-none p-3"
              />
            </div>

            {/* Row 3 */}
            <div className="space-y-1.5">
              <Label htmlFor="ownerNameKana" className="text-muted-foreground">
                飼主名(カナ) <span className="text-destructive">*</span>
              </Label>
              <Input
                id="ownerNameKana"
                placeholder="ハヤシ フミアキ"
                value={ownerData.ownerNameKana}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, ownerNameKana: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="address2" className="text-muted-foreground">
                住所2
              </Label>
              <Input
                id="address2"
                value={ownerData.address2}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, address2: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="homeAddress1" className="text-muted-foreground">
                住所1
              </Label>
              <Input
                id="homeAddress1"
                value={ownerData.homeAddress1}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, homeAddress1: e.target.value })
                }
              />
            </div>

            {/* Row 4 */}
            <div className="space-y-1.5">
              <Label htmlFor="birthDate" className="text-muted-foreground">
                飼主生年月日
              </Label>
              <Input
                id="birthDate"
                type="date"
                value={ownerData.birthDate}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, birthDate: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="email" className="text-muted-foreground">
                メールアドレス
              </Label>
              <Input
                id="email"
                type="email"
                value={ownerData.email}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, email: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="homeAddress2" className="text-muted-foreground">
                住所2
              </Label>
              <Input
                id="homeAddress2"
                value={ownerData.homeAddress2}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, homeAddress2: e.target.value })
                }
              />
            </div>

            {/* Row 5 */}
            <div className="space-y-1.5">
              <Label htmlFor="phone" className="text-muted-foreground">
                電話番号 <span className="text-destructive">*</span>
              </Label>
              <Input
                id="phone"
                placeholder="090-1234-5678"
                value={ownerData.phone}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, phone: e.target.value })
                }
              />
            </div>
            <div className="space-y-1.5 col-span-1 lg:col-span-2">
              <Label htmlFor="companyPhone" className="text-muted-foreground">
                会社 電話番号
              </Label>
              <Input
                id="companyPhone"
                value={ownerData.companyPhone}
                onChange={(e) =>
                  setOwnerData({ ...ownerData, companyPhone: e.target.value })
                }
              />
            </div>
          </div>
        </div>

        {/* Pet Information */}
        <div className="mb-4 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-bold flex items-center gap-2">
              <PawPrint className="h-4 w-4 text-muted-foreground" />
              ペット情報
            </h2>
            <Button
              size="sm"
              onClick={handleAddPet}
              className="gap-1.5 px-4"
            >
              <Plus className="size-4" />
              ペット追加
            </Button>
          </div>

          <div
            className="rounded-lg bg-card overflow-hidden shadow-sm border border-border"
          >
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent bg-muted/50 border-b border-border h-12">
                  <TableHead className="w-[120px] text-sm font-medium text-muted-foreground h-12">
                    ペット番号
                  </TableHead>
                  <TableHead className="w-[150px] text-sm font-medium text-muted-foreground h-12">
                    ペット名
                  </TableHead>
                  <TableHead className="w-[60px] text-sm font-medium text-muted-foreground h-12">
                    生死
                  </TableHead>
                  <TableHead className="w-[60px] text-sm font-medium text-muted-foreground h-12">
                    種別
                  </TableHead>
                  <TableHead className="w-[60px] text-sm font-medium text-muted-foreground h-12">
                    性別
                  </TableHead>
                  <TableHead className="w-[100px] text-sm font-medium text-muted-foreground h-12">
                    生年月日
                  </TableHead>
                  <TableHead className="w-[80px] text-sm font-medium text-muted-foreground h-12">
                    毛色
                  </TableHead>
                  <TableHead className="w-[80px] text-sm font-medium text-muted-foreground h-12">
                    体重
                  </TableHead>
                  <TableHead className="w-[150px] text-sm font-medium text-muted-foreground h-12">
                    環境
                  </TableHead>
                  <TableHead className="w-[200px] text-sm font-medium text-muted-foreground h-12">
                    備考
                  </TableHead>
                  <TableHead className="w-[80px] text-sm font-medium text-muted-foreground h-12">
                    操作
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pets.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={11} className="text-center py-8 text-sm text-muted-foreground">
                      ペット情報がありません。「ペット追加」ボタンから追加してください。
                    </TableCell>
                  </TableRow>
                ) : (
                  pets.map((pet) => (
                    <TableRow
                      key={pet.id}
                      className="transition-colors border-b-border hover:bg-muted/50 h-12"
                    >
                      <TableCell className="font-mono text-sm py-2">
                        {pet.petNumber}
                      </TableCell>
                      <TableCell className="text-sm py-2">{pet.petName}</TableCell>
                      <TableCell className="text-sm py-2">{pet.status}</TableCell>
                      <TableCell className="text-sm py-2">{pet.species}</TableCell>
                      <TableCell className="text-sm py-2">{pet.gender}</TableCell>
                      <TableCell className="font-mono text-sm py-2">
                        {pet.birthDate}
                      </TableCell>
                      <TableCell className="text-sm py-2">{pet.color}</TableCell>
                      <TableCell className="font-mono text-sm py-2">
                        {pet.weight}
                      </TableCell>
                      <TableCell className="text-sm py-2">{pet.environment}</TableCell>
                      <TableCell className="text-sm truncate max-w-[200px] py-2">
                        {pet.remarks}
                      </TableCell>
                      <TableCell className="py-2">
                        <div className="flex gap-1 justify-end">
                          <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                  <Button
                                      variant="ghost"
                                      size="icon"
                                  >
                                      <span className="sr-only">Open menu</span>
                                      <MoreHorizontal className="size-5" />
                                  </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                  <DropdownMenuLabel>操作</DropdownMenuLabel>
                                  <DropdownMenuItem onClick={() => handleEditPet(pet)}>
                                      <Edit className="mr-2 h-4 w-4" />
                                      詳細・編集
                                  </DropdownMenuItem>
                                  <DropdownMenuItem onClick={() => navigate(`/reservations?petId=${pet.id}`)}>
                                      <Calendar className="mr-2 h-4 w-4" />
                                      予約作成
                                  </DropdownMenuItem>
                                  <DropdownMenuItem onClick={() => navigate(`/medical-records/new?petId=${pet.id}`, { state: { from: ownerId ? `/owners/${ownerId}` : "/owners" } })}>
                                      <FileText className="mr-2 h-4 w-4" />
                                      カルテ作成
                                  </DropdownMenuItem>
                                  <DropdownMenuItem onClick={() => navigate(`/trimming/new?petId=${pet.id}`, { state: { from: ownerId ? `/owners/${ownerId}` : "/owners" } })}>
                                      <Scissors className="mr-2 h-4 w-4" />
                                      トリミング
                                  </DropdownMenuItem>
                                  <DropdownMenuItem onClick={() => navigate(`/hospitalization/new?petId=${pet.id}`, { state: { from: ownerId ? `/owners/${ownerId}` : "/owners" } })}>
                                      <Bed className="mr-2 h-4 w-4" />
                                      入院登録
                                  </DropdownMenuItem>
                                  <DropdownMenuItem onClick={() => navigate(`/accounting/new?petId=${pet.id}`, { state: { from: ownerId ? `/owners/${ownerId}` : "/owners" } })}>
                                      <CreditCard className="mr-2 h-4 w-4" />
                                      会計登録
                                  </DropdownMenuItem>
                              </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>

        <PetEditModal
          open={petModalOpen}
          onOpenChange={setPetModalOpen}
          ownerName={ownerData.ownerName || "飼主名"}
          petData={
            editingPet
              ? {
                  petNumber: editingPet.petNumber,
                  petName: editingPet.petName,
                  petNameKana: editingPet.petNameKana || "",
                  species: editingPet.species,
                  gender: editingPet.gender,
                  birthDate: editingPet.birthDate,
                  breed: editingPet.breed || "",
                  color: editingPet.color,
                  neuteredDate: editingPet.neuteredDate || "",
                  acquisitionType: editingPet.acquisitionType || "",
                  dangerLevel: editingPet.dangerLevel || "",
                  food: editingPet.food || "",
                  remarks: editingPet.remarks,
                }
              : undefined
          }
          onSave={handleSavePet}
        />
    </PageLayout>
  );
}
