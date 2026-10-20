// EMR-474: 再診問診回答の read-only 表示（最新回答・履歴で共用）。入力欄へは流し込まない。
import type { RevisitIntakeFullAnswers } from "../../types/models";
import {
  REVISIT_INTAKE_FIELD_LABELS,
  REVISIT_INTAKE_FIELD_ORDER,
  formatRevisitIntakeAnswer,
} from "../../lib/revisit-intake-options";

export function RevisitIntakeAnswerView({ answers }: { answers: RevisitIntakeFullAnswers }) {
  return (
    <dl className="divide-y divide-noah-border-light text-sm">
      {REVISIT_INTAKE_FIELD_ORDER.map((field) => {
        const value = formatRevisitIntakeAnswer(answers, field);
        return (
          <div key={field} className="py-2">
            <dt className="text-xs text-noah-text-sub">{REVISIT_INTAKE_FIELD_LABELS[field]}</dt>
            <dd
              className={`mt-0.5 whitespace-pre-wrap break-words ${
                value === null ? "text-noah-text-faint" : "text-noah-text"
              }`}
            >
              {value ?? "未回答"}
            </dd>
          </div>
        );
      })}
    </dl>
  );
}
