import { Logger, Transporter } from "@unipro-tech/node-logger";

export const baseLogger = new Logger(
  "UniQUE-Front", // ロガー名
  [Transporter.ConsoleTransporter()],
  {
    errorKey: "error",
  },
);
