"use client";

import { usePathname } from "next/navigation";
import { SnackbarProvider, useSnackbar, type VariantType } from "notistack";
import { useEffect } from "react";
import { replacePath } from "@/libs/replacePathAction";

export interface SnackbarData {
  message: string;
  variant: VariantType;
}

function InnerSnackRunner({
  snacks,
  replaceDuration,
}: {
  snacks: SnackbarData[];
  replaceDuration: number;
}) {
  const path = usePathname();
  const { enqueueSnackbar } = useSnackbar();
  useEffect(() => {
    snacks.forEach((snack) => {
      enqueueSnackbar(snack.message, { variant: snack.variant });
    });
    if (snacks.length > 0) {
      replacePath(path, replaceDuration);
    }
  }, [snacks, replaceDuration, path, enqueueSnackbar]);
  return null;
}

/**
 * # TemporarySnackProvider
 * 最初だけsnackbarを表示するコンポーネント
 */
export default function TemporarySnackProvider({
  snacks,
  replaceDuration = 0,
}: {
  snacks: SnackbarData[];
  replaceDuration?: number;
}) {
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <InnerSnackRunner snacks={snacks} replaceDuration={replaceDuration} />
    </SnackbarProvider>
  );
}
