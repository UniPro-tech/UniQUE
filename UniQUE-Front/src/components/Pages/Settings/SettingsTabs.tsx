"use client";

import { Box, Tab, Tabs } from "@mui/material";
import React from "react";

type Props = {
  labels: string[];
  children: React.ReactNode;
  defaultIndex?: number;
};

function a11yProps(index: number) {
  return {
    id: `settings-tab-${index}`,
    "aria-controls": `settings-tabpanel-${index}`,
  };
}

export default function SettingsTabs({
  labels,
  children,
  defaultIndex = 0,
}: Props) {
  const [value, setValue] = React.useState(defaultIndex);

  const handleChange = (_event: React.SyntheticEvent, newValue: number) => {
    setValue(newValue);
  };

  const childrenArray = React.Children.toArray(children);

  return (
    <Box>
      <Tabs
        value={value}
        onChange={handleChange}
        variant="scrollable"
        scrollButtons="auto"
        aria-label="settings tabs"
        sx={{ mb: 2 }}
      >
        {labels.map((label, i) => (
          <Tab key={label} label={label} {...a11yProps(i)} />
        ))}
      </Tabs>

      {childrenArray.map((child, i) => (
        <div
          key={labels[i]}
          role="tabpanel"
          id={`settings-tabpanel-${i}`}
          aria-labelledby={`settings-tab-${i}`}
          hidden={value !== i}
        >
          {value === i && <Box sx={{ pt: 2, pb: 2, px: 0 }}>{child}</Box>}
        </div>
      ))}
    </Box>
  );
}
