import { useNavigate } from "react-router";
import { FileText } from "lucide-react";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { paths } from "@/config/paths";

export function InterviewTemplateSettings() {
  const navigate = useNavigate();

  return (
    <PageLayout
      title="問診テンプレートマスタ"
      icon={<FileText className="size-5 text-[#37352F]" />}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-full"
    >
      <div className="flex items-center justify-center h-96">
        <p className="text-muted-foreground">
          問診テンプレートマスタ管理はまだ実装されていません。
          <br />
          バックエンド API の実装をお待ちください。
        </p>
      </div>
    </PageLayout>
  );
}
