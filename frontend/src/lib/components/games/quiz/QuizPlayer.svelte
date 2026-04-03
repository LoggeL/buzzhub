<script lang="ts">
	let { phase, data, timerVal, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let answered = $state(false);

	$effect(() => {
		if (phase === 'question') answered = false;
	});

	function answer(idx: number) {
		if (answered) return;
		answered = true;
		sendAction('answer', { answer: idx });
	}
</script>

<div class="quiz fade-in">
	{#if phase === 'question' && data}
		<div class="question-header">
			<span class="q-num">Frage {data.questionNum}/{data.totalQuestions}</span>
		</div>
		<h2 class="question-text">{data.question}</h2>

		{#if answered}
			<div class="answered-msg">
				<p>Antwort abgegeben!</p>
				{#if data.answeredCount}
					<p class="count">{data.answeredCount}/{data.totalPlayers} haben geantwortet</p>
				{/if}
			</div>
		{:else}
			<div class="options">
				{#each data.options as option, i}
					<button
						class="option-btn option-{i}"
						onclick={() => answer(i)}
					>
						{option}
					</button>
				{/each}
			</div>
		{/if}

	{:else if phase === 'reveal' && data}
		<div class="reveal">
			<h3>Richtige Antwort:</h3>
			<div class="correct-answer">{data.correctText}</div>
			{#if data.roundScores}
				<div class="round-scores">
					{#each Object.entries(data.roundScores) as [pid, pts]}
						<span class="score-chip">+{pts}</span>
					{/each}
				</div>
			{/if}
		</div>

	{:else if phase === 'scoreboard' && data}
		<div class="scoreboard">
			<h2>{data.final ? 'Endergebnis' : 'Zwischenstand'}</h2>
			{#if data.scores}
				{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [pid, score], i}
					<div class="score-row" class:top={i === 0}>
						<span class="rank">#{i + 1}</span>
						<span class="name">{pid}</span>
						<span class="pts">{score}</span>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</div>

<style>
	.quiz {
		width: 100%;
	}

	.question-header {
		text-align: center;
		margin-bottom: 0.5rem;
	}

	.q-num {
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	.question-text {
		text-align: center;
		font-size: 1.3rem;
		margin-bottom: 1.5rem;
		line-height: 1.4;
	}

	.options {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.75rem;
	}

	.option-btn {
		padding: 1.25rem 0.75rem;
		border-radius: var(--radius);
		font-weight: 600;
		font-size: 1rem;
		color: white;
		min-height: 4rem;
		transition: transform 0.1s;
	}

	.option-btn:active {
		transform: scale(0.97);
	}

	.option-0 { background: #e74c3c; }
	.option-1 { background: #3498db; }
	.option-2 { background: #2ecc71; }
	.option-3 { background: #f39c12; }

	.answered-msg {
		text-align: center;
		padding: 2rem;
	}

	.answered-msg p {
		font-size: 1.3rem;
		font-weight: 600;
	}

	.count {
		color: var(--text-muted);
		font-size: 0.9rem !important;
		font-weight: 400 !important;
		margin-top: 0.5rem;
	}

	.reveal {
		text-align: center;
		padding: 1rem 0;
	}

	.correct-answer {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--success);
		margin: 1rem 0;
	}

	.round-scores {
		display: flex;
		gap: 0.5rem;
		justify-content: center;
		flex-wrap: wrap;
	}

	.score-chip {
		background: rgba(46, 204, 113, 0.2);
		color: var(--success);
		padding: 0.25rem 0.75rem;
		border-radius: 1rem;
		font-weight: 600;
	}

	.scoreboard {
		width: 100%;
	}

	.scoreboard h2 {
		text-align: center;
		margin-bottom: 1rem;
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
