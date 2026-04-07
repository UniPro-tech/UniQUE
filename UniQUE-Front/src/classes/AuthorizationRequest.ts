import { createApiClient } from "@/libs/apiClient";
import { toCamelcase } from "@/libs/snakeCamelUtil";
import { Application } from "./Application";

export interface AuthorizationRequestData {
  id: string;
  client_id: string;
  code_challenge: string;
  code_challenge_method: string;
  nonce: string;
  prompt: string;
  redirect_uri: string;
  response_type: string;
  scope: string;
  state: string;
}

export class AuthorizationRequest {
  id: string;
  client_id: string;
  code_challenge: string;
  code_challenge_method: string;
  nonce: string;
  prompt: string;
  redirect_uri: string;
  response_type: string;
  scope: string;
  state: string;

  constructor(data: AuthorizationRequestData) {
    this.id = data.id;
    this.client_id = data.client_id;
    this.code_challenge = data.code_challenge;
    this.code_challenge_method = data.code_challenge_method;
    this.nonce = data.nonce;
    this.prompt = data.prompt;
    this.redirect_uri = data.redirect_uri;
    this.response_type = data.response_type;
    this.scope = data.scope;
    this.state = data.state;
  }

  // ------ Converter Methods ------

  static fromJson(data: AuthorizationRequestData): AuthorizationRequest {
    return new AuthorizationRequest(data);
  }

  toJson(): AuthorizationRequestData {
    return {
      id: this.id,
      client_id: this.client_id,
      code_challenge: this.code_challenge,
      code_challenge_method: this.code_challenge_method,
      nonce: this.nonce,
      prompt: this.prompt,
      redirect_uri: this.redirect_uri,
      response_type: this.response_type,
      scope: this.scope,
      state: this.state,
    };
  }

  // ------ Static Methods ------

  static async getById(id: string): Promise<AuthorizationRequest> {
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const response = await apiClient.get(
      `/internal/auth-requests/${encodeURIComponent(id)}`,
    );
    if (!response.ok) {
      throw new Error(`Failed to fetch authorization request with id ${id}`);
    }
    const data = toCamelcase<AuthorizationRequestData>(await response.json());
    return AuthorizationRequest.fromJson(data);
  }

  static async autoConsentedById(id: string): Promise<boolean> {
    const authReq = await AuthorizationRequest.getById(id);
    return authReq.autoConsented();
  }

  // ------ Instance Methods -----

  async autoConsented(): Promise<boolean> {
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    if (
      this.prompt.includes("consent") ||
      this.prompt.includes("select_account") ||
      this.prompt.includes("login")
    ) {
      return false; // User must explicitly consent
    }
    const response = await apiClient.get(
      `/internal/auth-requests/${encodeURIComponent(this.id)}/consented`,
    );
    if (!response.ok) {
      throw new Error(
        `Failed to check auto-consent for authorization request with id ${this.id}`,
      );
    }
    return true;
  }

  async getApplication(): Promise<Application> {
    const app = await Application.getById(this.client_id);
    return app;
  }
}
