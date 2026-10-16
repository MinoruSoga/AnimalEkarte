import type { ChangeEvent } from "react";

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
  /** EMR-412 X14: トリミング新規のみコースラベルに必須印。既定 false。 */
  courseRequiredMark?: boolean;
}

export interface TrimmingMiddleColumnProps {
  formData: TrimmingFormData;
  completedImagePreview: string | null;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onCompletedImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onRemoveCompletedImage: () => void;
}

// PR679 F7: 検索/ソート/日付フィルタは共有 PastRecordHistoryPanel 内蔵のため、
// 親経由の外部フィルタ props・ハンドラは削除済み。
export interface TrimmingRightColumnProps {
  sortedHistory: TrimmingHistoryItem[];
  isHistoryLoading: boolean;
  onHistoryClick: (updates: Partial<TrimmingFormData>) => void;
}
