/**
 * FE-RC-214: line-reserve は @/lib を直接 import しない。
 * 通貨フォーマットの正本は src/lib/format/number.ts。ここは再 export のみ。
 */
export { formatCurrency } from "@/lib/format/number";
