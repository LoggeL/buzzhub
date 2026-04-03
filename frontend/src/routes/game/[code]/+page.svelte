<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getSocket, getPlayerId } from '$lib/socket';
	import { lobby } from '$lib/stores/lobby';
	import { currentGame, gamePhase, gameData, timer, startTimer } from '$lib/stores/game';
	import QuizPlayer from '$lib/components/games/quiz/QuizPlayer.svelte';
	import VotingPlayer from '$lib/components/games/voting/VotingPlayer.svelte';
	import BluffPlayer from '$lib/components/games/bluff/BluffPlayer.svelte';
	import DrawingPlayer from '$lib/components/games/drawing/DrawingPlayer.svelte';

	let gameId = $state('');
	let phase = $state('');
	let data = $state<any>(null);
	let timerVal = $state(0);

	const gameComponents: Record<string, any> = {
		quiz: QuizPlayer,
		voting: VotingPlayer,
		bluff: BluffPlayer,
		drawing: DrawingPlayer,
	};

	onMount(() => {
		const socket = getSocket();
		const code = $page.params.code;

		currentGame.subscribe(g => { if (g) gameId = g.id; });
		gamePhase.subscribe(p => phase = p);
		gameData.subscribe(d => data = d);
		timer.subscribe(t => timerVal = t);

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

		socket.on('lobby:state', (msg: any) => {
			if (msg.lobby) {
				lobby.set(msg.lobby);
				if (msg.lobby.status === 'waiting') {
					goto(`/lobby/${code}`);
				}
			}
		});

		if (!gameId) {
			goto(`/lobby/${code}`);
		}

		return () => {
			socket.off('game:phase');
			socket.off('game:update');
			socket.off('game:timer');
			socket.off('game:end');
			socket.off('lobby:state');
		};
	});

	function sendAction(type: string, actionData: Record<string, any> = {}) {
		getSocket().emit('player:action', { type, ...actionData });
	}
</script>

<div class="page">
	{#if timerVal > 0}
		<div class="timer-bar">
			<div class="timer-bar-fill" style="width: {(timerVal / 15) * 100}%"></div>
		</div>
		<div class="timer-text">{timerVal}s</div>
	{/if}

	{#if phase === 'end'}
		<div class="end-screen card fade-in">
			<h2>Spiel beendet!</h2>
			{#if data?.scores}
				<div class="final-scores">
					{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [pid, score], i}
						<div class="score-row" class:winner={i === 0}>
							<span class="rank">#{i + 1}</span>
							<span class="score-name">{pid}</span>
							<span class="score-val">{score}</span>
						</div>
					{/each}
				</div>
			{/if}
			<p class="return-text">Zurueck zur Lobby...</p>
		</div>
	{:else if gameId && gameComponents[gameId]}
		{@const GameComponent = gameComponents[gameId]}
		<GameComponent
			{phase}
			{data}
			{timerVal}
			{sendAction}
		/>
	{:else}
		<div class="loading">Lade Spiel...</div>
	{/if}
</div>

<style>
	.timer-text {
		text-align: center;
		font-size: 1.2rem;
		font-weight: 700;
		color: var(--primary);
		margin-bottom: 0.5rem;
	}

	.end-screen {
		text-align: center;
		width: 100%;
		margin-top: 2rem;
	}

	.end-screen h2 {
		font-size: 1.8rem;
		margin-bottom: 1rem;
	}

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

	.rank {
		font-weight: 700;
		color: var(--text-muted);
		min-width: 2rem;
	}

	.score-name {
		flex: 1;
		font-weight: 500;
	}

	.score-val {
		font-weight: 700;
		color: var(--primary);
	}

	.return-text {
		color: var(--text-muted);
		margin-top: 1rem;
	}

	.loading {
		padding: 4rem 0;
		color: var(--text-muted);
	}
</style>
