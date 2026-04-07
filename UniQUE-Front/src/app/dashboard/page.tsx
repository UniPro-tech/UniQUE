import AccountBalanceIcon from "@mui/icons-material/AccountBalance";
import EventIcon from "@mui/icons-material/Event";
import { Box, Button, Grid, Stack, Typography } from "@mui/material";
import Image from "next/image";
import Link from "next/link";
import { Announcement } from "@/classes/Announcement";
import OtherServicesCard, {
  type OtherServiceCardProps,
} from "@/components/Cards/OtherServices";
import AnnouncementsList from "@/components/Lists/AnnouncementsList";

export const metadata = {
  title: "ダッシュボード",
  description: "ユーザーダッシュボードページ",
};

export const OTHER_SERVICES: OtherServiceCardProps[] = [
  {
    serviceName: "Event Manager",
    url: "https://event.uniproject.jp",
    icon: <EventIcon fontSize="large" sx={{ fontSize: 50 }} />,
  },
  {
    serviceName: "役員会サイト",
    url: "https://boardmembers.uniproject.jp",
    icon: <AccountBalanceIcon fontSize="large" sx={{ fontSize: 50 }} />,
  },
  {
    serviceName: "UniWiki",
    url: "https://wiki.uniproject.jp",
    iconURL: "/otherServices/growi.svg",
  },
  {
    serviceName: "Proxmox VE",
    url: "https://pm.uniproject.jp",
    iconURL: "/otherServices/proxmox.png",
  },
  {
    serviceName: "蔵雲",
    url: "https://cloud.uniproject.jp",
    icon: (
      <Image
        src="/otherServices/nextcloud.png"
        alt="蔵雲 icon"
        width={70}
        height={50}
      />
    ),
  },
  {
    serviceName: "Netbox",
    url: "https://netbox.uniproject.jp",
    iconURL: "/otherServices/netbox.png",
  },
];

export default async function DashboardPage() {
  const anns = (await Announcement.getAll()).map((a) => a.toJson());

  return (
    <Stack>
      <Typography variant="h4" gutterBottom>
        おかえりなさい！
      </Typography>
      <Typography variant="body1">
        ダッシュボードへようこそ。左のサイドバーからナビゲートしてください。
      </Typography>
      <Stack mt={4} spacing={2}>
        <Typography variant="h5">サービス一覧</Typography>
        <Grid
          container
          columns={12}
          mt={4}
          justifyContent={"start"}
          alignItems={"start"}
          justifyItems={"start"}
          spacing={2}
          id="other-services"
          maxWidth={`${160 * 6 + 16 * 5}px`}
        >
          {OTHER_SERVICES.map((service) => (
            <OtherServicesCard
              key={service.serviceName}
              serviceName={service.serviceName}
              url={service.url}
              icon={service.icon}
              iconURL={service.iconURL}
            />
          ))}
        </Grid>
      </Stack>
      <Stack mt={4} spacing={2} id="announce">
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
        >
          <Typography variant="h5">最近のお知らせ</Typography>
          <Link
            href="/dashboard/announcements"
            style={{ textDecoration: "none" }}
          >
            <Button size="small" variant="outlined">
              お知らせ管理
            </Button>
          </Link>
        </Stack>
        <Box>
          <AnnouncementsList initial={anns} canUpdate={false} />
        </Box>
        <div
          style={{ display: "flex", justifyContent: "center", marginTop: 8 }}
        >
          <Link
            href="/dashboard/announcements"
            style={{ textDecoration: "none" }}
          >
            <Button size="small" variant="outlined">
              もっと見る
            </Button>
          </Link>
        </div>
      </Stack>
    </Stack>
  );
}
