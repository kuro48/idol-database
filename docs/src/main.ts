import { createApiReference } from '@scalar/api-reference'

createApiReference('#app', {
  url: '/openapi.json',
  theme: 'default',
  layout: 'modern',
  defaultHttpClient: {
    targetKey: 'shell',
    clientKey: 'curl',
  },
  authentication: {
    preferredSecurityScheme: 'ApiKeyAuth',
    apiKey: {
      token: 'ik_live_YOUR_API_KEY',
    },
  },
  metaData: {
    title: 'Idol API Reference',
    description:
      'アイドル・グループ・事務所・イベント情報を提供する REST API のリファレンスドキュメント',
    ogDescription: 'Idol API — Developer Reference',
  },
  hideModels: false,
  hideDownloadButton: false,
  showSidebar: true,
})
