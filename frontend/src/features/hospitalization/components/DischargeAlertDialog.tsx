import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "../../../components/ui/alert-dialog";
import { H_STYLES } from "../styles";

interface DischargeAlertDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: () => void;
}

export const DischargeAlertDialog = ({ open, onOpenChange, onConfirm }: DischargeAlertDialogProps) => {
    return (
        <AlertDialog open={open} onOpenChange={onOpenChange}>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>退院処理を行いますか？</AlertDialogTitle>
                    <AlertDialogDescription>
                        ステータスを「退院済」に変更し、退院日を本日に設定します。<br />
                        この操作は取り消せません。
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel className={H_STYLES.button.action}>キャンセル</AlertDialogCancel>
                    <AlertDialogAction onClick={onConfirm} className={`bg-red-600 hover:bg-red-700 ${H_STYLES.button.action}`}>
                        退院処理を実行
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
};
