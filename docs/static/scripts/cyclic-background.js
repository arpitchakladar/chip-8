// Function to change the background of any slide
function cycleBackground() {
	// 1. Find the section elements
	const titleSlides = document.querySelectorAll("section[data-cycling-bg]");

	titleSlides.forEach((titleSlide) => {
		// 2. Retrieve the data for that particular section
		const backgroundImages = titleSlide
			.getAttribute("data-cycling-bg")
			.split(",")
			.map((assetName) => `/assets/${assetName}`);
		let currentIndex = parseInt(
			titleSlide.getAttribute("data-cycling-bg-index") || 0,
		);
		currentIndex = (currentIndex + 1) % backgroundImages.length;
		titleSlide.setAttribute("data-cycling-bg-index", currentIndex);

		// 2. Update the background image attribute
		titleSlide.setAttribute(
			"data-background-image",
			backgroundImages[currentIndex],
		);

		// 3. Tell Reveal.js to refresh the background layer
		if (window.Reveal && Reveal.getRevealElement) {
			Reveal.getRevealElement()
				.querySelector(".backgrounds")
				.querySelectorAll(".slide-background")[0] // Targets first slide
				.querySelector(".slide-background-content").style.backgroundImage =
				`url("${backgroundImages[currentIndex]}")`;
		}
	});
}

// Change every 3 seconds (3000ms)
document.addEventListener("DOMContentLoaded", () => {
	setInterval(cycleBackground, 2000);
});
