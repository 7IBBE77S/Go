
## Basics

Phase 1: Basic setup/movement
[/] set up go server with websocket support
[/] basic html/pixijs frontend with a moveable player sprite
[/] establish websocket communication
[/] test basic player movement syncing

Phase 2: Multiplayer foundations
[/] add player sessions/connections management
[/] implement game state broadcasting
[/] show other players moving in real time
[/] add basic collision detection

Phase 3: Combat system
[/] implement shooting mechanics
[/] add bullet physics
[/] add simple damage system
[/] create basic weapon pickups

Phase 4: Game features
[/] add different weapon types
[/] implement power ups
[/] add player health/respawning
[/] create basic game modes (deathmatch, etc.)

Phase 5: Polish/infrastructure
[ ] add game rooms/lobbies
[ ] implement score tracking
[/] set up proper database
[ ] add basic matchmaking

Phase 6: Kubernetes deployment
[/] containerize the application
[/] set up database persistence
[/] configure networking/ingress
[/] deploy to your cluster


Core Gameplay:
- Players view the game from above (top down perspective)
- WASD movement mouse to aim and shoot
- Players start with a basic weapon
- Better weapons and power ups spawn on the map
- Players can carry two weapons and switch between them
- Getting hit reduces health, reaching 0 means respawn

Weapons Could Include:
- Pistol (basic starter)
- Shotgun (spread shot)
- Machine Gun (rapid fire)
- Rocket Launcher (splash damage)
- Laser Gun (continuous beam)

Power ups:
- Health packs
- Speed boost
- Shield
- Damage boost
- Rapid fire

Game Modes (starting with just the first)
- Deathmatch (every player for themselves)
- Team deathmatch (future)
- Capture the flag (future)
- King of the hill (future)

