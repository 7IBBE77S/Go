import { Weapon } from './weapon.js';

export class Player {
    constructor(container, network, gameEngine) {
        // core dependencies
        this.container = container; // pixi container for rendering
        this.network = network; // server connection
        this.gameEngine = gameEngine; // and reference to the main game engine
        
        this.shieldGraphic = new PIXI.Graphics();
        container.addChild(this.shieldGraphic);

        // death handling
        this.isDead = false;
        this.deathTime = 0;
        this.respawnDelay = 3000; // 3 seconds to get back synced with server

        // weapon setup 
        this.currentWeapon = "pistol";
        this.isShooting = false;
        this.lastShot = 0;
        
        this.isIncognito = /Incognito/.test(navigator.userAgent);
        this.lastAutoFire = 0;

        this.weaponProperties = {
            pistol: {
                fireRate: 500,    
                requireClick: true 
            },
            shotgun: {
                fireRate: 1000,  
                requireClick: true
            },
            machine_gun: {
                fireRate: 100,   
                requireClick: false 
            }
        };

        // creating our player sprite 
        this.sprite = new PIXI.Graphics();
        this.sprite.visible = false; // hide it until we get a position
        this.sprite.beginFill(0xff0000);
        this.setColor(0xff0000);
        // placeholder triangle creation
        this.sprite.drawPolygon([
            -15, -15, // top left
            15, 0,    // right tip
            -15, 15   // bottom left
        ]);
        this.sprite.endFill();
        this.radius = 15; // collision radius based on triangle size
        container.addChild(this.sprite);

        // puts the player in the center of the screen to start
        this.sprite.x = window.innerWidth / 2;
        this.sprite.y = window.innerHeight / 2;

        // health bar setup
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

        // death state persistence don't want to lose everything on refresh
        const MAX_DEATH_AGE = 5000;
        const deathState = localStorage.getItem('playerDeathState');
        if (deathState) {
            const { deathTime } = JSON.parse(deathState);
            if (Date.now() - deathTime > MAX_DEATH_AGE) {
                localStorage.removeItem('playerDeathState'); // clean up the old deaths
            }
        }

        // clean up on page exit
        window.addEventListener('beforeunload', () => {
            if (this.deathTimer) clearTimeout(this.deathTimer);
            localStorage.removeItem('pendingRespawn');
        });

        // restores any of the pending respawn states
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

        // weapon system init
        this.weapon = new Weapon('pistol', this);
    }

    // position setter w/ some safety checks
    setPosition(position) {
        if (!position || typeof position.x !== 'number' || typeof position.y !== 'number') {
            console.error("invalid position:", position);
            return;
        }
        this.sprite.visible = true;
        this.healthBar.visible = true;
        this.sprite.x = position.x;
        this.sprite.y = position.y;
        this.sprite.rotation = position.rotation;
    }

 
    tryShoot() {
        if (this.isDead) return; // no ghost shooting allowed
        
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

    // Simple collision detection - are we bumping into someone?
    checkCollision(otherPlayer) {
        const dx = this.sprite.x - otherPlayer.x;
        const dy = this.sprite.y - otherPlayer.y;
        const distance = Math.sqrt(dx * dx + dy * dy);
        return distance < this.radius + this.radius;
    }

    // Color changer for that personal touch
    setColor(color) {
        this.sprite.clear();
        this.sprite.beginFill(color);
        this.sprite.drawPolygon([-15, -15, 15, 0, -15, 15]);
        this.sprite.endFill();
    }

    // Shoot handler with basic rate limiting
    handleShoot(e) {
        if (Date.now() - this.lastShot > this.fireRate) {
            this.network.sendShoot(this.sprite.rotation);
            this.lastShot = Date.now();
        }
    }

    // Redraw our fancy shield effect
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

    // Health management - keep that bar updated!
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

    // Death sequence - time to take a breather
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

        // Reset all those cool power-ups
        this.teleportAvailable = false;
        this.forceFieldActive = false;
        this.shield = 0;
        this.healthRegenActive = false;
        
        gsap.to(this.sprite, { alpha: 0.5, duration: 0.5 });
        this.gameEngine.hud.showRespawnCountdown(3, position);
        
        // Hide HUD icons
        this.gameEngine.hud.hidePowerUp("teleportation");
        this.gameEngine.hud.hidePowerUp("force_field");
        this.gameEngine.hud.hidePowerUp("health_regen");
    }

    // Back in action with a fresh start
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
        
        // Reset HUD
        this.gameEngine.hud.hidePowerUp("teleportation");
        this.gameEngine.hud.hidePowerUp("force_field");
        this.gameEngine.hud.hidePowerUp("health_regen");
    }

    // Main update loop - where the magic happens
    update(input, otherPlayers) {
        if (this.isDead || this.disableInput) {
            console.log('Skipping update - player is dead or input disabled');
            return;
        }

        // Movement with boundary checking
        const speed = 5;
        let newX = this.sprite.x;
        let newY = this.sprite.y;
        if (input.keys.w) newY -= speed;
        if (input.keys.s) newY += speed;
        if (input.keys.a) newX -= speed;
        if (input.keys.d) newX += speed;
        newX = Math.max(0, Math.min(newX, window.innerWidth));
        newY = Math.max(0, Math.min(newY, window.innerHeight));
        this.sprite.x = newX;
        this.sprite.y = newY;

        // Aim towards the mouse
        const dx = input.mouse.x - this.sprite.x;
        const dy = input.mouse.y - this.sprite.y;
        this.sprite.rotation = Math.atan2(dy, dx);

        // Keep the server in sync
        this.network.sendPlayerState({
            x: this.sprite.x,
            y: this.sprite.y,
            rotation: this.sprite.rotation
        });

        this.handleWeaponInput(input.weaponInputState);
        this.healthBar.position.set(this.sprite.x, this.sprite.y);
        this.redrawShield();

        // Weapon updates when we're alive and kicking
        if (!this.isDead && !this.disableInput) {
            this.weapon.update(input);
        }
    }

    // Switch weapons like a pro
    setWeapon(type) {
        this.currentWeapon = type;
        this.weapon = new Weapon(type, this);
        console.log("Weapon changed to:", type);
    }

    // Smart weapon input handling
    handleWeaponInput(weaponInput) {
        const now = Date.now();
        const weaponProps = this.weaponProperties[this.currentWeapon];

        if (!weaponInput.isHolding) return;

        if (weaponProps.requireClick) {
            if (now - this.lastShot >= 50 && 
                now - weaponInput.lastClickTime < 50) {
                this.tryShoot();
            }
        } else {
            if (now - this.lastShot >= weaponProps.fireRate) {
                this.tryShoot();
            }
        }
    }

    // Quick position reset with persistence
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

    // Fire away!
    shoot() {
        const bullet = {
            x: this.sprite.x,
            y: this.sprite.y,
            rotation: this.sprite.rotation,
        };
        this.network.sendShoot(bullet);
    }
}