import { useLiff } from '@/shared-liff/use-liff';
import { ErrorPage, type ErrorPageTheme } from '@/shared-liff/ErrorPage';
import { LIFF_ID } from './lib/liff-config';
import { LoadingPage } from './pages/LoadingPage';
import { PetHealthPage } from './pages/PetHealthPage';
import { LiffLinkPage } from './pages/LiffLinkPage';

const isLinkFlow = new URLSearchParams(window.location.search).has('token');

const ERROR_PAGE_THEME: ErrorPageTheme = {
  bg: 'bg-liff-brand-bg',
  heading: 'text-gray-800',
  body: 'text-gray-500',
  button: 'bg-liff-brand',
  buttonHover: 'hover:bg-liff-brand-dark',
};

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
    return <ErrorPage message="LINE認証に失敗しました" theme={ERROR_PAGE_THEME} />;
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

  return <ErrorPage message="ログイン情報が取得できませんでした" theme={ERROR_PAGE_THEME} />;
}
