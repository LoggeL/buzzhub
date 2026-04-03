<script lang="ts">
	import { onMount } from 'svelte';

	let { phase, data, timerVal, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let guess = $state('');
	let foundWords = $state<string[]>([]);
	let lastResult = $state<{ word: string; points: number; order: number } | null>(null);
	let wrongFlash = $state(false);
	let totalFound = $state(0);

	$effect(() => {
		if (phase === 'playing') {
			foundWords = [];
			totalFound = 0;
			lastResult = null;
		}
	});

	$effect(() => {
		if (data?.correct) {
			if (!foundWords.includes(data.correct)) {
				foundWords = [...foundWords, data.correct];
			}
			lastResult = { word: data.correct, points: data.points, order: data.order };
			setTimeout(() => { if (lastResult?.word === data.correct) lastResult = null; }, 2000);
		}
		if (data?.wrong) {
			wrongFlash = true;
			setTimeout(() => wrongFlash = false, 500);
		}
		if (data?.totalFound !== undefined) {
			totalFound = data.totalFound;
		}
		if (data?.found && typeof data.found === 'object' && data.found.word) {
			totalFound = data.totalFound || totalFound;
		}
	});

	function submitGuess() {
		const word = guess.trim().toUpperCase();
		if (!word) return;
		sendAction('guess', { word });
		guess = '';
	}

	function getOrderLabel(order: number): string {
		if (order === 1) return 'ERSTER!';
		if (order === 2) return 'Zweiter';
		if (order === 3) return 'Dritter';
		return `${order}.`;
	}
</script>

<div class="crossword fade-in">
	{#if phase === 'playing' && data}
		<div class="stats-bar">
			<span class="stat">{totalFound}/{data.wordCount} gefunden</span>
			<span class="stat found-count">Du: {foundWords.length}</span>
		</div>

		<!-- Grid -->
		<div class="grid-container">
			<div
				class="letter-grid"
				style="grid-template-columns: repeat({data.gridSize}, 1fr);"
			>
				{#each data.grid as row}
					{#each row.split('') as letter}
						<div class="cell">{letter}</div>
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

		<!-- Input -->
		<form class="guess-form" class:wrong={wrongFlash} onsubmit={e => { e.preventDefault(); submitGuess(); }}>
			<input
				class="input"
				type="text"
				placeholder="Wort eingeben..."
				bind:value={guess}
				maxlength="20"
				autocomplete="off"
				autocapitalize="characters"
				style="text-transform: uppercase;"
			/>
			<button class="btn btn-primary submit-btn" type="submit" disabled={!guess.trim()}>
				OK
			</button>
		</form>

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

	.grid-container {
		width: 100%;
		overflow: hidden;
		border: 2px solid #333;
		border-radius: 2px;
	}

	.letter-grid {
		display: grid;
		gap: 0;
		width: 100%;
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

	.guess-form {
		display: flex;
		gap: 0.5rem;
		margin-top: 0.75rem;
		transition: transform 0.1s;
	}

	.guess-form.wrong {
		animation: shake 0.3s;
	}

	@keyframes shake {
		0%, 100% { transform: translateX(0); }
		25% { transform: translateX(-5px); }
		75% { transform: translateX(5px); }
	}

	.guess-form .input {
		flex: 1;
		font-family: monospace;
		letter-spacing: 0.05em;
	}

	.submit-btn {
		width: auto !important;
		padding: 0.875rem 1.25rem !important;
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
