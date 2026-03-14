// React/Framework
import { useState } from "react";

// External
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export interface VitalData {
  temperature?: number;   // 体温 (℃)
  heartRate?: number;     // 心拍数 (bpm)
  respiratoryRate?: number; // 呼吸数 (breaths/min)
}

interface VitalInputDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  petName: string;
  onSave: (vitals: VitalData) => void;
}

interface FormState {
  temperature: string;
  heartRate: string;
  respiratoryRate: string;
}

const EMPTY_FORM: FormState = {
  temperature: "",
  heartRate: "",
  respiratoryRate: "",
};

export function VitalInputDialog({
  open,
  onOpenChange,
  petName,
  onSave,
}: VitalInputDialogProps) {
  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [validationError, setValidationError] = useState("");
  const [prevOpen, setPrevOpen] = useState(false);

  // Reset form when dialog opens
  if (open !== prevOpen) {
    setPrevOpen(open);
    if (open) {
      setForm(EMPTY_FORM);
      setValidationError("");
    }
  }

  const handleSave = () => {
    const hasAnyValue =
      form.temperature !== "" ||
      form.heartRate !== "" ||
      form.respiratoryRate !== "";

    if (!hasAnyValue) {
      setValidationError("少なくとも1つのバイタル値を入力してください。");
      return;
    }

    const vitals: VitalData = {
      temperature: form.temperature !== "" ? Number(form.temperature) : undefined,
      heartRate: form.heartRate !== "" ? Number(form.heartRate) : undefined,
      respiratoryRate: form.respiratoryRate !== "" ? Number(form.respiratoryRate) : undefined,
    };

    onSave(vitals);
    toast.success("バイタルを記録しました");
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>バイタル記録</DialogTitle>
          <DialogDescription>
            {petName} のバイタルサインを記録してください。
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="vital-temperature">体温 (℃)</Label>
            <Input
              id="vital-temperature"
              type="number"
              step="0.1"
              placeholder="例: 38.5"
              value={form.temperature}
              onChange={(e) => {
                setForm({ ...form, temperature: e.target.value });
                setValidationError("");
              }}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="vital-heart-rate">心拍数 (bpm)</Label>
            <Input
              id="vital-heart-rate"
              type="number"
              placeholder="例: 80"
              value={form.heartRate}
              onChange={(e) => {
                setForm({ ...form, heartRate: e.target.value });
                setValidationError("");
              }}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="vital-respiratory-rate">呼吸数</Label>
            <Input
              id="vital-respiratory-rate"
              type="number"
              placeholder="例: 20"
              value={form.respiratoryRate}
              onChange={(e) => {
                setForm({ ...form, respiratoryRate: e.target.value });
                setValidationError("");
              }}
            />
          </div>
        </div>

        {validationError && (
          <p className="text-sm text-red-500 -mt-2 mb-2">{validationError}</p>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            キャンセル
          </Button>
          <Button onClick={handleSave}>記録する</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
