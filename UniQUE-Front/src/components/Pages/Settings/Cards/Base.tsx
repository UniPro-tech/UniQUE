import { Card } from "@mui/material";
import { SnackbarProvider } from "notistack";

export interface FormStatus {
  status: "error" | "success" | "default" | "warning" | "info" | undefined;
  message: string;
}

export default function Base({
  sid,
  action,
  children,
  csrfToken,
}: {
  sid: string;
  action: (formData: FormData) => void | Promise<void>;
  isPending: boolean;
  children: React.ReactNode;
  csrfToken: string;
}) {
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <Card
        component={"form"}
        variant="outlined"
        action={action}
        sx={{ display: "flex", p: 2, flexDirection: "column", gap: 2 }}
      >
        <input type="hidden" name="csrfToken" value={csrfToken} />
        <input type="hidden" name="sid" value={sid} />
        {children}
      </Card>
    </SnackbarProvider>
  );
}
