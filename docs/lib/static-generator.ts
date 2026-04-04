import * as fs from "fs";
import * as path from "path";
import * as cheerio from "cheerio";
import { fileURLToPath } from "url";
import { minify } from "html-minifier-terser";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function injectRemoteCode($: cheerio.CheerioAPI) {
	const codeBlocks = $("[data-load-code]");
	const baseUrl =
		"https://raw.githubusercontent.com/arpitchakladar/chip-8/refs/heads/master/";

	console.log(
		`\n--- Fetching Remote Assets (${codeBlocks.length} found) ---`,
	);

	await Promise.all(
		codeBlocks.toArray().map(async (el) => {
			const $block = $(el);
			const fileName = $block.attr("data-load-code");
			const url = `${baseUrl}${fileName}`;

			try {
				const response = await fetch(url);
				if (!response.ok) throw new Error(`HTTP ${response.status}`);

				const text = await response.text();

				// Inject the raw text (Cheerio handles encoding < and >)
				$block.text(text);

				// Remove attribute to clean up production HTML
				// $block.removeAttr("data-load-code");

				console.log(`  ✓ Fetched: ${fileName}`);
			} catch (err) {
				console.error(`  × Failed: ${fileName} (${(err as any).message})`);
				$block.text(`// Error loading remote code: ${fileName}`);
			}
		}),
	);
}

export async function buildPresentation() {
	const layoutPath = path.join(__dirname, "../static", "index.html");

	const templateHtml = fs.readFileSync(layoutPath, "utf8");
	const $ = cheerio.load(templateHtml);

	// 1. Process Slides
	const slidesDir = path.join(__dirname, "../static", "slides");
	const slideFiles = fs.readdirSync(slidesDir).sort();
	const slidesContainer = $(".slides");
	slidesContainer.empty();

	slideFiles.forEach((file) => {
		const content = fs.readFileSync(path.join(slidesDir, file), "utf8");
		slidesContainer.append(content);
	});

	// 2. Populated all of the code snippets from github
	await injectRemoteCode($);

	// 3. Process CSS (Inlining into <style data-last-style>)
	const stylesDir = path.join(__dirname, "../static", "styles");
	const styleFiles = fs.readdirSync(stylesDir);
	const targetStyleTag = $("style[data-last-style]");

	styleFiles.forEach((file) => {
		const cssContent = fs.readFileSync(path.join(stylesDir, file), "utf8");
		const styleTag = `<style>\n${cssContent}\n</style>`;
		if (targetStyleTag.length > 0) {
			// Prepend ensures new CSS is above the "last" CSS content
			targetStyleTag.before(styleTag);
		} else {
			$("head").append(styleTag);
		}
	});

	// 4. Process JS (Inlining into <script data-last-script>)
	const scriptsDir = path.join(__dirname, "../static", "scripts");
	const scriptFiles = fs.readdirSync(scriptsDir);
	const targetScriptTag = $("script[data-last-script]");

	scriptFiles.forEach((file) => {
		const jsContent = fs.readFileSync(path.join(scriptsDir, file), "utf8");
		const scriptTag = `<script>\n${jsContent}\n</script>`;
		if (targetScriptTag.length > 0) {
			// Prepend content so original content stays at the bottom
			targetScriptTag.before(scriptTag);
		} else {
			$("body").append(scriptTag);
		}
	});

	// 5. Final Output Logic
	const rawHtml = $.html();

	// 6. Minify the output html
	const minifiedHtml = await minify(rawHtml, {
		collapseWhitespace: true,
		removeComments: true,
		minifyJS: true, // This minifies code inside <script> tags
		minifyCSS: true, // This minifies code inside <style> tags
		processConditionalComments: true,
		removeEmptyAttributes: true,
		decodeEntities: true,
	});

	return minifiedHtml;
}
