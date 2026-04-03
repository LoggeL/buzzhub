<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	let { phase, data, timerVal, sendAction }: {
		phase: string;
		data: any;
		timerVal: number;
		sendAction: (type: string, data?: Record<string, any>) => void;
	} = $props();

	let canvas = $state<HTMLCanvasElement>(undefined!);
	let ctx: CanvasRenderingContext2D | null = null;
	let drawing = $state(false);
	let color = $state('#ffffff');
	let brushSize = $state(4);
	let guess = $state('');
	let guessedCorrectly = $state(false);
	let currentStroke: { x: number; y: number }[] = [];
	let strokeBatch: any[] = [];
	let batchTimer: ReturnType<typeof setInterval> | null = null;

	const colors = ['#ffffff', '#e74c3c', '#3498db', '#2ecc71', '#f39c12', '#9b59b6'];

	$effect(() => {
		if (phase === 'draw') {
			guessedCorrectly = false;
			guess = '';
			if (canvas) initCanvas();
		}
	});

	$effect(() => {
		if (data?.stroke && data.role !== 'drawer') {
			drawRemoteStroke(data.stroke);
		}
		if (data?.correct) {
			guessedCorrectly = true;
		}
	});

	onMount(() => {
		return () => {
			if (batchTimer) clearInterval(batchTimer);
		};
	});

	function initCanvas() {
		if (!canvas) return;
		ctx = canvas.getContext('2d');
		if (!ctx) return;

		const rect = canvas.parentElement!.getBoundingClientRect();
		canvas.width = rect.width;
		canvas.height = rect.width;
		ctx.fillStyle = '#1a1a2e';
		ctx.fillRect(0, 0, canvas.width, canvas.height);
		ctx.lineCap = 'round';
		ctx.lineJoin = 'round';

		// Start batch timer
		if (batchTimer) clearInterval(batchTimer);
		batchTimer = setInterval(flushStrokes, 100);
	}

	function getPos(e: PointerEvent) {
		const rect = canvas.getBoundingClientRect();
		return {
			x: (e.clientX - rect.left) / rect.width,
			y: (e.clientY - rect.top) / rect.height,
		};
	}

	function onPointerDown(e: PointerEvent) {
		if (data?.role !== 'drawer') return;
		drawing = true;
		canvas.setPointerCapture(e.pointerId);
		currentStroke = [getPos(e)];
		const pos = getPos(e);
		if (ctx) {
			ctx.beginPath();
			ctx.strokeStyle = color;
			ctx.lineWidth = brushSize;
			ctx.moveTo(pos.x * canvas.width, pos.y * canvas.height);
		}
	}

	function onPointerMove(e: PointerEvent) {
		if (!drawing || data?.role !== 'drawer') return;
		const pos = getPos(e);
		currentStroke.push(pos);
		if (ctx) {
			ctx.lineTo(pos.x * canvas.width, pos.y * canvas.height);
			ctx.stroke();
		}
	}

	function onPointerUp() {
		if (!drawing) return;
		drawing = false;
		if (currentStroke.length > 0) {
			strokeBatch.push({
				points: currentStroke,
				color,
				size: brushSize,
			});
		}
		currentStroke = [];
	}

	function flushStrokes() {
		if (strokeBatch.length === 0) return;
		sendAction('draw', { stroke: strokeBatch });
		strokeBatch = [];
	}

	function drawRemoteStroke(strokes: any[]) {
		if (!ctx || !canvas || !Array.isArray(strokes)) return;
		for (const s of strokes) {
			if (!s.points || s.points.length < 2) continue;
			ctx.beginPath();
			ctx.strokeStyle = s.color || '#fff';
			ctx.lineWidth = s.size || 4;
			ctx.moveTo(s.points[0].x * canvas.width, s.points[0].y * canvas.height);
			for (let i = 1; i < s.points.length; i++) {
				ctx.lineTo(s.points[i].x * canvas.width, s.points[i].y * canvas.height);
			}
			ctx.stroke();
		}
	}

	function submitGuess() {
		if (!guess.trim() || guessedCorrectly) return;
		sendAction('guess', { guess: guess.trim() });
		guess = '';
	}

	function clearCanvas() {
		if (!ctx || !canvas) return;
		ctx.fillStyle = '#1a1a2e';
		ctx.fillRect(0, 0, canvas.width, canvas.height);
	}
</script>

<div class="drawing fade-in">
	{#if phase === 'draw' && data}
		<div class="turn-info">Runde {data.turnNum}/{data.totalTurns}</div>

		{#if data.role === 'drawer'}
			<div class="word-display">
				<span class="word-label">Dein Wort:</span>
				<span class="word">{data.word}</span>
			</div>

			<div class="canvas-container">
				<canvas
					bind:this={canvas}
					onpointerdown={onPointerDown}
					onpointermove={onPointerMove}
					onpointerup={onPointerUp}
					onpointerleave={onPointerUp}
					style="touch-action: none;"
				></canvas>
			</div>

			<div class="toolbar">
				<div class="color-picker">
					{#each colors as c}
						<button
							class="color-btn"
							class:active={color === c}
							style="background: {c}"
							onclick={() => color = c}
							aria-label="Farbe {c}"
						></button>
					{/each}
				</div>
				<div class="brush-sizes">
					<button
						class="size-btn"
						class:active={brushSize === 3}
						onclick={() => brushSize = 3}
						aria-label="Duenner Pinsel"
					>
						<span class="dot" style="width:6px;height:6px"></span>
					</button>
					<button
						class="size-btn"
						class:active={brushSize === 8}
						onclick={() => brushSize = 8}
						aria-label="Dicker Pinsel"
					>
						<span class="dot" style="width:12px;height:12px"></span>
					</button>
				</div>
				<button class="btn btn-ghost clear-btn" onclick={clearCanvas}>Loeschen</button>
			</div>

		{:else}
			<div class="hint-display">
				<span class="hint">{data.hint}</span>
			</div>

			<div class="canvas-container">
				<canvas bind:this={canvas} style="touch-action: none;"></canvas>
			</div>

			{#if guessedCorrectly}
				<div class="guessed-msg">
					<p>Richtig geraten! +{data.points || 100} Punkte</p>
				</div>
			{:else}
				<form class="guess-form" onsubmit={e => { e.preventDefault(); submitGuess(); }}>
					<input
						class="input"
						type="text"
						placeholder="Dein Tipp..."
						bind:value={guess}
						maxlength="50"
						autocomplete="off"
					/>
					<button class="btn btn-primary" type="submit" disabled={!guess.trim()}>
						Raten
					</button>
				</form>
			{/if}

			{#if data.chatGuess}
				<div class="chat-msg fade-in">
					<strong>{data.chatGuess.playerId}:</strong> {data.chatGuess.guess}
				</div>
			{/if}
		{/if}

	{:else if phase === 'reveal' && data}
		<div class="reveal">
			<h3>Das Wort war:</h3>
			<div class="revealed-word">{data.word}</div>
		</div>

	{:else if phase === 'scoreboard' && data}
		<div class="scoreboard">
			<h2>{data.final ? 'Endergebnis' : 'Zwischenstand'}</h2>
			{#if data.scores}
				{#each Object.entries(data.scores).sort((a, b) => (b[1] as number) - (a[1] as number)) as [pid, score], i}
					<div class="score-row" class:top={i === 0}>
						<span class="rank">#{i + 1}</span>
						<span class="name">{pid}</span>
						<span class="pts">{score}</span>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</div>

<style>
	.drawing { width: 100%; }

	.turn-info {
		text-align: center;
		color: var(--text-muted);
		font-size: 0.85rem;
		margin-bottom: 0.5rem;
	}

	.word-display, .hint-display {
		text-align: center;
		margin-bottom: 1rem;
	}

	.word-label {
		color: var(--text-muted);
		display: block;
		font-size: 0.85rem;
	}

	.word {
		font-size: 1.8rem;
		font-weight: 700;
		color: var(--primary);
	}

	.hint {
		font-size: 1.5rem;
		font-weight: 700;
		letter-spacing: 0.3em;
		font-family: monospace;
	}

	.canvas-container {
		width: 100%;
		aspect-ratio: 1;
		border-radius: var(--radius);
		overflow: hidden;
		border: 2px solid #333;
	}

	canvas {
		width: 100%;
		height: 100%;
		display: block;
	}

	.toolbar {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-top: 0.75rem;
		flex-wrap: wrap;
	}

	.color-picker {
		display: flex;
		gap: 0.4rem;
	}

	.color-btn {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		border: 2px solid transparent;
		transition: border-color 0.2s;
	}

	.color-btn.active {
		border-color: white;
	}

	.brush-sizes {
		display: flex;
		gap: 0.4rem;
		align-items: center;
	}

	.size-btn {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: var(--bg-input);
		display: flex;
		align-items: center;
		justify-content: center;
		border: 2px solid transparent;
	}

	.size-btn.active {
		border-color: var(--primary);
	}

	.dot {
		display: block;
		border-radius: 50%;
		background: white;
	}

	.clear-btn {
		padding: 0.5rem 1rem;
		font-size: 0.85rem;
		width: auto;
	}

	.guess-form {
		display: flex;
		gap: 0.5rem;
		margin-top: 0.75rem;
	}

	.guess-form .input {
		flex: 1;
	}

	.guess-form .btn {
		width: auto;
		padding: 0.875rem 1.25rem;
	}

	.guessed-msg {
		text-align: center;
		padding: 1rem;
		color: var(--success);
		font-weight: 600;
		font-size: 1.1rem;
	}

	.chat-msg {
		margin-top: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: var(--bg-card);
		border-radius: var(--radius-sm);
		font-size: 0.9rem;
	}

	.reveal {
		text-align: center;
		padding: 2rem 0;
	}

	.revealed-word {
		font-size: 2rem;
		font-weight: 700;
		color: var(--success);
		margin-top: 0.5rem;
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
	.pts { font-weight: 700; color: var(--primary); }
</style>
