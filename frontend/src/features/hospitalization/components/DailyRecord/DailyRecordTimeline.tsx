// External
import { memo } from "react";

// Internal
import { Badge } from "@/components/ui/badge";

// Relative
import { H_STYLES } from "@/features/hospitalization/styles";

// Types
import type { TimelineItem } from "@/features/hospitalization/types";
import { C } from "@/lib/design-tokens";

interface DailyRecordTimelineProps {
    items: TimelineItem[];
}

export const DailyRecordTimeline = memo(function DailyRecordTimeline({ items }: DailyRecordTimelineProps) {
    if (items.length === 0) {
        return (
            <div className={`${H_STYLES.text.base} ${C.text40} text-center py-4`}>
                今日の記録はまだありません
            </div>
        );
    }

    return (
        <div className={`flex flex-col ${H_STYLES.gap.default}`}>
            {items.map((item, idx) => (
                <div key={idx} className={`flex ${H_STYLES.gap.default} ${H_STYLES.text.base} group`}>
                    <div className={`font-mono ${C.text60} w-12 pt-0.5 ${H_STYLES.text.base}`}>{item.time}</div>
                    <div className={`flex-1 pb-2 border-b ${C.borderDivider} group-last:border-0`}>
                        <div className={`flex items-center ${H_STYLES.gap.default} mb-1`}>
                            {item.kind === 'vital' ? <Badge variant="outline" className={`${H_STYLES.text.xs} px-2 h-6 ${C.bgRedLight} ${C.danger} ${C.borderRedBadge}`}>バイタル</Badge> : null}
                            {item.kind === 'log' ? (
                                <Badge variant="outline" className={`
                                    ${H_STYLES.text.xs} px-2 h-6
                                    ${item.type === 'food' ? `${C.bgDiscountLight} ${C.textDiscount} ${C.borderOrangeBadge}` :
                                      item.type === 'excretion' ? `${C.bgNotice} ${C.textNotice} ${C.borderNotice}` :
                                      item.type === 'medicine' ? `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentBadge}` :
                                      item.type === 'treatment' ? `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderPurpleLight}` :
                                      `${C.bgStatusGray} ${C.textStatusGray} ${C.borderMuted}`}
                                `}>
                                    {item.type === 'food' ? '食事' :
                                     item.type === 'excretion' ? '排泄' :
                                     item.type === 'medicine' ? '投薬' :
                                     item.type === 'treatment' ? '処置' :
                                     'その他'}
                                </Badge>
                            ) : null}
                            <span className={`font-medium ${C.text} ${H_STYLES.text.base}`}>{item.value}</span>
                        </div>
                        {(item.notes || item.content) ? (
                            <div className={`${C.text80} pl-0.5 ${H_STYLES.text.base} leading-snug`}>{item.notes || item.content}</div>
                        ) : null}
                        {item.kind === 'vital' ? (
                            <div className={`grid grid-cols-4 ${H_STYLES.gap.tight} mt-1 ${H_STYLES.text.sm} ${C.text60} ${C.bgPage} ${H_STYLES.padding.tight} rounded font-medium`}>
                                {item.temperature ? <div>T: {item.temperature}℃</div> : null}
                                {item.heartRate ? <div>HR: {item.heartRate}</div> : null}
                                {item.respirationRate ? <div>RR: {item.respirationRate}</div> : null}
                                {item.weight ? <div>BW: {item.weight}kg</div> : null}
                            </div>
                        ) : null}
                    </div>
                </div>
            ))}
        </div>
    );
});
