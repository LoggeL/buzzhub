<script lang="ts">
	let { phase, data, timerVal, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let selecting = $state(false);
	let selectedPath = $state<number[][]>([]);
	let foundWords = $state<string[]>([]);
	let foundCells = $state<Set<string>>(new Set());
	let globalFoundWords = $state<Set<string>>(new Set());
	let lastResult = $state<{ word: string; points: number; order: number } | null>(null);
	let wrongFlash = $state(false);
	let totalFound = $state(0);

	$effect(() => {
		if (phase === 'playing') {
			foundWords = [];
			foundCells = new Set();
			globalFoundWords = new Set();
			totalFound = 0;
			lastResult = null;
		}
	});

	$effect(() => {
		if (data?.correct && data?.path) {
			if (!foundWords.includes(data.correct)) {
				foundWords = [...foundWords, data.correct];
			}
			// Add path cells to found set
			const newCells = new Set(foundCells);
			for (const [r, c] of data.path) {
				newCells.add(`${r}-${c}`);
			}
			foundCells = newCells;
			globalFoundWords = new Set([...globalFoundWords, data.correct]);
			lastResult = { word: data.correct, points: data.points, order: data.order };
			const w = data.correct;
			setTimeout(() => { if (lastResult?.word === w) lastResult = null; }, 2000);
		}
		if (data?.wrong) {
			wrongFlash = true;
			setTimeout(() => wrongFlash = false, 400);
		}
		if (data?.totalFound !== undefined) {
			totalFound = data.totalFound;
		}
		// Handle broadcast found events (from other players)
		if (data?.found && typeof data.found === 'object' && data.found.word) {
			globalFoundWords = new Set([...globalFoundWords, data.found.word]);
			if (data.found.path) {
				const newCells = new Set(foundCells);
				for (const [r, c] of data.found.path) {
					newCells.add(`${r}-${c}`);
				}
				foundCells = newCells;
			}
			totalFound = data.totalFound || totalFound;
		}
	});

	function cellKey(r: number, c: number): string {
		return `${r}-${c}`;
	}

	function isSelected(r: number, c: number): boolean {
		return selectedPath.some(p => p[0] === r && p[1] === c);
	}

	function isFound(r: number, c: number): boolean {
		return foundCells.has(cellKey(r, c));
	}

	function isAdjacent(a: number[], b: number[]): boolean {
		const dr = Math.abs(a[0] - b[0]);
		const dc = Math.abs(a[1] - b[1]);
		return dr <= 1 && dc <= 1 && !(dr === 0 && dc === 0);
	}

	function startSelect(r: number, c: number, e: PointerEvent) {
		e.preventDefault();
		(e.target as HTMLElement)?.setPointerCapture?.(e.pointerId);
		selecting = true;
		selectedPath = [[r, c]];
	}

	function moveSelect(e: PointerEvent) {
		if (!selecting) return;
		e.preventDefault();
		const el = document.elementFromPoint(e.clientX, e.clientY);
		if (!el || !('dataset' in el)) return;
		const ds = (el as HTMLElement).dataset;
		if (ds.row === undefined || ds.col === undefined) return;
		const r = parseInt(ds.row);
		const c = parseInt(ds.col);
		addToPath(r, c);
	}

	function addToPath(r: number, c: number) {
		const last = selectedPath[selectedPath.length - 1];
		if (last[0] === r && last[1] === c) return;

		// Allow backtracking to previous cell
		if (selectedPath.length >= 2) {
			const prev = selectedPath[selectedPath.length - 2];
			if (prev[0] === r && prev[1] === c) {
				selectedPath = selectedPath.slice(0, -1);
				return;
			}
		}

		// Don't revisit cells
		if (selectedPath.some(p => p[0] === r && p[1] === c)) return;

		// Must be adjacent to last cell
		if (!isAdjacent(last, [r, c])) return;

		selectedPath = [...selectedPath, [r, c]];
	}

	function endSelect(e: PointerEvent) {
		if (!selecting) return;
		e.preventDefault();
		selecting = false;
		if (selectedPath.length >= 3) {
			sendAction('guess', { path: selectedPath });
		}
		setTimeout(() => {
			if (!selecting) selectedPath = [];
		}, 200);
	}

	function getOrderLabel(order: number): string {
		if (order === 1) return 'ERSTER!';
		if (order === 2) return 'Zweiter';
		if (order === 3) return 'Dritter';
		return `${order}.`;
	}

	function getSelectedWord(): string {
		if (!data?.grid || selectedPath.length === 0) return '';
		return selectedPath.map(([r, c]) => {
			const row = data.grid[r];
			return row ? row[c] || '' : '';
		}).join('');
	}
</script>

<div class="crossword fade-in">
	{#if phase === 'playing' && data}
		<div class="stats-bar">
			<span class="stat">{totalFound}/{data.wordCount} gefunden</span>
			<span class="stat found-count">Du: {foundWords.length}</span>
		</div>

		<!-- Word list -->
		{#if data.words}
			<div class="word-list">
				{#each data.words as word}
					<span class="word-tag" class:word-found={globalFoundWords.has(word)}>
						{word}
					</span>
				{/each}
			</div>
		{/if}

		<!-- Selected word preview -->
		{#if selectedPath.length > 0}
			<div class="selection-preview" class:too-short={selectedPath.length < 3}>
				{getSelectedWord()}
			</div>
		{/if}

		<!-- Grid -->
		<div
			class="grid-container"
			class:wrong-shake={wrongFlash}
			onpointermove={moveSelect}
			onpointerup={endSelect}
			onpointercancel={endSelect}
		>
			<div
				class="letter-grid"
				style="grid-template-columns: repeat({data.gridSize}, 1fr); touch-action: none;"
			>
				{#each data.grid as row, r}
					{#each row.split('') as letter, c}
						<div
							class="cell"
							class:cell-selected={isSelected(r, c)}
							class:cell-found={isFound(r, c)}
							data-row={r}
							data-col={c}
							onpointerdown={(e) => startSelect(r, c, e)}
						>
							{letter}
						</div>
					{/each}
				{/each}
			</div>
		</div>

		<!-- Result flash -->
		{#if lastResult}
			<div class="result-flash fade-in" class:first={lastResult.order === 1}>
				<span class="result-word">{lastResult.word}</span>
				<span class="result-order">{getOrderLabel(lastResult.order)}</span>
				<span class="result-points">+{lastResult.points}</span>
			</div>
		{/if}

		<!-- Found words -->
		{#if foundWords.length > 0}
			<div class="found-list">
				{#each foundWords as word}
					<span class="found-chip">{word}</span>
				{/each}
			</div>
		{/if}

	{:else if phase === 'results' && data}
		<div class="results">
			<h2>Ergebnis</h2>

			{#if data.words}
				<div class="all-words">
					<h3>Alle Woerter ({data.words.length})</h3>
					<div class="word-chips">
						{#each data.words as word}
							{@const wasFound = data.found?.some((f: any) => f.word === word)}
							<span class="word-chip" class:found={wasFound} class:missed={!wasFound}>
								{word}
							</span>
						{/each}
					</div>
				</div>
			{/if}

			{#if data.scores}
				<div class="scoreboard">
					{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [pid, score], i}
						<div class="score-row" class:top={i === 0}>
							<span class="rank">#{i + 1}</span>
							<span class="name">{pid}</span>
							<span class="pts">{score}</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.crossword { width: 100%; }

	.stats-bar {
		display: flex;
		justify-content: space-between;
		margin-bottom: 0.5rem;
		font-size: 0.85rem;
	}

	.stat {
		color: var(--text-muted);
	}

	.found-count {
		color: var(--primary);
		font-weight: 600;
	}

	.word-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.3rem;
		margin-bottom: 0.5rem;
	}

	.word-tag {
		padding: 0.15rem 0.4rem;
		border-radius: 2px;
		font-size: 0.65rem;
		font-weight: 700;
		font-family: monospace;
		background: var(--bg-card);
		color: var(--text-muted);
		border: 1px solid #333;
		transition: all 0.3s;
	}

	.word-tag.word-found {
		background: rgba(46, 204, 113, 0.15);
		color: #2ecc71;
		border-color: #2ecc71;
		text-decoration: line-through;
	}

	.selection-preview {
		text-align: center;
		font-family: monospace;
		font-weight: 800;
		font-size: 1.1rem;
		letter-spacing: 0.1em;
		padding: 0.3rem;
		color: var(--primary);
		margin-bottom: 0.25rem;
	}

	.selection-preview.too-short {
		color: var(--text-muted);
	}

	.grid-container {
		width: 100%;
		overflow: hidden;
		border: 2px solid #333;
		border-radius: 2px;
		transition: transform 0.1s;
	}

	.grid-container.wrong-shake {
		animation: shake 0.3s;
	}

	@keyframes shake {
		0%, 100% { transform: translateX(0); }
		25% { transform: translateX(-5px); }
		75% { transform: translateX(5px); }
	}

	.letter-grid {
		display: grid;
		gap: 0;
		width: 100%;
		user-select: none;
		-webkit-user-select: none;
	}

	.cell {
		aspect-ratio: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: 0.9rem;
		font-family: monospace;
		background: var(--bg-card);
		border: 1px solid #222;
		color: var(--text);
		user-select: none;
		-webkit-user-select: none;
		touch-action: none;
		cursor: pointer;
		transition: background 0.15s, border-color 0.15s;
	}

	.cell.cell-selected {
		background: rgba(255, 144, 0, 0.35);
		border-color: var(--primary);
		color: #fff;
		transform: scale(1.05);
		z-index: 1;
	}

	.cell.cell-found {
		background: rgba(46, 204, 113, 0.2);
		border-color: rgba(46, 204, 113, 0.4);
	}

	.cell.cell-found.cell-selected {
		background: rgba(255, 144, 0, 0.35);
		border-color: var(--primary);
	}

	.result-flash {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 0.5rem;
		margin: 0.5rem 0;
		border-radius: 2px;
		background: rgba(255, 144, 0, 0.1);
		border: 1px solid var(--primary);
	}

	.result-flash.first {
		background: rgba(255, 144, 0, 0.2);
	}

	.result-word {
		font-weight: 700;
		font-size: 1.1rem;
	}

	.result-order {
		font-size: 0.8rem;
		color: var(--primary);
		font-weight: 700;
	}

	.result-points {
		font-weight: 700;
		color: var(--primary);
	}

	.found-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
		margin-top: 0.75rem;
	}

	.found-chip {
		background: rgba(255, 144, 0, 0.15);
		color: var(--primary);
		padding: 0.2rem 0.5rem;
		border-radius: 2px;
		font-size: 0.75rem;
		font-weight: 700;
		font-family: monospace;
	}

	.results { width: 100%; }

	.results h2 {
		text-align: center;
		margin-bottom: 1rem;
	}

	.all-words {
		margin-bottom: 1.5rem;
	}

	.all-words h3 {
		font-size: 0.9rem;
		color: var(--text-muted);
		margin-bottom: 0.5rem;
	}

	.word-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
	}

	.word-chip {
		padding: 0.2rem 0.5rem;
		border-radius: 2px;
		font-size: 0.75rem;
		font-weight: 700;
		font-family: monospace;
	}

	.word-chip.found {
		background: rgba(255, 144, 0, 0.15);
		color: var(--primary);
	}

	.word-chip.missed {
		background: rgba(255, 255, 255, 0.05);
		color: var(--text-muted);
		text-decoration: line-through;
	}

	.scoreboard { width: 100%; }

	.score-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		background: var(--bg-card);
		border-radius: 2px;
		margin-bottom: 0.5rem;
	}

	.score-row.top {
		background: rgba(255, 144, 0, 0.1);
		border: 1px solid var(--primary);
	}

	.rank { font-weight: 700; color: var(--text-muted); min-width: 2rem; }
	.name { flex: 1; font-weight: 500; }
	.pts { font-weight: 700; color: var(--primary); }
</style>
