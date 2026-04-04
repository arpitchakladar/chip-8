import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";
import { buildPresentation } from "../lib/static-generator.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

(async function () {
	// Define paths
	const distDir = path.join(__dirname, "../build");
	const distPath = path.join(distDir, "index.html");

	const revealSource = path.join(__dirname, "../node_modules/reveal.js/dist");
	const revealDest = path.join(distDir, "reveal.js");

	const assetsSource = path.join(__dirname, "../static/assets");
	const assetsDest = path.join(distDir, "assets");

	// 1. Generate the HTML content
	const finalHtml = await buildPresentation();

	// 2. Ensure the build directory exists
	if (!fs.existsSync(distDir)) {
		fs.mkdirSync(distDir, { recursive: true });
	}

	// 3. Write the index.html
	fs.writeFileSync(distPath, finalHtml);

	// 4. Copy reveal.js dist folder
	if (fs.existsSync(revealSource)) {
		fs.cpSync(revealSource, revealDest, { recursive: true });
		console.log("✓ Copied: reveal.js core files");
	} else {
		console.warn("⚠ Warning: node_modules/reveal.js/dist not found!");
	}

	// 5. Copy static assets
	if (fs.existsSync(assetsSource)) {
		fs.cpSync(assetsSource, assetsDest, { recursive: true });
		console.log("✓ Copied: static assets");
	}

	console.log("✓ Build complete! Check the /build folder.");
})();
