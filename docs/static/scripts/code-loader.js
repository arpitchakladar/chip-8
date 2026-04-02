async function loadCode() {
	const blocks = document.querySelectorAll("[data-load-code]");
	console.log(blocks);

	await Promise.all(
		Array.from(blocks).map(async (block) => {
			const url = `/code?file=${block.getAttribute("data-load-code")}`;

			try {
				block.textContent = "// Loading...";

				const res = await fetch(url);
				const text = await res.text();

				block.textContent = text;
			} catch (err) {
				block.textContent = "// Failed to load code";
				console.error(err);
			}
		}),
	);

	// Re-run syntax highlighting after all code is inserted
	if (window.Reveal && Reveal.highlight) {
		Reveal.highlight();
	}
}
