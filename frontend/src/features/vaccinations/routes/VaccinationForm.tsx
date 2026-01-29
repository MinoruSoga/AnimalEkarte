// React/Framework
import { useEffect } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import { Trash2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout";

// Relative
import { useVaccinationForm } from "../hooks/useVaccinationForm";
import { useMasterItems } from "@/hooks/use-master-items";

export const VaccinationForm = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");

  const { data: vaccineItems } = useMasterItems("vaccine");

  const {
      isEdit,
      petSelection,
      form,
      historyFilter,
      handleSave
  } = useVaccinationForm(id);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];

  useEffect(() => {
    // Redirect only if:
    // 1. No pet is selected
    // 2. Not in edit mode
    // 3. No petId in URL (if petId exists, we wait for it to load)
    if (!selectedPet && !isEdit && !petId) {
        navigate("/vaccinations/select-pet");
    }
  }, [selectedPet, isEdit, navigate, petId]);

  if (!selectedPet && !isEdit && petId) return null;
  if (!selectedPet && !isEdit) return null;

  const {
      vaccineName, setVaccineName,
      date, setDate,
      supplemental, setSupplemental,
      lot1, setLot1,
      lot2, setLot2,
      lot3, setLot3,
      lot4, setLot4,
      nextScheduleType, setNextScheduleType,
      nextDate, setNextDate,
      remarks, setRemarks
  } = form;

  const {
      filterStartDate, setFilterStartDate,
      filterEndDate, setFilterEndDate,
      historySearchTerm, setHistorySearchTerm,
      sortOrder, setSortOrder
  } = historyFilter;

  const handleBack = () => {
    if (location.state?.from) {
        navigate(location.state.from);
    } else {
        navigate("/vaccinations");
    }
  };

  // Mock Data for History (kept in component for now as hook doesn't provide it yet)
  const historyItems = [
    { id: 1, name: "フィラリア薬", date: "24/4/6", next: "24/4/6" },
    { id: 2, name: "ノミダニ予防薬", date: "25/6/15", next: "-" },
    { id: 3, name: "ノミダニ予防薬", date: "25/1/15", next: "25/6/15" },
    { id: 4, name: "ノミダニ予防薬", date: "25/1/15", next: "25/6/15" },
    { id: 5, name: "狂犬病ワクチン", date: "25/1/20", next: "25/6/20" },
    { id: 6, name: "ジステンパーワクチン", date: "25/1/25", next: "25/6/25" },
    { id: 7, name: "パルボウイルスワクチン", date: "25/1/30", next: "25/6/30" },
    { id: 8, name: "猫ウイルス性鼻気管炎ワクチン", date: "25/1/35", next: "25/6/35" },
    { id: 9, name: "猫クラシウイルスワクチン", date: "25/1/40", next: "25/6/40" },
    { id: 10, name: "レプトスピラワクチン", date: "25/1/45", next: "25/6/45" },
    { id: 11, name: "コロナウイルスワクチン", date: "25/1/50", next: "25/6/50" },
    { id: 12, name: "フェライン・レトロウイルスワクチン", date: "25/1/55", next: "25/6/55" },
    { id: 13, name: "パルボウイルスキャリアワクチン", date: "25/1/60", next: "25/6/60" },
  ];

  return (
    <PageLayout
      title={isEdit ? "予防接種詳細・編集" : "新規予防接種登録"}
      onBack={handleBack}
      maxWidth="max-w-[1440px]"
      headerAction={
        <div className="flex gap-2">
            {isEdit && (
                <Button 
                    variant="ghost"
                    className="text-red-600 hover:text-red-700 hover:bg-red-50 px-4 h-10 text-sm"
                >
                    <Trash2 className="mr-1.5 size-4" />
                    削除
                </Button>
            )}
            <Button 
                onClick={handleSave}
                className="bg-[#37352F] hover:bg-[#37352F]/90 text-white shadow-sm px-6 h-10 text-sm"
            >
                保存
            </Button>
        </div>
      }
    >
        {selectedPet && (
            <PatientInfoCard
              ownerName={selectedPet.ownerName}
              petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
              petNumber={selectedPet.petNumber || selectedPet.id}
              weight={selectedPet.weight || "-"}
              staffName="医師A"
              serviceType="予防接種"
              petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
              insuranceName={selectedPet.insuranceName || "保険情報未登録"}
              insuranceDetails={selectedPet.insuranceDetails || "-"}
              nextVisitDate="-"
              nextVisitContent="-"
              className="mb-4"
            />
        )}

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 pb-16">
            {/* Left Column: Form */}
                    <div className="col-span-1 lg:col-span-6 flex flex-col gap-4">
                        <div className="bg-white rounded-lg p-4 border border-[rgba(55,53,47,0.16)] shadow-sm space-y-4">
                            {/* Vaccination Form Fields (Matching MedicalRecordVaccination) */}
                            <div className="grid grid-cols-2 gap-4">
                                <div className="flex flex-col gap-1.5">
                                    <Label className="text-sm text-[#37352F]/60">予防接種名</Label>
                                    <Select value={vaccineName} onValueChange={setVaccineName}>
                                    <SelectTrigger className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]">
                                        <SelectValue placeholder="選択してください" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {vaccineItems.map((item) => (
                                            <SelectItem key={item.id} value={item.name}>
                                                {item.name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                    </Select>
                                </div>
                                <div className="flex flex-col gap-1.5">
                                    <Label className="text-sm text-[#37352F]/60">予防接種日</Label>
                                    <Input
                                    type="date"
                                    value={date}
                                    onChange={(e) => setDate(e.target.value)}
                                    className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]"
                                    />
                                </div>
                            </div>

                            {/* Supplemental */}
                            <div className="flex flex-col gap-1.5">
                            <Label className="text-sm text-[#37352F]/60">補助説明</Label>
                            <Input
                                value={supplemental}
                                onChange={(e) => setSupplemental(e.target.value)}
                                className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]"
                            />
                            </div>

                            {/* LOT Numbers */}
                            <div className="grid grid-cols-4 gap-2">
                            <div className="flex flex-col gap-1.5">
                                <Label className="text-sm text-[#37352F]/60">LOT1</Label>
                                <Input
                                value={lot1}
                                onChange={(e) => setLot1(e.target.value)}
                                className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]"
                                />
                            </div>
                            <div className="flex flex-col gap-1.5">
                                <Label className="text-sm text-[#37352F]/60">LOT2</Label>
                                <Input
                                value={lot2}
                                onChange={(e) => setLot2(e.target.value)}
                                className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]"
                                />
                            </div>
                            <div className="flex flex-col gap-1.5">
                                <Label className="text-sm text-[#37352F]/60">LOT3</Label>
                                <Input
                                value={lot3}
                                onChange={(e) => setLot3(e.target.value)}
                                className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]"
                                />
                            </div>
                            <div className="flex flex-col gap-1.5">
                                <Label className="text-sm text-[#37352F]/60">LOT4</Label>
                                <Input
                                value={lot4}
                                onChange={(e) => setLot4(e.target.value)}
                                className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]"
                                />
                            </div>
                            </div>

                            {/* Next Schedule Type */}
                            <div className="flex flex-col gap-1.5">
                            <Label className="text-sm text-[#37352F]/60">次回予防接種予定設定</Label>
                            <RadioGroup
                                value={nextScheduleType}
                                onValueChange={setNextScheduleType}
                                className="flex flex-row gap-4 flex-wrap pt-1"
                            >
                                <div className="flex items-center space-x-1.5">
                                <RadioGroupItem value="3weeks" id="r1" className="text-[#37352F]" />
                                <Label htmlFor="r1" className="font-normal text-sm text-[#37352F]">3週間後</Label>
                                </div>
                                <div className="flex items-center space-x-1.5">
                                <RadioGroupItem value="4weeks" id="r2" className="text-[#37352F]" />
                                <Label htmlFor="r2" className="font-normal text-sm text-[#37352F]">4週間後</Label>
                                </div>
                                <div className="flex items-center space-x-1.5">
                                <RadioGroupItem value="1year" id="r3" className="text-[#37352F]" />
                                <Label htmlFor="r3" className="font-normal text-sm text-[#37352F]">1年後</Label>
                                </div>
                                <div className="flex items-center space-x-1.5">
                                <RadioGroupItem value="other" id="r4" className="text-[#37352F]" />
                                <Label htmlFor="r4" className="font-normal text-sm text-[#37352F]">以外</Label>
                                </div>
                            </RadioGroup>
                            </div>

                            {/* Next Date */}
                            <div className="flex flex-col gap-1.5">
                            <Label className="text-sm text-[#37352F]/60">次回予定日</Label>
                            <Input
                                type="date"
                                value={nextDate}
                                onChange={(e) => setNextDate(e.target.value)}
                                className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]"
                            />
                            </div>

                            {/* Remarks */}
                            <div className="flex flex-col gap-1.5 flex-1">
                            <Label className="text-sm text-[#37352F]/60">備考</Label>
                            <Textarea
                                value={remarks}
                                onChange={(e) => setRemarks(e.target.value)}
                                className="resize-none bg-white border-[rgba(55,53,47,0.16)] min-h-[100px] text-sm text-[#37352F]"
                            />
                            </div>
                        </div>
                    </div>

                    {/* Right Column: History */}
                    <div className="col-span-1 lg:col-span-6 flex flex-col gap-3">
                        <h2 className="text-sm font-bold text-[#37352F]">予防接種履歴</h2>

                        {/* Filters */}
                        <div className="space-y-3 bg-white p-3 rounded-lg border border-[rgba(55,53,47,0.16)]">
                            <div className="flex flex-col gap-1.5">
                                <Label className="text-sm text-[#37352F]/60">実施日</Label>
                                <div className="flex items-center gap-2">
                                    <Input 
                                        type="date"
                                        value={filterStartDate}
                                        onChange={(e) => setFilterStartDate(e.target.value)}
                                        className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
                                    />
                                    <span className="text-[#37352F] text-sm">〜</span>
                                    <Input 
                                        type="date"
                                        value={filterEndDate}
                                        onChange={(e) => setFilterEndDate(e.target.value)}
                                        className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-col gap-1.5">
                                <Label className="text-sm text-[#37352F]/60">検索単語</Label>
                                <div className="flex gap-2">
                                    <Input
                                        value={historySearchTerm}
                                        onChange={(e) => setHistorySearchTerm(e.target.value)}
                                        className="flex-1 bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
                                        placeholder="検索..."
                                    />
                                    <Button variant="outline" className="h-10 text-sm bg-black text-white hover:bg-black/90 hover:text-white border-none">
                                        クリア
                                    </Button>
                                    <Select value={sortOrder} onValueChange={setSortOrder}>
                                        <SelectTrigger className="w-[80px] h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="desc">降順</SelectItem>
                                            <SelectItem value="asc">昇順</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>
                        </div>

                        {/* Table */}
                        <div className="border border-[rgba(55,53,47,0.16)] rounded-lg bg-white overflow-hidden flex flex-col min-h-[400px]">
                            {/* Header */}
                            <div className="flex items-center border-b border-[rgba(55,53,47,0.16)] bg-white text-sm font-bold text-[#0f172a] h-10 shrink-0 bg-[#F7F6F3]">
                                <div className="flex-1 px-3 text-center">予防接種名</div>
                                <div className="w-[100px] px-2 text-center border-l border-[rgba(55,53,47,0.16)]">実施日</div>
                                <div className="w-[100px] px-2 text-center border-l border-[rgba(55,53,47,0.16)]">次予定</div>
                                <div className="w-[80px] px-2 text-center border-l border-[rgba(55,53,47,0.16)]">操作</div>
                            </div>

                            {/* Scrollable Rows */}
                            <div className="flex-1 overflow-y-auto">
                                {selectedPets.length > 0 ? (
                                    historyItems.map((item) => (
                                        <div key={item.id} className="flex items-center border-b border-[rgba(55,53,47,0.16)] bg-white text-sm text-[#0f172a] h-10 hover:bg-gray-50">
                                            <div className="flex-1 px-3 truncate">{item.name}</div>
                                            <div className="w-[100px] px-2 text-center border-l border-[rgba(55,53,47,0.16)] font-mono">{item.date}</div>
                                            <div className="w-[100px] px-2 text-center border-l border-[rgba(55,53,47,0.16)] font-mono">{item.next}</div>
                                            <div className="w-[80px] px-2 flex justify-center border-l border-[rgba(55,53,47,0.16)]">
                                                <Button 
                                                    variant="outline" 
                                                    size="sm" 
                                                    className="h-10 w-[50px] text-sm bg-black text-white hover:bg-black/90 hover:text-white border-none px-0"
                                                >
                                                    複製
                                                </Button>
                                            </div>
                                        </div>
                                    ))
                                ) : (
                                    <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
                                        ペットを選択してください
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
    </PageLayout>
  );
}
