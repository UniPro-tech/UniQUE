export type Prompt = "consent" | "none";

export type TokenSet = {
  access_token: string;
  refresh_token: string;
  id_token: string;
};

export type Milestone = {
  title: string;
  details: unknown;
};

export type LoginState = {
  state: string;
  nonce: string;
  verifier: string;
  expiresAt: number;
  stage: number;
  path?: "rotated" | "direct";
  prompt?: Prompt;
  events: Milestone[];
  tokens?: TokenSet;
  error?: string;
};

export type RequestLog = {
  method: string;
  url: string;
  headers?: Record<string, string>;
  body?: Record<string, string>;
};
