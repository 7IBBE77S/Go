export class Weapon {
    static Types = {
        PISTOL: 'pistol',
        SHOTGUN: 'shotgun',
        MACHINE_GUN: 'machine_gun'
    };
    constructor(type, owner) {
        this.type = type;
        this.owner = owner;
        this.lastFireTime = 0;
        // ACTUAL maximum fire rates that can't be exceeded
        this.properties = {
            [Weapon.Types.PISTOL]: {
                fireRate: 500, // max of ~3 shots per second no matter what
                requireClick: true
            },
            [Weapon.Types.SHOTGUN]: {
                fireRate: 1000, // maximum of 1 shot per second, period.
                requireClick: true
            },
            [Weapon.Types.MACHINE_GUN]: {
                fireRate: 100, // 10 shots per second when held
                requireClick: false
            }
        };
        // queue to track rapid inputs used to detect and limit spam
        this.inputQueue = [];
    }
    update(input) {
        const now = Date.now();
        // enforce absolute fire rate limits (not really working though as intended)
        if (now - this.lastFireTime < this.properties[this.type].fireRate) {
            return; // too soon, no matter what inputs happened
        }
        // for machine gun simple auto fire when held
        if (!this.properties[this.type].requireClick) {
            if (input.weaponInputState.isHolding) {
                this.shoot();
            }
            return;
        }
        // for pistol/shotgun
        if (input.weaponInputState.isNewPress) {
            // Clear old inputs from queue
            this.inputQueue = this.inputQueue.filter(time =>
                now - time < 1000 // only keep last second of inputs
            );
            // strict rate limiting based on weapon type
            if (this.type === Weapon.Types.SHOTGUN) {
                // for shotgun absolutely enforce 1 second between shots
                if (this.inputQueue.length === 0 || now - this.inputQueue[this.inputQueue.length - 1] >= 1000) {
                    this.shoot();
                    this.inputQueue.push(now);
                }
            } else if (this.type === Weapon.Types.PISTOL) {
                // Frr pistol, enforce maximum of 3 shots per second
                if (this.inputQueue.length < 3 || now - this.inputQueue[this.inputQueue.length - 3] >= 1000) {
                    this.shoot();
                    this.inputQueue.push(now);
                }
            }
        }
    }
    shoot() {
        if (!this.owner.network || !this.owner.network.getState().isInitialized) return;
        this.owner.network.sendShoot(this.owner.sprite.rotation);
        this.lastFireTime = Date.now();
    }
}