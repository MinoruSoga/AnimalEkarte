import { Input } from "../../../components/ui/input";
import { Label } from "../../../components/ui/label";
import { Textarea } from "../../../components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";
import { RadioGroup, RadioGroupItem } from "../../../components/ui/radio-group";
import { Calendar } from "../../../components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "../../../components/ui/popover";
import { Calendar as CalendarIcon } from "lucide-react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { cn } from "../../../components/ui/utils";
import { ReservationAppointment } from "../../../types";
import { useMasterItems } from "../../master/hooks/useMasterItems";

interface ReservationFormFieldsProps {
  formData: Partial<ReservationAppointment>;
  onChange: (data: Partial<ReservationAppointment>) => void;
}

export const ReservationFormFields = ({
  formData,
  onChange,
}: ReservationFormFieldsProps) => {
  const { data: serviceTypes } = useMasterItems("serviceType");

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label className="text-sm text-[#37352F]/60">日付</Label>
          <Popover>
            <PopoverTrigger asChild>
              <button
                type="button"
                className={cn(
                  "flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-1 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 border-[rgba(55,53,47,0.16)] text-[#37352F]",
                  !formData.start && "text-muted-foreground"
                )}
              >
                <span className="flex items-center">
                  <CalendarIcon className="mr-2 h-4 w-4" />
                  {formData.start ? (
                    format(formData.start, "yyyy/MM/dd (E)", { locale: ja })
                  ) : (
                    <span>日付を選択</span>
                  )}
                </span>
              </button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0">
              <Calendar
                mode="single"
                selected={formData.start}
                onSelect={(date) => {
                  if (!date) return;
                  const newStart = new Date(date);
                  const newEnd = new Date(date);
                  
                  if (formData.start) {
                    newStart.setHours(formData.start.getHours(), formData.start.getMinutes());
                  }
                  if (formData.end) {
                    newEnd.setHours(formData.end.getHours(), formData.end.getMinutes());
                  }
                  
                  onChange({ ...formData, start: newStart, end: newEnd });
                }}
                initialFocus
              />
            </PopoverContent>
          </Popover>
        </div>

        <div className="space-y-1.5">
          <Label className="text-sm text-[#37352F]/60">時間</Label>
          <div className="flex items-center gap-2">
            <Select
              value={formData.start ? format(formData.start, "H:mm") : "10:00"}
              onValueChange={(v) => {
                if (!formData.start) return;
                const [h, m] = v.split(":").map(Number);
                const newStart = new Date(formData.start);
                newStart.setHours(h, m);
                onChange({ ...formData, start: newStart });
              }}
            >
              <SelectTrigger className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="max-h-[200px]">
                {Array.from({ length: 48 }).map((_, i) => {
                  const h = Math.floor(i / 2);
                  const m = i % 2 === 0 ? "00" : "30";
                  const time = `${h}:${m}`;
                  return (
                    <SelectItem key={time} value={time}>
                      {time}
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
            <span className="text-sm text-[#37352F]">〜</span>
            <Select
              value={formData.end ? format(formData.end, "H:mm") : "11:00"}
              onValueChange={(v) => {
                if (!formData.end) return;
                const [h, m] = v.split(":").map(Number);
                const newEnd = new Date(formData.end);
                newEnd.setHours(h, m);
                onChange({ ...formData, end: newEnd });
              }}
            >
              <SelectTrigger className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="max-h-[200px]">
                {Array.from({ length: 48 }).map((_, i) => {
                  const h = Math.floor(i / 2);
                  const m = i % 2 === 0 ? "00" : "30";
                  const time = `${h}:${m}`;
                  return (
                    <SelectItem key={time} value={time}>
                      {time}
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label className="text-sm text-[#37352F]/60">予約区分</Label>
          <Select
            value={formData.type}
            onValueChange={(v: any) => onChange({ ...formData, type: v })}
          >
            <SelectTrigger className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {serviceTypes.map((item) => (
                <SelectItem key={item.id} value={item.name}>
                  {item.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-1.5">
          <Label className="text-sm text-[#37352F]/60">初診/再診</Label>
          <RadioGroup
            value={formData.visitType}
            onValueChange={(v: any) => onChange({ ...formData, visitType: v })}
            className="flex gap-4 pt-1"
          >
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="first" id="first" className="text-[#37352F]" />
              <Label htmlFor="first" className="text-sm text-[#37352F] cursor-pointer">初診</Label>
            </div>
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="revisit" id="revisit" className="text-[#37352F]" />
              <Label htmlFor="revisit" className="text-sm text-[#37352F] cursor-pointer">再診</Label>
            </div>
          </RadioGroup>
        </div>
      </div>

      <div className="space-y-1.5">
        <Label className="text-sm text-[#37352F]/60">担当者</Label>
        <Select
          value={formData.doctor}
          onValueChange={(v) => onChange({ ...formData, doctor: v })}
        >
          <SelectTrigger className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]">
            <SelectValue placeholder="担当者を選択" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="医師A">医師A</SelectItem>
            <SelectItem value="医師B">医師B</SelectItem>
            <SelectItem value="スタッフA">スタッフA</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1.5">
        <Label className="text-sm text-[#37352F]/60">メモ</Label>
        <Textarea
          value={formData.notes || ""}
          onChange={(e) => onChange({ ...formData, notes: e.target.value })}
          placeholder="詳細や備考を入力..."
          className="min-h-[80px] text-sm resize-none bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]"
        />
      </div>
    </div>
  );
};
