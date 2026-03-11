export function logout(): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(() => resolve(), 200);
  });
}
