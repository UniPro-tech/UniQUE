import type { UniQUE_Error } from "./base";
import { FrontendErrors } from "./frontend-errors";

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
