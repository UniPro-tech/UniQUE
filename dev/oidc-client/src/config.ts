export const issuer = process.env.OIDC_ISSUER ?? "http://localhost:8000";
export const issuerInternal = process.env.OIDC_ISSUER_INTERNAL ?? issuer;
export const clientId =
  process.env.OIDC_CLIENT_ID ?? "01J00000000000000000000001";
export const redirectUri =
  process.env.OIDC_REDIRECT_URI ?? "http://localhost:3002/callback";
export const port = Number(process.env.PORT ?? 3002);
