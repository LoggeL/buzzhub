<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { getSocket, saveSession, clearSession } from '$lib/socket';
	import { lobby, playerId, error } from '$lib/stores/lobby';

	let mode = $state<'home' | 'create' | 'join'>('home');
	let name = $state('');
	let code = $state('');
	let loading = $state(false);

	const previews = [
		{ title: 'Heisser Quickie', meta: '1.2M Aufrufe', tag: 'HOT' },
		{ title: 'Fake Lust', meta: '2.1M Aufrufe', tag: 'NEW' },
		{ title: 'Buchstaben-Tease', meta: '3.9M Aufrufe', tag: 'TOP' },
	];

	const stats = [
		{ value: '17', label: 'Modi' },
		{ value: '16', label: 'Spieler' },
		{ value: '0', label: 'Downloads' },
	];

	const categories = ['Quickies', 'After Dark', 'Teamplay', 'Mindgames', 'Zeichnen', 'Wortspiele'];

	const nowPlaying = [
		'LEVP schaut gerade Red Flags Only',
		'MBUG ist live mit Buchstaben-Tease',
		'Host gesucht fuer Dirty Gericht',
	];

	onMount(() => {
		clearSession();
		const socket = getSocket();

		socket.on('lobby:created', (data: any) => {
			loading = false;
			lobby.set(data.lobby);
			playerId.set(data.playerId);
			saveSession(data.token, data.playerId, data.lobby.code);
			goto(`/lobby/${data.lobby.code}`);
		});

		socket.on('lobby:joined', (data: any) => {
			loading = false;
			lobby.set(data.lobby);
			playerId.set(data.playerId);
			saveSession(data.token, data.playerId, data.lobby.code);
			goto(`/lobby/${data.lobby.code}`);
		});

		socket.on('lobby:error', (data: any) => {
			loading = false;
			error.set(data.message);
			setTimeout(() => error.set(''), 3000);
		});

		return () => {
			socket.off('lobby:created');
			socket.off('lobby:joined');
			socket.off('lobby:error');
		};
	});

	function createLobby() {
		if (!name.trim()) return;
		loading = true;
		error.set('');
		getSocket().emit('lobby:create', { name: name.trim() });
	}

	function joinLobby() {
		if (!name.trim() || !code.trim()) return;
		loading = true;
		error.set('');
		getSocket().emit('lobby:join', { code: code.trim().toUpperCase(), name: name.trim() });
	}
</script>

<div class="page home-page">
	<div class="hero">
		<div class="hero-topline">
			<span>Live Party Channel</span>
			<span>18+ Vibes</span>
		</div>
		<h1 class="logo">Buzz<span class="hub">Hub</span></h1>
		<p class="tagline">After-Dark Partyspiele fuer den Gruppenchat, die Couch und den grossen Bildschirm.</p>
		<div class="preview-strip" aria-label="Spielvorschau">
			{#each previews as item}
				<div class="preview-card">
					<div class="preview-thumb">
						<span>{item.tag}</span>
					</div>
					<div>
						<strong>{item.title}</strong>
						<small>{item.meta}</small>
					</div>
				</div>
			{/each}
		</div>
	</div>

	{#if $error}
		<div class="error-msg fade-in">{$error}</div>
	{/if}

	{#if mode === 'home'}
		<div class="home-dashboard fade-in">
			<div class="stat-row" aria-label="BuzzHub Statistiken">
				{#each stats as stat}
					<div class="stat-item">
						<strong>{stat.value}</strong>
						<span>{stat.label}</span>
					</div>
				{/each}
			</div>

			<div class="category-rail" aria-label="Kategorien">
				{#each categories as category}
					<span>{category}</span>
				{/each}
			</div>

			<div class="live-feed" aria-label="Live Feed">
				<div class="live-dot"></div>
				<div class="feed-lines">
					{#each nowPlaying as line}
						<span>{line}</span>
					{/each}
				</div>
			</div>
		</div>

		<div class="actions fade-in">
			<button class="menu-action primary-action" onclick={() => mode = 'create'}>
				<span class="action-kicker">Host werden</span>
				<span class="action-title">Privaten Raum starten</span>
				<span class="action-copy">Code teilen, Modi waehlen, loslegen.</span>
			</button>
			<button class="menu-action secondary-action" onclick={() => mode = 'join'}>
				<span class="action-kicker">Code bereit?</span>
				<span class="action-title">In die Lobby rein</span>
				<span class="action-copy">Namen setzen und direkt beitreten.</span>
			</button>
		</div>
	{:else if mode === 'create'}
		<form class="form-card card fade-in" onsubmit={e => { e.preventDefault(); createLobby(); }}>
			<h2>Neues Spiel</h2>
			<input
				class="input"
				type="text"
				placeholder="Dein Name"
				bind:value={name}
				maxlength="20"
				autocomplete="off"
			/>
			<button class="btn btn-primary" type="submit" disabled={!name.trim() || loading}>
				{loading ? 'Wird erstellt...' : 'Lobby erstellen'}
			</button>
			<button class="btn btn-ghost" type="button" onclick={() => mode = 'home'}>
				Zurueck
			</button>
		</form>
	{:else}
		<form class="form-card card fade-in" onsubmit={e => { e.preventDefault(); joinLobby(); }}>
			<h2>Beitreten</h2>
			<input
				class="input"
				type="text"
				placeholder="Dein Name"
				bind:value={name}
				maxlength="20"
				autocomplete="off"
			/>
			<input
				class="input code-input"
				type="text"
				placeholder="Raumcode (z.B. ABCD)"
				bind:value={code}
				maxlength="4"
				autocomplete="off"
				style="text-transform: uppercase; letter-spacing: 0.3em; text-align: center; font-size: 1.5rem; font-weight: 700;"
			/>
			<button class="btn btn-primary" type="submit" disabled={!name.trim() || code.trim().length !== 4 || loading}>
				{loading ? 'Verbinde...' : 'Beitreten'}
			</button>
			<button class="btn btn-ghost" type="button" onclick={() => mode = 'home'}>
				Zurueck
			</button>
		</form>
	{/if}
</div>

<style>
	.home-page {
		position: relative;
		justify-content: center;
		max-width: 560px;
		padding-top: 2rem;
		padding-bottom: 2rem;
	}

	.home-page::before {
		content: '';
		position: fixed;
		inset: 0;
		pointer-events: none;
		background:
			linear-gradient(135deg, rgba(255, 144, 0, 0.18), transparent 32%),
			radial-gradient(circle at 78% 18%, rgba(255, 144, 0, 0.2), transparent 28%),
			linear-gradient(180deg, transparent 0%, rgba(255, 144, 0, 0.08) 100%);
	}

	.hero {
		position: relative;
		text-align: center;
		width: 100%;
		padding: 2.25rem 0 1.5rem;
	}

	.hero-topline {
		display: flex;
		justify-content: center;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.hero-topline span {
		padding: 0.25rem 0.5rem;
		background: rgba(255, 144, 0, 0.12);
		border: 1px solid rgba(255, 144, 0, 0.35);
		border-radius: 2px;
		color: var(--primary);
		font-size: 0.68rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}

	.logo {
		font-size: clamp(3rem, 14vw, 5.4rem);
		font-weight: 900;
		letter-spacing: 0;
		color: #fff;
		line-height: 0.95;
		text-transform: uppercase;
		text-shadow: 0 16px 40px rgba(0,0,0,0.65);
	}

	.hub {
		background: var(--primary);
		color: #000;
		padding: 0 0.2em;
		border-radius: 4px;
		margin-left: 0.05em;
	}

	.tagline {
		color: #d7d7d7;
		margin: 1rem auto 0;
		font-size: 1.05rem;
		line-height: 1.45;
		max-width: 28rem;
	}

	.preview-strip {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.5rem;
		margin-top: 1.5rem;
	}

	.preview-card {
		min-width: 0;
		background: #161616;
		border: 1px solid #282828;
		border-radius: var(--radius);
		overflow: hidden;
		text-align: left;
	}

	.preview-thumb {
		aspect-ratio: 16/9;
		display: flex;
		align-items: flex-end;
		justify-content: flex-start;
		padding: 0.4rem;
		background:
			linear-gradient(145deg, rgba(255, 144, 0, 0.95), rgba(121, 35, 0, 0.85)),
			#222;
	}

	.preview-thumb span {
		background: #000;
		color: var(--primary);
		border-radius: 2px;
		padding: 0.12rem 0.28rem;
		font-size: 0.58rem;
		font-weight: 900;
	}

	.preview-card div:last-child {
		padding: 0.45rem;
		display: grid;
		gap: 0.15rem;
	}

	.preview-card strong {
		font-size: 0.72rem;
		line-height: 1.2;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.preview-card small {
		color: var(--text-muted);
		font-size: 0.62rem;
	}

	.actions {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		width: 100%;
		margin-top: 1rem;
	}

	.home-dashboard {
		position: relative;
		width: 100%;
		display: grid;
		gap: 0.75rem;
	}

	.stat-row {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.5rem;
	}

	.stat-item {
		min-width: 0;
		padding: 0.75rem 0.5rem;
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid #2a2a2a;
		border-radius: var(--radius);
		text-align: center;
	}

	.stat-item strong {
		display: block;
		color: var(--primary);
		font-size: 1.4rem;
		font-weight: 900;
		line-height: 1;
	}

	.stat-item span {
		display: block;
		margin-top: 0.28rem;
		color: var(--text-muted);
		font-size: 0.68rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.category-rail {
		display: flex;
		gap: 0.4rem;
		overflow-x: auto;
		padding-bottom: 0.1rem;
		scrollbar-width: none;
	}

	.category-rail::-webkit-scrollbar {
		display: none;
	}

	.category-rail span {
		flex: 0 0 auto;
		padding: 0.42rem 0.58rem;
		background: #141414;
		border: 1px solid #303030;
		border-radius: 999px;
		color: #ededed;
		font-size: 0.72rem;
		font-weight: 800;
	}

	.live-feed {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 0.65rem;
		align-items: center;
		padding: 0.75rem;
		background: linear-gradient(135deg, rgba(255, 144, 0, 0.14), rgba(255, 255, 255, 0.04));
		border: 1px solid rgba(255, 144, 0, 0.26);
		border-radius: var(--radius);
		overflow: hidden;
	}

	.live-dot {
		width: 0.65rem;
		height: 0.65rem;
		border-radius: 50%;
		background: var(--primary);
		box-shadow: 0 0 0 5px rgba(255, 144, 0, 0.15);
	}

	.feed-lines {
		min-width: 0;
		display: flex;
		gap: 1.25rem;
		overflow: hidden;
		white-space: nowrap;
		color: #f0f0f0;
		font-size: 0.78rem;
		font-weight: 700;
	}

	.feed-lines span {
		flex: 0 0 auto;
	}

	.menu-action {
		width: 100%;
		display: grid;
		gap: 0.15rem;
		padding: 1rem;
		border-radius: var(--radius);
		text-align: left;
		color: var(--text);
		border: 1px solid #2a2a2a;
		background: #171717;
		transition: transform 0.18s, border-color 0.18s, background 0.18s;
	}

	.menu-action:hover {
		transform: translateY(-1px);
		border-color: rgba(255, 144, 0, 0.7);
	}

	.primary-action {
		background: linear-gradient(135deg, var(--primary), #d65f00);
		color: #050505;
		border-color: var(--primary);
	}

	.secondary-action {
		background: linear-gradient(135deg, #1d1d1d, #111);
	}

	.action-kicker {
		font-size: 0.68rem;
		font-weight: 900;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		opacity: 0.78;
	}

	.action-title {
		font-size: 1.12rem;
		font-weight: 900;
	}

	.action-copy {
		font-size: 0.86rem;
		color: inherit;
		opacity: 0.72;
	}

	.form-card {
		position: relative;
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 1rem;
	}

	.form-card h2 {
		text-align: center;
		font-size: 1.3rem;
		margin-bottom: 0.5rem;
	}

	@media (max-width: 420px) {
		.preview-strip {
			grid-template-columns: 1fr;
		}

		.preview-card {
			display: grid;
			grid-template-columns: 5rem 1fr;
		}

		.feed-lines {
			font-size: 0.72rem;
		}
	}
</style>
