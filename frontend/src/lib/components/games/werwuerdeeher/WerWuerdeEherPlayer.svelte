<script lang="ts">
	import { lobby } from '$lib/stores/lobby';

	let { phase, data, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let voted = $state(false);

	$effect(() => {
		if (phase === 'pick') voted = false;
	});

	function vote(playerId: string) {
		if (voted) return;
		voted = true;
		sendAction('vote', { playerId });
	}

	function playerName(playerId: string): string {
		return $lobby?.players.find(player => player.id === playerId)?.name ?? playerId;
	}
</script>

<div class="would-rather fade-in">
	{#if phase === 'pick' && data}
		<div class="round-info">Runde {data.roundNum}/{data.totalRounds}</div>
		<h2 class="prompt">{data.prompt}</h2>

		{#if voted}
			<div class="submitted-msg">
				<p>Abgestimmt!</p>
				{#if data.votedCount}
					<p class="count">{data.votedCount}/{data.totalPlayers}</p>
				{/if}
			</div>
		{:else if data.choices}
			<div class="player-options">
				{#each data.choices as playerId}
					<button class="player-btn" onclick={() => vote(playerId)}>
						<span>{playerName(playerId)}</span>
					</button>
				{/each}
			</div>
		{/if}

	{:else if phase === 'results' && data}
		<div class="round-info">Runde {data.roundNum}/{data.totalRounds}</div>
		<h3 class="prompt">{data.prompt}</h3>
		<div class="results">
			{#if data.winners?.length}
				<div class="winner-box">
					<div class="winner-label">Mehrheit</div>
					<div class="winner-names">
						{#each data.winners as playerId, i}
							{playerName(playerId)}{i < data.winners.length - 1 ? ', ' : ''}
						{/each}
					</div>
				</div>
			{/if}

			{#if data.voteCounts}
				{#each Object.entries(data.voteCounts).sort((a, b) => (b[1] as number) - (a[1] as number)) as [playerId, count]}
					<div class="result-row">
						<span>{playerName(playerId)}</span>
						<strong>{count} Stimme{count === 1 ? '' : 'n'}</strong>
					</div>
				{/each}
			{/if}
		</div>

	{:else if phase === 'scoreboard' && data}
		<div class="scoreboard">
			<h2>{data.final ? 'Endergebnis' : 'Zwischenstand'}</h2>
			{#if data.scores}
				{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [playerId, score], i}
					<div class="score-row" class:top={i === 0}>
						<span class="rank">#{i + 1}</span>
						<span class="name">{playerName(playerId)}</span>
						<span class="pts">{score}</span>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</div>

<style>
	.would-rather { width: 100%; }

	.round-info {
		text-align: center;
		color: var(--text-muted);
		font-size: 0.85rem;
		margin-bottom: 0.5rem;
	}

	.prompt {
		text-align: center;
		font-size: 1.25rem;
		line-height: 1.35;
		margin-bottom: 1.5rem;
	}

	.player-options {
		display: grid;
		gap: 0.75rem;
	}

	.player-btn {
		width: 100%;
		min-height: 3.25rem;
		padding: 0.85rem 1rem;
		background: var(--bg-card);
		color: var(--text);
		border: 2px solid transparent;
		border-radius: var(--radius-sm);
		font-size: 1.05rem;
		font-weight: 700;
		text-align: left;
		transition: transform 0.15s, border-color 0.15s, background 0.15s;
	}

	.player-btn:hover {
		background: rgba(255, 176, 0, 0.08);
		border-color: var(--primary);
		transform: translateY(-1px);
	}

	.submitted-msg {
		text-align: center;
		padding: 2rem;
	}

	.submitted-msg p {
		font-size: 1.3rem;
		font-weight: 700;
	}

	.count {
		color: var(--text-muted);
		font-size: 0.9rem !important;
		font-weight: 400 !important;
		margin-top: 0.5rem;
	}

	.results {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.winner-box {
		padding: 1rem;
		border: 2px solid var(--primary);
		border-radius: var(--radius);
		background: rgba(255, 176, 0, 0.12);
		text-align: center;
	}

	.winner-label {
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		margin-bottom: 0.35rem;
	}

	.winner-names {
		color: var(--primary);
		font-size: 1.35rem;
		font-weight: 900;
	}

	.result-row, .score-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		background: var(--bg-card);
		border-radius: var(--radius-sm);
	}

	.result-row {
		justify-content: space-between;
	}

	.result-row strong, .pts {
		color: var(--primary);
	}

	.scoreboard h2 {
		text-align: center;
		margin-bottom: 1rem;
	}

	.score-row {
		margin-bottom: 0.5rem;
	}

	.score-row.top {
		background: rgba(233, 69, 96, 0.15);
		border: 1px solid var(--primary);
	}

	.rank {
		font-weight: 700;
		color: var(--text-muted);
		min-width: 2rem;
	}

	.name {
		flex: 1;
		font-weight: 600;
	}

	.pts {
		font-weight: 800;
	}
</style>
