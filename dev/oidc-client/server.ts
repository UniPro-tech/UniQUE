import { serve } from "bun";
import { app } from "./src/app";
import { port } from "./src/config";

serve({ port, fetch: app.fetch });
