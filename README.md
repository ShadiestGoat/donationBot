# Donation Bot

This is a Discord Bot works with my [donation API](https://github.com/ShadiestGoat/donations).

## Config

This project can be configured through a `.env` file. For documentation on values, see the `template.env` file. Note, that some values don't have a default value. This is denoted through the option not being set in `template.env`, such as the `GUILDS` value not having an `=""` appended after it.

You can also add a `status.json` file in the cwd, and it will be used for the bot's status. [The data format should be like this](https://discord.com/developers/docs/topics/gateway-events#update-presence-gateway-presence-update-structure)

## Prerequisites 

You should already have the [Donation API](https://github.com/ShadiestGoat/donations) running, and you should already have created & loaded an application in the `auths.json` file for this bot. 

Note: The api MUST be running under `https`!

## Usage

1. Clone this repo
2. Configure
3. `go build`
4. `donationBot`
