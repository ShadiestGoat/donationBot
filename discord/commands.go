package discord

import "github.com/bwmarrin/discordgo"

var commands = map[string]*discordgo.ApplicationCommand{}

var t = true

// These are points to false/true values, do not change <3
var DefaultFalse = new(bool)
var DefaultTrue = &t

// Register a command and it's handler
func RegisterCommand(cmd *discordgo.ApplicationCommand, handler HandlerCommand) {
	commands[cmd.Name] = cmd
	commandHandlers[cmd.Name] = handler
}
