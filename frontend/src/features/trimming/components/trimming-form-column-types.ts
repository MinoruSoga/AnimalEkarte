import type { ChangeEvent } from "react";

import type { SortOrder } from "@/types";
import type { TrimmingFormData } from "@/types/trimming";

interface TrimmingMasterItem {
  id: string;
  name: string;
  price?: number;
  status?: string;
}

export interface TrimmingHistoryItem {
  id: string;
  date: string;
  styleRequest: string;
  courseId: string;
  optionIds: string[];
  usedShampoo: string;
  usedRibbon: string;
  remarks: string;
  staff: string;
}

export interface TrimmingLeftColumnProps {
  formData: TrimmingFormData;
  courses: TrimmingMasterItem[];
  options: TrimmingMasterItem[];
  styleImagePreview: string | null;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onCourseModalOpen: () => void;
  onStyleImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onRemoveStyleImage: () => void;
  courseError?: string;
  /** #233: カルテ画面から直接新規作成する場合のみ true。登録時ステータス選択の表示可否。 */
  showInitialStatusSelector: boolean;
}

export interface TrimmingMiddleColumnProps {
  formData: TrimmingFormData;
  completedImagePreview: string | null;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onCompletedImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onRemoveCompletedImage: () => void;
}

export interface TrimmingRightColumnProps {
  sortedHistory: TrimmingHistoryItem[];
  isHistoryLoading: boolean;
  historySearchTerm: string;
  historySortOrder: SortOrder;
  historyDateRange: { from: string; to: string };
  onSearchTermChange: (value: string) => void;
  onSortOrderChange: (value: SortOrder) => void;
  onClear: () => void;
  onFilterStartDateChange: (value: string) => void;
  onFilterEndDateChange: (value: string) => void;
  onHistoryClick: (updates: Partial<TrimmingFormData>) => void;
}
