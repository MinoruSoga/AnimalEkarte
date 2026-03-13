// Medical records feature types

/** Interview (問診) history list item */
export interface InterviewHistoryItem {
  id: string;
  date: string;
  author: string;
  type: string;
  title: string;
  content: string;
}
