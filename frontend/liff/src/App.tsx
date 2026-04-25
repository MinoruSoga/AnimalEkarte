import { useLiff } from './hooks/use-liff';
import { LIFF_ID } from './lib/liff-config';
import { LoadingPage } from './pages/LoadingPage';
import { ErrorPage } from './pages/ErrorPage';
import { PetHealthPage } from './pages/PetHealthPage';

export function App() {
  const { idToken, displayName, pictureUrl, isReady, initError } = useLiff(LIFF_ID);

  if (!isReady) {
    return <LoadingPage />;
  }

  if (initError) {
    return <ErrorPage message="LINE認証に失敗しました" />;
  }

  if (idToken) {
    return (
      <PetHealthPage
        idToken={idToken}
        displayName={displayName}
        pictureUrl={pictureUrl}
      />
    );
  }

  return <ErrorPage message="ログイン情報が取得できませんでした" />;
}
