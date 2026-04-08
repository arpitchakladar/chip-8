// Global variables to store the ROM data, VM instance, button element, and canvas
let snakeGameRom = null;
let snakeGameVM = null;
let snakeGameButton = null;
let snakeGameCanvas = null;

// Initializes and runs the Snake game ROM
async function runSnakeGame() {
	// Create a new Emulator instance with the canvas and 10kHz clock speed
	snakeGameVM = await chip_8.Emulator(snakeGameCanvas, 10000);

	// Fetch the Snake game assembly source from the repository if not already cached
	if (!snakeGameRom) {
		const response = await fetch(
			"https://raw.githubusercontent.com/arpitchakladar/chip-8/refs/heads/master/games/snake.asm",
		);
		const asmCode = await response.text();

		// Create an assembler and compile the assembly code to ROM bytecode
		const assembler = await chip_8.Assembler(asmCode);
		snakeGameRom = await assembler.assemble();
	}

	// Load the compiled ROM into the emulator
	await snakeGameVM.loadROM(snakeGameRom);
	// Start the emulator execution
	snakeGameVM.run();
	// Attach keyboard event handlers for input
	await snakeGameVM.handleKeyboard();
}

// Destroys the current VM instance and restarts the game
async function restartSnakeGame() {
	if (snakeGameVM) {
		// Destroy the current emulator instance to stop execution and release resources
		snakeGameVM.destroy();
		snakeGameVM = null;
		// Reset canvas dimensions (clears the display)
		snakeGameCanvas.setAttribute("width", "640");
		snakeGameCanvas.setAttribute("height", "320");
	} else {
		// First run: update button text to indicate restart capability
		snakeGameButton.innerHTML = "RESTART";
	}
	// Start the game fresh
	await runSnakeGame();
	// Focus the canvas to capture keyboard input immediately
	snakeGameCanvas.focus();
}

// Sets up event listeners for the Snake game UI
async function setupSnakeGame() {
	// Get references to the button and canvas elements from the DOM
	snakeGameButton = document.querySelector("button#snake-game-button");
	snakeGameCanvas = document.querySelector("canvas#snake-game");
	// Attach click handler to restart the game when button is pressed
	snakeGameButton.addEventListener("click", restartSnakeGame);
}
