export class Bullet {
    constructor(x, y, rotation, speed) {
        this.sprite = new PIXI.Graphics();
        this.sprite.beginFill(0xFFFFFF);
        this.sprite.drawCircle(0, 0, 3); // small circle for bullet
        this.sprite.endFill();
        this.sprite.x = x;
        this.sprite.y = y;
        this.sprite.rotation = rotation;
        this.speed = speed;
    }

    update() {
        this.sprite.x += Math.cos(this.sprite.rotation) * this.speed;
        this.sprite.y += Math.sin(this.sprite.rotation) * this.speed;
    }

    isOffScreen() {
        return (
            this.sprite.x < 0 ||
            this.sprite.x > window.innerWidth ||
            this.sprite.y < 0 ||
            this.sprite.y > window.innerHeight
        );
    }

    destroy() {
        this.app.stage.removeChild(this.sprite);
    }
}