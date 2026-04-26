<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getSocket, getPlayerId } from '$lib/socket';
	import { lobby, playerId, error } from '$lib/stores/lobby';
	import { currentGame } from '$lib/stores/game';
	import type { Lobby } from '$lib/stores/lobby';

	const GAMES = [
		{ id: 'quiz', name: 'Heisser Quickie', desc: 'Mehrere Spieler druecken gleichzeitig den richtigen Buzzer.', icon: '🔥', min: 2, duration: '6:30', views: '1.2M', thumb: '#e74c3c' },
		{ id: 'voting', name: 'Abstimmung XXL', desc: 'Wer hat die laengste Antwort? Alle bewerten sich gegenseitig.', icon: '🍆', min: 3, duration: '8:15', views: '856K', thumb: '#f39c12' },
		{ id: 'bluff', name: 'Fake Lust', desc: 'Verfuehre die Runde mit falschen Antworten und erkenne die Wahrheit.', icon: '🥵', min: 3, duration: '10:02', views: '2.1M', thumb: '#9b59b6' },
		{ id: 'drawing', name: 'Nacktes Zeichnen', desc: 'Alle starren auf deinen Stift. Zeichne schnell, bevor die Zeit kommt.', icon: '🫦', min: 3, duration: '12:45', views: '943K', thumb: '#2ecc71' },
		{ id: 'crossword', name: 'Versteckte Woerter', desc: 'Finde die geheimen Woerter im Gitter, bevor es die anderen tun.', icon: '💦', min: 2, duration: '3:00', views: '3.4M', thumb: '#e67e22', configurable: true },
		{ id: 'wordtrails', name: 'Buchstaben-Tease', desc: 'Verbinde Buchstaben zu passenden Woertern und fuelle alle Slots.', icon: '👅', min: 2, duration: '3:00', views: '3.9M', thumb: '#1abc9c' },
		{ id: 'codenames', name: 'Geheime Signale', desc: 'Zwei Teams, ein heisser Hinweis, viele verdeckte Karten.', icon: '🕵️', min: 4, duration: '15:00', views: '4.7M', thumb: '#c0392b' },
		{ id: 'headlines', name: 'Clickbait After Dark', desc: 'Ergaenze schamlose Schlagzeilen und vote den groessten Teaser.', icon: '😈', min: 3, duration: '7:00', views: '1.8M', thumb: '#d35400' },
		{ id: 'redflags', name: 'Red Flags Only', desc: 'Schreibe Warnsignale fuer Profile, Chats und fragwuerdige Anzeigen.', icon: '🚩', min: 3, duration: '7:00', views: '2.6M', thumb: '#e74c3c' },
		{ id: 'courtroom', name: 'Dirty Gericht', desc: 'Verteidige oder zerlege absurde Party-Vergehen vor der Runde.', icon: '⚖️', min: 3, duration: '7:30', views: '1.5M', thumb: '#34495e' },
		{ id: 'emoji', name: 'Emoji Beichte', desc: 'Deute zweideutige Geschichten aus reinem Emoji-Chaos.', icon: '😏', min: 3, duration: '6:30', views: '3.1M', thumb: '#f1c40f' },
		{ id: 'werwuerdeeher', name: 'Wer Wuerde Eher', desc: 'Waehlt direkt, wer zur schmutzigen Frage am besten passt.', icon: '👀', min: 3, duration: '5:30', views: '2.9M', thumb: '#8e44ad' },
		{ id: 'cvlies', name: 'Sexy Lebenslauf', desc: 'Erfinde fragwuerdige Qualifikationen fuer noch fragwuerdigere Jobs.', icon: '💼', min: 3, duration: '7:00', views: '1.1M', thumb: '#16a085' },
		{ id: 'memecourt', name: 'Meme Lust', desc: 'Schreibe Captions fuer Bilder, die niemand erklaeren will.', icon: '📸', min: 3, duration: '7:00', views: '4.2M', thumb: '#2980b9' },
		{ id: 'therapy', name: 'Aftercare Therapie', desc: 'Gib fragwuerdige Ratschlaege fuer kleine Dramen nach der Eskalation.', icon: '🛋️', min: 3, duration: '7:30', views: '2.0M', thumb: '#27ae60' },
		{ id: 'passwordpanic', name: 'Passwort Panik', desc: 'Errate Login-Geheimnisse, die besser verborgen geblieben waeren.', icon: '🔑', min: 3, duration: '6:30', views: '1.7M', thumb: '#7f8c8d' },
		{ id: 'lastwords', name: 'Letzter Hoehepunkt', desc: 'Schreibe den letzten Satz direkt vor der kompletten Eskalation.', icon: '💬', min: 3, duration: '7:00', views: '3.8M', thumb: '#c0392b' },
	];

	let lobbyData = $state<Lobby | null>(null);
	let myId = $state('');
	let isHost = $derived(lobbyData?.hostId === myId);
	let gameSettings = $state<Record<string, any>>({ gridSize: 10, duration: 90 });
	let showSettings = $derived(isHost && lobbyData?.gameId === 'crossword');

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
		// Send settings before starting if configurable
		const selectedGame = GAMES.find(g => g.id === lobbyData?.gameId);
		if (selectedGame?.configurable) {
			getSocket().emit('lobby:configure-game', { settings: gameSettings });
		}
		getSocket().emit('lobby:start-game', {});
	}

	function updateSetting(key: string, value: number) {
		gameSettings = { ...gameSettings, [key]: value };
	}

	function kickPlayer(pid: string) {
		if (!isHost) return;
		getSocket().emit('lobby:kick', { playerId: pid });
	}

	function leaveLobby() {
		getSocket().emit('lobby:leave', {});
		goto('/');
	}

	function aboutText(gameId: string): string {
		const texts: Record<string, string> = {
			quiz: 'Kurze Wissensrunden mit Tempo: Wer schnell richtig antwortet, sammelt die Punkte.',
			voting: 'Alle schreiben Antworten, danach entscheidet die Gruppe, welche am besten trifft.',
			bluff: 'Erfinde plausible Fake-Antworten und locke andere weg von der echten Loesung.',
			drawing: 'Eine Person zeichnet, der Rest raet live. Je schneller der Treffer, desto besser.',
			crossword: 'Durchsuche ein Buchstabengitter nach Woertern. Fruehe Funde bringen mehr Punkte.',
			wordtrails: 'Baue Woerter aus einem gemeinsamen Buchstabenpool und fuelle die versteckten Slots.',
			codenames: 'Teams geben Hinweise und decken Karten auf. Falsche Treffer helfen der Gegenseite.',
			headlines: 'Fuellt Luecken in uebertriebene Headlines und waehlt den klickstaerksten Unsinn.',
			redflags: 'Zu jedem Profil wird die schlimmste Warnung gesucht. Die Gruppe votet den Treffer.',
			courtroom: 'Alle liefern Statements zu absurden Anklagen. Das ueberzeugendste Plaedoyer gewinnt.',
			emoji: 'Aus Emoji-Ketten entstehen wilde Interpretationen. Die beste Deutung holt Stimmen.',
			werwuerdeeher: 'Alle zeigen auf eine Person. Mehrheitswahl und richtige Einschaetzung geben Punkte.',
			cvlies: 'Erfinde Qualifikationen fuer absurde Jobs. Die glaubwuerdigste Luege gewinnt.',
			memecourt: 'Alle schreiben Captions zu fiktiven Bildern. Die viralste Zeile bekommt die Stimmen.',
			therapy: 'Die Runde therapiert Alltagsdramen mit maximal fragwuerdigen Ratschlaegen.',
			passwordpanic: 'Errate das wahrscheinlichste Katastrophen-Passwort fuer absurde Accounts.',
			lastwords: 'Zu jeder Szene wird der letzte Satz vor dem Chaos gesucht. Beste Pointe gewinnt.',
		};
		return texts[gameId] ?? 'Kurzer Party-Modus fuer schnelle Runden mit der ganzen Lobby.';
	}

	function gameImage(gameId: string): string {
		return `/game-images/${gameId}.webp`;
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

		<!-- Game Selection -->
		<div class="video-section">
			<div class="video-grid">
				{#each GAMES as g}
					{@const disabled = !isHost || lobbyData.players.length < g.min}
					<button
						class="video-card"
						class:selected={lobbyData.gameId === g.id}
						class:disabled
						onclick={() => !disabled && selectGame(g.id)}
					>
						<div class="thumb" style="background-color: {g.thumb}; background-image: linear-gradient(180deg, rgba(0,0,0,0.08), rgba(0,0,0,0.48)), url('{gameImage(g.id)}');">
							<span class="thumb-icon">{g.icon}</span>
							<span class="info-badge" aria-label="Info zu {g.name}">i</span>
							<span class="about-popover">{aboutText(g.id)}</span>
							<span class="thumb-duration">{g.duration}</span>
							{#if lobbyData.gameId === g.id}
								<div class="thumb-selected-badge">AUSGEWAEHLT</div>
							{/if}
							{#if disabled && isHost}
								<div class="thumb-overlay">Min. {g.min} Spieler</div>
							{/if}
						</div>
						<div class="video-info">
							<div class="video-title">{g.name}</div>
							<div class="video-channel">BuzzHub Originals</div>
							<div class="video-about">{aboutText(g.id)}</div>
							<div class="video-meta">{g.views} Aufrufe</div>
						</div>
					</button>
				{/each}
			</div>
		</div>

		{#if showSettings}
			<div class="card settings-card fade-in">
				<h3>Einstellungen</h3>
				<div class="setting-row">
					<div class="setting-label">Spielfeldgroesse</div>
					<div class="setting-options">
						{#each [{v:8,l:'Klein (8x8)'},{v:10,l:'Mittel (10x10)'},{v:12,l:'Gross (12x12)'},{v:14,l:'Riesig (14x14)'}] as opt}
							<button
								class="setting-btn"
								class:active={gameSettings.gridSize === opt.v}
								onclick={() => updateSetting('gridSize', opt.v)}
							>{opt.l}</button>
						{/each}
					</div>
				</div>
				<div class="setting-row">
					<div class="setting-label">Zeitlimit</div>
					<div class="setting-options">
						{#each [{v:60,l:'60s'},{v:90,l:'90s'},{v:120,l:'2 Min'},{v:180,l:'3 Min'}] as opt}
							<button
								class="setting-btn"
								class:active={gameSettings.duration === opt.v}
								onclick={() => updateSetting('duration', opt.v)}
							>{opt.l}</button>
						{/each}
					</div>
				</div>
			</div>
		{/if}

		{#if isHost}
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
	.page {
		max-width: 1040px;
	}

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

	.players-card, .waiting-card {
		width: 100%;
		margin-top: 1rem;
	}

	.players-card h3 {
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

	.video-section {
		width: 100%;
		margin-top: 1rem;
	}

	.video-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.75rem;
	}

	.video-card {
		display: flex;
		flex-direction: column;
		background: none;
		color: var(--text);
		text-align: left;
		padding: 0;
		border: none;
		transition: opacity 0.2s;
	}

	.video-card.disabled {
		opacity: 0.35;
		cursor: not-allowed;
	}

	.thumb {
		position: relative;
		aspect-ratio: 16/9;
		border-radius: 2px;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		background-size: cover;
		background-position: center;
	}

	.thumb-icon {
		font-size: 2.5rem;
		filter: drop-shadow(0 2px 8px rgba(0,0,0,0.5));
		opacity: 0.92;
		z-index: 1;
	}

	.thumb-duration {
		position: absolute;
		bottom: 4px;
		right: 4px;
		background: rgba(0,0,0,0.8);
		color: #fff;
		font-size: 0.7rem;
		font-weight: 700;
		padding: 1px 4px;
		border-radius: 2px;
	}

	.info-badge {
		position: absolute;
		top: 6px;
		right: 6px;
		width: 1.25rem;
		height: 1.25rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(0,0,0,0.72);
		color: var(--primary);
		border: 1px solid rgba(255, 144, 0, 0.55);
		border-radius: 50%;
		font-size: 0.78rem;
		font-weight: 900;
		font-style: italic;
		z-index: 3;
	}

	.about-popover {
		position: absolute;
		left: 6px;
		right: 6px;
		bottom: 28px;
		padding: 0.5rem 0.55rem;
		background: rgba(0,0,0,0.86);
		color: #fff;
		border: 1px solid rgba(255, 144, 0, 0.45);
		border-radius: var(--radius-sm);
		font-size: 0.68rem;
		font-weight: 700;
		line-height: 1.25;
		opacity: 0;
		transform: translateY(4px);
		pointer-events: none;
		transition: opacity 0.16s, transform 0.16s;
		z-index: 2;
	}

	.video-card:hover .about-popover,
	.video-card:focus-visible .about-popover {
		opacity: 1;
		transform: translateY(0);
	}

	.thumb-selected-badge {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		background: var(--primary);
		color: #000;
		font-size: 0.6rem;
		font-weight: 800;
		padding: 2px 0;
		text-align: center;
		letter-spacing: 0.05em;
	}

	.thumb-overlay {
		position: absolute;
		inset: 0;
		background: rgba(0,0,0,0.7);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--primary);
	}

	.video-card.selected .thumb {
		outline: 2px solid var(--primary);
	}

	.video-info {
		padding: 0.35rem 0.1rem;
	}

	.video-title {
		font-size: 0.9rem;
		font-weight: 700;
		line-height: 1.3;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.video-channel {
		font-size: 0.7rem;
		color: var(--text-muted);
		margin-top: 0.15rem;
	}

	.video-about {
		margin-top: 0.25rem;
		color: #d7d7d7;
		font-size: 0.68rem;
		line-height: 1.25;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.video-meta {
		font-size: 0.65rem;
		color: var(--text-muted);
		margin-top: 0.18rem;
	}

	@media (max-width: 520px) {
		.page {
			max-width: 100%;
		}

		.video-grid {
			gap: 0.5rem;
		}

		.video-title {
			font-size: 0.72rem;
		}

		.video-channel {
			font-size: 0.62rem;
		}

		.video-about {
			display: none;
		}

		.video-meta {
			font-size: 0.58rem;
		}

		.thumb-icon {
			font-size: 1.8rem;
		}

		.info-badge {
			width: 1.05rem;
			height: 1.05rem;
			font-size: 0.68rem;
			top: 4px;
			right: 4px;
		}

		.about-popover {
			left: 4px;
			right: 4px;
			bottom: 22px;
			padding: 0.35rem 0.4rem;
			font-size: 0.58rem;
			line-height: 1.18;
		}
	}

	.settings-card {
		width: 100%;
		margin-top: 1rem;
	}

	.settings-card h3 {
		font-size: 0.9rem;
		color: var(--primary);
		margin-bottom: 0.75rem;
	}

	.setting-row {
		margin-bottom: 0.75rem;
	}

	.setting-label {
		display: block;
		font-size: 0.8rem;
		color: var(--text-muted);
		margin-bottom: 0.35rem;
	}

	.setting-options {
		display: flex;
		gap: 0.35rem;
		flex-wrap: wrap;
	}

	.setting-btn {
		padding: 0.4rem 0.6rem;
		font-size: 0.75rem;
		background: var(--bg-input);
		color: var(--text);
		border-radius: 2px;
		border: 1px solid #333;
		font-weight: 500;
		transition: all 0.15s;
	}

	.setting-btn.active {
		background: var(--primary);
		color: #000;
		border-color: var(--primary);
		font-weight: 700;
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
