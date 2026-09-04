import { useRouteError, isRouteErrorResponse, Link } from "react-router";
import { AlertCircle, Home } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ICON, C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

export function RouteErrorBoundary() {
  const error = useRouteError();

  let title = "エラーが発生しました";
  let message = "予期しないエラーが発生しました。再度お試しください。";

  if (isRouteErrorResponse(error)) {
    if (error.status === 404) {
      title = "ページが見つかりません";
      message = "お探しのページは存在しないか、移動した可能性があります。";
    } else {
      title = `エラー ${error.status}`;
      message = error.statusText || message;
    }
  }

  return (
    <div className={`flex-1 p-6 flex flex-col items-center justify-center gap-4 ${C.bgPage}`}>
      <AlertCircle className={`size-12 ${C.danger}`} />
      <h1 className={`text-xl font-bold ${C.text}`}>{title}</h1>
      <p className={`${C.text50} text-center max-w-md`}>{message}</p>
      <Button asChild variant="outline">
        <Link to={paths.home.getHref()}>
          <Home className={`${ICON.action} mr-2`} />
          ダッシュボードへ戻る
        </Link>
      </Button>
    </div>
  );
}

export function RootErrorBoundary() {
  const error = useRouteError();

  let message = "アプリケーションで予期しないエラーが発生しました。";
  if (isRouteErrorResponse(error)) {
    message = `${error.status}: ${error.statusText}`;
  }

  return (
    <div className={`min-h-screen flex flex-col items-center justify-center gap-4 ${C.bgPage} p-4`}>
      <AlertCircle className={`size-16 ${C.danger}`} />
      <h1 className={`text-heading-3 font-bold ${C.text}`}>エラーが発生しました</h1>
      <p className={`${C.text50} text-center max-w-md`}>{message}</p>
      <Button onClick={() => (window.location.href = paths.home.getHref())} variant="outline">
        <Home className={`${ICON.action} mr-2`} />
        再読み込み
      </Button>
    </div>
  );
}
