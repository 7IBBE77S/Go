// Manages all the other players in the game (not the local player).
// its like the stage manager for everyone else handling their sprites,
// health bars, bullets, weapons, and powerups, making sure they look right on screen.
export class OtherPlayers {
    constructor(container) {
        //stage where all the visual elements will live.
        this.container = container;
        // maps tack everything by id
        this.players = new Map();       // stores player sprites, keyed by their unique ID.
        this.bullets = new Map();       // keeps track of bullets zipping around.
        this.healthBars = new Map();    // health bars floating above players’ heads.
        this.weapons = new Map();       // weapons lying around waiting to be picked up.
        this.powerups = new Map();      // cool powerup items scattered on the map.
    }

    // helper to give each player a unique color based on their id. should be replaced with actual sprite graphics
    // No two players should look the same
    getPlayerColor(id) {
        // id string into a number with a simple hash nothing crazy just works.
        let hash = 0;
        for (let i = 0; i < id.length; i++) {
            hash = ((hash << 5) - hash) + id.charCodeAt(i);
            hash = hash & hash; // Keeps it an integer avoids weirdness.
        }

        // map hash to hue and create a bright color
        const hue = Math.abs(hash) % 360;
        const color = PIXI.utils.rgb2hex([
            Math.cos(hue * Math.PI / 180) * 0.5 + 0.5,         // red component.
            Math.cos((hue + 120) * Math.PI / 180) * 0.5 + 0.5, // Green offset by 120 degress.
            Math.cos((hue + 240) * Math.PI / 180) * 0.5 + 0.5  // Blue, offset by 240.
        ]);

        return color; // returns a hex color code for pixijs to use.
    }

    // updates the health bar keeps it in sync with their current health.
    // https://pixijs.com/8.x/guides/advanced/
    updateHealth(id, health) {
        console.log(`updating health for player ${id} to ${health}`);
        let healthBar = this.healthBars.get(id);
        if (!healthBar) {
            // no health bar yet
            const container = new PIXI.Container();

            const bg = new PIXI.Graphics();
            bg.beginFill(0x000000);
            bg.drawRect(-17, -25, 34, 6); // positioned above the player.
            container.addChild(bg);

            // the green foreground shows the actual health scales with the value.
            const fg = new PIXI.Graphics();
            fg.beginFill(0x00FF00);
            fg.drawRect(-16, -24, 32 * (health / 100), 4); 
            container.addChild(fg);

            // add it to the stage and store it for later.
            this.container.addChild(container);
            this.healthBars.set(id, container);
            healthBar = container;
        }

        // then update the green bar to reflect the new health value.
        const fg = healthBar.children[1]; 
        fg.clear();
        fg.beginFill(0x00FF00);
        fg.drawRect(-16, -24, 32 * (health / 100), 4);

        // move the health bar so it follows the player.
        const player = this.players.get(id);
        if (player) {
            healthBar.position.set(player.x, player.y);
        }

        // If they’re out of health, make them look ghostly to show theyre dead.
        const playerSprite = this.players.get(id);
        if (playerSprite) {
            playerSprite.alpha = health <= 0 ? 0.5 : 1; // 50% opacity if dead
        }
    }

    // updates a players position and rotation keeps them moving smoothly.
    updatePlayer(id, position, color) {
        let playerSprite = this.players.get(id);

        if (!playerSprite) {
            console.log(`Creating new sprite for player ${id}`);
            playerSprite = new PIXI.Graphics();
            playerSprite.beginFill(color);
            playerSprite.drawPolygon([
                -15, -15,
                15, 0,   
                -15, 15  
            ]);
            playerSprite.endFill();
            this.container.addChild(playerSprite);
            this.players.set(id, playerSprite);
        }

        // syncs the sprite with the servers position and rotation data.
        playerSprite.x = position.x;
        playerSprite.y = position.y;
        playerSprite.rotation = position.rotation;

        // and keep the health bar tagging along too.
        const healthBar = this.healthBars.get(id);
        if (healthBar) {
            healthBar.position.set(position.x, position.y);
        }

        console.log(`Updated player ${id} to position:`, position);
    }

    // removes a player from the game.
    removePlayer(id) {
        const playerSprite = this.players.get(id);
        if (playerSprite) {
            // Kick them off the stage and out of our records (client side).
            this.container.removeChild(playerSprite);
            this.players.delete(id);
            console.log(`Removed player ${id}`);
        }
    }

    updateBullet(id, position) {
        let bullet = this.bullets.get(id);
        if (!bullet) {
            bullet = new PIXI.Graphics();
            bullet.beginFill(0xFF0000);  // Yellow for visibility
            bullet.drawCircle(0, 0, 3);   // Smaller bullet
            bullet.endFill();
            this.container.addChild(bullet);
            this.bullets.set(id, bullet);
        }
        bullet.x = position.x;
        bullet.y = position.y;
    }

    // bullet’s done its job clean it up.
    removeBullet(id) {
        const bullet = this.bullets.get(id);
        if (bullet) {
            this.container.removeChild(bullet);
            this.bullets.delete(id);
        }
    }

    // drops a weapon on the map ready for someone to grab
    addWeapon(id, position, type) {
        console.log("Adding weapon:", { id, position, type });
        const container = new PIXI.Container();

        const sprite = new PIXI.Graphics();
        sprite.lineStyle(2, 0x000000); 
        switch (type) {
            case 'shotgun':
                sprite.beginFill(0x8B4513); 
                sprite.drawRect(-15, -5, 30, 10);
                break;
            case 'machine_gun':
                sprite.beginFill(0x4169E1);
                sprite.drawRect(-20, -4, 40, 8);
                break;
            default: // pistol
                sprite.beginFill(0x808080);
                sprite.drawRect(-10, -3, 20, 6);
        }
        sprite.endFill();

        // add a glowing effect—makes it pop on the screen!
        const glow = new PIXI.Graphics();
        glow.beginFill(0xFFFF00, 0.3); 
        glow.drawCircle(0, 0, 25);
        glow.endFill();

        container.addChild(glow);
        container.addChild(sprite);

        // Place it where the server says it should be.
        container.x = position.x;
        container.y = position.y;

        this.container.addChild(container);
        this.weapons.set(id, container);

        // Let’s make it float up and down—classic pickup animation!
        const startY = position.y;
        const animate = () => {
            const container = this.weapons.get(id);
            if (container) {
                container.y = startY + Math.sin(Date.now() / 500) * 5; // Sine wave magic.
                requestAnimationFrame(animate); // Keep it looping.
            }
        };
        animate();
    }

    // weapons been picked up so remove it from the scene.
    removeWeapon(id) {
        const weapon = this.weapons.get(id);
        if (weapon) {
            this.container.removeChild(weapon);
            this.weapons.delete(id);
        }
    }

    // brings a player back to life
    handlePlayerRespawn(id, position, health) {
        const playerSprite = this.players.get(id);
        if (playerSprite) {
            playerSprite.alpha = 1; // No more ghost mode.
            playerSprite.x = position.x; // New spawn spot.
            playerSprite.y = position.y;
        }
        this.updateHealth(id, health);
    }

    // marks a player as down
    handlePlayerDeath(id) {
        const playerSprite = this.players.get(id);
        if (playerSprite) {
            playerSprite.alpha = 0.5; 
            console.log(`Player ${id} marked as dead`);
        }
    }

    // drops a powerup on the map
    addPowerUp(id, position, type) {
        console.log("Adding powerup:", { id, position, type });
        const container = new PIXI.Container();
        const graphics = new PIXI.Graphics();

        // Different shapes and colors for each powerup type
        if (type === "teleportation") {
            graphics.beginFill(0xAA00FF); 
            graphics.drawCircle(0, 0, 12);
        } else if (type === "force_field") {
            graphics.lineStyle(3, 0x0000FF); 
            graphics.drawCircle(0, 0, 20);
        } else if (type === "health_regen") {
            graphics.beginFill(0x00FF00); 
            graphics.drawCircle(0, 0, 10);
        }
        graphics.endFill();
        container.addChild(graphics);
        container.x = position.x;
        container.y = position.y;
        this.container.addChild(container);
        this.powerups.set(id, container);
    }

    removePowerUp(id) {
        const pu = this.powerups.get(id);
        if (pu) {
            this.container.removeChild(pu);
            this.powerups.delete(id);
        }
    }
}