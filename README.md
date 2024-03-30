# csgo-discord-rpc
Adds Discord RPC integration to CS2 and CS:GO

## Known bugs

It doesn't disables after CS:GO closes. I can't fix it. Use Bill2's Process Manager and fix it manually.

### Manually installation
First you need to build the executable:
```
$ go build -ldflags -H=windowsgui
```

*Optionally: Move the executable into the Window's startup folder*

Then you move the *gamestate_integration_discordrpc.cfg* file into the [Counter-Strike cfg folder](https://developer.valvesoftware.com/wiki/Counter-Strike:_Global_Offensive_Game_State_Integration#Locating_CS:GO_Install_Directory)
