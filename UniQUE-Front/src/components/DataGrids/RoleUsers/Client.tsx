"use client";
import DeleteIcon from "@mui/icons-material/Delete";
import PersonAddIcon from "@mui/icons-material/PersonAdd";
import { Box, Button, darken, Paper, Stack } from "@mui/material";
import {
  DataGrid,
  type DataGridProps,
  GridActionsCellItem,
  type GridColDef,
  type GridRowId,
  type GridValidRowModel,
  gridClasses,
  useGridApiRef,
} from "@mui/x-data-grid";
import { jaJP } from "@mui/x-data-grid/locales";
import React from "react";
import type { RoleData } from "@/classes/Role";
import type { UserData } from "@/classes/types/User";
import AssignUserToRoleDialog from "@/components/Dialogs/AssignUserToRole";
import { unassignUserFromRole } from "@/components/Dialogs/AssignUserToRole/action";
import DeleteDialog from "@/components/Dialogs/Delete";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";

export default function RoleUsersDataGridClient({
  role,
  rows,
}: {
  role: RoleData;
  rows: UserData[];
}) {
  const apiRef = useGridApiRef();
  const [localRows, setLocalRows] = React.useState<UserData[]>(() => rows);
  const [assignDialogOpen, setAssignDialogOpen] = React.useState(false);

  React.useEffect(() => {
    setLocalRows(rows);
  }, [rows]);

  const [undeletedRows, setUndeletedRows] = React.useState<GridRowId | null>(
    null,
  );
  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false);

  const handleAssignSuccess = () => {
    // ページをリロードして最新のデータを取得
    window.location.reload();
  };

  const handleDelete = async (
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    _prevState: FormStatus | null,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    _formData: FormData | null,
  ) => {
    try {
      if (!undeletedRows) {
        return {
          status: "error",
          message: "削除対象が選択されていません",
        } as FormStatus;
      }

      const result = await unassignUserFromRole(
        role.id,
        undeletedRows as string,
      );

      if (!result.success) {
        return {
          status: "error",
          message: result.error || "削除に失敗しました",
        } as FormStatus;
      }

      setLocalRows((prev) =>
        prev.filter((r) => String(r.id) !== String(undeletedRows)),
      );
      apiRef.current?.updateRows([{ id: undeletedRows, _action: "delete" }]);
      setUndeletedRows(null);
      setDeleteDialogOpen(false);
      return {
        status: "success",
        message: "ロールの割り当てを削除しました",
      } as FormStatus;
    } catch (error) {
      return {
        status: "error",
        message: `削除に失敗しました: ${String(error)}`,
      } as FormStatus;
    }
  };

  const unsavedChangesRef = React.useRef<{
    unsavedRows: Record<GridRowId, GridValidRowModel>;
    rowsBeforeChange: Record<GridRowId, GridValidRowModel>;
  }>({
    unsavedRows: {},
    rowsBeforeChange: {},
  });

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
              key={"delete-row"}
              icon={<DeleteIcon />}
              label="削除"
              onClick={() => {
                setUndeletedRows(id);
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
        field: "displayName",
        headerName: "表示名",
        width: 180,
        flex: 1,
        valueGetter: (_value: unknown, row: UserData) =>
          row.profile?.displayName || row.customId || "",
      },
      {
        field: "customId",
        headerName: "カスタムID",
        width: 140,
        flex: 0.8,
      },
      {
        field: "email",
        headerName: "メールアドレス",
        width: 220,
        flex: 1,
      },
      {
        field: "status",
        headerName: "ステータス",
        width: 120,
      },
    ];
  }, []);

  const initialState = React.useMemo<
    NonNullable<DataGridProps["initialState"]>
  >(
    () => ({
      sorting: {
        sortModel: [{ field: "displayName", sort: "asc" }],
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

  const getRowClassName = React.useCallback<
    NonNullable<DataGridProps["getRowClassName"]>
  >(({ id }) => {
    const unsavedRow = unsavedChangesRef.current.unsavedRows[id];
    if (unsavedRow) {
      if (unsavedRow._action === "delete") return "row--removed";
      return "row--edited";
    }
    return "";
  }, []);

  return (
    <Stack spacing={2}>
      <Box sx={{ display: "flex", justifyContent: "flex-end" }}>
        <Button
          variant="contained"
          startIcon={<PersonAddIcon />}
          onClick={() => setAssignDialogOpen(true)}
        >
          ユーザーを追加
        </Button>
      </Box>
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
            getRowClassName={getRowClassName}
            pageSizeOptions={[5, 10, 25, 50]}
            density="comfortable"
            sx={{
              [`& .${gridClasses.row}.row--removed`]: {
                backgroundColor: (theme) => {
                  if (theme.palette.mode === "light")
                    return "rgba(255, 170, 170, 0.3)";
                  return darken("rgba(255, 170, 170, 1)", 0.7);
                },
              },
              [`& .${gridClasses.row}.row--edited`]: {
                backgroundColor: (theme) => {
                  if (theme.palette.mode === "light")
                    return "rgba(255, 254, 176, 0.3)";
                  return darken("rgba(255, 254, 176, 1)", 0.6);
                },
              },
            }}
          />
        </Box>
      </Paper>
      <DeleteDialog
        open={deleteDialogOpen}
        dataAction={handleDelete}
        handleClose={() => setDeleteDialogOpen(false)}
        title="ロールの割り当て"
      />
      <AssignUserToRoleDialog
        open={assignDialogOpen}
        roleId={role.id}
        roleName={role.name}
        currentUserIds={localRows.map((u) => u.id)}
        handleClose={() => setAssignDialogOpen(false)}
        onSuccess={handleAssignSuccess}
      />
    </Stack>
  );
}
