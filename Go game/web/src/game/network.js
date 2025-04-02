export function setupNetwork() {
	const sessionId = (() => {
	  let id = localStorage.getItem("sessionId");
	  if (!id) {
		id = `session-${crypto.randomUUID()}`;
		localStorage.setItem("sessionId", id);
	  }
	  return id;
	})();
  
	if (!window.location.search.includes("sessionId")) {
	  const newUrl = new URL(window.location);
	  newUrl.searchParams.set("sessionId", sessionId);
	  window.history.replaceState({}, "", newUrl);
	}
  
	const storedDeathState = localStorage.getItem("playerDeathState");
	const initialDeathState = storedDeathState
	  ? JSON.parse(storedDeathState)
	  : {
		  isDead: false,
		  deathTime: null,
		  position: null,
		};
	let ws = null;
	let baseReconnectDelay = 1000;
	const MAX_DELAY = 30000;
	let reconnectAttempt = 0;
	const MAX_RECONNECT_ATTEMPTS = 5;
  
	const gameState = {
	  playerId: null,
	  health: 100,
	  weapon: "pistol",
	  isDead: initialDeathState.isDead,
	  position: initialDeathState.position,
	  deathTime: initialDeathState.deathTime,
	  matchActive: false,
	};
	let isInitialized = false;
	const callbacks = {
	  playerInit: null,
	  playerPosition: null,
	  playerDeath: null,
	  playerRespawn: null,
	  healthUpdate: null,
	  weaponPickup: null,
	  bulletUpdate: null,
	  bulletHit: null,
	  weaponSpawn: null,
	  weaponDespawn: null,
	  powerupSpawn: null,
	  powerupPickup: null,
	  playerDisconnect: null,
	  lobbyUpdate: null,
	  matchStarted: null,
	};
  
	// async function detectIncognito() {
	//   try {
	// 	const fs = await navigator.storage.estimate();
	// 	return fs.quota < 120000000;
	//   } catch {
	// 	return false;
	//   }
	// }
	// main network 
	function connect() {
	  if (ws?.readyState === WebSocket.OPEN) return;
	  const delay = Math.min(baseReconnectDelay * 2 ** reconnectAttempt, MAX_DELAY);
	  baseReconnectDelay *= 1.3;
	  setTimeout(() => {
		ws = new WebSocket("ws://localhost:8080/ws");
		ws.onopen = () => {
		  console.log("Connected to server with session:", sessionId);
		  reconnectAttempt = 0;
		  ws.send(
			JSON.stringify({
			  type: "session_init",
			  session_id: sessionId,
			})
		  );
		};
		ws.onmessage = (event) => {
		  try {
			if (!event.data || event.data === "") {
			  console.log("Empty message received, skipping");
			  return;
			}
			let message;
			try {
			  message = JSON.parse(event.data);
			} catch (e) {
			  console.warn("Invalid JSON received:", event.data);
			  return;
			}
			if (!message || !message.type) {
			  console.warn("Invalid message format:", message);
			  return;
			}
			switch (message.type) {
			  case "player_init":
				gameState.playerId = message.player_id;
				gameState.health = message.health;
				gameState.weapon = message.weapon;
				gameState.isDead = message.isDead;
				gameState.position = message.position;
				if (message.position) {
				  gameState.position = message.position;
				  isInitialized = true;
				  if (callbacks.playerInit) {
					callbacks.playerInit(
					  message.player_id,
					  message.color,
					  message.position || { x: 0, y: 0 },
					  message.health,
					  message.weapon,
					  message.isDead,
					  message.teleportAvailable,
					  message.forceFieldActive,
					  message.healthRegenActive
					);
				  }
				}
				if (gameState.isDead && callbacks.playerDeath) {
				  callbacks.playerDeath(message.player_id, message.position);
				}
				if (message.isDead) {
				  localStorage.setItem(
					"playerDeathState",
					JSON.stringify({
					  isDead: true,
					  deathTime: message.deathTime,
					  position: message.position,
					})
				  );
				} else {
				  localStorage.removeItem("playerDeathState");
				}
				break;
			  case "position_update":
				if (callbacks.playerPosition) {
				  callbacks.playerPosition(
					message.player_id,
					message.position,
					message.color
				  );
				}
				break;
			  case "bullet_hit":
				if (callbacks.bulletHit) {
				  callbacks.bulletHit(message.bullet_id);
				}
				break;
			  case "bullet_update":
				if (callbacks.bulletUpdate) {
				  callbacks.bulletUpdate(message.bullet_id, message.position);
				}
				break;
			  case "health_update":
				if (message.player_id === gameState.playerId) {
				  gameState.health = message.health;
				}
				if (callbacks.healthUpdate) {
				  callbacks.healthUpdate(message.player_id, message.health);
				}
				break;
			  case "weapon_spawn":
				if (callbacks.weaponSpawn) {
				  callbacks.weaponSpawn(
					message.weapon_id,
					message.position,
					message.weapon
				  );
				}
				break;
			  case "powerup_spawn":
				if (callbacks.powerupSpawn) {
				  console.log("Received powerup_spawn:", message);
				  callbacks.powerupSpawn(
					message.powerup_id,
					message.position,
					message.powerup
				  );
				}
				break;
			  case "powerup_pickup":
				if (callbacks.powerupPickup) {
				  callbacks.powerupPickup(
					message.player_id,
					message.powerup_id,
					message.powerup
				  );
				}
				break;
			  case "teleport":
				if (callbacks.teleport) {
				  callbacks.teleport(message.player_id, message.position);
				}
				break;
			  case "weapon_pickup":
				if (message.player_id === gameState.playerId) {
				  gameState.weapon = message.weapon;
				}
				if (callbacks.weaponPickup) {
				  callbacks.weaponPickup(
					message.player_id,
					message.weapon_id,
					message.weapon
				  );
				}
				break;
			  case "player_death":
				if (message.player_id === gameState.playerId) {
				  gameState.isDead = true;
				  gameState.health = 0;
				  gameState.deathTime = message.deathTime;
				  gameState.position = message.position;
				  localStorage.setItem(
					"playerDeathState",
					JSON.stringify({
					  isDead: true,
					  deathTime: message.deathTime,
					})
				  );
				}
				if (callbacks.playerDeath) {
				  callbacks.playerDeath(message.player_id, message.position);
				}
				break;
			  case "player_respawn":
				if (message.player_id === gameState.playerId) {
				  gameState.isDead = false;
				  gameState.health = message.health;
				  gameState.weapon = message.weapon;
				  gameState.deathTime = null;
				  gameState.position = message.position;
				  localStorage.removeItem("playerDeathState");
				}
				if (callbacks.playerRespawn) {
				  callbacks.playerRespawn(
					message.player_id,
					message.position,
					message.health,
					message.weapon
				  );
				}
				break;
			  case "weapon_despawn":
				if (callbacks.weaponDespawn) {
				  callbacks.weaponDespawn(message.weapon_id);
				}
				break;
			  case "player_disconnect":
				if (callbacks.playerDisconnect) {
				  callbacks.playerDisconnect(message.player_id);
				}
				break;
			  case "lobby_update":
				if (callbacks.lobbyUpdate) {
				  callbacks.lobbyUpdate(message.players);
				}
				break;
			  case "match_started":
				gameState.matchActive = true;
				if (callbacks.matchStarted) {
				  callbacks.matchStarted();
				}
				break;
			  default:
				console.warn("Unhandled message type:", message.type);
			}
		  } catch (error) {
			console.error("Error handling message:", error);
			console.log("Raw message:", event.data);
		  }
		};
  
		ws.onclose = () => {
		  if (reconnectAttempt < MAX_RECONNECT_ATTEMPTS) {
			reconnectAttempt++;
			connect();
		  }
		};
		ws.onerror = (error) => {
		  console.error("WebSocket error:", error);
		};
	  }, delay);
	}
  
	connect();
  
	return {
	  sendPlayerState: (state) => {
		if (ws?.readyState === WebSocket.OPEN && !gameState.isDead && isInitialized) {
		  ws.send(
			JSON.stringify({
			  type: "position",
			  position: state,
			})
		  );
		}
	  },
	  sendShoot: (rotation) => {
		if (ws?.readyState === WebSocket.OPEN && !gameState.isDead && isInitialized) {
		  ws.send(
			JSON.stringify({
			  type: "shoot",
			  rotation: rotation,
			})
		  );
		}
	  },
	  forceReconnect: () => {
		if (ws) {
		  ws.close();
		  reconnectAttempt = 0;
		  connect();
		}
	  },
	  sendTeleport: (cursorPos) => {
		if (ws?.readyState === WebSocket.OPEN && !gameState.isDead && isInitialized) {
		  ws.send(
			JSON.stringify({
			  type: "teleport",
			  cursorPos: cursorPos,
			})
		  );
		}
	  },
	  sendJoinMatch: () => {
		if (ws?.readyState === WebSocket.OPEN && !gameState.matchActive && isInitialized) {
		  ws.send(JSON.stringify({ type: "join_match" }));
		}
	  },
	  sendStartMatch: () => {
		if (ws?.readyState === WebSocket.OPEN && !gameState.matchActive && isInitialized) {
		  ws.send(JSON.stringify({ type: "start_match" }));
		}
	  },
	  onPlayerInit: (cb) => {
		callbacks.playerInit = cb;
	  },
	  onPlayerDeath: (cb) => {
		callbacks.playerDeath = cb;
	  },
	  onPlayerRespawn: (cb) => {
		callbacks.playerRespawn = cb;
	  },
	  onHealthUpdate: (cb) => {
		callbacks.healthUpdate = cb;
	  },
	  onWeaponPickup: (cb) => {
		callbacks.weaponPickup = cb;
	  },
	  onBulletUpdate: (cb) => {
		callbacks.bulletUpdate = cb;
	  },
	  onBulletHit: (cb) => {
		callbacks.bulletHit = cb;
	  },
	  onWeaponSpawn: (cb) => {
		callbacks.weaponSpawn = cb;
	  },
	  onWeaponDespawn: (cb) => {
		callbacks.weaponDespawn = cb;
	  },
	  onPlayerDisconnect: (cb) => {
		callbacks.playerDisconnect = cb;
	  },
	  onPlayerPosition: (cb) => {
		callbacks.playerPosition = cb;
	  },
	  onPowerupSpawn: (cb) => {
		callbacks.powerupSpawn = cb;
	  },
	  onPowerupPickup: (cb) => {
		callbacks.powerupPickup = cb;
	  },
	  onTeleport: (cb) => {
		callbacks.teleport = cb;
	  },
	  onLobbyUpdate: (cb) => {
		callbacks.lobbyUpdate = cb;
	  },
	  onMatchStarted: (cb) => {
		callbacks.matchStarted = cb;
	  },
	  isConnected: () => ws?.readyState === WebSocket.OPEN,
	  getState: () => ({ ...gameState }),
	};
  }
  