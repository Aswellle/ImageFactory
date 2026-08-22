/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly IF_API_BASE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
