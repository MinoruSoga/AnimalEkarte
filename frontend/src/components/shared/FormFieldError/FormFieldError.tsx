import { C } from "@/lib/design-tokens";

interface FormFieldErrorProps {
  message: string | undefined | null;
  id?: string;
}

export function FormFieldError({ message, id }: FormFieldErrorProps) {
  if (!message) return null;

  return (
    <p id={id} role="alert" className={`text-sm ${C.danger} mt-1`}>
      {message}
    </p>
  );
}
