import { UserManager, WebStorageStateStore } from 'oidc-client-ts'

export const userManager = new UserManager({
  authority: import.meta.env.VITE_OIDC_ISSUER,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  post_logout_redirect_uri: import.meta.env.VITE_OIDC_POST_LOGOUT_URI,
  scope: 'openid profile email idol.read idol.write idol.admin idol.contribute',
  response_type: 'code',
  userStore: new WebStorageStateStore({ store: window.localStorage }),
})
