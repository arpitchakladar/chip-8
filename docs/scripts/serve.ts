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

/**
 * API: Fetch code from GitHub
 * Example:
 * /code?file=emulator/cpu.go&start=for {&end=}
 */
app.get("/code", async (req, res) => {
	try {
		const { file, start, end } = req.query;

		if (!file) {
			return res.status(400).send("Missing 'file' parameter");
		}

		const GITHUB_RAW_BASE =
			"https://raw.githubusercontent.com/arpitchakladar/chip-8/refs/heads/master/";

		const url = GITHUB_RAW_BASE + file;
		console.log(url);

		const response = await fetch(url);
		if (!response.ok) {
			throw new Error("Failed to fetch file");
		}

		let text = await response.text();

		// slicing
		if (start) {
			const startIdx = text.indexOf(start);
			if (startIdx !== -1) {
				let endIdx = text.length;

				if (end) {
					const tempEnd = text.indexOf(end, startIdx + start.length);
					if (tempEnd !== -1) {
						endIdx = tempEnd + end.length;
					}
				}

				text = text.slice(startIdx, endIdx);
			}
		}

		res.setHeader("Content-Type", "text/plain");
		res.send(text);
	} catch (err) {
		console.error(err);
		res.status(500).send("Error fetching code");
	}
});

// 404 handler
app.use((req, res) => {
	res.status(404).sendFile(path.join(STATIC_DIR, "404.html"));
});

app.listen(PORT, () => {
	console.log(`Server running at http://localhost:${PORT}/`);
});
