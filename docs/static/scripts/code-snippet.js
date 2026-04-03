async function prepareCodeFragments() {
	// Find all containers that might have these definitions
	const containers = document.querySelectorAll("[data-code-snippet]");

	containers.forEach((container) => {
		// Select only direct children with the data attribute
		const templates = container.querySelectorAll("[data-code-fragment]");

		templates.forEach((el, sequence) => {
			// 1. Extract attributes
			const indexes =
				el.getAttribute("data-code-fragment-indexes") || "0";
			const title =
				el.getAttribute("data-code-fragment-title") || "No title";
			const text =
				el.getAttribute("data-code-fragment-text") || "No text";
			const color = sequence % 2 === 0 ? "var(--blue)" : "var(--red)";
			const pos = "";

			indexes.split(",").forEach((fragmentIndex) => {
				// 2. Create the fragment popup
				const popup = document.createElement("div");
				popup.className = "fragment fade-in-then-out";
				popup.setAttribute("data-fragment-index", fragmentIndex.trim());
				popup.setAttribute("data-code-fragment-element", "");

				// 3. Apply Styling
				popup.style.cssText = `
					--code-fragment-element-color: ${color}
				`;

				popup.innerHTML = `
					<h5 style="margin: 0 0 10px 0; color: ${color}; font-size: 0.5em;">${title}</h5>
					<p style="font-size: 0.5em; margin: 0; line-height: 1.4;">${text}</p>
				`;

				// 4. Move to container and cleanup
				container.appendChild(popup);
			});
			el.remove();
		});
	});
}
