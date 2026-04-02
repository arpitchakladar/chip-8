// List of GIFs you want to cycle through
const backgrounds = [
	'/assets/RPS.gif'
];

let currentIndex = 0;

// Function to change the background of the first slide
function cycleBackground() {
	currentIndex = (currentIndex + 1) % backgrounds.length;

	// 1. Find the section element
	const titleSlide = document.querySelector('section[data-id-bg="cycling-bg"]');

	// 2. Update the attribute
	titleSlide.setAttribute('data-background-image', backgrounds[currentIndex]);

	// 3. Tell Reveal.js to refresh the background layer
	Reveal.getRevealElement().querySelector('.backgrounds')
		.querySelectorAll('.slide-background')[0] // Targets first slide
		.querySelector('.slide-background-content')
		.style.backgroundImage = `url("${backgrounds[currentIndex]}")`;
}

// Change every 5 seconds (5000ms)
setInterval(cycleBackground, 5000);
