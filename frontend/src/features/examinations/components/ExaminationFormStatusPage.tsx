import type { ComponentProps, ReactNode } from "react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LAYOUT } from "@/lib/design-tokens";

interface ExaminationFormStatusPageProps {
  resource: ComponentProps<typeof PageLayout>["resource"];
  onBack: () => void;
  children: ReactNode;
}

// FE-RC-045/046: ExaminationForm.tsx の loading/not-found/error 早期 return が
// 共有していた PageLayout の骨格（BUG-016）を1箇所へ集約する。
// resource は呼び出し元から受け取る（TASK-444-S1: generated/models の新規 import 追加禁止）。
export function ExaminationFormStatusPage({
  resource,
  onBack,
  children,
}: ExaminationFormStatusPageProps) {
  return (
    <PageLayout
      title="検査"
      resource={resource}
      onBack={onBack}
      maxWidth={LAYOUT.pageContentMaxWidth.formMid}
      align="left"
    >
      {children}
    </PageLayout>
  );
}
