# BuzzHub 🔥

**Der wildeste Partyspiel-Server im ganzen Internet.**

Stell dir vor: Jackbox, aber auf Deutsch, selbst gehostet, und komplett unzensiert. BuzzHub ist die Plattform fuer Leute, die es leid sind, dass Partyspiele immer nur auf Englisch existieren und man 30 Euro fuer ein Spiel zahlt, das man dreimal spielt.

**Live:** [buzzhub.logge.top](https://buzzhub.logge.top)

---

## So funktioniert's

1. Ein Host erstellt eine Lobby (am besten auf dem grossen Bildschirm)
2. Alle anderen oeffnen `buzzhub.logge.top` auf ihrem Handy
3. Code eingeben, Namen waehlen, los geht's
4. Der Host waehlt ein Spiel und drueckt Start
5. Chaos.

---

## Die Spiele

### 🍑 Heisses Quiz
Klassisches Quizduell. Vier Antworten, eine richtig, alle druecken gleichzeitig. Wer falsch liegt, schaut dumm.

### 🍆 Abstimmung XXL
Kreative Fragen, noch kreativere Antworten. Alle bewerten sich gegenseitig. Demokratie in ihrer reinsten Form.

### 🥵 Bluff Master
Eine Frage, eine echte Antwort, und jede Menge erfundener Quatsch. Wer schluckt die falsche Antwort? Wer taeuscht die anderen?

### 🫦 Nacktes Zeichnen
Einer zeichnet, alle raten. Klingt einfach? Versuch mal "Bundeskanzler" mit dem Finger auf dem Handy zu malen.

### 💦 Woerter Suche
Ein Buchstabengitter voller versteckter Woerter. Aber hier wird nicht getippt — du draggst deinen Finger ueber die Buchstaben wie ein Wort-Ninja. Woerter koennen auch um die Ecke gehen (L-foermig). Wer zuerst findet, kriegt die meisten Punkte.

### 🕵️ Geheime Woerter
Codenames auf Deutsch. Zwei Teams, ein Spymaster pro Team, 25 Woerter auf dem Tisch. Der Spymaster gibt einen Hinweis — EIN Wort und eine Zahl. Das Team muss raten, welche Woerter zu ihnen gehoeren. Aber Vorsicht: ein falscher Klick auf die schwarze Karte und dein Team ist sofort raus.

---

## Tech Stack

| Was | Womit |
|-----|-------|
| Backend | Go + Socket.IO |
| Frontend | SvelteKit (SSG) |
| State | Redis |
| Hosting | Docker + Dokploy |
| CDN/DNS | Cloudflare |

**Architektur:**
```
Handy/Browser → Cloudflare → Traefik → Go-Server ←→ Redis
                                          ↕
                                    Socket.IO (Echtzeit)
```

Jedes Spiel ist ein eigenstaendiges Go-Package das sich per `init()` selbst registriert. Neues Spiel = neuer Ordner, fertig. Das Frontend ist eine SvelteKit SPA die sich per WebSocket mit dem Backend verbindet.

---

## Entwicklung

```bash
# Backend
cd backend && go run ./cmd/server

# Frontend
cd frontend && npm install && npm run dev

# Oder alles zusammen
docker compose up --build
```

---

## Eigene Spiele bauen

1. Erstelle `backend/internal/games/deinspiel/deinspiel.go`
2. Implementiere das `game.Game` Interface
3. Registriere dich in `init()` mit `game.Register("deinspiel", ...)`
4. Blanko-Import in `cmd/server/main.go`
5. Frontend-Komponente unter `frontend/src/lib/components/games/deinspiel/`
6. In die Component-Map und GAMES-Array eintragen
7. Fertig. Kein Framework, kein Boilerplate, kein Bullshit.

---

## Lizenz

Mach damit was du willst. Aber wenn du's irgendwo hostest, lad mich zur Party ein. 🍻
