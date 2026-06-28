export class UniQUE_Error extends Error {
  code: string;
  type: UniQUE_ErrorType;
  isConfidential: boolean = false;

  constructor(
    message: string,
    option: {
      code: string;
      type: UniQUE_ErrorType;
      isConfidential?: boolean;
    },
  ) {
    super(message);
    this.code = option.code;
    this.isConfidential = option.isConfidential ?? false;
    this.type = option.type;
    this.name = `UniQUE ${this.type}`;
  }
}

export enum UniQUE_ErrorType {
  AUTHENTICATION_ERROR = "Authentication Error",
  AUTHORIZATION_ERROR = "Authorization Error",
  AUTH_SERVER_ERROR = "Auth Server Error",
  FRONTEND_ERROR = "Frontend Error",
  FORM_REQUEST_ERROR = "Form Request Error",
  RESOURCE_API_ERROR = "Resource API Error",
}
