import { fileURLToPath } from "url";
import express from "express";
import path from "path";
import { buildPresentation } from "../lib/static-generator";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const app = express();
const PORT = process.env.PORT || 8080;

// Define directory paths
const STATIC_DIR = path.join(__dirname, "../static");
const REVEAL_DIR = path.join(__dirname, "../node_modules/reveal.js/dist");

// Serve reveal.js files
app.use("/reveal.js", express.static(REVEAL_DIR));

app.get("/", async (_req, res) => {
	return res.send(await buildPresentation());
});

// Serve static files
app.use(express.static(STATIC_DIR));

// 404 handler
app.use((_req, res) => {
	res.status(404).sendFile(path.join(STATIC_DIR, "404.html"));
});

app.listen(PORT, () => {
	console.log(`Server running at http://localhost:${PORT}/`);
});
