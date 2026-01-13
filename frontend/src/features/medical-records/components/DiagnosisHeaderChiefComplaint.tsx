import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../../../components/ui/card";
import { ScrollArea } from "../../../components/ui/scroll-area";
import { FileText } from "lucide-react";

export const DiagnosisHeaderChiefComplaint = React.memo(function DiagnosisHeaderChiefComplaint() {
  return (
    <div className="col-span-3 flex flex-col min-h-0">
      <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
        <CardHeader className="p-0 pb-2">
          <CardTitle className="text-sm font-bold text-[#37352F] flex items-center gap-2">
            <FileText className="size-4" />
            主訴情報
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0 flex-1 min-h-0">
          <ScrollArea className="h-full rounded-md border border-[rgba(55,53,47,0.16)] bg-[#F7F6F3] p-3 text-sm text-[#37352F]">
            <div className="space-y-3">
              <div>
                <span className="font-bold block text-muted-foreground mb-1 text-sm">
                  # どんな症状
                </span>
                <p className="text-sm leading-snug">
                  嘔吐2回、下痢なし。食欲低下。
                </p>
              </div>
              <div>
                <span className="font-bold block text-muted-foreground mb-1 text-sm">
                  # どこが
                </span>
                <p className="text-sm leading-snug">自宅のリビングで</p>
              </div>
              <div>
                <span className="font-bold block text-muted-foreground mb-1 text-sm">
                  # いつから
                </span>
                <p className="text-sm leading-snug">昨晩から</p>
              </div>
              <div>
                <span className="font-bold block text-muted-foreground mb-1 text-sm">
                  # その他
                </span>
                <p className="text-sm leading-snug">異物誤飲の可能性なし</p>
              </div>
            </div>
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  );
});
