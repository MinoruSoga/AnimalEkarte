export function MaintenancePage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-noah-teal-light">
      <div className="max-w-md mx-auto px-4 text-center">
        <div className="text-6xl mb-4" aria-hidden="true">
          🔧
        </div>
        <h1 className="text-xl font-bold text-noah-teal-dark mb-2">メンテナンス中</h1>
        <p className="text-noah-text-sub">
          ただいまシステムメンテナンス中です。
          <br />
          しばらく経ってから再度お試しください。
        </p>
      </div>
    </div>
  );
}
