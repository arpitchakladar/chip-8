import express from "express";
import path from "path";

const app = express();
const PORT = process.env.PORT || 8080;

// Define directory paths
const ASSETS_DIR = path.join(__dirname, "../assets");
const STATIC_DIR = path.join(__dirname, "../pages");
const REVEAL_DIR = path.join(__dirname, "../node_modules/reveal.js/dist");

// Serve reveal.js files under the /reveal.js prefix
app.use("/reveal.js", express.static(REVEAL_DIR));

// Serve assets under the /assets prefix
app.use("/assets", express.static(ASSETS_DIR));

// Serve your main pages (index.html, etc.) from the root
app.use(express.static(STATIC_DIR));

// Custom 404 Handler
app.use((req, res) => {
	res.status(404).sendFile(path.join(STATIC_DIR, "404.html"));
});

app.listen(PORT, () => {
	console.log(`Server running at http://localhost:${PORT}/`);
});
