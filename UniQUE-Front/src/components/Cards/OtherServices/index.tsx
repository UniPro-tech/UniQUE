"use client";
import DesignServicesIcon from "@mui/icons-material/DesignServices";
import { Card, CardContent, Grid, Typography } from "@mui/material";
import Image from "next/image";
import Link from "next/link";

export interface OtherServiceCardProps {
  serviceName: string;
  url: string;
  iconURL?: string | null;
  icon?: React.ReactNode;
}

export default function OtherServicesCard({
  serviceName,
  url,
  iconURL,
  icon,
}: OtherServiceCardProps) {
  return (
    <Grid
      size={{ xs: 6, sm: 4, md: 3, lg: 2 }}
      component={Link}
      href={url}
      target={"_blank"}
      style={{ textDecoration: "none" }}
      sx={{
        width: 160,
        height: 180,
      }}
    >
      <Card
        variant="outlined"
        sx={{
          width: 160,
          height: 180,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <CardContent
          sx={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            gap: 1,
            padding: 2,
            width: "100%",
          }}
        >
          {icon ? (
            icon
          ) : iconURL ? (
            <Image
              src={iconURL}
              alt={`${serviceName} icon`}
              width={50}
              height={50}
            />
          ) : (
            <DesignServicesIcon fontSize="large" sx={{ fontSize: 50 }} />
          )}
          <Typography variant={"subtitle1"}>{serviceName}</Typography>
        </CardContent>
      </Card>
    </Grid>
  );
}
