<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { getSocket } from '$lib/socket';
	import { gamePhase, gameData, timer, startTimer } from '$lib/stores/game';

	let phase = $state('');
	let data = $state<any>(null);
	let timerVal = $state(0);
	let lobbyData = $state<any>(null);

	onMount(() => {
		const socket = getSocket();
		const code = $page.params.code;

		socket.emit('host:join', { code });

		gamePhase.subscribe(p => phase = p);
		gameData.subscribe(d => data = d);
		timer.subscribe(t => timerVal = t);

		socket.on('lobby:state', (msg: any) => {
			if (msg.lobby) lobbyData = msg.lobby;
		});

		socket.on('game:phase', (msg: any) => {
			gamePhase.set(msg.phase);
			gameData.set(msg.data);
		});

		socket.on('game:update', (msg: any) => {
			gameData.update(d => ({ ...d, ...msg }));
		});

		socket.on('game:timer', (msg: any) => {
			startTimer(Math.floor(msg.duration));
		});

		socket.on('game:end', (msg: any) => {
			gameData.set(msg);
			gamePhase.set('end');
		});

		return () => {
			socket.off('lobby:state');
			socket.off('game:phase');
			socket.off('game:update');
			socket.off('game:timer');
			socket.off('game:end');
		};
	});
</script>

<div class="host-page">
	<div class="host-header">
		<h1 class="host-logo">Buzz<span class="hub">Hub</span></h1>
		{#if timerVal > 0}
			<div class="host-timer">{timerVal}s</div>
		{/if}
	</div>

	{#if !phase && lobbyData}
		<div class="waiting-screen">
			<div class="room-code">{lobbyData.code}</div>
			<p>Verbinde dich auf buzzhub.logge.top</p>
			<div class="player-grid">
				{#each lobbyData.players as player}
					<div class="player-chip">{player.name}</div>
				{/each}
			</div>
		</div>

	{:else if phase === 'question' && data}
		<div class="host-question">
			<div class="q-num">Frage {data.questionNum}/{data.totalQuestions}</div>
			<h2>{data.question}</h2>
			<div class="host-options">
				{#each data.options as option, i}
					<div class="host-option option-{i}">{option}</div>
				{/each}
			</div>
		</div>

	{:else if phase === 'reveal' && data}
		<div class="host-reveal">
			<h2>Richtige Antwort</h2>
			<div class="correct">{data.correctText}</div>
		</div>

	{:else if phase === 'prompt' && data}
		<div class="host-prompt">
			<div class="round-num">Runde {data.roundNum}/{data.totalRounds}</div>
			<h2>{data.prompt}</h2>
			<p class="sub">Spieler schreiben ihre Antworten...</p>
		</div>

	{:else if phase === 'vote' && data}
		<div class="host-vote">
			<h2>Abstimmung</h2>
			{#if data.submissions}
				<div class="submission-grid">
					{#each data.submissions as sub}
						<div class="submission-card">{sub.answer}</div>
					{/each}
				</div>
			{/if}
		</div>

	{:else if phase === 'write' && data}
		<div class="host-prompt">
			<div class="round-num">Runde {data.roundNum}/{data.totalRounds}</div>
			<h2>{data.question}</h2>
			<p class="sub">Spieler erfinden falsche Antworten...</p>
		</div>

	{:else if phase === 'guess' && data}
		<div class="host-vote">
			<h2>Welche ist die Wahrheit?</h2>
			{#if data.options}
				<div class="submission-grid">
					{#each data.options as opt}
						<div class="submission-card">{opt.answer}</div>
					{/each}
				</div>
			{/if}
		</div>

	{:else if phase === 'draw' && data}
		<div class="host-drawing">
			<div class="hint-text">{data.hint}</div>
			<div class="host-canvas-container">
				<canvas id="hostCanvas"></canvas>
			</div>
		</div>

	{:else if phase === 'playing' && data}
		<div class="host-crossword">
			<h2>Woertersuche</h2>
			{#if data.grid}
				<div class="host-grid" style="grid-template-columns: repeat({data.gridSize}, 1fr);">
					{#each data.grid as row}
						{#each row.split('') as letter}
							<div class="host-cell">{letter}</div>
						{/each}
					{/each}
				</div>
			{/if}
			{#if data.words}
				<div class="host-word-list">
					{#each data.words as word}
						<span class="host-word">{word}</span>
					{/each}
				</div>
			{/if}
		</div>

	{:else if phase === 'assign' && data}
		<div class="host-codenames">
			<h2>Team-Zuordnung</h2>
			<div class="cn-teams">
				<div class="cn-team cn-red">
					<h3>Team Rot</h3>
					{#if data.teams}
						{#each Object.entries(data.teams) as [pid, team]}
							{#if team === 'red'}
								<div class="cn-member">
									{pid}
									{#if data.spymasters?.includes(pid)}
										<span class="cn-spy">SPY</span>
									{/if}
								</div>
							{/if}
						{/each}
					{/if}
				</div>
				<div class="cn-team cn-blue">
					<h3>Team Blau</h3>
					{#if data.teams}
						{#each Object.entries(data.teams) as [pid, team]}
							{#if team === 'blue'}
								<div class="cn-member">
									{pid}
									{#if data.spymasters?.includes(pid)}
										<span class="cn-spy">SPY</span>
									{/if}
								</div>
							{/if}
						{/each}
					{/if}
				</div>
			</div>
		</div>

	{:else if (phase === 'hint' || phase === 'guess') && data?.cards}
		<div class="host-codenames">
			<div class="cn-status">
				<span class="cn-turn cn-{data.currentTeam}">
					{data.currentTeam === 'red' ? 'ROT' : 'BLAU'} ist dran
				</span>
				<span class="cn-remaining">
					<span class="cn-r">{data.redLeft}</span> / <span class="cn-b">{data.blueLeft}</span>
				</span>
			</div>
			{#if phase === 'guess' && data.hint}
				<div class="cn-hint-display">
					<span class="cn-hint-word">{data.hint}</span>
					<span class="cn-hint-num">{data.hintNum}</span>
					<span class="cn-guesses">({data.guessesLeft} uebrig)</span>
				</div>
			{/if}
			<div class="cn-grid">
				{#each data.cards as card}
					<div class="cn-card cn-{card.color}" class:cn-revealed={card.revealed}>
						<span>{card.word}</span>
					</div>
				{/each}
			</div>
		</div>

	{:else if (phase === 'results' || phase === 'scoreboard' || phase === 'end') && data}
		<div class="host-scores">
			<h2>{data.final ? 'Endergebnis' : 'Zwischenstand'}</h2>
			{#if data.scores}
				<div class="score-list">
					{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [pid, score], i}
						<div class="host-score-row" class:winner={i === 0}>
							<span class="rank">#{i + 1}</span>
							<span class="name">{pid}</span>
							<span class="score">{score}</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{:else}
		<div class="waiting-screen">
			<p>Warte auf Spiel...</p>
		</div>
	{/if}
</div>

<style>
	.host-page {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		text-align: center;
		background: var(--bg);
	}

	.host-header {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 2rem;
		z-index: 10;
	}

	.host-logo {
		font-size: 1.5rem;
		font-weight: 900;
		color: #fff;
	}

	.hub {
		background: var(--primary);
		color: #000;
		padding: 0 0.15em;
		border-radius: 3px;
	}

	.host-timer {
		font-size: 2rem;
		font-weight: 700;
		color: var(--primary);
	}

	.waiting-screen {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1.5rem;
	}

	.room-code {
		font-size: 6rem;
		font-weight: 900;
		color: var(--primary);
		letter-spacing: 0.2em;
	}

	.player-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		justify-content: center;
		max-width: 600px;
	}

	.player-chip {
		background: var(--bg-card);
		padding: 0.5rem 1.25rem;
		border-radius: 2rem;
		font-weight: 600;
		font-size: 1.1rem;
	}

	.host-question, .host-reveal, .host-prompt, .host-vote, .host-drawing, .host-scores {
		max-width: 800px;
		width: 100%;
	}

	.q-num, .round-num {
		color: var(--text-muted);
		margin-bottom: 1rem;
		font-size: 1.1rem;
	}

	.host-question h2, .host-prompt h2, .host-vote h2 {
		font-size: 2.2rem;
		margin-bottom: 2rem;
		line-height: 1.3;
	}

	.host-options {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.host-option {
		padding: 1.5rem;
		border-radius: var(--radius);
		font-size: 1.3rem;
		font-weight: 600;
		color: white;
	}

	.option-0 { background: #e74c3c; }
	.option-1 { background: #3498db; }
	.option-2 { background: #2ecc71; }
	.option-3 { background: #f39c12; }

	.correct {
		font-size: 3rem;
		font-weight: 700;
		color: var(--success);
	}

	.sub {
		color: var(--text-muted);
		font-size: 1.2rem;
	}

	.submission-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		margin-top: 1rem;
	}

	.submission-card {
		background: var(--bg-card);
		padding: 1.5rem;
		border-radius: var(--radius);
		font-size: 1.2rem;
	}

	.hint-text {
		font-size: 2.5rem;
		font-weight: 700;
		letter-spacing: 0.3em;
		font-family: monospace;
		margin-bottom: 1.5rem;
	}

	.host-canvas-container {
		width: 500px;
		height: 500px;
		margin: 0 auto;
		border: 2px solid #333;
		border-radius: var(--radius);
		overflow: hidden;
	}

	.host-canvas-container canvas {
		width: 100%;
		height: 100%;
	}

	.host-crossword {
		max-width: 700px;
		width: 100%;
	}

	.host-crossword h2 {
		font-size: 2rem;
		margin-bottom: 1rem;
	}

	.host-grid {
		display: grid;
		gap: 0;
		max-width: 500px;
		margin: 0 auto;
	}

	.host-cell {
		aspect-ratio: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: 1.2rem;
		font-family: monospace;
		background: var(--bg-card);
		border: 1px solid #222;
	}

	.host-word-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		justify-content: center;
		margin-top: 1.5rem;
	}

	.host-word {
		background: var(--bg-card);
		padding: 0.4rem 0.8rem;
		border-radius: 2px;
		font-family: monospace;
		font-weight: 700;
		font-size: 1rem;
	}

	.host-scores h2 {
		font-size: 2rem;
		margin-bottom: 1.5rem;
	}

	.score-list {
		max-width: 500px;
		margin: 0 auto;
	}

	.host-score-row {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem 1.5rem;
		background: var(--bg-card);
		border-radius: var(--radius);
		margin-bottom: 0.75rem;
		font-size: 1.3rem;
	}

	.host-score-row.winner {
		background: rgba(233, 69, 96, 0.15);
		border: 2px solid var(--primary);
		font-size: 1.5rem;
	}

	.rank { font-weight: 700; color: var(--text-muted); min-width: 3rem; }
	.name { flex: 1; font-weight: 600; text-align: left; }
	.score { font-weight: 700; color: var(--primary); }

	/* Codenames Host */
	.host-codenames {
		max-width: 800px;
		width: 100%;
	}

	.host-codenames h2 {
		font-size: 2rem;
		margin-bottom: 1.5rem;
	}

	.cn-teams {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 2rem;
	}

	.cn-team {
		padding: 1.5rem;
		border-radius: var(--radius);
		background: var(--bg-card);
	}

	.cn-team h3 {
		font-size: 1.5rem;
		margin-bottom: 1rem;
	}

	.cn-red h3 { color: #e74c3c; }
	.cn-blue h3 { color: #3498db; }

	.cn-member {
		padding: 0.75rem 1rem;
		background: rgba(255,255,255,0.05);
		border-radius: var(--radius-sm);
		margin-bottom: 0.5rem;
		font-size: 1.2rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.cn-spy {
		font-size: 0.7rem;
		background: #f39c12;
		color: #000;
		padding: 0.15rem 0.5rem;
		border-radius: 3px;
		font-weight: 800;
	}

	.cn-status {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
		font-size: 1.5rem;
	}

	.cn-turn {
		font-weight: 800;
		padding: 0.4rem 1rem;
		border-radius: var(--radius-sm);
	}

	.cn-turn.cn-red { background: rgba(231,76,60,0.2); color: #e74c3c; }
	.cn-turn.cn-blue { background: rgba(52,152,219,0.2); color: #3498db; }

	.cn-remaining { font-weight: 700; }
	.cn-r { color: #e74c3c; }
	.cn-b { color: #3498db; }

	.cn-hint-display {
		text-align: center;
		margin-bottom: 1rem;
		padding: 0.75rem;
		background: var(--bg-card);
		border-radius: var(--radius-sm);
	}

	.cn-hint-word {
		font-size: 2rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.cn-hint-num {
		font-size: 2rem;
		font-weight: 800;
		color: var(--primary);
		margin-left: 0.75rem;
	}

	.cn-guesses {
		font-size: 1rem;
		color: var(--text-muted);
		margin-left: 0.5rem;
	}

	.cn-grid {
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 6px;
		max-width: 650px;
		margin: 0 auto;
	}

	.cn-card {
		aspect-ratio: 1.4;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 4px;
		font-weight: 700;
		font-size: 1rem;
		text-align: center;
		padding: 4px;
		border: 3px solid transparent;
	}

	.cn-card.cn-red { background: rgba(231,76,60,0.2); border-color: #e74c3c; }
	.cn-card.cn-blue { background: rgba(52,152,219,0.2); border-color: #3498db; }
	.cn-card.cn-neutral { background: rgba(189,183,160,0.2); border-color: #bdb7a0; }
	.cn-card.cn-assassin { background: rgba(30,30,30,0.8); border-color: #555; color: #fff; }

	.cn-card.cn-revealed.cn-red { background: #e74c3c; color: #fff; }
	.cn-card.cn-revealed.cn-blue { background: #3498db; color: #fff; }
	.cn-card.cn-revealed.cn-neutral { background: #8b8472; color: #fff; }
	.cn-card.cn-revealed.cn-assassin { background: #1a1a1a; color: #e74c3c; }
</style>
