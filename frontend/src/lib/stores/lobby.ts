import { writable } from 'svelte/store';

export interface Player {
	id: string;
	name: string;
	connected: boolean;
	score: number;
}

export interface Lobby {
	code: string;
	hostId: string;
	gameId: string;
	status: string;
	maxPlayers: number;
	players: Player[];
}

export const lobby = writable<Lobby | null>(null);
export const playerId = writable<string>('');
export const error = writable<string>('');
