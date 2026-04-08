let snakeGameRom = null;
let snakeGameVM = null;
let snakeGameButton = null;
let snakeGameCanvas = null;

async function runSnakeGame() {
	snakeGameVM = await chip_8.Emulator(snakeGameCanvas, 10000);

	if (!snakeGameRom) {
		const response = await fetch(
			"https://raw.githubusercontent.com/arpitchakladar/chip-8/refs/heads/master/games/snake.asm",
		);
		const asmCode = await response.text();

		const assembler = await chip_8.Assembler(asmCode);
		snakeGameRom = await assembler.assemble();
	}

	await snakeGameVM.loadROM(snakeGameRom);
	snakeGameVM.run();
	await snakeGameVM.handleKeyboard();
}

async function restartSnakeGame() {
	if (snakeGameVM) {
		snakeGameVM.destroy();
		snakeGameVM = null;
		snakeGameCanvas.setAttribute("width", "640");
		snakeGameCanvas.setAttribute("height", "320");
	} else {
		snakeGameButton.innerHTML = "RESTART";
	}
	await runSnakeGame();
	snakeGameCanvas.focus();
}

async function setupSnakeGame() {
	snakeGameButton = document.querySelector("button#snake-game-button");
	snakeGameCanvas = document.querySelector("canvas#snake-game");
	snakeGameButton.addEventListener("click", restartSnakeGame);
}
