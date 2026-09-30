import { serve } from "bun";
import { app } from "./src/app";
import { hostname, port } from "./src/config";

serve({ hostname, port, fetch: app.fetch });
