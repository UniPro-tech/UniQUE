"use client";

import { SnackbarProvider } from "notistack";

export default function LayoutWrapperClient({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      {children}
    </SnackbarProvider>
  );
}
