import { FrontendErrors } from "./frontend-errors";

export class UniQUE_Error extends Error {
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
    Object.setPrototypeOf(this, new.target.prototype);
    this.cause = option.code;
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

export const GetErrorMessageClient = (error: Error) => {
  console.log(error.name);
  if (error.name.startsWith("UniQUE")) {
    if (!(error as UniQUE_Error).isConfidential) {
      return `[${(error as UniQUE_Error).cause}] ${error.message}`;
    }
    console.log("This is confidential unique log");
    console.log(error);
  } else {
    console.log("This is not unique log");
    console.log(error);
    return `[${FrontendErrors.UnhandledException.cause}] ${FrontendErrors.UnhandledException.message}`;
  }
};
