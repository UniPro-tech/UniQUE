import { Card } from "@mui/material";

export interface FormStatus {
  status: "error" | "success" | "default" | "warning" | "info" | undefined;
  message: string;
}

export default function Base({
  sid,
  action,
  children,
}: {
  sid: string;
  action: (formData: FormData) => void | Promise<void>;
  isPending: boolean;
  children: React.ReactNode;
}) {
  return (
    <Card
      component={"form"}
      variant="outlined"
      action={action}
      sx={{ display: "flex", p: 2, flexDirection: "column", gap: 2 }}
    >
      <input type="hidden" name="sid" value={sid} />
      {children}
    </Card>
  );
}
