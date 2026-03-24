// React/Framework
import { useNavigate } from "react-router";

// External
import { ClipboardCheck, ArrowRight } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";

export function CheckupsList() {
  const navigate = useNavigate();

  return (
    <PageLayout title="定期健診">
      <div className="flex flex-col items-center justify-center py-20 gap-6 text-center">
        <ClipboardCheck className="size-16 text-muted-foreground/40" />
        <div className="space-y-2">
          <h2 className="text-lg font-semibold text-foreground">
            定期健診一覧（準備中）
          </h2>
          <p className="text-sm text-muted-foreground max-w-sm">
            クリニック横断の定期健診一覧は現在開発中です。
            <br />
            各カルテの「定期健診」タブから健診記録を確認・登録できます。
          </p>
        </div>
        <Button
          variant="outline"
          onClick={() => navigate(paths.medicalRecords.getHref())}
          className="gap-2"
        >
          カルテ一覧を開く
          <ArrowRight className="size-4" />
        </Button>
      </div>
    </PageLayout>
  );
}
