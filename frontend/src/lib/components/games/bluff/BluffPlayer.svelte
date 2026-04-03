<script lang="ts">
	let { phase, data, timerVal, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let submitted = $state(false);
	let guessed = $state(false);
	let fakeAnswer = $state('');

	$effect(() => {
		if (phase === 'write') { submitted = false; fakeAnswer = ''; }
		if (phase === 'guess') { guessed = false; }
	});

	function submitFake() {
		if (!fakeAnswer.trim() || submitted) return;
		submitted = true;
		sendAction('submit', { answer: fakeAnswer.trim() });
	}

	function guess(answerId: string) {
		if (guessed) return;
		guessed = true;
		sendAction('guess', { answerId });
	}
</script>

<div class="bluff fade-in">
	{#if phase === 'write' && data}
		<div class="round-info">Runde {data.roundNum}/{data.totalRounds}</div>
		<h2 class="question">{data.question}</h2>
		<p class="instruction">Erfinde eine glaubwuerdige falsche Antwort!</p>

		{#if submitted}
			<div class="submitted-msg">
				<p>Abgesendet!</p>
				{#if data.submittedCount}
					<p class="count">{data.submittedCount}/{data.totalPlayers}</p>
				{/if}
			</div>
		{:else}
			<form class="answer-form" onsubmit={e => { e.preventDefault(); submitFake(); }}>
				<input
					class="input"
					type="text"
					placeholder="Deine falsche Antwort..."
					bind:value={fakeAnswer}
					maxlength="100"
					autocomplete="off"
				/>
				<button class="btn btn-primary" type="submit" disabled={!fakeAnswer.trim()}>
					Absenden
				</button>
			</form>
		{/if}

	{:else if phase === 'guess' && data}
		<h2 class="question">{data.question}</h2>
		<p class="instruction">Welche Antwort ist die richtige?</p>

		{#if guessed}
			<div class="submitted-msg">
				<p>Geraten!</p>
				{#if data.guessedCount}
					<p class="count">{data.guessedCount}/{data.totalPlayers}</p>
				{/if}
			</div>
		{:else if data.options}
			<div class="guess-options">
				{#each data.options as opt}
					<button class="guess-btn" onclick={() => guess(opt.id)}>
						{opt.answer}
					</button>
				{/each}
			</div>
		{/if}

	{:else if phase === 'reveal' && data}
		<h3 class="question">{data.question}</h3>
		<div class="reveal-answer">
			<div class="real-label">Richtige Antwort:</div>
			<div class="real-answer">{data.realAnswer}</div>
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
	.bluff { width: 100%; }

	.round-info {
		text-align: center;
		color: var(--text-muted);
		font-size: 0.85rem;
		margin-bottom: 0.5rem;
	}

	.question {
		text-align: center;
		font-size: 1.2rem;
		margin-bottom: 0.5rem;
		line-height: 1.4;
	}

	.instruction {
		text-align: center;
		color: var(--text-muted);
		margin-bottom: 1.5rem;
	}

	.answer-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.submitted-msg {
		text-align: center;
		padding: 2rem;
	}

	.submitted-msg p {
		font-size: 1.3rem;
		font-weight: 600;
	}

	.count {
		color: var(--text-muted);
		font-size: 0.9rem !important;
		font-weight: 400 !important;
		margin-top: 0.5rem;
	}

	.guess-options {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.guess-btn {
		padding: 1rem;
		background: var(--bg-card);
		color: var(--text);
		border-radius: var(--radius);
		font-size: 1.05rem;
		text-align: left;
		border: 2px solid transparent;
		transition: all 0.2s;
	}

	.guess-btn:hover {
		border-color: var(--primary);
	}

	.reveal-answer {
		text-align: center;
		margin-top: 1rem;
	}

	.real-label {
		color: var(--text-muted);
		margin-bottom: 0.5rem;
	}

	.real-answer {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--success);
	}

	.scoreboard { width: 100%; }
	.scoreboard h2 { text-align: center; margin-bottom: 1rem; }

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
