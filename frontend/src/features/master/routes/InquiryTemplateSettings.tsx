import { useNavigate } from "react-router";
import { ClipboardList } from "lucide-react";
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { STYLE, LAYOUT, C } from "@/lib/design-tokens";

export function InquiryTemplateSettings() {
  const navigate = useNavigate();

  return (
    <PageLayout
      title="問診テンプレート"
      icon={<ClipboardList className="size-5 text-[#37352F]" />}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-3xl"
    >
      <div className="flex flex-col items-center justify-center py-24 gap-3">
        <div className={STYLE.pageIcon}>
          <ClipboardList className={LAYOUT.pageIcon.innerIcon} />
        </div>
        <p className={`text-sm ${C.text45}`}>問診テンプレート機能は準備中です</p>
      </div>
    </PageLayout>
  );
}
