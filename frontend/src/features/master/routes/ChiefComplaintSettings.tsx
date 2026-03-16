import { useNavigate } from "react-router";
import { MessageSquareText } from "lucide-react";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { paths } from "@/config/paths";

export function ChiefComplaintSettings() {
  const navigate = useNavigate();

  return (
    <PageLayout
      title="主訴マスタ"
      icon={<MessageSquareText className="size-5 text-[#37352F]" />}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-full"
    >
      <div className="flex items-center justify-center h-96">
        <p className="text-muted-foreground">
          主訴マスタ管理はまだ実装されていません。
          <br />
          バックエンド API の実装をお待ちください。
        </p>
      </div>
    </PageLayout>
  );
}
