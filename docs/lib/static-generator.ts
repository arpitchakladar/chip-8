import * as fs from "fs";
import * as path from "path";
import * as cheerio from "cheerio";
import { fileURLToPath } from "url";
import { minify } from "html-minifier-terser";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Finds all <link> and <script> tags pointing to reveal.js/ paths,
 * reads the files from node_modules/reveal.js/dist, and inlines them.
 */
function inlineRevealAssets($: cheerio.CheerioAPI) {
	const revealDistPath = path.join(
		__dirname,
		"../node_modules/reveal.js/dist",
	);

	// 1. Handle Reveal.js CSS
	$('link[href^="reveal.js/"]').each((_, el) => {
		const $el = $(el);
		const fileName = $el.attr("href")!.replace("reveal.js/", "");
		const filePath = path.join(revealDistPath, fileName);

		if (fs.existsSync(filePath)) {
			const css = fs.readFileSync(filePath, "utf8");
			$el.replaceWith(`<style>\n${css}\n</style>`);
			console.log(`	✓ Inlined Reveal CSS: ${fileName}`);
		} else {
			console.warn(`	× Reveal CSS not found: ${filePath}`);
		}
	});

	// 2. Handle Reveal.js JS
	$('script[src^="reveal.js/"]').each((_, el) => {
		const $el = $(el);
		const fileName = $el.attr("src")!.replace("reveal.js/", "");
		const filePath = path.join(revealDistPath, fileName);

		if (fs.existsSync(filePath)) {
			const js = fs.readFileSync(filePath, "utf8");
			$el.replaceWith(`<script>\n${js}\n</script>`);
			console.log(`	✓ Inlined Reveal JS: ${fileName}`);
		} else {
			console.warn(`	× Reveal JS not found: ${filePath}`);
		}
	});
}

/**
 * Pre-generates code fragment popups and GitHub links at build time.
 * NOTE: Should be called before injectRemoteCode
 */
function prepareCodeFragments($: cheerio.CheerioAPI) {
	const containers = $("[data-code-snippet]");

	containers.each((_, container) => {
		const $container = $(container);

		// 1. Generate the GitHub Link
		const codeSnippetElement = $container.find("[data-load-code]");
		const codePath = codeSnippetElement.attr("data-load-code");

		if (codePath) {
			const githubUrl = `https://github.com/arpitchakladar/chip-8/blob/master/${codePath}`;
			const codeFilePathLink = `
				<a href="${githubUrl}" class="secondary code-path">
				${codePath}
				</a>
			`;

			$container.prepend(codeFilePathLink);
		} else {
			$container.addClass("no-code-link");
		}

		// 2. Process Fragment Templates
		const templates = $container.find("[data-code-fragment]");

		templates.each((sequence, el) => {
			const $el = $(el);

			// Extract attributes
			const indexes = $el.attr("data-code-fragment-indexes") || "0";
			const title = $el.attr("data-code-fragment-title") || "No title";
			const text = $el.attr("data-code-fragment-text") || "No text";

			// Determine color (alternating)
			const color = sequence % 2 === 0 ? "var(--blue)" : "var(--red)";

			// Create popups for each index
			indexes.split(",").forEach((fragmentIndex) => {
				const cleanIndex = fragmentIndex.trim();

				const popup = `
					<div class="fragment fade-in-then-out"
						data-fragment-index="${cleanIndex}"
						data-code-fragment-element=""
						style="--code-fragment-element-color: ${color}"
					>
						<h5 style="margin: 0 0 10px 0; color: ${color}; font-size: 0.5em;">${title}</h5>
						<p style="font-size: 0.5em; margin: 0; line-height: 1.4;">${text}</p>
					</div>
				`;

				$container.append(popup);
			});

			// Remove the original template element from the final HTML
			$el.remove();
		});
	});
}

/**
 * Pre-loads code snippets from github at build time/serverside
 * NOTE: Should be called after prepareCodeFragments
 */
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
				// Padd the end with 20 empty lines (for better scrolling)
				$block.text(text + "\n".repeat(20));

				// Remove attribute to clean up production HTML
				$block.removeAttr("data-load-code");

				console.log(`	✓ Fetched: ${fileName}`);
			} catch (err) {
				console.error(
					`	× Failed: ${fileName} (${(err as any).message})`,
				);
				$block.text(`// Error loading remote code: ${fileName}`);
			}
		}),
	);
}

/**
 * Takes the contents of static/slides, static/styles and static/scripts
 * and injects them into static/index.html and returns the final html
 */
export async function buildPresentation() {
	const layoutPath = path.join(__dirname, "../static", "index.html");

	const templateHtml = fs.readFileSync(layoutPath, "utf8");
	const $ = cheerio.load(templateHtml);

	// 1. Inline the reveal assets
	inlineRevealAssets($);

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
	prepareCodeFragments($);
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
