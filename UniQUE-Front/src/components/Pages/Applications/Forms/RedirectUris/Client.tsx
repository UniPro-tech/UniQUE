"use client";

import DeleteIcon from "@mui/icons-material/Delete";
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import {
  DataGrid,
  type DataGridProps,
  GridActionsCellItem,
  type GridColDef,
  type GridRowId,
} from "@mui/x-data-grid";
import { jaJP } from "@mui/x-data-grid/locales";
import { SnackbarProvider, useSnackbar } from "notistack";
import { useEffect, useMemo, useState } from "react";
import type { ApplicationData } from "@/classes/Application";
import { addRedirectUri, deleteRedirectUri } from "./action";

export default function RedirectUrisClient({
  application,
  initialUris,
}: {
  application: ApplicationData;
  initialUris?: string[];
}) {
  const [uris, setUris] = useState<string[]>(initialUris ?? []);
  const [newUri, setNewUri] = useState("");
  const [saving, setSaving] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [uriToDelete, setUriToDelete] = useState("");

  const onDelete = (uri: string) => {
    setUriToDelete(uri);
    setDeleteDialogOpen(true);
  };

  useEffect(() => {
    // keep initialUris as source of truth on mount
    setUris(initialUris ?? []);
  }, [initialUris]);

  const { enqueueSnackbar } = useSnackbar();

  const addUri = async () => {
    if (!newUri) return;
    await addRedirectUri(application, newUri);
    setSaving(true);
    try {
      setUris((s) => [...s, newUri]);
      setNewUri("");
      enqueueSnackbar("追加しました", { variant: "success" });
    } catch {
      enqueueSnackbar("追加に失敗しました", { variant: "error" });
    } finally {
      setSaving(false);
    }
  };

  const rows = uris.map((u) => ({ id: u, uri: u }));

  const columns = useMemo<GridColDef[]>(() => {
    return [
      {
        field: "actions",
        headerName: "操作",
        type: "actions",
        width: 100,
        sortable: false,
        filterable: false,
        getActions: ({ id }: { id: GridRowId }) => {
          return [
            <GridActionsCellItem
              key={"delete-row"}
              icon={<DeleteIcon />}
              label="削除"
              onClick={() => onDelete(String(id))}
            />,
          ];
        },
      },
      {
        field: "uri",
        headerName: "Redirect URI",
        width: 600,
        flex: 1,
        sortable: true,
        renderCell: (params) => {
          return (
            <Typography variant="body2" sx={{ wordBreak: "break-all" }}>
              {params.value as string}
            </Typography>
          );
        },
      },
    ];
  }, [onDelete]);

  const initialState = useMemo<NonNullable<DataGridProps["initialState"]>>(
    () => ({
      sorting: { sortModel: [{ field: "uri", sort: "asc" }] },
      pagination: { paginationModel: { pageSize: 10, page: 0 } },
    }),
    [],
  );

  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={500}>
      <Stack spacing={2}>
        <Typography variant="subtitle1">Redirect URIs</Typography>

        <Paper sx={{ width: "100%", p: 2 }}>
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={1}
            alignItems="center"
          >
            <TextField
              label="Redirect URI"
              value={newUri}
              size="small"
              onChange={(e) => setNewUri(e.target.value)}
              sx={{ minWidth: 360, flex: 1 }}
            />
            <Button
              variant="contained"
              onClick={addUri}
              disabled={saving || !newUri}
              sx={{ whiteSpace: "nowrap" }}
            >
              追加
            </Button>
          </Stack>

          <Box sx={{ mt: 2 }}>
            <DataGrid
              rows={rows}
              columns={columns}
              disableRowSelectionOnClick
              initialState={initialState}
              pageSizeOptions={[5, 10, 25, 50]}
              density="comfortable"
              localeText={jaJP.components.MuiDataGrid.defaultProps.localeText}
              autoHeight
              rowHeight={56}
              sx={{
                borderRadius: 1,
                "& .MuiDataGrid-cell": { py: 1.25 },
                "& .MuiDataGrid-columnHeaders": (theme) => ({
                  backgroundColor: theme.palette.grey[50],
                }),
              }}
            />
          </Box>
        </Paper>
        <DeleteRedirectUriDialog
          open={deleteDialogOpen}
          uri={uriToDelete}
          application={application}
          setUris={setUris}
          onClose={() => setDeleteDialogOpen(false)}
        />
      </Stack>
    </SnackbarProvider>
  );
}

function DeleteRedirectUriDialog({
  open,
  uri,
  application,
  setUris,
  onClose,
}: {
  open: boolean;
  uri: string;
  application: ApplicationData;
  setUris: React.Dispatch<React.SetStateAction<string[]>>;
  onClose: () => void;
}) {
  const { enqueueSnackbar } = useSnackbar();
  const onConfirm = async (uri: string) => {
    try {
      await deleteRedirectUri(application, uri);
      setUris((s) => s.filter((u) => u !== uri));
      enqueueSnackbar("削除しました", { variant: "success" });
    } catch {
      enqueueSnackbar("削除に失敗しました", { variant: "error" });
    }
  };
  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Redirect URIの削除</DialogTitle>
      <DialogContent>
        <Typography>
          &quot;{uri}&quot;
          を本当に削除していいですか？この操作は取り消せません。
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>キャンセル</Button>
        <Button
          color="error"
          onClick={() => {
            onConfirm(uri);
            onClose();
          }}
          variant="contained"
        >
          削除
        </Button>
      </DialogActions>
    </Dialog>
  );
}
