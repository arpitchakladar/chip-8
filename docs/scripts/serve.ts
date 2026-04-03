import express from "express";
import path from "path";

const app = express();
const PORT = process.env.PORT || 8080;

// Define directory paths
const STATIC_DIR = path.join(__dirname, "../static");
const REVEAL_DIR = path.join(__dirname, "../node_modules/reveal.js/dist");

// Serve reveal.js files
app.use("/reveal.js", express.static(REVEAL_DIR));

// Serve static files
app.use(express.static(STATIC_DIR));

// 404 handler
app.use((req, res) => {
	res.status(404).sendFile(path.join(STATIC_DIR, "404.html"));
});

app.listen(PORT, () => {
	console.log(`Server running at http://localhost:${PORT}/`);
});
