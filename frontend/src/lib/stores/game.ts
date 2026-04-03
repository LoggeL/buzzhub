import { writable } from 'svelte/store';

export interface GameInfo {
	id: string;
	name: string;
	description: string;
	minPlayers: number;
	maxPlayers: number;
	icon: string;
}

export const currentGame = writable<GameInfo | null>(null);
export const gamePhase = writable<string>('');
export const gameData = writable<any>(null);
export const timer = writable<number>(0);

let timerInterval: ReturnType<typeof setInterval> | null = null;

export function startTimer(duration: number) {
	timer.set(duration);
	if (timerInterval) clearInterval(timerInterval);
	timerInterval = setInterval(() => {
		timer.update(t => {
			if (t <= 0) {
				if (timerInterval) clearInterval(timerInterval);
				return 0;
			}
			return t - 1;
		});
	}, 1000);
}

export function stopTimer() {
	if (timerInterval) clearInterval(timerInterval);
	timer.set(0);
}
