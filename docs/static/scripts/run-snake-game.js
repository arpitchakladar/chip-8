async function runSnakeGame() {
	const canvas = document.querySelector("canvas#snake-game");
	const vm = await chip_8.Emulator(canvas, 10000);

	const response = await fetch(
		"https://raw.githubusercontent.com/arpitchakladar/chip-8/refs/heads/master/games/snake.asm",
	);
	const asmCode = await response.text();

	const assembler = await chip_8.Assembler(asmCode);
	const romData = await assembler.assemble();

	await vm.loadROM(romData);
	vm.run();
	await vm.handleKeyboard();
}
