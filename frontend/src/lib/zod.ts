import { z } from "zod";

// Required string field with Japanese error message
export function requiredString(fieldName: string) {
  return z.string().min(1, `${fieldName}は必須です`);
}

// Optional string field
export function optionalString() {
  return z.string().optional();
}

// Required email field
export function requiredEmail() {
  return z.string().email("有効なメールアドレスを入力してください");
}

// Required phone number (Japanese format)
export function requiredPhone() {
  return z
    .string()
    .regex(/^0\d{9,10}$/, "有効な電話番号を入力してください（ハイフンなし）");
}

// Optional phone number (Japanese format)
export function optionalPhone() {
  return z
    .string()
    .regex(/^0\d{9,10}$/, "有効な電話番号を入力してください（ハイフンなし）")
    .optional()
    .or(z.literal(""));
}

// Postal code (Japanese format: 7 digits)
export function postalCode() {
  return z
    .string()
    .regex(/^\d{7}$/, "郵便番号は7桁の数字で入力してください（ハイフンなし）");
}
