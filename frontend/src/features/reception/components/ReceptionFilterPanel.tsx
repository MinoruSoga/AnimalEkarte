import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C } from "@/lib/design-tokens";

interface DoctorOption {
  id: string;
  name: string;
}

interface ReceptionFilterPanelProps {
  selectedVisitTypes: string[];
  selectedDoctor: string;
  doctors: DoctorOption[];
  isTrimmingOnly: boolean;
  onToggleVisitType: (type: string) => void;
  onSelectedDoctorChange: (doctor: string) => void;
  onTrimmingOnlyChange: (value: boolean) => void;
}

export function ReceptionFilterPanel({
  selectedVisitTypes,
  selectedDoctor,
  doctors,
  isTrimmingOnly,
  onToggleVisitType,
  onSelectedDoctorChange,
  onTrimmingOnlyChange,
}: ReceptionFilterPanelProps) {
  return (
    <div className={`${C.bgWhite} border-b border-border px-6 py-4 animate-in slide-in-from-top-1 fade-in duration-200`}>
      <div className="flex flex-wrap gap-8">
        <div className="space-y-2">
          <h4 className={`font-bold text-base ${C.text}`}>診察区分</h4>
          <div className="flex gap-4">
            {["初診", "再診"].map((type) => (
              <div key={type} className="flex items-center space-x-2">
                <Checkbox
                  id={`visit-${type}`}
                  checked={selectedVisitTypes.includes(type)}
                  onCheckedChange={() => onToggleVisitType(type)}
                  className="size-4"
                />
                <Label htmlFor={`visit-${type}`} className={`text-base font-normal cursor-pointer ${C.text}`}>
                  {type}
                </Label>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-2">
          <h4 className={`font-bold text-base ${C.text}`}>指名</h4>
          <Select value={selectedDoctor} onValueChange={onSelectedDoctorChange}>
            <SelectTrigger className={`w-[200px] h-10 text-base ${C.bgWhite} border-input`}>
              <SelectValue placeholder="指名を選択" />
            </SelectTrigger>
            <SelectContent>
              {doctors.map((doctor) => (
                <SelectItem key={doctor.id} value={doctor.id}>
                  {doctor.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <h4 className={`font-bold text-base ${C.text}`}>種類</h4>
          <div className="flex items-center space-x-2 pt-0.5">
            <Checkbox
              id="trimming-only"
              checked={isTrimmingOnly}
              onCheckedChange={(checked) => onTrimmingOnlyChange(checked === true)}
              className="size-4"
            />
            <Label htmlFor="trimming-only" className={`text-base font-normal cursor-pointer ${C.text}`}>
              トリミングのみ表示
            </Label>
          </div>
        </div>
      </div>
    </div>
  );
}
