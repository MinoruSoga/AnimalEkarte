import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "../../../../components/ui/button";
import { H_STYLES } from "../../styles";

interface DateNavigationProps {
    date: Date;
    onPrev: () => void;
    onNext: () => void;
}

export const DateNavigation = ({ date, onPrev, onNext }: DateNavigationProps) => {
    return (
        <div className={`flex items-center justify-between bg-white ${H_STYLES.padding.box} rounded-lg border border-[rgba(55,53,47,0.16)] shadow-sm`}>
            <Button variant="ghost" size="icon" onClick={onPrev} className="h-10 w-10">
                <ChevronLeft className="h-6 w-6" />
            </Button>
            <div className={`font-bold text-[#37352F] ${H_STYLES.text.lg}`}>
                {format(date, "yyyy年 M月 d日 (EEE)", { locale: ja })}
            </div>
            <Button variant="ghost" size="icon" onClick={onNext} className="h-10 w-10">
                <ChevronRight className="h-6 w-6" />
            </Button>
        </div>
    );
};
