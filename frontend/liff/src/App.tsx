import { useLiff } from "@/shared-liff/use-liff";
import { ErrorPage } from "@/shared-liff/ErrorPage";
import { LIFF_ID } from "./lib/liff-config";
import { LoadingPage } from "./pages/LoadingPage";
import { PetHealthPage } from "./pages/PetHealthPage";
import { LiffLinkPage } from "./pages/LiffLinkPage";

const isLinkFlow = new URLSearchParams(window.location.search).has("token");

export function App() {
  if (isLinkFlow) {
    return <LiffLinkPage />;
  }

  return <HealthCardApp />;
}

function HealthCardApp() {
  const { idToken, displayName, pictureUrl, isReady, initError } = useLiff(LIFF_ID);

  if (!isReady) {
    return <LoadingPage />;
  }

  if (initError) {
    return <ErrorPage message="LINE認証に失敗しました" />;
  }

  if (idToken) {
    return <PetHealthPage idToken={idToken} displayName={displayName} pictureUrl={pictureUrl} />;
  }

  return <ErrorPage message="ログイン情報が取得できませんでした" />;
}
