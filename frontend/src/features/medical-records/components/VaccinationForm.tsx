// React/Framework
import React from "react";

// Internal
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

interface VaccinationFormProps {
  vaccineName: string;
  setVaccineName: (v: string) => void;
  date: string;
  setDate: (v: string) => void;
  supplemental: string;
  setSupplemental: (v: string) => void;
  lot1: string;
  setLot1: (v: string) => void;
  lot2: string;
  setLot2: (v: string) => void;
  lot3: string;
  setLot3: (v: string) => void;
  lot4: string;
  setLot4: (v: string) => void;
  nextScheduleType: string;
  setNextScheduleType: (v: string) => void;
  nextDate: string;
  setNextDate: (v: string) => void;
  remarks: string;
  setRemarks: (v: string) => void;
}

export const VaccinationForm = React.memo(function VaccinationForm({
  vaccineName,
  setVaccineName,
  date,
  setDate,
  supplemental,
  setSupplemental,
  lot1,
  setLot1,
  lot2,
  setLot2,
  lot3,
  setLot3,
  lot4,
  setLot4,
  nextScheduleType,
  setNextScheduleType,
  nextDate,
  setNextDate,
  remarks,
  setRemarks,
}: VaccinationFormProps) {
  return (
    <div className="col-span-6 flex flex-col gap-4">
      {/* Row 1: Name and Date */}
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm font-medium text-[#37352F]/60">
            予防接種名
          </Label>
          <Select value={vaccineName} onValueChange={setVaccineName}>
            <SelectTrigger className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]">
              <SelectValue placeholder="選択してください" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="esophagitis">食道炎</SelectItem>
              <SelectItem value="rabies">狂犬病ワクチン</SelectItem>
              <SelectItem value="distemper">ジステンパー</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm font-medium text-[#37352F]/60">
            予防接種日
          </Label>
          <Input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
          />
        </div>
      </div>

      {/* Supplemental */}
      <div className="flex flex-col gap-1.5">
        <Label className="text-sm font-medium text-[#37352F]/60">
          補助説明
        </Label>
        <Input
          value={supplemental}
          onChange={(e) => setSupplemental(e.target.value)}
          className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
        />
      </div>

      {/* LOT Numbers */}
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm font-medium text-[#37352F]/60">
            LOT1
          </Label>
          <Input
            value={lot1}
            onChange={(e) => setLot1(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm font-medium text-[#37352F]/60">
            LOT2
          </Label>
          <Input
            value={lot2}
            onChange={(e) => setLot2(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm font-medium text-[#37352F]/60">
            LOT3
          </Label>
          <Input
            value={lot3}
            onChange={(e) => setLot3(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm font-medium text-[#37352F]/60">
            LOT4
          </Label>
          <Input
            value={lot4}
            onChange={(e) => setLot4(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
          />
        </div>
      </div>

      {/* Next Schedule Type */}
      <div className="flex flex-col gap-1.5">
        <Label className="text-sm font-medium text-[#37352F]/60">
          次回予防接種予定設定
        </Label>
        <RadioGroup
          value={nextScheduleType}
          onValueChange={setNextScheduleType}
          className="flex flex-row gap-6 pt-2"
        >
          <div className="flex items-center space-x-2">
            <RadioGroupItem
              value="3weeks"
              id="r1"
              className="border-gray-400 text-[#37352F]"
            />
            <Label
              htmlFor="r1"
              className="font-normal text-[#37352F] text-sm cursor-pointer"
            >
              3週間後
            </Label>
          </div>
          <div className="flex items-center space-x-2">
            <RadioGroupItem
              value="4weeks"
              id="r2"
              className="border-gray-400 text-[#37352F]"
            />
            <Label
              htmlFor="r2"
              className="font-normal text-[#37352F] text-sm cursor-pointer"
            >
              4週間後
            </Label>
          </div>
          <div className="flex items-center space-x-2">
            <RadioGroupItem
              value="1year"
              id="r3"
              className="border-gray-400 text-[#37352F]"
            />
            <Label
              htmlFor="r3"
              className="font-normal text-[#37352F] text-sm cursor-pointer"
            >
              1年後
            </Label>
          </div>
          <div className="flex items-center space-x-2">
            <RadioGroupItem
              value="other"
              id="r4"
              className="border-gray-400 text-[#37352F]"
            />
            <Label
              htmlFor="r4"
              className="font-normal text-[#37352F] text-sm cursor-pointer"
            >
              以外
            </Label>
          </div>
        </RadioGroup>
      </div>

      {/* Next Date */}
      <div className="flex flex-col gap-1.5">
        <Label className="text-sm font-medium text-[#37352F]/60">
          次回予定日
        </Label>
        <Input
          type="date"
          value={nextDate}
          onChange={(e) => setNextDate(e.target.value)}
          className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
        />
      </div>

      {/* Remarks */}
      <div className="flex flex-col gap-1.5 flex-1 min-h-0">
        <Label className="text-sm font-medium text-[#37352F]/60">
          備考
        </Label>
        <Textarea
          value={remarks}
          onChange={(e) => setRemarks(e.target.value)}
          className="flex-1 resize-none bg-white border-[rgba(55,53,47,0.16)] p-3 text-sm text-[#37352F] leading-relaxed"
        />
      </div>
    </div>
  );
});
