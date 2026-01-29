import { Toaster } from "@/components/ui/sonner";

interface AppProviderProps {
  children: React.ReactNode;
}

export function AppProvider({ children }: AppProviderProps) {
  return (
    <>
      {children}
      <Toaster />
    </>
  );
}
