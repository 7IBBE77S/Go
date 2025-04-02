export function setupRenderer() {
    const app = new PIXI.Application({
        width: window.innerWidth,
        height: window.innerHeight,
        backgroundColor: 0x444444,  
        resolution: window.devicePixelRatio || 1,
        autoDensity: true,
    });
    document.body.appendChild(app.view);

    // view takes up full screen
    app.renderer.view.style.position = 'absolute';
    app.renderer.view.style.display = 'block';
    app.view.style.top = '0';
    app.view.style.left = '0';

    // create the grid
    const grid = new PIXI.Graphics();
    const gridSize = 50; // Size of each grid square (adjustable)
    const gridColor = 0x666666; // Lighter grey for grid lines
    const lineThickness = 0.5; // Thickness of grid lines

    // vertical lines
    for (let x = 0; x <= window.innerWidth; x += gridSize) {
        grid.lineStyle(lineThickness, gridColor);
        grid.moveTo(x, 0);
        grid.lineTo(x, window.innerHeight);
    }

    // horizontal lines
    for (let y = 0; y <= window.innerHeight; y += gridSize) {
        grid.lineStyle(lineThickness, gridColor);
        grid.moveTo(0, y);
        grid.lineTo(window.innerWidth, y);
    }

    // add grid to the stage (background layer)
    app.stage.addChild(grid);

    // so grid stays behind other game elements
    app.stage.setChildIndex(grid, 0);

    return app;
}


// Simple weapon sprites for now
export function createWeaponSprite(type) {
    const sprite = new PIXI.Graphics();
    
    switch(type) {
        case 'shotgun':
            sprite.beginFill(0x8B4513);  // brown
            sprite.drawRect(-15, -5, 30, 10);
            break;
        default:
            sprite.beginFill(0x808080);  // Gray
            sprite.drawRect(-10, -3, 20, 6);
    }
    
    sprite.endFill();
    return sprite;
}

// function getTextureForWeapon(type) {
//     // Load appropriate texture based on weapon type
//     // You'll need to implement actual texture loading
//     return PIXI.Texture.WHITE; // Placeholder
// }