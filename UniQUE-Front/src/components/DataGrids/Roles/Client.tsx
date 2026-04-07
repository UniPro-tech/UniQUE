"use client";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import { Box, Paper } from "@mui/material";
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
import type { RoleData } from "@/classes/Role";
import DeleteDialog from "@/components/Dialogs/Delete";
import { deleteRole } from "@/components/Forms/RoleEditForm/action";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";

export default function RolesDataGridClient({ rows }: { rows: RoleData[] }) {
  const apiRef = useGridApiRef();
  const [localRows, setLocalRows] = React.useState<RoleData[]>(() => rows);
  const [undeletedRole, setUndeletedRole] = React.useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false);

  React.useEffect(() => {
    setLocalRows(rows);
  }, [rows]);

  const handleDelete = async (
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    _prevState: FormStatus | null,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    _formData: FormData | null,
  ) => {
    try {
      if (!undeletedRole) {
        return {
          status: "error",
          message: "削除対象が選択されていません",
        } as FormStatus;
      }

      const result = await deleteRole(undeletedRole);

      if (!result.success) {
        return {
          status: "error",
          message: result.error || "削除に失敗しました",
        } as FormStatus;
      }

      setLocalRows((prev) =>
        prev.filter((r) => String(r.id) !== String(undeletedRole)),
      );
      apiRef.current?.updateRows([{ id: undeletedRole, _action: "delete" }]);
      setUndeletedRole(null);
      setDeleteDialogOpen(false);
      return {
        status: "success",
        message: "ロールを削除しました",
      } as FormStatus;
    } catch (error) {
      return {
        status: "error",
        message: `削除に失敗しました: ${String(error)}`,
      } as FormStatus;
    }
  };

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
                redirect(`./roles/${id}`, RedirectType.push);
              }}
            />,
            <GridActionsCellItem
              key={"delete-row"}
              icon={<DeleteIcon />}
              label="削除"
              onClick={() => {
                setUndeletedRole(String(id));
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
        headerName: "ロール名",
        width: 180,
        flex: 1,
      },
      {
        field: "customId",
        headerName: "カスタムID",
        width: 140,
        flex: 0.8,
      },
      {
        field: "description",
        headerName: "説明",
        width: 250,
        flex: 1,
      },
      {
        field: "permissionBitmask",
        headerName: "権限ビットマスク",
        width: 160,
        type: "number",
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
        title="ロール"
      />
    </Paper>
  );
}
