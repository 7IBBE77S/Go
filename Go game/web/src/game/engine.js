// the conductor that orchestrates everything client side
// It handles rendering, input, networking, and game state all in one place.
import { setupRenderer } from "/src/game/render.js";
import { setupInput } from "/src/game/input.js";
import { setupNetwork } from "/src/game/network.js";
import { Player } from "/src/game/entities/player.js";
import { OtherPlayers } from "/src/game/entities/otherPlayers.js";
import { HUD } from "/src/ui/hub.js";
// import { createLobby } from '/src/ui/lobby.js';

export class GameEngine {
	constructor() {
		this.app = setupRenderer();

		// camera container and add it to the stage.
		this.camera = new PIXI.Container();
		this.app.stage.addChild(this.camera);

		this.debugConnections = new PIXI.Graphics();
		this.camera.addChild(this.debugConnections);

		this.debugTexts = new PIXI.Container();
		this.camera.addChild(this.debugTexts);

		this.debugEnabled = false; // initially off

		window.addEventListener("keydown", (e) => {
			if (e.key.toLowerCase() === "p") {
				this.debugEnabled = !this.debugEnabled;
				console.log("Debug connections toggled:", this.debugEnabled);
			}
		});

		this.input = setupInput();
		this.network = setupNetwork();
		this.hud = new HUD(this.app, this.network);
		this.player = new Player(this.camera, this.network, this);
		this.otherPlayers = new OtherPlayers(this.camera);
		this.serverPosition = null;
		this.playerId = null;

		// Define the viewport boundaries and maximum size.
		this.viewport = {
			left: 0,
			top: 0,
			width: window.innerWidth,
			height: window.innerHeight,
			right: window.innerWidth,
			bottom: window.innerHeight
		};
		this.maxWidth = 1600;  // maximum width in pixels
		this.maxHeight = 1200; // maximum height in pixels
		this.app.renderer.resize(this.viewport.width, this.viewport.height);
		this.app.view.style.position = 'absolute';
		this.app.view.style.left = '0';
		this.app.view.style.top = '0';

		setInterval(() => {
			if (!this.network.isConnected()) {
				console.log("Network disconnected");
				// could add visual indicator here
			}
		}, 1000);





		this.network.onPlayerInit((id, color, position, health, weapon, isDead, teleportAvailable, forceFieldActive, healthRegenActive) => {
			console.log("Initializing player - ID:", id, "Dead:", isDead, "Position:", position);
			this.playerId = id;
			this.player.setColor(color);
			this.player.setPosition(position);
			this.player.updateHealth(health);
			this.player.setWeapon(weapon);
			this.player.teleportAvailable = teleportAvailable || false;
			this.player.forceFieldActive = forceFieldActive || false;
			this.player.redrawShield();
			this.player.healthRegenActive = healthRegenActive || false;
		
			if (teleportAvailable) this.hud.showPowerUp("teleportation");
			if (forceFieldActive) this.hud.showPowerUp("force_field");
			if (healthRegenActive) this.hud.showPowerUp("health_regen");
		
			if (isDead) {
				this.player.handleDeath(position);
				const deathTime = this.network.getState().deathTime;
				if (deathTime) {
					const respawnDelay = Math.max(0, 3000 - (Date.now() - deathTime * 1000));
					this.hud.showRespawnCountdown(respawnDelay / 1000, position);
				}
			}
		});

		// handles network messages for other players
		this.network.onPlayerPosition((id, position, color) => {
			if (id === this.playerId) {
				// Force our position to match server during collisions
				this.player.setPosition(position);
				// Save the position to maintain it through reloads
				sessionStorage.setItem("playerPosition", JSON.stringify(position));
			} else {
				this.otherPlayers.updatePlayer(id, position, color);
			}
		});

		this.network.onBulletUpdate((id, position) => {
			this.otherPlayers.updateBullet(id, position);
		});

		this.network.onBulletHit((bulletId) => {
			this.otherPlayers.removeBullet(bulletId);
		});

		this.network.onPlayerDisconnect((id) => {
			console.log("Player disconnected:", id);
			this.otherPlayers.removePlayer(id);
		});
		this.network.onHealthUpdate((id, health) => {
			console.log(`Received health update for player ${id}: ${health}`);
			if (id === this.playerId) {
				this.player.updateHealth(health);
			} else {
				this.otherPlayers.updateHealth(id, health);
			}
		});

		this.network.onWeaponSpawn((id, position, type) => {
			console.log("Weapon spawned:", { id, position, type });
			this.otherPlayers.addWeapon(id, position, type);
		});

		this.network.onWeaponPickup((playerID, weaponID, weaponType) => {
			console.log("Weapon picked up:", { playerID, weaponID, weaponType });
			if (playerID === this.playerId) {
				this.player.setWeapon(weaponType);
			}
			this.otherPlayers.removeWeapon(weaponID);
		});

		this.network.onPlayerRespawn((id, position, health, weapon) => {
			console.log("Player respawn:", id);

			if (!position || typeof position.x !== "number" || typeof position.y !== "number") {
				console.error("Invalid respawn position:", position);
				return;
			}

			if (typeof health !== "number" || health < 0 || health > 100) {
				console.error("Invalid health value:", health);
				return;
			}

			const effectiveWeapon = weapon || "pistol";
			if (!weapon) {
				console.warn("Weapon not specified during respawn. Defaulting to pistol.");
			}

			if (id === this.playerId) {
				this.player.handleRespawn(position);
				this.player.setWeapon(effectiveWeapon);
				this.player.updateHealth(health);
			} else {
				this.otherPlayers.handlePlayerRespawn(id, position, health);
			}
		});





		this.network.onPlayerDeath((id) => {
			console.log("Player death:", id);
			if (id === this.playerId) {
				this.player.handleDeath();
				this.hud.showRespawnCountdown(5);
			} else {
				this.otherPlayers.handlePlayerDeath(id); 
			}
		});


		this.network.onWeaponDespawn((id) => {
			console.log("Weapon despawn:", id);
			this.otherPlayers.removeWeapon(id);
		});

		this.network.onPowerupSpawn((id, position, type) => {
			console.log("Received powerup_spawn:", { id, position, type });
			this.otherPlayers.addPowerUp(id, position, type);
		});


		this.player.teleportAvailable = false;

		window.addEventListener("keydown", (e) => {
			if (e.key.toLowerCase() === "t") {
				if (this.player.teleportAvailable) {
					console.log("Activating teleport...");
					const cursorPos = { x: this.input.mouse.x, y: this.input.mouse.y };
					this.network.sendTeleport(cursorPos)
				
				}
			}
		});



		this.network.onPowerupPickup((playerID, powerupID, type) => {
			console.log("Powerup picked up:", { playerID, powerupID, type });
			// removes the powerup sprite from the world.
			this.otherPlayers.removePowerUp(powerupID);
			if (playerID === this.playerId) {
				this.player.teleportAvailable = false;
				this.player.forceFieldActive = false;
				this.player.healthRegenActive = false;
				this.hud.hidePowerUp("teleportation");
				this.hud.hidePowerUp("force_field");
				this.hud.hidePowerUp("health_regen");

				switch (type) {
					case "teleportation":
						this.player.teleportAvailable = true;
						this.hud.showPowerUp("teleportation");
						break;
					case "force_field":
						this.player.forceFieldActive = true;
						this.player.shield = 100;
						this.hud.showPowerUp("force_field");
						break;
					case "health_regen":
						this.player.healthRegenActive = true;
						this.hud.showPowerUp("health_regen");
						break;
				}
			}
		});


		this.network.onTeleport((playerID, newPosition) => {
			console.log("Teleport message received:", { playerID, newPosition });
			if (playerID === this.playerId) {
				this.player.setPosition(newPosition);
			} else {
				this.otherPlayers.updatePlayer(playerID, newPosition, this.otherPlayers.getPlayerColor(playerID));
			}
		});


		// handle window resizing
		window.addEventListener("resize", () => {
			// on a browser resize you might want to update only if the browser
			// window is larger than your current viewport.
			const newWidth = Math.max(window.innerWidth, this.viewport.right - this.viewport.left);
			const newHeight = Math.max(window.innerHeight, this.viewport.bottom - this.viewport.top);
			this.app.renderer.resize(newWidth, newHeight);
		});
	}

	//not working
	checkViewportExpansion() {
		const threshold = 50; // when player is within 50px of edge
		const expansionStep = 100; // Expands by 100px at a time
		let changed = false;

		// get player's position relative to current viewport size
		const playerX = this.player.sprite.x;
		const playerY = this.player.sprite.y;

		// check right edge
		if (playerX > this.viewport.width - threshold) {
			const newWidth = Math.min(this.viewport.width + expansionStep, this.maxWidth);
			if (newWidth !== this.viewport.width) {
				this.viewport.width = newWidth;
				changed = true;
			}
		}

		// check bottom edge
		if (playerY > this.viewport.height - threshold) {
			const newHeight = Math.min(this.viewport.height + expansionStep, this.maxHeight);
			if (newHeight !== this.viewport.height) {
				this.viewport.height = newHeight;
				changed = true;
			}
		}

		// If viewport changed then update renderer size
		if (changed) {
			// resize the renderer
			this.app.renderer.resize(this.viewport.width, this.viewport.height);

			console.log('Viewport expanded:', {
				width: this.viewport.width,
				height: this.viewport.height
			});
		}
	}



	// interpolates from green to red.
	getConnectionColor(dist, threshold) {
		const half = threshold / 2;
		let t = 0;
		if (dist <= half) {
			t = 0;
		} else if (dist >= threshold) {
			t = 1;
		} else {
			t = (dist - half) / half;
		}
		const r = Math.floor(255 * t);
		const g = Math.floor(255 * (1 - t));
		const b = 0;
		return { r, g, b };
	}

	//fun debug function
	// draw connections among objects and to the window edges.
	drawConnections() {
		const objectThreshold = 550; // max distance for connection between objects.
		const edgeMargin = 50;         // distance from window edge to draw an edge connection.

		// clear previous debug lines and texts.
		this.debugConnections.clear();
		this.debugTexts.removeChildren();

		// just an array of objects to connect.
		const objects = [];
		// the local player.
		objects.push({
			x: this.player.sprite.x,
			y: this.player.sprite.y,
			label: "player"
		});
		// other players.
		for (const [id, sprite] of this.otherPlayers.players) {
			objects.push({
				x: sprite.x,
				y: sprite.y,
				label: "other"
			});
		}
		// weapon pickups.
		for (const [id, container] of this.otherPlayers.weapons) {
			objects.push({
				x: container.x,
				y: container.y,
				label: "weapon"
			});
		}

		//for powerups
		for (const [id, container] of this.otherPlayers.powerups) {
			objects.push({
				x: container.x,
				y: container.y,
				label: "powerup"
			});
		}

		// draws connections among objects.
		for (let i = 0; i < objects.length; i++) {
			for (let j = i + 1; j < objects.length; j++) {
				const dx = objects[i].x - objects[j].x;
				const dy = objects[i].y - objects[j].y;
				const dist = Math.sqrt(dx * dx + dy * dy);
				if (dist < objectThreshold) {
					const color = this.getConnectionColor(dist, objectThreshold);
					const hexColor = (color.r << 16) | (color.g << 8) | color.b;
					this.debugConnections.lineStyle(2, hexColor, 1.0);
					this.debugConnections.moveTo(objects[i].x, objects[i].y);
					this.debugConnections.lineTo(objects[j].x, objects[j].y);
					// compute and draw midpoint coordinates.
					const midX = (objects[i].x + objects[j].x) / 2;
					const midY = (objects[i].y + objects[j].y) / 2;
					const coordText = new PIXI.Text(
						`(x: ${midX.toFixed(0)}, y: ${midY.toFixed(0)})`,
						new PIXI.TextStyle({
							fontFamily: "Arial",
							fontSize: 12,
							fill: hexColor,
							stroke: 0x000000,
							strokeThickness: 1,
						})
					);
					coordText.x = midX;
					coordText.y = midY;
					this.debugTexts.addChild(coordText);
				}
			}
		}

		// draw connections from objects to window edges.
		const winWidth = window.innerWidth;
		const winHeight = window.innerHeight;
		for (const obj of objects) {
			// left edge.
			if (obj.x < edgeMargin) {
				const edgeColor = { r: 10, g: 10, b: 235 }; 
				const hexEdgeColor = (edgeColor.r << 16) | (edgeColor.g << 8) | edgeColor.b;
				this.debugConnections.lineStyle(2, hexEdgeColor, 1.0);
				this.debugConnections.moveTo(obj.x, obj.y);
				this.debugConnections.lineTo(0, obj.y);
				const midX = (obj.x + 0) / 2;
				const midY = obj.y;
				const text = new PIXI.Text(
					`(${midX.toFixed(0)}, ${midY.toFixed(0)})`,
					new PIXI.TextStyle({
						fontFamily: "Arial",
						fontSize: 12,
						fill: hexEdgeColor,
						stroke: 0x000000,
						strokeThickness: 1,
					})
				);
				text.x = midX;
				text.y = midY;
				this.debugTexts.addChild(text);
			}
			// right edge.
			if (obj.x > winWidth - edgeMargin) {
				const edgeColor = { r: 10, g: 10, b: 235 };
				const hexEdgeColor = (edgeColor.r << 16) | (edgeColor.g << 8) | edgeColor.b;
				this.debugConnections.lineStyle(2, hexEdgeColor, 1.0);
				this.debugConnections.moveTo(obj.x, obj.y);
				this.debugConnections.lineTo(winWidth, obj.y);
				const midX = (obj.x + winWidth) / 2;
				const midY = obj.y;
				const text = new PIXI.Text(
					`(${midX.toFixed(0)}, ${midY.toFixed(0)})`,
					new PIXI.TextStyle({
						fontFamily: "Arial",
						fontSize: 12,
						fill: hexEdgeColor,
						stroke: 0x000000,
						strokeThickness: 1,
					})
				);
				text.x = midX;
				text.y = midY;
				this.debugTexts.addChild(text);
			}
			// top edge.
			if (obj.y < edgeMargin) {
				const edgeColor = { r: 10, g: 10, b: 235 };
				const hexEdgeColor = (edgeColor.r << 16) | (edgeColor.g << 8) | edgeColor.b;
				this.debugConnections.lineStyle(2, hexEdgeColor, 1.0);
				this.debugConnections.moveTo(obj.x, obj.y);
				this.debugConnections.lineTo(obj.x, 0);
				const midX = obj.x;
				const midY = obj.y / 2;
				const text = new PIXI.Text(
					`(${midX.toFixed(0)}, ${midY.toFixed(0)})`,
					new PIXI.TextStyle({
						fontFamily: "Arial",
						fontSize: 12,
						fill: hexEdgeColor,
						stroke: 0x000000,
						strokeThickness: 1,
					})
				);
				text.x = midX;
				text.y = midY;
				this.debugTexts.addChild(text);
			}
			// bottom edge.
			if (obj.y > winHeight - edgeMargin) {
				const edgeColor = { r: 10, g: 10, b: 235 };
				const hexEdgeColor = (edgeColor.r << 16) | (edgeColor.g << 8) | edgeColor.b;
				this.debugConnections.lineStyle(2, hexEdgeColor, 1.0);
				this.debugConnections.moveTo(obj.x, obj.y);
				this.debugConnections.lineTo(obj.x, winHeight);
				const midX = obj.x;
				const midY = (obj.y + winHeight) / 2;
				const text = new PIXI.Text(
					`(${midX.toFixed(0)}, ${midY.toFixed(0)})`,
					new PIXI.TextStyle({
						fontFamily: "Arial",
						fontSize: 12,
						fill: hexEdgeColor,
						stroke: 0x000000,
						strokeThickness: 1,
					})
				);
				text.x = midX;
				text.y = midY;
				this.debugTexts.addChild(text);
			}
		}
	}


	start() {
		this.app.ticker.add(() => {
			if (!this.player.isDead) {
				this.gameLoop();
			}
		});
	}

	gameLoop() {

		if (!this.player.isDead) {
			const previousPosition = {
				x: this.player.sprite.x,
				y: this.player.sprite.y,
				rotation: this.player.sprite.rotation
			};

			this.player.update(this.input, this.otherPlayers);
			this.input.weaponInputState.isNewPress = false;


			// only sends if the position changed
			if (previousPosition.x !== this.player.sprite.x ||
				previousPosition.y !== this.player.sprite.y ||
				previousPosition.rotation !== this.player.sprite.rotation) {

				this.network.sendPlayerState({
					x: this.player.sprite.x,
					y: this.player.sprite.y,
					rotation: this.player.sprite.rotation,
				});
			}
		}

		// this.checkViewportExpansion();
		if (this.debugEnabled) {
			this.drawConnections();
		} else {
			this.debugConnections.clear();
			this.debugTexts.removeChildren();
		}


	}
}

