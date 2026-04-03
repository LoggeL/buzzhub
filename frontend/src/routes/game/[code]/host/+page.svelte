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
</style>
