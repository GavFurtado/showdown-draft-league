/// <reference types="vite/client" />

declare module '*.{png,jpg,jpeg,gif,svg}' {
  const content: string;
  export default content;
}
