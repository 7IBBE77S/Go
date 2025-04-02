import { Weapon } from './weapon.js';

export class Player {
    constructor(container, network, gameEngine) {
        this.container = container;
        this.network = network;
        this.gameEngine = gameEngine;
        this.shieldGraphic = new PIXI.Graphics();
        container.addChild(this.shieldGraphic);

        this.isDead = false;
        this.deathTime = 0;
        this.respawnDelay = 3000;

        this.currentWeapon = "pistol";
        this.isShooting = false;
        this.lastShot = 0;
        this.isIncognito = /Incognito/.test(navigator.userAgent);
        this.lastAutoFire = 0;

        this.weaponProperties = {
            pistol: { fireRate: 500, requireClick: true },
            shotgun: { fireRate: 1000, requireClick: true },
            machine_gun: { fireRate: 100, requireClick: false }
        };

        this.numPoints = 24;
        this.baseRadius = 15; 
        this.shapePoints = [];
        this.color = 0xff0000; 

        const seed = this.network.getState().playerId || Math.random().toString();
        Math.seedrandom(seed);
        for (let i = 0; i < this.numPoints; i++) {
            const theta = i * 2 * Math.PI / this.numPoints;
            const radiusVariation = (Math.random() - 0.5) * 2;
            const baseRadius = this.baseRadius + radiusVariation;
            const x = Math.cos(theta) * baseRadius;
            const y = Math.sin(theta) * baseRadius;
            this.shapePoints.push({ baseX: x, baseY: y, x, y });
        }

        this.sprite = new PIXI.Graphics();
        this.sprite.visible = false;
        this.drawShape();
        this.radius = 15; 
        container.addChild(this.sprite);

        this.sprite.x = window.innerWidth / 2;
        this.sprite.y = window.innerHeight / 2;

        this.healthBar = new PIXI.Container();
        this.healthBar.visible = false;
        const bg = new PIXI.Graphics();
        bg.beginFill(0x000000);
        bg.drawRect(-17, -25, 34, 6);
        this.healthBar.addChild(bg);

        const fg = new PIXI.Graphics();
        fg.beginFill(0x00FF00);
        fg.drawRect(-16, -24, 32, 4);
        this.healthBar.addChild(fg);

        container.addChild(this.healthBar);
        this.health = 100;

        this.weapon = new Weapon('pistol', this);

        const MAX_DEATH_AGE = 5000;
        const deathState = localStorage.getItem('playerDeathState');
        if (deathState) {
            const { deathTime } = JSON.parse(deathState);
            if (Date.now() - deathTime > MAX_DEATH_AGE) {
                localStorage.removeItem('playerDeathState');
            }
        }

        window.addEventListener('beforeunload', () => {
            if (this.deathTimer) {
                clearTimeout(this.deathTimer);
            }
            localStorage.removeItem('pendingRespawn');
        });

        const pendingRespawn = localStorage.getItem('pendingRespawn');
        if (pendingRespawn) {
            try {
                const { position, deathTime } = JSON.parse(pendingRespawn);
                const timeSinceDeath = Date.now() - deathTime;
                if (timeSinceDeath < 3000) {
                    this.handleDeath(position);
                    this.gameEngine.hud.showRespawnCountdown(
                        Math.ceil((3000 - timeSinceDeath) / 1000),
                        position
                    );
                }
            } catch (e) {
                console.error('Error restoring respawn state:', e);
            } finally {
                localStorage.removeItem('pendingRespawn');
            }
        }
    }

    generateOrganicShape() {
        const points = [];
        const numPoints = 12;
        const baseRadius = 15;
        const variation = 5;
        const seed = this.network.getState().playerId || Math.random().toString();
        const random = Math.seedrandom(seed, { global: false });
        for (let i = 0; i < numPoints; i++) {
            const angle = (i / numPoints) * 2 * Math.PI;
            const radius = baseRadius + (random() - 0.5) * variation;
            const x = Math.cos(angle) * radius;
            const y = Math.sin(angle) * radius;
            points.push({ x, y, baseX: x, baseY: y });
        }
        return points;
    }

    drawShape() {
        this.sprite.clear();
        this.sprite.beginFill(this.color);
        this.sprite.moveTo(this.shapePoints[0].x, this.shapePoints[0].y);
        for (let i = 1; i < this.numPoints; i++) {
            this.sprite.lineTo(this.shapePoints[i].x, this.shapePoints[i].y);
        }
        this.sprite.lineTo(this.shapePoints[0].x, this.shapePoints[0].y);
        this.sprite.endFill();
    }

    setPosition(position) {
        if (!position || typeof position.x !== 'number' || typeof position.y !== 'number') {
            console.error("Invalid position:", position);
            return;
        }
        this.sprite.visible = true;
        this.healthBar.visible = true;
        this.sprite.x = position.x;
        this.sprite.y = position.y;
        this.sprite.rotation = position.rotation;
    }

    tryShoot() {
        if (this.isDead) return;
        const now = Date.now();
        const weaponProps = this.weaponProperties[this.currentWeapon];
        if (weaponProps.requireClick) {
            this.network.sendShoot(this.sprite.rotation);
            this.lastShot = now;
        } else {
            if (now - this.lastShot >= weaponProps.fireRate) {
                this.network.sendShoot(this.sprite.rotation);
                this.lastShot = now;
            }
        }
    }

    checkCollision(otherPlayer) {
        const dx = this.sprite.x - otherPlayer.x;
        const dy = this.sprite.y - otherPlayer.y;
        const distance = Math.sqrt(dx * dx + dy * dy);
        return distance < this.radius + this.radius;
    }

    setColor(color) {
        this.color = color;
        this.drawShape();
    }

    handleShoot(e) {
        if (Date.now() - this.lastShot > this.fireRate) {
            this.network.sendShoot(this.sprite.rotation);
            this.lastShot = Date.now();
        }
    }

    redrawShield() {
        this.shieldGraphic.clear();
        if (this.forceFieldActive) {
            this.shieldGraphic.lineStyle(4, 0x0000FF, 0.7);
            this.shieldGraphic.drawCircle(0, 0, this.radius + 10);
            this.shieldGraphic.visible = true;
        } else {
            this.shieldGraphic.visible = false;
        }
    }

    updateHealth(health) {
        this.health = health;
        const fg = this.healthBar.children[1];
        fg.clear();
        fg.beginFill(0x00FF00);
        fg.drawRect(-16, -24, 32 * (health / 100), 4);
        if (health <= 0 && !this.isDead) {
            this.handleDeath();
        }
    }

    handleDeath(position) {
        if (this.isIncognito) {
            sessionStorage.setItem('tmpDeathState', JSON.stringify({
                position: this.network.getState().position,
                timestamp: Date.now()
            }));
        }
        this.network.getState().position = position;
        this.isDead = true;
        this.health = 0;
        this.disableInput = true;
        this.teleportAvailable = false;
        this.forceFieldActive = false;
        this.shield = 0;
        this.healthRegenActive = false;
        gsap.to(this.sprite, { alpha: 0.5, duration: 0.5 });
        this.gameEngine.hud.showRespawnCountdown(3, position);
        this.gameEngine.hud.hidePowerUp("teleportation");
        this.gameEngine.hud.hidePowerUp("force_field");
        this.gameEngine.hud.hidePowerUp("health_regen");
    }

    handleRespawn(position, health, weapon = 'pistol') {
        this.isDead = false;
        this.health = health;
        this.disableInput = false;
        this.currentWeapon = weapon;
        this.teleportAvailable = false;
        this.forceFieldActive = false;
        this.shield = 0;
        this.healthRegenActive = false;
        this.sprite.alpha = 1;
        this.setPosition(position);
        this.updateHealth(health);
        this.gameEngine.hud.hidePowerUp("teleportation");
        this.gameEngine.hud.hidePowerUp("force_field");
        this.gameEngine.hud.hidePowerUp("health_regen");
    }

    update(input, otherPlayers) {
        if (this.isDead || this.disableInput) {
            console.log('Skipping update - player is dead or input disabled');
            return;
        }

        const speed = 5;
        let newX = this.sprite.x;
        let newY = this.sprite.y;
        if (input.keys.w) newY -= speed;
        if (input.keys.s) newY += speed;
        if (input.keys.a) newX -= speed;
        if (input.keys.d) newX += speed;
        newX = Math.max(0, Math.min(newX, window.innerWidth));
        newY = Math.max(0, Math.min(newY, window.innerHeight));

        const dx = newX - this.sprite.x;
        const dy = newY - this.sprite.y;
        const velocityMagnitude = Math.sqrt(dx * dx + dy * dy);

        this.sprite.x = newX;
        this.sprite.y = newY;

        const dxMouse = input.mouse.x - this.sprite.x;
        const dyMouse = input.mouse.y - this.sprite.y;
        this.sprite.rotation = Math.atan2(dyMouse, dxMouse);

        this.network.sendPlayerState({
            x: this.sprite.x,
            y: this.sprite.y,
            rotation: this.sprite.rotation
        });

        this.handleWeaponInput(input.weaponInputState);
        this.healthBar.position.set(this.sprite.x, this.sprite.y);

        const time = Date.now() / 1000; 
        for (let i = 0; i < this.numPoints; i++) {
            const point = this.shapePoints[i];
            const theta = i * 2 * Math.PI / this.numPoints;
            const offset1 = 2 * Math.sin(time * 1 + theta * 2);
            const offset2 = 1 * Math.sin(time * 3 + theta * 4);
            const totalOffset = offset1 + offset2;
            point.x = point.baseX + totalOffset * (point.baseX / this.baseRadius);
            point.y = point.baseY + totalOffset * (point.baseY / this.baseRadius);
        }

        if (velocityMagnitude > 0) {
            const vx = dx / velocityMagnitude; 
            const vy = dy / velocityMagnitude;
            const stretchFactor = velocityMagnitude * 0.05; 
            for (let i = 0; i < this.numPoints; i++) {
                const point = this.shapePoints[i];

                const dot = (point.baseX / this.baseRadius) * vx + (point.baseY / this.baseRadius) * vy;
                point.x += dot * stretchFactor * vx;
                point.y += dot * stretchFactor * vy;
            }
        }

        this.drawShape();

        if (this.forceFieldActive) {
            this.shieldGraphic.clear();
            this.shieldGraphic.lineStyle(4, 0x0000FF, 0.7);
            this.shieldGraphic.drawCircle(this.sprite.x, this.sprite.y, this.radius + 10);
            this.shieldGraphic.visible = true;
        } else {
            if (this.shieldGraphic) {
                this.shieldGraphic.visible = false;
            }
        }

        this.redrawShield();

        if (!this.isDead && !this.disableInput) {
            this.weapon.update(input);
        }
    }


    setWeapon(type) {
        this.currentWeapon = type;
        this.weapon = new Weapon(type, this);
        console.log("Weapon changed to:", type);
    }

    handleWeaponInput(weaponInput) {
        const now = Date.now();
        const weaponProps = this.weaponProperties[this.currentWeapon];
        if (!weaponInput.isHolding) return;
        if (weaponProps.requireClick) {
            if (now - this.lastShot >= 50) {
                if (now - weaponInput.lastClickTime < 50) {
                    this.tryShoot();
                }
            }
        } else {
            if (now - this.lastShot >= weaponProps.fireRate) {
                this.tryShoot();
            }
        }
    }

    resetPosition() {
        const stored = sessionStorage.getItem("playerPosition");
        if (stored) {
            const pos = JSON.parse(stored);
            this.setPosition(pos);
        } else {
            this.sprite.x = window.innerWidth / 2;
            this.sprite.y = window.innerHeight / 2;
        }
    }

    shoot() {
        const bullet = {
            x: this.sprite.x,
            y: this.sprite.y,
            rotation: this.sprite.rotation,
        };
        this.network.sendShoot(bullet);
    }
}