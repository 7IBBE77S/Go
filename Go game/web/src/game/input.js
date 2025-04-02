export function setupInput() {
    const input = {
        keys: {},
        mouse: { x: 0, y: 0 },
        isShooting: false,
        weaponInputState: {
            isHolding: false,
            lastClickTime: 0,
            inputMethod: null,
            clicked: false,
            isNewPress: false
        }
    };
    let spacePressed = false;
    let mousePressed = false;
    // input for processing frame
    let lastFrameTime = performance.now();
    function processFrame(timestamp) {
        const dt = timestamp - lastFrameTime;
        lastFrameTime = timestamp;
        // clears new press flag after one frame
        input.weaponInputState.isNewPress = false;
        requestAnimationFrame(processFrame);
    }
    requestAnimationFrame(processFrame);
    // mouse/touch pad handlers
    window.addEventListener('mousedown', (e) => {
        mousePressed = true;
        input.weaponInputState.isHolding = true;
        input.weaponInputState.lastClickTime = Date.now();
        input.weaponInputState.inputMethod = 'mouse';
        input.weaponInputState.isNewPress = true;
        input.isShooting = true;
    });
    window.addEventListener('keydown', (e) => {
        const key = e.key.toLowerCase();
        input.keys[key] = true; // important for WASD movement
        if (key === ' ' && !spacePressed) {
            spacePressed = true;
            input.weaponInputState.isHolding = true;
            input.weaponInputState.lastClickTime = Date.now();
            input.weaponInputState.inputMethod = 'space';
            input.weaponInputState.clicked = true;
            input.weaponInputState.isNewPress = true;
            input.isShooting = true;
        }
    });
    window.addEventListener('keyup', (e) => {
        const key = e.key.toLowerCase();
        input.keys[key] = false; 
        if (key === ' ') {
            spacePressed = false;
            if (!mousePressed) {
                input.weaponInputState.isHolding = false;
                input.isShooting = false;
            }
            input.weaponInputState.clicked = false;
            input.weaponInputState.isNewPress = false;
        }
    });
    window.addEventListener('mouseup', (e) => {
        mousePressed = false;
        if (!spacePressed) {
            input.weaponInputState.isHolding = false;
            input.isShooting = false;
        }
        input.weaponInputState.clicked = false;
    });
    window.addEventListener('mousemove', (e) => {
        input.mouse.x = e.clientX;
        input.mouse.y = e.clientY;
    });
    window.addEventListener('touchstart', (e) => {
        input.weaponInputState.isHolding = true;
        input.weaponInputState.lastClickTime = Date.now();
        input.weaponInputState.inputMethod = 'touch';
        input.isShooting = true;
    });
    window.addEventListener('touchend', (e) => {
        if (!spacePressed) {
            input.weaponInputState.isHolding = false;
            input.isShooting = false;
        }
    });
    window.addEventListener('blur', () => {
        input.keys = {};
        input.isShooting = false;
        input.weaponInputState.isHolding = false;
        spacePressed = false;
        mousePressed = false;
    });
    return input;
}