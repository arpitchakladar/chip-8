async function runSnakeGame() {
	const canvas = document.querySelector("canvas#snake-game");
	const vm = new chip_8.Emulator(canvas, 100000);

	const response = await fetch(
		"https://raw.githubusercontent.com/arpitchakladar/chip-8/refs/heads/master/games/snake.asm",
	);
	const asmCode = await response.text();

	const assembler = new chip_8.Assembler(asmCode);
	const romData = assembler.assemble();

	vm.loadROM(romData);
	vm.run();
	vm.handleKeyboard();
}
