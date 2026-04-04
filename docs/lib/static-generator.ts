import * as fs from "fs";
import * as path from "path";
import * as cheerio from "cheerio";

import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export async function buildPresentation() {
	const layoutPath = path.join(__dirname, "../static", "index.html"); // Your template
	const distPath = path.join(__dirname, "../build/index.html"); // Output destination

	// 1. Read the template
	const templateHtml = fs.readFileSync(layoutPath, "utf8");
	const $ = cheerio.load(templateHtml);

	// 2. Process Slides (Inject into .slides)
	const slidesDir = path.join(__dirname, "../static", "slides");
	const slideFiles = fs.readdirSync(slidesDir).sort(); // Sorts 01, 02, etc.

	const slidesContainer = $(".slides");
	slidesContainer.empty(); // Clear existing placeholders

	slideFiles.forEach((file) => {
		const content = fs.readFileSync(path.join(slidesDir, file), "utf8");
		slidesContainer.append(content);
		console.log(`✓ Injected Slide: ${file}`);
	});

	// 3. Process CSS (Inject into <head>)
	const stylesDir = path.join(__dirname, "../static", "styles");
	const styleFiles = fs.readdirSync(stylesDir);

	styleFiles.forEach((file) => {
		// We use the relative path for the href
		$("head").append(`<link rel="stylesheet" href="styles/${file}">\n`);
	});

	// 4. Process JS Scripts (Inject into <body>)
	const scriptsDir = path.join(__dirname, "../static", "scripts");
	const scriptFiles = fs.readdirSync(scriptsDir);

	scriptFiles.forEach((file) => {
		$("body").append(`<script src="scripts/${file}"></script>\n`);
	});

	// 5. Final Reveal.js Initialization
	// We add this at the very end so it runs after all scripts are loaded
	$("body").append(`
		<script defer>
			(async function () {
				await Promise.all([
					loadCode(),
					cycleBackground(),
					prepareCodeFragments(),
				]);

				Reveal.initialize({
					hash: true,
					keyboard: true,
					controls: false,
					plugins: [RevealMarkdown, RevealHighlight],
				});
			})();
		</script>
	`);

	return $.html();

	// 6. Write the final file
	if (!fs.existsSync(path.dirname(distPath)))
		fs.mkdirSync(path.dirname(distPath));
	fs.writeFileSync(distPath, $.html());

	console.log("\nBuild complete! Check the dist/ folder.");
}
