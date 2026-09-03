import type {
  CreateInquiryTemplateRequest,
  UpdateInquiryTemplateRequest,
} from "../api/inquiry-templates";
import type { InterviewTemplateFormData } from "../lib/interview-template-side-panel-model";

export function buildInterviewTemplateCreateRequest(
  data: InterviewTemplateFormData,
): CreateInquiryTemplateRequest {
  return {
    category: data.category,
    title: data.title,
    content: data.content,
    is_active: data.isActive,
  };
}

export function buildInterviewTemplateUpdateRequest(
  data: InterviewTemplateFormData,
): UpdateInquiryTemplateRequest {
  return {
    category: data.category,
    title: data.title,
    content: data.content,
    is_active: data.isActive,
  };
}
