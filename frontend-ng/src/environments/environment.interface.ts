export interface Environment {
  production: boolean;
  apiUrl: string;
  // Enables the dev-only user impersonation panel (requires server's ENV=dev routes)
  devTools: boolean;
}
