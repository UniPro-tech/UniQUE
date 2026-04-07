"use client";
import AdminPanelSettingsIcon from "@mui/icons-material/AdminPanelSettings";
import AppsIcon from "@mui/icons-material/Apps";
import CampaignIcon from "@mui/icons-material/Campaign";
import ChevronLeftIcon from "@mui/icons-material/ChevronLeft";
import ChevronRightIcon from "@mui/icons-material/ChevronRight";
import HomeIcon from "@mui/icons-material/Home";
import MenuIcon from "@mui/icons-material/Menu";
import PeopleIcon from "@mui/icons-material/People";
import PersonAddIcon from "@mui/icons-material/PersonAdd";
import { Stack } from "@mui/material";
import MuiAppBar, {
  type AppBarProps as MuiAppBarProps,
} from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import CssBaseline from "@mui/material/CssBaseline";
import Divider from "@mui/material/Divider";
import MuiDrawer from "@mui/material/Drawer";
import IconButton from "@mui/material/IconButton";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import {
  type CSSObject,
  styled,
  type Theme,
  useTheme,
} from "@mui/material/styles";
import Toolbar from "@mui/material/Toolbar";
import Image from "next/image";
import Link from "next/link";
import * as React from "react";
import type { RoleData } from "@/classes/Role";
import type { UserData } from "@/classes/types/User";
import { PermissionBitsFields } from "@/constants/Permission";
import DiscordLinkAlert from "./DiscordLinkAlert";
import AccountButton from "./Sidebar/AccountButton";

const drawerWidth = 240;

const openedMixin = (theme: Theme): CSSObject => ({
  width: drawerWidth,
  transition: theme.transitions.create("width", {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.enteringScreen,
  }),
  overflowX: "hidden",
});

const closedMixin = (theme: Theme): CSSObject => ({
  transition: theme.transitions.create("width", {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  overflowX: "hidden",
  width: `calc(${theme.spacing(7)} + 1px)`,
  [theme.breakpoints.up("sm")]: {
    width: `calc(${theme.spacing(8)} + 1px)`,
  },
});

const DrawerHeader = styled("div")(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-end",
  padding: theme.spacing(0, 1),
  ...theme.mixins.toolbar,
}));

interface AppBarProps extends MuiAppBarProps {
  open?: boolean;
}

const AppBar = styled(MuiAppBar, {
  shouldForwardProp: (prop) => prop !== "open",
})<AppBarProps>(({ theme }) => ({
  zIndex: theme.zIndex.drawer + 1,
  transition: theme.transitions.create(["width", "margin"], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  variants: [
    {
      props: ({ open }) => open,
      style: {
        marginLeft: drawerWidth,
        width: `calc(100% - ${drawerWidth}px)`,
        transition: theme.transitions.create(["width", "margin"], {
          easing: theme.transitions.easing.sharp,
          duration: theme.transitions.duration.enteringScreen,
        }),
      },
    },
  ],
}));

const Drawer = styled(MuiDrawer, {
  shouldForwardProp: (prop) => prop !== "open",
})(({ theme }) => {
  return {
    width: drawerWidth,
    whiteSpace: "nowrap",
    boxSizing: "border-box",
    variants: [
      {
        props: ({ open }) => open,
        style: {
          ...openedMixin(theme),
          "& .MuiDrawer-paper": openedMixin(theme),
        },
      },
      {
        props: ({ open }) => !open,
        style: {
          ...closedMixin(theme),
          "& .MuiDrawer-paper": closedMixin(theme),
          "@media (max-width: 600px)": {
            display: "none",
          },
        },
      },
    ],
  };
});

interface NavLink {
  text: string;
  href: string;
  icon: React.ReactNode;
}

export default function MiniDrawer({
  user,
  userRoles,
  children,
}: {
  user: UserData | null;
  userRoles?: RoleData[];
  children: React.ReactNode;
}) {
  /**
   * 権限チェックヘルパー関数
   * 複数の権限幅を持つロールの場合、オブジェクトで渡して（OR結合で）チェック
   */
  const hasPermission = (requiredFlag: bigint): boolean => {
    return (
      userRoles?.some(
        (role) => (BigInt(role.permissionBitmask) & requiredFlag) !== 0n,
      ) ?? false
    );
  };

  const canManageRequests = hasPermission(PermissionBitsFields.USER_CREATE);
  const canManageRoles = hasPermission(PermissionBitsFields.ROLE_MANAGE);

  const NAVIGSTION_LINKS: NavLink[][] = [
    [
      { text: "ダッシュボード", href: "/dashboard", icon: <HomeIcon /> },
      {
        text: "アナウンス",
        href: "/dashboard/announcements",
        icon: <CampaignIcon />,
      },
      {
        text: "メンバー管理",
        href: "/dashboard/members",
        icon: <PeopleIcon />,
      },
      {
        text: "アプリケーション管理",
        href: "/dashboard/applications",
        icon: <AppsIcon />,
      },
    ],
    ...(canManageRequests || canManageRoles
      ? [
          [
            ...(canManageRequests
              ? [
                  {
                    text: "メンバー申請管理",
                    href: "/dashboard/requests",
                    icon: <PersonAddIcon />,
                  },
                ]
              : []),
            ...(canManageRoles
              ? [
                  {
                    text: "ロール管理",
                    href: "/dashboard/roles",
                    icon: <AdminPanelSettingsIcon />,
                  },
                ]
              : []),
          ],
        ]
      : []),
  ];

  const theme = useTheme();
  const [open, setOpen] = React.useState(false);

  const handleDrawerOpen = () => {
    setOpen(true);
  };

  const handleDrawerClose = () => {
    setOpen(false);
  };

  return (
    <Box sx={{ display: "flex" }}>
      <CssBaseline />
      <AppBar
        position="fixed"
        open={open}
        style={{
          background: "#e6ecec",
          color: "#000000",
        }}
        variant="outlined"
        suppressHydrationWarning
      >
        <Toolbar>
          <Stack
            direction="row"
            spacing={2}
            alignItems="center"
            justifyContent={"center"}
            justifyItems={"center"}
            margin={0}
            padding={0}
          >
            <IconButton
              color="inherit"
              aria-label="open drawer"
              onClick={handleDrawerOpen}
              edge="start"
              sx={[open && { display: "none" }]}
            >
              <MenuIcon />
            </IconButton>
            <Link href="/dashboard">
              <Image
                src="/unique_logotype.png"
                alt="UniQUE Logo"
                width={128}
                height={42}
                style={{ display: "block", objectFit: "contain" }}
              />
            </Link>
          </Stack>
          <Box sx={{ flexGrow: 1 }} />
          <AccountButton user={user} />
        </Toolbar>
        <DiscordLinkAlert />
      </AppBar>
      <Drawer variant="permanent" open={open}>
        <DrawerHeader>
          <IconButton onClick={handleDrawerClose}>
            {theme.direction === "rtl" ? (
              <ChevronRightIcon />
            ) : (
              <ChevronLeftIcon />
            )}
          </IconButton>
        </DrawerHeader>
        <Divider />
        {NAVIGSTION_LINKS.map((section, index) => (
          <div key={index}>
            <List key={index}>
              {section.map((link) => (
                <ListItem
                  key={link.text}
                  disablePadding
                  sx={{ display: "block" }}
                >
                  <ListItemButton
                    href={link.href}
                    sx={{
                      minHeight: 48,
                      justifyContent: open ? "initial" : "center",
                      px: 2.5,
                    }}
                  >
                    <ListItemIcon
                      sx={{
                        minWidth: 0,
                        mr: open ? 1 : "auto",
                        justifyContent: "center",
                      }}
                    >
                      {link.icon}
                    </ListItemIcon>
                    <ListItemText
                      primary={link.text}
                      sx={{ opacity: open ? 1 : 0 }}
                    />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
            {index < NAVIGSTION_LINKS.length - 1 && (
              <Divider key={`divider-${index}`} />
            )}
          </div>
        ))}
      </Drawer>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: { xs: 3, sm: 2, md: 3 },
          width: { xs: "100%", sm: "auto" },
          maxWidth: "100%",
          overflow: "auto",
          boxSizing: "border-box",
        }}
      >
        <DrawerHeader />
        {children}
      </Box>
    </Box>
  );
}
