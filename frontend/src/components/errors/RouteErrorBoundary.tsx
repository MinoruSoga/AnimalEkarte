import { useRouteError, isRouteErrorResponse, Link } from "react-router";
import { AlertCircle, Home } from "lucide-react";
import { Button } from "@/components/ui/button";

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
    <div className="flex-1 p-5 flex flex-col items-center justify-center gap-4">
      <AlertCircle className="size-12 text-destructive" />
      <h1 className="text-xl font-bold text-foreground">{title}</h1>
      <p className="text-muted-foreground text-center max-w-md">{message}</p>
      <Button asChild variant="outline">
        <Link to="/">
          <Home className="size-4 mr-2" />
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
    <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-[#F7F6F3] p-4">
      <AlertCircle className="size-16 text-destructive" />
      <h1 className="text-2xl font-bold">エラーが発生しました</h1>
      <p className="text-muted-foreground text-center max-w-md">{message}</p>
      <Button onClick={() => window.location.href = "/"} variant="outline">
        <Home className="size-4 mr-2" />
        再読み込み
      </Button>
    </div>
  );
}
