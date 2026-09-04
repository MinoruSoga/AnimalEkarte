import type { InquiryTemplate } from "../api/inquiry-templates";

export interface InterviewTemplateFormData {
  category: string;
  title: string;
  content: string;
  isActive: boolean;
}

export function interviewTemplateToFormData(
  item: InquiryTemplate | null,
): InterviewTemplateFormData {
  return {
    category: item?.category ?? "",
    title: item?.title ?? "",
    content: item?.content ?? "",
    isActive: item?.isActive ?? true,
  };
}
