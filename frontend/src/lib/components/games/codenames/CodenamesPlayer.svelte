<script lang="ts">
	import { lobby } from '$lib/stores/lobby';

	let { phase, data, timerVal, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let selectedTeam = $state('');

	$effect(() => {
		if (data?.teams) {
			// Could update local state based on server
		}
	});

	function joinTeam(team: string) {
		selectedTeam = team;
		sendAction('join-team', { team });
	}

	function setSpymaster() {
		sendAction('set-spymaster', {});
	}

	function giveHint(e: SubmitEvent) {
		e.preventDefault();
		const form = e.target as HTMLFormElement;
		const word = (form.elements.namedItem('hintWord') as HTMLInputElement).value.trim();
		const count = parseInt((form.elements.namedItem('hintCount') as HTMLSelectElement).value);
		if (!word || !count) return;
		sendAction('give-hint', { word, count });
	}

	function guessCard(index: number) {
		sendAction('guess', { index });
	}

	function endTurn() {
		sendAction('end-turn', {});
	}

	function cardColor(card: any): string {
		if (!card.revealed && !card.color) return '';
		switch (card.color) {
			case 'red': return 'card-red';
			case 'blue': return 'card-blue';
			case 'neutral': return 'card-neutral';
			case 'assassin': return 'card-assassin';
			default: return '';
		}
	}

	function isMyTurn(d: any): boolean {
		return d?.myTeam === d?.currentTeam;
	}

	function playerName(pid: string): string {
		return $lobby?.players.find(player => player.id === pid)?.name ?? pid;
	}
</script>

<div class="codenames">
	{#if phase === 'assign' && data}
		<div class="assign-phase fade-in">
			<h2>Team waehlen</h2>
			<div class="team-columns">
				<div class="team-col team-red">
					<h3>Team Rot</h3>
					<button class="team-btn red" class:active={selectedTeam === 'red'} onclick={() => joinTeam('red')}>
						Beitreten
					</button>
					{#if data.teams}
						{#each Object.entries(data.teams) as [pid, team]}
							{#if team === 'red'}
								<div class="team-member">
									{pid}
									{#if data.spymasters?.includes(pid)}
										<span class="spy-badge">SPYMASTER</span>
									{/if}
								</div>
							{/if}
						{/each}
					{/if}
				</div>
				<div class="team-col team-blue">
					<h3>Team Blau</h3>
					<button class="team-btn blue" class:active={selectedTeam === 'blue'} onclick={() => joinTeam('blue')}>
						Beitreten
					</button>
					{#if data.teams}
						{#each Object.entries(data.teams) as [pid, team]}
							{#if team === 'blue'}
								<div class="team-member">
									{pid}
									{#if data.spymasters?.includes(pid)}
										<span class="spy-badge">SPYMASTER</span>
									{/if}
								</div>
							{/if}
						{/each}
					{/if}
				</div>
			</div>
			{#if selectedTeam}
				<button class="btn btn-primary spy-btn" onclick={setSpymaster}>
					Ich bin Spymaster!
				</button>
			{/if}
		</div>

	{:else if phase === 'hint' && data}
		<div class="game-phase fade-in">
			<div class="status-bar">
				<span class="team-indicator {data.currentTeam}">
					{data.currentTeam === 'red' ? 'Rot' : 'Blau'} ist dran
				</span>
				<span class="counts">
					<span class="count-red">{data.redLeft}</span> /
					<span class="count-blue">{data.blueLeft}</span>
				</span>
			</div>

			<div class="grid">
				{#each data.cards as card, i}
					<div class="card-cell {cardColor(card)}" class:revealed={card.revealed}>
						<span class="card-word">{card.word}</span>
					</div>
				{/each}
			</div>

			{#if data.isSpymaster && isMyTurn(data)}
				<form class="hint-form" onsubmit={giveHint}>
					<input name="hintWord" type="text" placeholder="Hinwort..." class="hint-input" autocomplete="off" />
					<select name="hintCount" class="hint-select">
						{#each [1,2,3,4,5,6,7,8,9] as n}
							<option value={n}>{n}</option>
						{/each}
					</select>
					<button type="submit" class="btn btn-primary">Senden</button>
				</form>
			{:else if data.isSpymaster}
				<p class="wait-text">Warte auf den anderen Spymaster...</p>
			{:else}
				<p class="wait-text">Spymaster denkt nach...</p>
			{/if}
		</div>

	{:else if phase === 'guess' && data}
		<div class="game-phase fade-in">
			<div class="status-bar">
				<span class="team-indicator {data.currentTeam}">
					{data.currentTeam === 'red' ? 'Rot' : 'Blau'} ist dran
				</span>
				<span class="counts">
					<span class="count-red">{data.redLeft}</span> /
					<span class="count-blue">{data.blueLeft}</span>
				</span>
			</div>

			<div class="hint-display">
				<span class="hint-word">{data.hint}</span>
				<span class="hint-num">{data.hintNum}</span>
				{#if data.guessesLeft != null}
					<span class="guesses-left">({data.guessesLeft} uebrig)</span>
				{/if}
			</div>

			<div class="grid">
				{#each data.cards as card, i}
					{@const canClick = !card.revealed && isMyTurn(data) && !data.isSpymaster}
					<button
						class="card-cell {cardColor(card)}"
						class:revealed={card.revealed}
						class:clickable={canClick}
						onclick={() => canClick && guessCard(i)}
						disabled={!canClick}
					>
						<span class="card-word">{card.word}</span>
					</button>
				{/each}
			</div>

			{#if isMyTurn(data) && !data.isSpymaster}
				<button class="btn btn-ghost end-turn-btn" onclick={endTurn}>
					Zug beenden
				</button>
			{:else if data.isSpymaster && isMyTurn(data)}
				<p class="wait-text">Dein Team raet...</p>
			{:else}
				<p class="wait-text">Anderes Team ist dran...</p>
			{/if}
		</div>

	{:else if phase === 'results' && data}
		<div class="results-phase fade-in">
			<h2>Spiel beendet!</h2>
			<div class="winner-display {data.winner}">
				Team {data.winner === 'red' ? 'Rot' : 'Blau'} gewinnt!
			</div>
			<div class="grid grid-small">
				{#each data.cards as card, i}
					<div class="card-cell {cardColor(card)} revealed">
						<span class="card-word">{card.word}</span>
					</div>
				{/each}
			</div>
		</div>

	{:else if phase === 'scoreboard' && data}
		<div class="results-phase fade-in">
			<h2>{data.final ? 'Endergebnis' : 'Zwischenstand'}</h2>
			{#if data.scores}
				<div class="final-scores">
					{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [pid, score], i}
						<div class="score-row" class:winner={i === 0}>
							<span class="rank">#{i + 1}</span>
							<span class="score-name">{playerName(pid)}</span>
							<span class="score-val">{score}</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.codenames {
		width: 100%;
	}

	.assign-phase {
		text-align: center;
	}

	.assign-phase h2 {
		font-size: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.team-columns {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.team-col {
		padding: 1rem;
		border-radius: var(--radius);
		background: var(--bg-card);
	}

	.team-col h3 {
		font-size: 1rem;
		margin-bottom: 0.75rem;
	}

	.team-red h3 { color: #e74c3c; }
	.team-blue h3 { color: #3498db; }

	.team-btn {
		width: 100%;
		padding: 0.6rem;
		border-radius: var(--radius-sm);
		font-weight: 700;
		font-size: 0.85rem;
		margin-bottom: 0.75rem;
		border: 2px solid transparent;
		background: var(--bg-input);
		color: var(--text);
	}

	.team-btn.red:hover, .team-btn.red.active { background: rgba(231,76,60,0.2); border-color: #e74c3c; }
	.team-btn.blue:hover, .team-btn.blue.active { background: rgba(52,152,219,0.2); border-color: #3498db; }

	.team-member {
		padding: 0.4rem 0.6rem;
		background: var(--bg-input);
		border-radius: var(--radius-sm);
		font-size: 0.85rem;
		margin-bottom: 0.35rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.spy-badge {
		font-size: 0.6rem;
		background: #f39c12;
		color: #000;
		padding: 0.1rem 0.35rem;
		border-radius: 3px;
		font-weight: 800;
	}

	.spy-btn {
		margin-top: 0.5rem;
	}

	.game-phase, .results-phase {
		width: 100%;
	}

	.status-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}

	.team-indicator {
		font-weight: 700;
		font-size: 1rem;
		padding: 0.3rem 0.75rem;
		border-radius: var(--radius-sm);
	}

	.team-indicator.red { background: rgba(231,76,60,0.2); color: #e74c3c; }
	.team-indicator.blue { background: rgba(52,152,219,0.2); color: #3498db; }

	.counts {
		font-weight: 700;
		font-size: 1rem;
	}

	.count-red { color: #e74c3c; }
	.count-blue { color: #3498db; }

	.hint-display {
		text-align: center;
		margin-bottom: 0.75rem;
		padding: 0.5rem;
		background: var(--bg-card);
		border-radius: var(--radius-sm);
	}

	.hint-word {
		font-size: 1.3rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.hint-num {
		font-size: 1.3rem;
		font-weight: 800;
		color: var(--primary);
		margin-left: 0.5rem;
	}

	.guesses-left {
		font-size: 0.85rem;
		color: var(--text-muted);
		margin-left: 0.5rem;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 4px;
		margin-bottom: 1rem;
	}

	.grid-small {
		max-width: 400px;
		margin: 1rem auto;
	}

	.card-cell {
		aspect-ratio: 1.4;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-card);
		border-radius: 3px;
		border: 2px solid #333;
		padding: 2px;
		transition: all 0.2s;
	}

	.card-cell.clickable {
		cursor: pointer;
	}

	.card-cell.clickable:hover {
		border-color: var(--primary);
		transform: scale(1.05);
	}

	.card-cell.revealed, .card-cell.card-red.revealed {
		opacity: 0.85;
	}

	.card-cell.card-red { background: rgba(231,76,60,0.25); border-color: #e74c3c; }
	.card-cell.card-blue { background: rgba(52,152,219,0.25); border-color: #3498db; }
	.card-cell.card-neutral { background: rgba(189,183,160,0.25); border-color: #bdb7a0; }
	.card-cell.card-assassin { background: rgba(30,30,30,0.9); border-color: #555; }

	.card-cell.card-red.revealed { background: #e74c3c; }
	.card-cell.card-blue.revealed { background: #3498db; }
	.card-cell.card-neutral.revealed { background: #8b8472; }
	.card-cell.card-assassin.revealed { background: #1a1a1a; }

	.card-word {
		font-size: 0.6rem;
		font-weight: 700;
		text-transform: uppercase;
		text-align: center;
		line-height: 1.1;
		word-break: break-all;
	}

	.hint-form {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.hint-input {
		flex: 1;
		padding: 0.6rem 0.75rem;
		background: var(--bg-input);
		border: 1px solid #333;
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: 0.9rem;
	}

	.hint-select {
		padding: 0.6rem;
		background: var(--bg-input);
		border: 1px solid #333;
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: 0.9rem;
		width: 3.5rem;
	}

	.wait-text {
		text-align: center;
		color: var(--text-muted);
		font-style: italic;
	}

	.end-turn-btn {
		width: 100%;
	}

	.results-phase {
		text-align: center;
	}

	.results-phase h2 {
		font-size: 1.5rem;
		margin-bottom: 1rem;
	}

	.winner-display {
		font-size: 1.3rem;
		font-weight: 800;
		padding: 1rem;
		border-radius: var(--radius);
		margin-bottom: 1rem;
	}

	.winner-display.red { background: rgba(231,76,60,0.2); color: #e74c3c; }
	.winner-display.blue { background: rgba(52,152,219,0.2); color: #3498db; }

	.final-scores {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin: 1rem 0;
	}

	.score-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		background: var(--bg-input);
		border-radius: var(--radius-sm);
	}

	.score-row.winner {
		background: rgba(233, 69, 96, 0.15);
		border: 1px solid var(--primary);
	}

	.rank { font-weight: 700; color: var(--text-muted); min-width: 2rem; }
	.score-name { flex: 1; font-weight: 500; }
	.score-val { font-weight: 700; color: var(--primary); }
</style>
