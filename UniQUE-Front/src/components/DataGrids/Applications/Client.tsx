"use client";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import { Box, Link as MuiLink, Paper } from "@mui/material";
import {
  DataGrid,
  type DataGridProps,
  GridActionsCellItem,
  type GridColDef,
  type GridRowId,
  useGridApiRef,
} from "@mui/x-data-grid";
import { jaJP } from "@mui/x-data-grid/locales";
import { RedirectType, redirect } from "next/navigation";
import React from "react";
import type { ApplicationData } from "@/classes/Application";
import DeleteDialog from "@/components/Dialogs/Delete";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";
import { deleteApplicationById } from "./deleteAction";

export default function ApplicationsDataGridClient({
  rows,
}: {
  rows: ApplicationData[];
}) {
  const apiRef = useGridApiRef();
  const [localRows, setLocalRows] = React.useState<ApplicationData[]>(
    () => rows,
  );

  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false);
  const [deletedAppId, setDeletedAppId] = React.useState<string | null>(null);

  const handleDelete = async (
    _prevState: FormStatus | null,
    _formData: FormData | null,
  ): Promise<FormStatus | null> => {
    void _prevState;
    void _formData;
    try {
      if (!deletedAppId) {
        return { status: "error", message: "削除対象が選択されていません" };
      }
      const result = await deleteApplicationById(deletedAppId);
      if (!result.success) {
        return {
          status: "error",
          message: result.error || "削除に失敗しました",
        };
      }
      setLocalRows((prev) => prev.filter((r) => r.id !== deletedAppId));
      setDeletedAppId(null);
      setDeleteDialogOpen(false);
      return { status: "success", message: "アプリケーションを削除しました" };
    } catch (err) {
      console.error(err);
      return { status: "error", message: String(err) };
    }
  };

  React.useEffect(() => {
    setLocalRows(rows);
  }, [rows]);

  const columns = React.useMemo<GridColDef[]>(() => {
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
              key={"edit-row"}
              icon={<EditIcon />}
              label="編集"
              onClick={() => {
                redirect(`./applications/${id}`, RedirectType.push);
              }}
            />,
            <GridActionsCellItem
              key={"delete-row"}
              icon={<DeleteIcon />}
              label="削除"
              onClick={() => {
                setDeletedAppId(String(id));
                setDeleteDialogOpen(true);
              }}
            />,
          ];
        },
      },
      {
        field: "id",
        headerName: "ID",
        width: 200,
        flex: 0.5,
      },
      {
        field: "name",
        headerName: "アプリケーション名",
        width: 200,
        flex: 1,
      },
      {
        field: "description",
        headerName: "説明",
        width: 250,
        flex: 1,
      },
      {
        field: "websiteUrl",
        headerName: "Webサイト",
        width: 200,
        flex: 1,
        renderCell: (params) => {
          if (!params.value) return null;
          return (
            <MuiLink
              href={params.value as string}
              target="_blank"
              rel="noopener noreferrer"
              sx={{ display: "flex", alignItems: "center", gap: 0.5 }}
            >
              <span
                style={{
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {params.value as string}
              </span>
              <OpenInNewIcon fontSize="small" />
            </MuiLink>
          );
        },
      },
      {
        field: "privacyPolicyUrl",
        headerName: "プライバシーポリシー",
        width: 200,
        flex: 1,
        renderCell: (params) => {
          if (!params.value) return null;
          return (
            <MuiLink
              href={params.value as string}
              target="_blank"
              rel="noopener noreferrer"
              sx={{ display: "flex", alignItems: "center", gap: 0.5 }}
            >
              <span
                style={{
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {params.value as string}
              </span>
              <OpenInNewIcon fontSize="small" />
            </MuiLink>
          );
        },
      },
      {
        field: "ownerDisplayName",
        headerName: "所有者",
        width: 200,
        flex: 0.8,
        renderCell: (params) => {
          const displayName = (params.row as ApplicationData).owner?.profile
            .displayName;
          const customId = (params.row as ApplicationData).owner?.customId;
          if (!displayName && !customId) return "不明";
          return `${displayName} (@${customId})`;
        },
      },
    ];
  }, []);

  const initialState = React.useMemo<
    NonNullable<DataGridProps["initialState"]>
  >(
    () => ({
      sorting: {
        sortModel: [{ field: "name", sort: "asc" }],
      },
      columns: {
        columnVisibilityModel: {
          id: false,
        },
      },
      pagination: {
        paginationModel: {
          pageSize: 10,
          page: 0,
        },
      },
    }),
    [],
  );

  return (
    <Paper sx={{ width: "100%", overflow: "hidden" }}>
      <Box sx={{ height: 500, width: "100%" }}>
        <DataGrid
          rows={localRows}
          columns={columns}
          apiRef={apiRef}
          disableRowSelectionOnClick
          initialState={initialState}
          ignoreValueFormatterDuringExport
          localeText={jaJP.components.MuiDataGrid.defaultProps.localeText}
          pageSizeOptions={[5, 10, 25, 50]}
          density="comfortable"
        />
      </Box>
      <DeleteDialog
        open={deleteDialogOpen}
        dataAction={handleDelete}
        handleClose={() => setDeleteDialogOpen(false)}
        title="アプリケーション"
      />
    </Paper>
  );
}
