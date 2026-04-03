<script lang="ts">
	let { phase, data, timerVal, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let submitted = $state(false);
	let voted = $state(false);
	let answer = $state('');

	$effect(() => {
		if (phase === 'prompt') { submitted = false; answer = ''; }
		if (phase === 'vote') { voted = false; }
	});

	function submit() {
		if (!answer.trim() || submitted) return;
		submitted = true;
		sendAction('submit', { answer: answer.trim() });
	}

	function vote(playerId: string) {
		if (voted) return;
		voted = true;
		sendAction('vote', { playerId });
	}
</script>

<div class="voting fade-in">
	{#if phase === 'prompt' && data}
		<div class="round-info">Runde {data.roundNum}/{data.totalRounds}</div>
		<h2 class="prompt">{data.prompt}</h2>

		{#if submitted}
			<div class="submitted-msg">
				<p>Abgesendet!</p>
				{#if data.submittedCount}
					<p class="count">{data.submittedCount}/{data.totalPlayers}</p>
				{/if}
			</div>
		{:else}
			<form class="answer-form" onsubmit={e => { e.preventDefault(); submit(); }}>
				<input
					class="input"
					type="text"
					placeholder="Deine Antwort..."
					bind:value={answer}
					maxlength="100"
					autocomplete="off"
				/>
				<button class="btn btn-primary" type="submit" disabled={!answer.trim()}>
					Absenden
				</button>
			</form>
		{/if}

	{:else if phase === 'vote' && data}
		<h2 class="prompt">{data.prompt}</h2>

		{#if voted}
			<div class="submitted-msg">
				<p>Abgestimmt!</p>
				{#if data.votedCount}
					<p class="count">{data.votedCount}/{data.totalPlayers}</p>
				{/if}
			</div>
		{:else if data.submissions}
			<div class="vote-options">
				{#each data.submissions as sub}
					<button class="vote-btn" onclick={() => vote(sub.playerId)}>
						{sub.answer}
					</button>
				{/each}
			</div>
		{/if}

	{:else if phase === 'results' && data}
		<h3 class="prompt">{data.prompt}</h3>
		<div class="results">
			{#if data.submissions && data.voteCounts}
				{#each Object.entries(data.submissions) as [pid, answer]}
					<div class="result-row">
						<div class="result-answer">{answer}</div>
						<div class="result-votes">{data.voteCounts[pid] || 0} Stimmen</div>
					</div>
				{/each}
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
	.voting { width: 100%; }

	.round-info {
		text-align: center;
		color: var(--text-muted);
		font-size: 0.85rem;
		margin-bottom: 0.5rem;
	}

	.prompt {
		text-align: center;
		font-size: 1.2rem;
		margin-bottom: 1.5rem;
		line-height: 1.4;
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

	.vote-options {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.vote-btn {
		padding: 1rem;
		background: var(--bg-card);
		color: var(--text);
		border-radius: var(--radius);
		font-size: 1.05rem;
		text-align: left;
		border: 2px solid transparent;
		transition: all 0.2s;
	}

	.vote-btn:hover {
		border-color: var(--primary);
	}

	.results {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.result-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem;
		background: var(--bg-card);
		border-radius: var(--radius-sm);
	}

	.result-answer { flex: 1; }
	.result-votes { color: var(--primary); font-weight: 600; }

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
