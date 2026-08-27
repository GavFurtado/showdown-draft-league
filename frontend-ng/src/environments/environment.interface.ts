export interface Environment {
  production: boolean;
  apiUrl: string;
  /** Enables the dev-only user impersonation panel (requires backend ENV=dev routes). */
  devTools: boolean;
}
