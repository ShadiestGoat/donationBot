package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/shadiestgoat/log"
)

type confItem struct {
	Env string

	// Has to be *string, *int, *float64, *bool, *[]string. Slices are space separated!
	// Bool is case insensitive and supports the following keys:
	// TRUE -> true,
	Res any

	// If not nil, is used to set the value of Res (Res must be a pointer!)
	Parser func(inp string) any

	Default  any
	Required bool

	// Format: 'ITEM' is not set, so CONSEQUENCE
	Consequence string
}

var (
	GUILD_IDS              = []string{}
	TOKEN                  = ""
	CHANNELS_ANNOUNCEMENTS = []string{}
	CHANNELS_FUND          = []string{}
	ROLES_PERSISTENT       = []*DonationRole{}
	ROLES_MONTHLY          = []*DonationRole{}
	LOCATION               = ""
	DEBUG_MENTION          = ""
	DEBUG_WEBHOOK          = ""
)

func Init() {
	godotenv.Load(".env", "conf.env")

	var confMap = []*confItem{
		{
			Env: "DEBUG_MENTION",
			Res: &DEBUG_MENTION,
		},
		{
			Env:         "DEBUG_WEBHOOK",
			Res:         &DEBUG_WEBHOOK,
			Consequence: "it will not send debug messages to discord",
		},

		{
			Env: "GUILDS",
			Res: &GUILD_IDS,
			Required: true,
		},

		{
			Env: "LOCATION",
			Res: &LOCATION,
			Required: true,
		},
		{
			Env:      "TOKEN",
			Res:      &TOKEN,
			Required: true,
		},

		{
			Env:      "CHANNELS_ANNOUNCEMENTS",
			Res:      &CHANNELS_ANNOUNCEMENTS,
			Default: []string{},
			Consequence: "there will be no announcements on a new donation!",
		},
		{
			Env:      "CHANNELS_FUND",
			Res:      &CHANNELS_FUND,
			Default:  []string{},
			Consequence: "there will be no announcements of new funds!",
		},

		{
			Env:      "ROLES_PERSISTENT",
			Res:      &ROLES_PERSISTENT,
			Parser:   parseDonationRole,
			Default:  []*DonationRole{},
			Consequence: "there will be no roles given to donors that donated more than a month ago!",
		},
		{
			Env:      "ROLES_MONTHLY",
			Res:      &ROLES_MONTHLY,
			Parser:   parseDonationRole,
			Default:  []*DonationRole{},
			Consequence: "there will be no special roles given to monthly donors!",
		},
	}

	for _, opt := range confMap {
		handleKey(opt)
	}
}

func handleKey(opt *confItem) {
	item := os.Getenv(opt.Env)

	if item == "" {
		if opt.Required {
			log.Fatal("'%s' is a needed variable, but is not present! Please read the README.md file for more info.", opt.Env)
		}

		if opt.Default == nil {
			if opt.Consequence != "" {
				log.Warn("'%s' is not set, so %s!", opt.Env, opt.Consequence)
			}
			return
		}

		switch opt.Res.(type) {
		case *[]string:
			*(opt.Res.(*[]string)) = opt.Default.([]string)
		case *string:
			*(opt.Res.(*string)) = opt.Default.(string)
		case *bool:
			*(opt.Res.(*bool)) = opt.Default.(bool)
		case *float64:
			*(opt.Res.(*float64)) = opt.Default.(float64)
		case *int:
			*(opt.Res.(*int)) = opt.Default.(int)
		default:
			log.Fatal("Couldn't set the default value for '%s'", opt.Env)
		}

		return
	}

	switch resP := opt.Res.(type) {
	case *[]string:
		*resP = strings.Split(item, " ")
	case *string:
		*resP = item
	case *bool:
		v, err := strconv.ParseBool(item)
		if err != nil {
			log.Fatal("'%s' is not a valid bool value!", opt.Env)
		}

		*resP = v
	case *float64:
		v, err := strconv.ParseFloat(item, 64)
		if err != nil {
			log.Fatal("'%s' is not a valid float value!", opt.Env)
		}

		*resP = v
	case *int:
		v, err := strconv.Atoi(item)
		if err != nil {
			log.Fatal("'%s' is not a valid int value!", opt.Env)
		}

		*resP = v
	default:
		if opt.Parser == nil {
			log.Fatal("Unknown type for '%s'", opt.Env)
		}

		v := opt.Parser(item)
		*(resP.(*any)) = v
	}
}
