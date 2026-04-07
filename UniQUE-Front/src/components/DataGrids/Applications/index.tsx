import { Application, type ApplicationData } from "@/classes/Application";
import ApplicationsDataGridClient from "./Client";

export default async function ApplicationsDataGrid({
  rows,
}: {
  rows?: Application[] | ApplicationData[];
}) {
  if (!rows) {
    const fetched = await Application.getAll();
    rows = await Promise.all(fetched.map((app) => app.toJson()));
  } else {
    rows = await Promise.all(
      rows.map((app) => {
        if (app instanceof Application) {
          return app.toJson();
        }
        return app;
      }),
    );
  }

  return <ApplicationsDataGridClient rows={rows} />;
}
