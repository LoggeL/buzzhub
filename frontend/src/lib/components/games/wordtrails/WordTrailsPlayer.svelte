<script lang="ts">
	import { lobby } from '$lib/stores/lobby';

	let { phase, data, timerVal, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let selected = $state<number[]>([]);
	let foundWords = $state<string[]>([]);
	let globalFound = $state<Set<string>>(new Set());
	let lastResult = $state<{ word: string; points: number; order: number } | null>(null);
	let wrongFlash = $state(false);
	let totalFound = $state(0);

	$effect(() => {
		if (phase === 'playing') {
			selected = [];
			foundWords = [];
			globalFound = new Set();
			lastResult = null;
			wrongFlash = false;
			totalFound = 0;
		}
	});

	$effect(() => {
		if (data?.correct) {
			if (!foundWords.includes(data.correct)) {
				foundWords = [...foundWords, data.correct];
			}
			globalFound = new Set([...globalFound, data.correct]);
			lastResult = { word: data.correct, points: data.points, order: data.order };
			const word = data.correct;
			setTimeout(() => { if (lastResult?.word === word) lastResult = null; }, 1800);
		}
		if (data?.wrong) {
			wrongFlash = true;
			setTimeout(() => wrongFlash = false, 350);
		}
		if (data?.found?.word) {
			globalFound = new Set([...globalFound, data.found.word]);
			totalFound = data.totalFound || totalFound;
		}
		if (data?.totalFound !== undefined) {
			totalFound = data.totalFound;
		}
	});

	function toggleLetter(index: number) {
		if (selected.includes(index)) return;
		selected = [...selected, index];
	}

	function currentWord(): string {
		if (!data?.letters) return '';
		return selected.map(index => data.letters[index]).join('');
	}

	function backspace() {
		selected = selected.slice(0, -1);
	}

	function clear() {
		selected = [];
	}

	function submit() {
		const word = currentWord();
		if (word.length < 3) return;
		sendAction('guess', { word });
		selected = [];
	}

	function orderLabel(order: number): string {
		if (order === 1) return 'ERSTER';
		if (order === 2) return 'ZWEITER';
		if (order === 3) return 'DRITTER';
		return `${order}.`;
	}

	function playerName(pid: string): string {
		return $lobby?.players.find(player => player.id === pid)?.name ?? pid;
	}
</script>

<div class="wordtrails fade-in">
	{#if phase === 'playing' && data}
		<div class="mode-name">{data.modeName}</div>
		<div class="stats-bar">
			<span>{data.theme}</span>
			<span>{totalFound}/{data.wordCount} geloest</span>
		</div>

		<div class="slots" aria-label="Wort-Slots">
			{#each data.slots as slot}
				<div class="slot" class:found={slot.found || globalFound.has(slot.word)}>
					<span>{slot.word}</span>
					<small>{slot.length}</small>
				</div>
			{/each}
		</div>

		<div class="current-word" class:too-short={currentWord().length < 3} class:wrong-shake={wrongFlash}>
			{currentWord() || 'WORT BAUEN'}
		</div>

		<div class="letter-wheel" role="group" aria-label="Buchstaben">
			{#each data.letters as letter, i}
				<button
					class="letter-tile"
					class:selected={selected.includes(i)}
					onclick={() => toggleLetter(i)}
					disabled={selected.includes(i)}
					aria-label={`Buchstabe ${letter}`}
				>
					{letter}
				</button>
			{/each}
		</div>

		<div class="controls">
			<button class="ctrl-btn" onclick={backspace} disabled={!selected.length}>Zurueck</button>
			<button class="ctrl-btn" onclick={clear} disabled={!selected.length}>Leeren</button>
			<button class="btn btn-primary submit-btn" onclick={submit} disabled={currentWord().length < 3}>Senden</button>
		</div>

		{#if lastResult}
			<div class="result-flash fade-in" class:first={lastResult.order === 1}>
				<span>{lastResult.word}</span>
				<strong>{orderLabel(lastResult.order)}</strong>
				<span>+{lastResult.points}</span>
			</div>
		{/if}

		{#if foundWords.length > 0}
			<div class="found-list">
				{#each foundWords as word}
					<span class="found-chip">{word}</span>
				{/each}
			</div>
		{/if}

	{:else if phase === 'results' && data}
		<div class="results">
			<h2>Buchstaben-Tease</h2>
			<p class="theme">{data.theme}</p>

			<div class="slots reveal">
				{#each data.slots as slot}
					<div class="slot found">
						<span>{slot.word}</span>
						<small>{slot.length}</small>
					</div>
				{/each}
			</div>

			{#if data.scores}
				<div class="scoreboard">
					{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [pid, score], i}
						<div class="score-row" class:top={i === 0}>
							<span class="rank">#{i + 1}</span>
							<span class="name">{playerName(pid)}</span>
							<span class="pts">{score}</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.wordtrails { width: 100%; }

	.mode-name {
		text-align: center;
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		margin-bottom: 0.35rem;
	}

	.stats-bar {
		display: flex;
		justify-content: space-between;
		color: var(--text-muted);
		font-size: 0.85rem;
		margin-bottom: 0.75rem;
	}

	.slots {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.4rem;
		margin-bottom: 1rem;
	}

	.slot {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		padding: 0.5rem 0.65rem;
		border-radius: var(--radius-sm);
		background: var(--bg-card);
		border: 1px solid #333;
		font-family: monospace;
		font-weight: 800;
		letter-spacing: 0.08em;
	}

	.slot small {
		color: var(--text-muted);
		font-family: inherit;
		letter-spacing: 0;
	}

	.slot.found {
		background: rgba(46, 204, 113, 0.15);
		border-color: #2ecc71;
		color: #2ecc71;
	}

	.current-word {
		min-height: 3rem;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 1rem;
		border-radius: var(--radius);
		background: rgba(255, 144, 0, 0.12);
		border: 1px solid var(--primary);
		color: var(--primary);
		font-family: monospace;
		font-size: 1.3rem;
		font-weight: 900;
		letter-spacing: 0.12em;
	}

	.current-word.too-short {
		color: var(--text-muted);
		border-color: #333;
		background: var(--bg-card);
	}

	.current-word.wrong-shake {
		animation: shake 0.3s;
	}

	@keyframes shake {
		0%, 100% { transform: translateX(0); }
		25% { transform: translateX(-5px); }
		75% { transform: translateX(5px); }
	}

	.letter-wheel {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 0.6rem;
		margin-bottom: 1rem;
	}

	.letter-tile {
		aspect-ratio: 1;
		border-radius: 50%;
		background: var(--bg-card);
		border: 2px solid #333;
		color: var(--text);
		font-size: 1.35rem;
		font-weight: 900;
		font-family: monospace;
	}

	.letter-tile.selected {
		background: var(--primary);
		border-color: var(--primary);
		color: #000;
	}

	.controls {
		display: grid;
		grid-template-columns: 1fr 1fr 1.4fr;
		gap: 0.5rem;
	}

	.ctrl-btn {
		padding: 0.75rem 0.5rem;
		background: var(--bg-card);
		color: var(--text);
		border-radius: var(--radius-sm);
		border: 1px solid #333;
		font-weight: 700;
	}

	.submit-btn {
		width: auto;
	}

	.result-flash {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 0.65rem;
		margin: 0.75rem 0;
		border-radius: var(--radius-sm);
		background: rgba(255, 144, 0, 0.1);
		border: 1px solid var(--primary);
	}

	.result-flash.first {
		background: rgba(255, 144, 0, 0.2);
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

	.results h2, .theme {
		text-align: center;
	}

	.theme {
		color: var(--text-muted);
		margin-bottom: 1rem;
	}

	.scoreboard {
		margin-top: 1rem;
	}

	.score-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		background: var(--bg-card);
		border-radius: var(--radius-sm);
		margin-bottom: 0.5rem;
	}

	.score-row.top {
		background: rgba(233, 69, 96, 0.15);
		border: 1px solid var(--primary);
	}

	.rank { font-weight: 700; color: var(--text-muted); min-width: 2rem; }
	.name { flex: 1; font-weight: 500; }
	.pts { font-weight: 700; color: var(--primary); }
</style>
