<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getSocket, getPlayerId } from '$lib/socket';
	import { lobby, playerId, error } from '$lib/stores/lobby';
	import { currentGame } from '$lib/stores/game';
	import type { Lobby } from '$lib/stores/lobby';

	const GAMES = [
		{ id: 'quiz', name: 'Quiz Battle', desc: 'Beantworte Fragen schneller als alle anderen!', icon: '🧠', min: 2 },
		{ id: 'voting', name: 'Abstimmung', desc: 'Schreibe die witzigste Antwort!', icon: '🏆', min: 3 },
		{ id: 'bluff', name: 'Bluff Master', desc: 'Taeuschen deine Mitspieler mit falschen Antworten!', icon: '🎭', min: 3 },
		{ id: 'drawing', name: 'Kritzelei', desc: 'Zeichne und lass die anderen raten!', icon: '✏️', min: 3 },
	];

	let lobbyData = $state<Lobby | null>(null);
	let myId = $state('');
	let isHost = $derived(lobbyData?.hostId === myId);

	onMount(() => {
		myId = getPlayerId() || '';
		const socket = getSocket();
		const code = $page.params.code;

		lobby.subscribe(l => { if (l) lobbyData = l; });

		socket.on('lobby:state', (data: any) => {
			if (data.lobby) lobby.set(data.lobby);
		});

		socket.on('lobby:player-joined', (data: any) => {
			if (data.lobby) lobby.set(data.lobby);
		});

		socket.on('lobby:player-left', (data: any) => {
			if (data.lobby) lobby.set(data.lobby);
			if (data.playerId === myId && data.kicked) {
				goto('/');
			}
		});

		socket.on('lobby:error', (data: any) => {
			error.set(data.message);
			setTimeout(() => error.set(''), 3000);
		});

		socket.on('game:start', (data: any) => {
			currentGame.set(data.game);
			goto(`/game/${code}`);
		});

		if (!lobbyData) {
			// Page reload — try rejoin
			const token = localStorage.getItem('buzzhub_token');
			if (token) {
				socket.emit('lobby:rejoin', { token });
			} else {
				goto('/');
			}
		}

		return () => {
			socket.off('lobby:state');
			socket.off('lobby:player-joined');
			socket.off('lobby:player-left');
			socket.off('lobby:error');
			socket.off('game:start');
		};
	});

	function selectGame(gameId: string) {
		if (!isHost) return;
		getSocket().emit('lobby:select-game', { gameId });
	}

	function startGame() {
		if (!isHost || !lobbyData?.gameId) return;
		getSocket().emit('lobby:start-game', {});
	}

	function kickPlayer(pid: string) {
		if (!isHost) return;
		getSocket().emit('lobby:kick', { playerId: pid });
	}

	function leaveLobby() {
		getSocket().emit('lobby:leave', {});
		goto('/');
	}
</script>

<div class="page">
	{#if lobbyData}
		<div class="lobby-header">
			<div class="room-code-label">Raumcode</div>
			<div class="room-code">{lobbyData.code}</div>
			<div class="player-count">{lobbyData.players.length} / {lobbyData.maxPlayers} Spieler</div>
		</div>

		{#if $error}
			<div class="error-msg fade-in">{$error}</div>
		{/if}

		<!-- Player List -->
		<div class="card players-card">
			<h3>Spieler</h3>
			<div class="player-list">
				{#each lobbyData.players as player}
					<div class="player-item" class:disconnected={!player.connected}>
						<span class="player-name">
							{player.name}
							{#if player.id === lobbyData.hostId}
								<span class="host-badge">HOST</span>
							{/if}
							{#if player.id === myId}
								<span class="you-badge">DU</span>
							{/if}
						</span>
						{#if !player.connected}
							<span class="status-dot offline"></span>
						{/if}
						{#if isHost && player.id !== myId}
							<button class="kick-btn" onclick={() => kickPlayer(player.id)}>✕</button>
						{/if}
					</div>
				{/each}
			</div>
		</div>

		<!-- Game Selection (Host only) -->
		{#if isHost}
			<div class="card game-select-card">
				<h3>Spiel waehlen</h3>
				<div class="game-grid">
					{#each GAMES as g}
						{@const disabled = lobbyData.players.length < g.min}
						<button
							class="game-tile"
							class:selected={lobbyData.gameId === g.id}
							class:disabled
							onclick={() => !disabled && selectGame(g.id)}
						>
							<span class="game-icon">{g.icon}</span>
							<span class="game-name">{g.name}</span>
							<span class="game-desc">{g.desc}</span>
							{#if disabled}
								<span class="game-min">Min. {g.min} Spieler</span>
							{/if}
						</button>
					{/each}
				</div>
			</div>

			<button
				class="btn btn-primary start-btn"
				disabled={!lobbyData.gameId}
				onclick={startGame}
			>
				Spiel starten
			</button>
		{:else}
			<div class="card waiting-card">
				<p class="waiting-text">Warte auf den Host...</p>
				{#if lobbyData.gameId}
					{@const selected = GAMES.find(g => g.id === lobbyData?.gameId)}
					{#if selected}
						<div class="selected-game">
							<span>{selected.icon}</span>
							<span>{selected.name}</span>
						</div>
					{/if}
				{/if}
			</div>
		{/if}

		<button class="btn btn-ghost leave-btn" onclick={leaveLobby}>
			Lobby verlassen
		</button>
	{:else}
		<div class="loading">Verbinde...</div>
	{/if}
</div>

<style>
	.lobby-header {
		text-align: center;
		padding: 2rem 0 1rem;
	}

	.room-code-label {
		color: var(--text-muted);
		font-size: 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
	}

	.room-code {
		font-size: 3rem;
		font-weight: 900;
		letter-spacing: 0.2em;
		color: var(--primary);
	}

	.player-count {
		color: var(--text-muted);
		font-size: 0.9rem;
		margin-top: 0.25rem;
	}

	.players-card, .game-select-card, .waiting-card {
		width: 100%;
		margin-top: 1rem;
	}

	.players-card h3, .game-select-card h3 {
		font-size: 1rem;
		color: var(--text-muted);
		margin-bottom: 0.75rem;
	}

	.player-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.player-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem 0.75rem;
		background: var(--bg-input);
		border-radius: var(--radius-sm);
	}

	.player-item.disconnected {
		opacity: 0.5;
	}

	.player-name {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 500;
	}

	.host-badge {
		font-size: 0.65rem;
		background: var(--primary);
		color: #000;
		padding: 0.15rem 0.4rem;
		border-radius: 3px;
		font-weight: 700;
	}

	.you-badge {
		font-size: 0.65rem;
		background: var(--accent);
		padding: 0.15rem 0.4rem;
		border-radius: 4px;
		font-weight: 700;
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.status-dot.offline {
		background: var(--danger);
	}

	.kick-btn {
		background: none;
		color: var(--text-muted);
		font-size: 1rem;
		padding: 0.25rem;
	}

	.kick-btn:hover {
		color: var(--danger);
	}

	.game-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.75rem;
	}

	.game-tile {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 0.25rem;
		padding: 1rem 0.5rem;
		background: var(--bg-input);
		border-radius: var(--radius);
		border: 2px solid transparent;
		color: var(--text);
		transition: all 0.2s;
	}

	.game-tile:hover:not(.disabled) {
		border-color: var(--primary);
	}

	.game-tile.selected {
		border-color: var(--primary);
		background: rgba(233, 69, 96, 0.1);
	}

	.game-tile.disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.game-icon {
		font-size: 2rem;
	}

	.game-name {
		font-weight: 600;
		font-size: 0.9rem;
	}

	.game-desc {
		font-size: 0.7rem;
		color: var(--text-muted);
		line-height: 1.3;
	}

	.game-min {
		font-size: 0.65rem;
		color: var(--warning);
	}

	.start-btn {
		margin-top: 1rem;
	}

	.waiting-card {
		text-align: center;
	}

	.waiting-text {
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.selected-game {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		margin-top: 0.75rem;
		font-size: 1.2rem;
		font-weight: 600;
	}

	.leave-btn {
		margin-top: 1.5rem;
	}

	.loading {
		padding: 4rem 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}
</style>
