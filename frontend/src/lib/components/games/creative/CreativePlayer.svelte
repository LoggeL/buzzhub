<script lang="ts">
	import { lobby } from '$lib/stores/lobby';

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
		if (phase === 'prompt') {
			submitted = false;
			answer = '';
		}
		if (phase === 'vote') {
			voted = false;
		}
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

	function playerName(pid: string): string {
		return $lobby?.players.find(player => player.id === pid)?.name ?? pid;
	}
</script>

<div class="creative fade-in">
	{#if phase === 'prompt' && data}
		<div class="mode-name">{data.modeName}</div>
		<div class="round-info">Runde {data.roundNum}/{data.totalRounds}</div>
		<div class="prompt-card">
			<div class="prompt-label">{data.promptLabel || 'Prompt'}</div>
			<h2>{data.prompt}</h2>
		</div>

		{#if submitted}
			<div class="submitted-msg">
				<p>Abgesendet!</p>
				{#if data.submittedCount}
					<p class="count">{data.submittedCount}/{data.totalPlayers}</p>
				{/if}
			</div>
		{:else}
			<form class="answer-form" onsubmit={e => { e.preventDefault(); submit(); }}>
				<label class="input-label" for="creative-answer">{data.inputLabel || 'Antwort'}</label>
				<textarea
					id="creative-answer"
					class="input answer-input"
					placeholder={data.placeholder || 'Deine Antwort...'}
					bind:value={answer}
					maxlength="140"
					autocomplete="off"
				></textarea>
				<button class="btn btn-primary" type="submit" disabled={!answer.trim()}>
					Absenden
				</button>
			</form>
		{/if}

	{:else if phase === 'vote' && data}
		<div class="mode-name">{data.modeName}</div>
		<h2 class="vote-title">{data.voteTitle || 'Beste Antwort?'}</h2>
		<p class="prompt-small">{data.prompt}</p>

		{#if voted}
			<div class="submitted-msg">
				<p>Abgestimmt!</p>
				{#if data.votedCount}
					<p class="count">{data.votedCount}/{data.totalPlayers}</p>
				{/if}
			</div>
		{:else if data.submissions?.length}
			<div class="vote-options">
				{#each data.submissions as sub}
					<button class="vote-btn" onclick={() => vote(sub.playerId)}>
						{sub.answer}
					</button>
				{/each}
			</div>
		{:else}
			<div class="submitted-msg">
				<p>Keine Antworten zum Abstimmen.</p>
			</div>
		{/if}

	{:else if phase === 'results' && data}
		<div class="mode-name">{data.modeName}</div>
		<h3 class="vote-title">{data.prompt}</h3>
		<div class="results">
			{#if data.submissions && data.voteCounts}
				{#each Object.entries(data.submissions) as [pid, text]}
					<div class="result-row">
						<div>
							<div class="result-answer">{text}</div>
							<div class="result-author">{playerName(pid)}</div>
						</div>
						<div class="result-votes">{data.voteCounts[pid] || 0} Stimmen</div>
					</div>
				{/each}
			{/if}
		</div>

	{:else if phase === 'scoreboard' && data}
		<div class="scoreboard">
			<h2>Endergebnis</h2>
			{#if data.scores}
				{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [pid, score], i}
					<div class="score-row" class:top={i === 0}>
						<span class="rank">#{i + 1}</span>
						<span class="name">{playerName(pid)}</span>
						<span class="pts">{score}</span>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</div>

<style>
	.creative { width: 100%; }

	.mode-name, .round-info, .prompt-label, .input-label, .prompt-small, .result-author {
		color: var(--text-muted);
	}

	.mode-name {
		text-align: center;
		font-size: 0.75rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		margin-bottom: 0.35rem;
	}

	.round-info {
		text-align: center;
		font-size: 0.85rem;
		margin-bottom: 0.75rem;
	}

	.prompt-card {
		background: var(--bg-card);
		border: 1px solid #333;
		border-radius: var(--radius);
		padding: 1rem;
		margin-bottom: 1rem;
	}

	.prompt-label, .input-label {
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.35rem;
		display: block;
	}

	.prompt-card h2, .vote-title {
		text-align: center;
		font-size: 1.2rem;
		line-height: 1.35;
	}

	.prompt-small {
		text-align: center;
		font-size: 0.9rem;
		margin: -0.75rem 0 1rem;
	}

	.answer-form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.answer-input {
		min-height: 5rem;
		resize: vertical;
		font-family: inherit;
		line-height: 1.35;
	}

	.submitted-msg {
		text-align: center;
		padding: 2rem 1rem;
	}

	.submitted-msg p {
		font-size: 1.2rem;
		font-weight: 600;
	}

	.count {
		font-size: 0.9rem !important;
		font-weight: 400 !important;
		color: var(--text-muted);
		margin-top: 0.5rem;
	}

	.vote-options, .results {
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
	}

	.vote-btn:hover {
		border-color: var(--primary);
	}

	.result-row {
		display: flex;
		justify-content: space-between;
		gap: 0.75rem;
		align-items: center;
		padding: 0.75rem;
		background: var(--bg-card);
		border-radius: var(--radius-sm);
	}

	.result-answer {
		font-weight: 600;
	}

	.result-author {
		font-size: 0.75rem;
		margin-top: 0.2rem;
	}

	.result-votes, .pts {
		color: var(--primary);
		font-weight: 700;
		white-space: nowrap;
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
</style>
