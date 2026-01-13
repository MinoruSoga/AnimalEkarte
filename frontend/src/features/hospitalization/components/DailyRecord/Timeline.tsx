import { Badge } from "../../../../components/ui/badge";
import { TimelineItem } from "../../types";
import { H_STYLES } from "../../styles";

interface TimelineProps {
    items: TimelineItem[];
}

export const Timeline = ({ items }: TimelineProps) => {
    if (items.length === 0) {
        return (
            <div className={`${H_STYLES.text.base} text-[#37352F]/40 text-center py-4`}>
                今日の記録はまだありません
            </div>
        );
    }

    return (
        <div className={`flex flex-col ${H_STYLES.gap.default}`}>
            {items.map((item, idx) => (
                <div key={idx} className={`flex ${H_STYLES.gap.default} ${H_STYLES.text.base} group`}>
                    <div className={`font-mono text-[#37352F]/60 w-12 pt-0.5 ${H_STYLES.text.base}`}>{item.time}</div>
                    <div className="flex-1 pb-2 border-b border-[rgba(55,53,47,0.06)] group-last:border-0">
                        <div className={`flex items-center ${H_STYLES.gap.default} mb-1`}>
                            {item.kind === 'vital' && <Badge variant="outline" className={`${H_STYLES.text.xs} px-2 h-6 bg-rose-50 text-rose-600 border-rose-200`}>バイタル</Badge>}
                            {item.kind === 'log' && (
                                <Badge variant="outline" className={`
                                    ${H_STYLES.text.xs} px-2 h-6
                                    ${item.type === 'food' ? 'bg-orange-50 text-orange-600 border-orange-200' : 
                                      item.type === 'excretion' ? 'bg-amber-50 text-amber-600 border-amber-200' :
                                      item.type === 'medicine' ? 'bg-blue-50 text-blue-600 border-blue-200' :
                                      item.type === 'treatment' ? 'bg-purple-50 text-purple-600 border-purple-200' :
                                      'bg-gray-50 text-gray-600 border-gray-200'}
                                `}>
                                    {item.type === 'food' ? '食事' : 
                                     item.type === 'excretion' ? '排泄' : 
                                     item.type === 'medicine' ? '投薬' : 
                                     item.type === 'treatment' ? '処置' : 
                                     'その他'}
                                </Badge>
                            )}
                            <span className={`font-medium text-[#37352F] ${H_STYLES.text.base}`}>{item.value}</span>
                        </div>
                        {(item.notes || item.content) && (
                            <div className={`text-[#37352F]/80 pl-0.5 ${H_STYLES.text.base} leading-snug`}>{item.notes || item.content}</div>
                        )}
                        {item.kind === 'vital' && (
                            <div className={`grid grid-cols-4 ${H_STYLES.gap.tight} mt-1 ${H_STYLES.text.sm} text-[#37352F]/60 bg-gray-50 ${H_STYLES.padding.tight} rounded font-medium`}>
                                {item.temperature && <div>T: {item.temperature}℃</div>}
                                {item.heartRate && <div>HR: {item.heartRate}</div>}
                                {item.respirationRate && <div>RR: {item.respirationRate}</div>}
                                {item.weight && <div>BW: {item.weight}kg</div>}
                            </div>
                        )}
                    </div>
                </div>
            ))}
        </div>
    );
};
