async function prepareCodeFragments() {
	// Find all containers that might have these definitions
	const containers = document.querySelectorAll("[data-code-snippet]");

	containers.forEach((container) => {
		// Select only direct children with the data attribute
		const templates = container.querySelectorAll("[data-code-fragment]");

		templates.forEach((el) => {
			// 1. Extract attributes
			const indexes =
				el.getAttribute("data-code-fragment-indexes") || "0";
			const title =
				el.getAttribute("data-code-fragment-title") || "No title";
			const text =
				el.getAttribute("data-code-fragment-text") || "No text";
			const color =
				el.getAttribute("data-code-fragment-color") || "white";
			const pos =
				el.getAttribute("data-code-fragment-position") ||
				"top: 10%; right: 5%;";

			indexes.split(",").forEach((index) => {
				// 2. Create the fragment popup
				const popup = document.createElement("div");
				popup.className = "fragment fade-in-then-out";
				popup.setAttribute("data-fragment-index", index.trim());

				// 3. Apply Styling
				popup.style.cssText = `
				position: absolute;
				${pos};
				width: 300px;
				background: rgba(34, 34, 34, 0.9);
				border: 2px solid ${color};
				padding: 15px;
				box-shadow: 10px 10px 20px rgba(0,0,0,0.5);
				z-index: 10;
				text-align: left;
				border-radius: 8px;
			`;

				popup.innerHTML = `
				<h4 style="margin: 0 0 10px 0; color: ${color}; font-size: 0.8em;">${title}</h4>
				<p style="font-size: 0.5em; margin: 0; line-height: 1.4;">${text}</p>
			`;

				// 4. Move to container and cleanup
				container.appendChild(popup);
			});
			el.remove();
		});
	});
}
