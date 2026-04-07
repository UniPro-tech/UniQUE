"use client";
import CheckIcon from "@mui/icons-material/Check";
import DeleteIcon from "@mui/icons-material/Delete";
import RestoreIcon from "@mui/icons-material/Restore";
import SaveIcon from "@mui/icons-material/Save";
import VisibilityIcon from "@mui/icons-material/Visibility";
import { Button, darken } from "@mui/material";
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
import { useRouter } from "next/navigation";
import { enqueueSnackbar } from "notistack";
import React, { useMemo } from "react";
import type { ProfileData } from "@/classes/Profile";
import type { UserData, UserStatus } from "@/classes/types/User";
import ApproveRegistApplyDialog from "@/components/DataGrids/Members/ApproveDialog";
import { deleteUserById } from "@/components/DataGrids/Members/actions/deleteAction";
import RejectDialog from "@/components/DataGrids/Members/RejectDialog";
import DeleteDialog from "@/components/Dialogs/Delete";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";
import {
  getSelectableAffiliationPeriods,
  USER_STATUS_OPTIONS,
} from "@/constants/UserConstants";
import { updateUserById } from "./actions/updateAction";

export default function MembersDataGrid({
  rows,
  beforeJoined = false,
  canDelete = false,
  canUpdate = false,
  canRead = false,
}: {
  rows: UserData[];
  beforeJoined?: boolean;
  canDelete?: boolean;
  canUpdate?: boolean;
  canRead?: boolean;
}) {
  const apiRef = useGridApiRef();
  const router = useRouter();
  const [hasUnsavedRows, setHasUnsavedRows] = React.useState(false);
  const [isSaving, setIsSaving] = React.useState(false);

  // 所属期のオプション配列をメモ化
  const affiliationPeriodValueOptionsArray = React.useMemo(
    () => getSelectableAffiliationPeriods(),
    [],
  );

  // ステータスのオプション配列をメモ化
  const statusValueOptionsArray = React.useMemo(
    () =>
      USER_STATUS_OPTIONS.map((opt) => ({
        value: opt.value,
        label: opt.label,
      })),
    [],
  );
  const [localRows, setLocalRows] = React.useState<UserData[]>(rows);
  React.useEffect(() => {
    setLocalRows(rows);
  }, [rows]);

  const [approveDialogOpen, setApproveDialogOpen] = React.useState(false);
  const [approvedUser, setApprovedUser] = React.useState<UserData | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false);
  const [deletedUserId, setDeletedUserId] = React.useState<null | string>(null);
  const [rejectDialogOpen, setRejectDialogOpen] = React.useState(false);
  const [rejectUser, setRejectUser] = React.useState<UserData | null>(null);
  const unsavedChangesRef = React.useRef<{
    unsavedRows: Record<GridRowId, GridValidRowModel>;
    rowsBeforeChange: Record<GridRowId, GridValidRowModel>;
  }>({
    unsavedRows: {},
    rowsBeforeChange: {},
  });

  /** 行データが完全な情報を持つかチェック (email, statusなどが含まれている) */
  const hasFullData = React.useCallback((row: UserData): boolean => {
    return row.email !== undefined && row.status !== undefined;
  }, []);

  const columns = useMemo<GridColDef[]>(() => {
    return [
      ...(beforeJoined
        ? ([
            {
              field: "actions",
              headerName: "操作",
              type: "actions",
              width: 140,
              getActions: ({ id }: { id: GridRowId }) => {
                const user = localRows.find((u) => String(u.id) === String(id));
                const canApprove =
                  user?.emailVerified === true && user?.discordLinked === true;
                return [
                  <GridActionsCellItem
                    key={"approve"}
                    icon={<CheckIcon />}
                    label="Approve"
                    disabled={!canApprove}
                    onClick={() => {
                      setApprovedUser(user || null);
                      setApproveDialogOpen(true);
                    }}
                  />,
                  <GridActionsCellItem
                    key={"reject"}
                    icon={<DeleteIcon />}
                    label="Reject"
                    onClick={() => {
                      setRejectUser(user || null);
                      setRejectDialogOpen(true);
                    }}
                  />,
                ];
              },
            },
          ] as GridColDef[])
        : ([
            {
              field: "actions",
              headerName: "操作",
              type: "actions",
              width: 140,
              getActions: ({ id }: { id: GridRowId }) => {
                const actions: React.ReactNode[] = [];
                actions.push(
                  <GridActionsCellItem
                    key={"view-detail"}
                    icon={<VisibilityIcon />}
                    label="View detail"
                    onClick={() => {
                      router.push(`/dashboard/members/${id}`);
                    }}
                  />,
                );
                if (canDelete) {
                  actions.push(
                    <GridActionsCellItem
                      key={"delete-row"}
                      icon={<DeleteIcon />}
                      label="削除"
                      onClick={() => {
                        setDeletedUserId(String(id));
                        setDeleteDialogOpen(true);
                      }}
                    />,
                  );
                }
                if (canUpdate) {
                  actions.push(
                    <GridActionsCellItem
                      key={"discard-changes"}
                      icon={<RestoreIcon />}
                      label="Discard changes"
                      disabled={
                        unsavedChangesRef.current.unsavedRows[id] === undefined
                      }
                      onClick={() => {
                        apiRef.current?.updateRows([
                          unsavedChangesRef.current.rowsBeforeChange[id],
                        ]);
                        delete unsavedChangesRef.current.rowsBeforeChange[id];
                        delete unsavedChangesRef.current.unsavedRows[id];
                        setHasUnsavedRows(
                          Object.keys(unsavedChangesRef.current.unsavedRows)
                            .length > 0,
                        );
                      }}
                    />,
                  );
                }
                return actions;
              },
            },
          ] as GridColDef[])),
      { field: "id", headerName: "ID", width: 250 },
      {
        field: "displayName",
        headerName: "表示名",
        width: 150,
        valueGetter: (_value: unknown, row: UserData) =>
          row.profile?.displayName || row.customId || "",
      },
      {
        field: "customId",
        headerName: "カスタムID",
        editable: !beforeJoined && canUpdate,
        width: 150,
      },
      // メール・ステータス等の機密情報は完全なデータがある場合のみ表示&編集可能
      ...(localRows.some((r) => hasFullData(r))
        ? [
            {
              field: "email",
              headerName: "メールアドレス",
              editable: !beforeJoined && canUpdate,
              width: 250,
            } as GridColDef,
            ...(canRead
              ? [
                  {
                    field: "externalEmail",
                    headerName: "外部メールアドレス",
                    width: 250,
                    editable: !beforeJoined && canUpdate,
                  } as GridColDef,
                ]
              : []),
            {
              field: "affiliationPeriod",
              headerName: "所属期",
              editable: !beforeJoined && canUpdate,
              width: 150,
              type: "singleSelect",
              valueOptions: affiliationPeriodValueOptionsArray,
            } as GridColDef,
            {
              field: "status",
              headerName: "ステータス",
              editable: !beforeJoined && canUpdate,
              width: 120,
              type: "singleSelect",
              valueOptions: statusValueOptionsArray,
            } as GridColDef,
            {
              field: "updatedAt",
              headerName: "更新日時",
              width: 150,
              type: "dateTime",
              valueGetter: (value: string) => {
                return value ? new Date(value) : null;
              },
            } as GridColDef,
          ]
        : []),
      {
        field: "joinedAt",
        headerName: "参加日",
        width: 150,
        type: "date",
        valueGetter: (_value: unknown, row: UserData) => {
          const d = row.profile?.joinedAt;
          return d ? new Date(d) : null;
        },
        valueSetter: (value, row) => {
          const date =
            value instanceof Date
              ? value
              : typeof value === "string"
                ? new Date(`${value}T00:00:00`)
                : new Date(NaN);

          if (Number.isNaN(date.getTime())) {
            return row; // 無効な日付の場合は変更を適用しない
          }

          return {
            ...row,
            profile: {
              ...row.profile,
              joinedAt: date.toLocaleDateString("sv-SE", {
                timeZone: "Asia/Tokyo",
              }),
            },
          };
        },
        editable: true,
      },
      ...(beforeJoined
        ? [
            {
              field: "emailVerified",
              headerName: "メール認証",
              width: 120,
              renderCell: (params) => {
                const verified = params.row.emailVerified === true;
                return verified ? "✓ 認証済み" : "✗ 未認証";
              },
            } as GridColDef,
            {
              field: "discordLinked",
              headerName: "Discord連携",
              width: 120,
              renderCell: (params) => {
                const linked = params.row.discordLinked === true;
                return linked ? "✓ 連携済み" : "✗ 未連携";
              },
            } as GridColDef,
          ]
        : []),
      {
        field: "createdAt",
        headerName: "作成日時",
        width: 150,
        type: "dateTime",
        valueGetter: (value: string | undefined) => {
          return value ? new Date(value) : null;
        },
      },
    ];
  }, [
    apiRef,
    beforeJoined,
    localRows,
    affiliationPeriodValueOptionsArray,
    statusValueOptionsArray,
    router,
    hasFullData,
    canDelete,
    canUpdate,
    canRead,
  ]);

  const handleDelete = async (
    _prevState: FormStatus | null,
    _formData: FormData | null,
  ) => {
    void _prevState;
    void _formData;
    try {
      if (!deletedUserId) {
        return {
          status: "error",
          message: "削除対象が選択されていません",
        } as FormStatus;
      }

      const result = await deleteUserById(deletedUserId);

      if (!result.success) {
        return {
          status: "error",
          message: result.error || "削除に失敗しました",
        } as FormStatus;
      }

      setLocalRows((prev) =>
        prev.filter((r) => String(r.id) !== deletedUserId),
      );
      apiRef.current?.updateRows([{ id: deletedUserId, _action: "delete" }]);
      setDeletedUserId(null);
      setDeleteDialogOpen(false);
      return {
        status: "success",
        message: "ユーザーを削除しました",
      } as FormStatus;
    } catch (error) {
      return {
        status: "error",
        message: `削除に失敗しました: ${String(error)}`,
      } as FormStatus;
    }
  };

  const processRowUpdate = React.useCallback<
    NonNullable<DataGridProps["processRowUpdate"]>
  >((newRow, oldRow) => {
    const rowId = newRow.id;
    unsavedChangesRef.current.unsavedRows[rowId] = newRow;
    if (!unsavedChangesRef.current.rowsBeforeChange[rowId]) {
      unsavedChangesRef.current.rowsBeforeChange[rowId] = oldRow;
    }
    setHasUnsavedRows(true);
    return newRow;
  }, []);

  const discardChanges = React.useCallback(() => {
    setHasUnsavedRows(false);
    Object.values(unsavedChangesRef.current.rowsBeforeChange).forEach((row) => {
      apiRef.current?.updateRows([row]);
    });
    unsavedChangesRef.current = {
      unsavedRows: {},
      rowsBeforeChange: {},
    };
  }, [apiRef]);

  const saveChanges = React.useCallback(async () => {
    try {
      setIsSaving(true);
      const unsaved = Object.values(unsavedChangesRef.current.unsavedRows);
      for (const row of unsaved) {
        try {
          // 機密情報（email など）が含まれているか確認
          if (!row.email) {
            console.warn(
              `User ${row.id} has no email data. Make sure you have access to edit this user.`,
            );
          }

          const payload: Partial<
            Omit<UserData, "id" | "createdAt" | "updatedAt" | "deletedAt">
          > = {
            customId: row.customId as string | undefined,
            email: row.email as string | undefined,
            externalEmail: row.externalEmail as string | undefined,
            affiliationPeriod: row.affiliationPeriod as string | undefined,
            status: row.status as UserStatus | undefined,
            profile: (row.profile as unknown as ProfileData) ?? undefined,
          };
          const savedPlain = await updateUserById(row.id, payload);
          apiRef.current?.updateRows([savedPlain]);
          setLocalRows((prev) =>
            prev.map((r) =>
              String(r.id) === String(savedPlain.id) ? savedPlain : r,
            ),
          );
        } catch (err) {
          console.error("Save row failed:", err);
        }
      }
      setIsSaving(false);
      setHasUnsavedRows(false);
      enqueueSnackbar("変更を保存しました", { variant: "success" });
      unsavedChangesRef.current = {
        unsavedRows: {},
        rowsBeforeChange: {},
      };
    } catch {
      setIsSaving(false);
    }
  }, [apiRef]);

  const getRowClassName = React.useCallback<
    NonNullable<DataGridProps["getRowClassName"]>
  >(({ id }) => {
    const unsavedRow = unsavedChangesRef.current.unsavedRows[id];
    if (unsavedRow) {
      if (unsavedRow._action === "delete") {
        return "row--removed";
      }
      return "row--edited";
    }
    return "";
  }, []);

  return (
    <div style={{ width: "100%" }}>
      {!beforeJoined && canUpdate && (
        <div style={{ marginBottom: 8 }}>
          <Button
            disabled={!hasUnsavedRows}
            loading={isSaving}
            onClick={saveChanges}
            startIcon={<SaveIcon />}
            loadingPosition="start"
          >
            <span>変更を保存</span>
          </Button>
          <Button
            disabled={!hasUnsavedRows || isSaving}
            onClick={discardChanges}
            startIcon={<RestoreIcon />}
          >
            変更を元に戻す
          </Button>
        </div>
      )}
      <div>
        <DataGrid
          rows={localRows}
          columns={columns}
          apiRef={apiRef}
          disableRowSelectionOnClick
          processRowUpdate={processRowUpdate}
          ignoreValueFormatterDuringExport
          initialState={{
            columns: {
              columnVisibilityModel: {
                id: false,
              },
            },
          }}
          showToolbar
          sx={{
            [`& .${gridClasses.row}.row--removed`]: {
              backgroundColor: (theme) => {
                if (theme.palette.mode === "light") {
                  return "rgba(255, 170, 170, 0.3)";
                }
                return darken("rgba(255, 170, 170, 1)", 0.7);
              },
            },
            [`& .${gridClasses.row}.row--edited`]: {
              backgroundColor: (theme) => {
                if (theme.palette.mode === "light") {
                  return "rgba(255, 254, 176, 0.3)";
                }
                return darken("rgba(255, 254, 176, 1)", 0.6);
              },
            },
          }}
          localeText={jaJP.components.MuiDataGrid.defaultProps.localeText}
          loading={isSaving}
          getRowClassName={getRowClassName}
          autoHeight
        />
      </div>
      <ApproveRegistApplyDialog
        open={approveDialogOpen}
        handleClose={() => setApproveDialogOpen(false)}
        user={approvedUser}
      />
      <RejectDialog
        open={rejectDialogOpen}
        user={rejectUser}
        handleClose={() => setRejectDialogOpen(false)}
      />
      {canDelete && (
        <DeleteDialog
          open={deleteDialogOpen}
          dataAction={handleDelete}
          handleClose={() => setDeleteDialogOpen(false)}
          title="ユーザー"
        />
      )}
    </div>
  );
}
