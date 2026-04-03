<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { getSocket, saveSession, clearSession } from '$lib/socket';
	import { lobby, playerId, error } from '$lib/stores/lobby';

	let mode = $state<'home' | 'create' | 'join'>('home');
	let name = $state('');
	let code = $state('');
	let loading = $state(false);

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

<div class="page">
	<div class="hero">
		<h1 class="logo">Buzz<span class="hub">Hub</span></h1>
		<p class="tagline">Party-Spiele mit Freunden</p>
	</div>

	{#if $error}
		<div class="error-msg fade-in">{$error}</div>
	{/if}

	{#if mode === 'home'}
		<div class="actions fade-in">
			<button class="btn btn-primary" onclick={() => mode = 'create'}>
				Spiel erstellen
			</button>
			<button class="btn btn-secondary" onclick={() => mode = 'join'}>
				Spiel beitreten
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
	.hero {
		text-align: center;
		padding: 3rem 0 2rem;
	}

	.logo {
		font-size: 3.5rem;
		font-weight: 900;
		letter-spacing: -0.02em;
		color: #fff;
	}

	.hub {
		background: var(--primary);
		color: #000;
		padding: 0 0.2em;
		border-radius: 4px;
		margin-left: 0.05em;
	}

	.tagline {
		color: var(--text-muted);
		margin-top: 0.5rem;
		font-size: 1.1rem;
	}

	.actions {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		width: 100%;
		margin-top: 1rem;
	}

	.form-card {
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
</style>
