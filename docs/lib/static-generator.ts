import * as fs from "fs";
import * as path from "path";
import * as cheerio from "cheerio";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

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

	// 2. Process CSS (Inlining into <style data-last-style>)
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

	// 3. Process JS (Inlining into <script data-last-script>)
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

	// 4. Final Output Logic
	const finalHtml = $.html();

	return finalHtml;
}
