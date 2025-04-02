export class HUD {
    constructor(app, network) {
      this.app = app;
      this.network = network;
      this.container = new PIXI.Container();
      app.stage.addChild(this.container);
      // respawn countdown text (unchanged)
      this.respawnText = new PIXI.Text('', {
        fontFamily: 'Arial',
        fontSize: 48,
        fill: 0xFF0000,
        align: 'center',
        stroke: 0x000000,
        strokeThickness: 4,
        dropShadow: true,
        dropShadowColor: '#000000',
        dropShadowBlur: 4,
        dropShadowDistance: 6
      });
      this.respawnText.anchor.set(0.5);
      this.respawnText.x = app.screen.width / 2;
      this.respawnText.y = app.screen.height / 2;
      this.respawnText.visible = false;
      this.container.addChild(this.respawnText);
  
      // container for power up icons.
      this.powerUpIcons = {};
      window.addEventListener('resize', () => {
        this.respawnText.x = app.screen.width / 2;
        this.respawnText.y = app.screen.height / 2;
      });
    }
  
    showRespawnCountdown(seconds, position) {
        this.respawnText.visible = true;
        // string for the position if provided
        let posText = "";
        if (position && typeof position.x === 'number' && typeof position.y === 'number') {
            posText = `\nPosition: ${Math.round(position.x)},${Math.round(position.y)}`;
        }
        this.respawnText.text = `Respawning in ${seconds}${posText}`;

        let remainingSeconds = seconds;
        const update = () => {
            remainingSeconds -= 1;
            if (remainingSeconds <= 0) {
                this.respawnText.visible = false;
            } else {
                let posText = "";
                if (position && typeof position.x === 'number' && typeof position.y === 'number') {
                    posText = `\nPosition: ${Math.round(position.x)},${Math.round(position.y)}`;
                }
                this.respawnText.text = `Respawning in ${remainingSeconds}${posText}`;
                setTimeout(update, 1000);
            }
        };
        setTimeout(update, 1000);
    }

    showPowerUp(type) {
        if (!this.powerUpIcons[type]) {
          let icon;
          switch (type) {
            case "teleportation":
              icon = new PIXI.Text("Teleport (T)", { fontFamily: "Arial", fontSize: 24, fill: 0xAA00FF });
              break;
            case "force_field":
              icon = new PIXI.Text("Shield", { fontFamily: "Arial", fontSize: 24, fill: 0x0000FF });
              break;
            case "health_regen":
              icon = new PIXI.Text("Regen", { fontFamily: "Arial", fontSize: 24, fill: 0x00FF00 });
              break;
            default:
              icon = new PIXI.Text(type, { fontFamily: "Arial", fontSize: 24, fill: 0xFFFFFF });
          }
          // Position the icon
          icon.x = 20;
          icon.y = 20 + Object.keys(this.powerUpIcons).length * 30;
          this.container.addChild(icon);
          this.powerUpIcons[type] = icon;
        }
      }
    
      hidePowerUp(type) {
        if (this.powerUpIcons[type]) {
          this.container.removeChild(this.powerUpIcons[type]);
          delete this.powerUpIcons[type];
        }
      }
    }