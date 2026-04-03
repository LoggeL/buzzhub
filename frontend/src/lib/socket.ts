import { io, type Socket } from 'socket.io-client';
import { browser } from '$app/environment';

let socket: Socket | null = null;

export function getSocket(): Socket {
	if (!browser) throw new Error('Socket only available in browser');
	if (!socket) {
		socket = io({
			autoConnect: true,
			reconnection: true,
			reconnectionDelay: 1000,
			reconnectionAttempts: 20,
		});

		socket.on('connect', () => {
			console.log('Connected to server');
			// Auto-rejoin if we have a token
			const token = localStorage.getItem('buzzhub_token');
			if (token) {
				socket!.emit('lobby:rejoin', { token });
			}
		});
	}
	return socket;
}

export function saveSession(token: string, playerId: string, lobbyCode: string) {
	localStorage.setItem('buzzhub_token', token);
	localStorage.setItem('buzzhub_playerId', playerId);
	localStorage.setItem('buzzhub_lobby', lobbyCode);
}

export function clearSession() {
	localStorage.removeItem('buzzhub_token');
	localStorage.removeItem('buzzhub_playerId');
	localStorage.removeItem('buzzhub_lobby');
}

export function getPlayerId(): string | null {
	return browser ? localStorage.getItem('buzzhub_playerId') : null;
}

export function getToken(): string | null {
	return browser ? localStorage.getItem('buzzhub_token') : null;
}
