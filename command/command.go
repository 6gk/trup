package command

import (
	"errors"
	"log"
	"net/url"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type Env struct {
	RoleMod         string
	ChannelShowcase string
}

type Context struct {
	Env     *Env
	Session *discordgo.Session
	Message *discordgo.Message
}

type Command struct {
	Exec          func(*Context, []string)
	Usage         string
	Help          string
	ModeratorOnly bool
}

var Commands = map[string]Command{
	"modping": {
		Exec:  modping,
		Usage: modpingUsage,
		Help:  "Pings online mods. Don't abuse.",
	},
	"fetch": {
		Exec:  fetch,
		Usage: fetchUsage,
	},
	"setfetch": {
		Exec: setFetch,
		Help: setFetchHelp,
	},
	"repo": {
		Exec: repo,
		Help: "Sends a link to the bot's repository.",
	},
	"move": {
		Exec: move,
		Help: moveUsage,
	},
	"git": {
		Exec:  git,
		Usage: gitUsage,
		Help:  "Adds a git link to your fetch",
	},
	"dotfiles": {
		Exec:  dotfiles,
		Usage: dotfilesUsage,
		Help:  "Adds a dotfiles link to your fetch",
	},
	"desc": {
		Exec:  desc,
		Usage: descUsage,
		Help:  "Sets or clears your description, displays it with fetch",
	},
	"note": moderatorOnly(Command{
		Exec:  note,
		Usage: noteUsage,
	}),
	"warn": moderatorOnly(Command{
		Exec:  warn,
		Usage: warnUsage,
	}),
}

var parseMentionRegexp = regexp.MustCompile(`<@!?(\d+)>`)

// parseMention takes a Discord mention string and returns the id
func parseMention(mention string) string {
	res := parseMentionRegexp.FindStringSubmatch(mention)
	if len(res) < 2 {
		return ""
	}
	return res[1]
}

var parseChannelMentionRegexp = regexp.MustCompile(`<#(\d+)>`)

func parseChannelMention(mention string) string {
	res := parseChannelMentionRegexp.FindStringSubmatch(mention)
	if len(res) < 2 {
		return ""
	}
	return res[1]
}

var userNotFound = errors.New("User not found")

func (ctx *Context) userFromString(str string) (*discordgo.Member, error) {
	if m := parseMention(str); m != "" {
		mem, err := ctx.Session.GuildMember(ctx.Message.GuildID, m)
		return mem, err
	}

	guild, err := ctx.Session.State.Guild(ctx.Message.GuildID)
	if err != nil {
		return nil, err
	}

	for _, m := range guild.Members {
		if str == m.User.Username || str == m.Nick {
			return m, nil
		}
	}

	return nil, userNotFound
}

func (ctx *Context) Reply(msg string) {
	_, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
	if err != nil {
		log.Printf("Failed to reply to message %s; Error: %s\n", ctx.Message.ID, err)
	}
}

func (ctx *Context) ReportError(msg string, err error) {
	log.Printf("Error Message ID: %s; ChannelID: %s; GuildID: %s; Author ID: %s; msg: %s; error: %s\n", ctx.Message.ID, ctx.Message.ChannelID, ctx.Message.GuildID, ctx.Message.Author.ID, msg, err)
	ctx.Reply(msg)
}

func moderatorOnly(cmd Command) Command {
	return Command{
		Exec: func(ctx *Context, args []string) {
			for _, r := range ctx.Message.Member.Roles {
				if r == ctx.Env.RoleMod {
					cmd.Exec(ctx, args)
					return
				}
			}

			ctx.Reply("this command is only for moderators.")
		},
		Usage:         cmd.Usage,
		ModeratorOnly: true,
	}
}

func (ctx *Context) isModerator() bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleMod {
			return true
		}
	}
	return false
}

func isValidUrl(toTest string) bool {
	u, err := url.Parse(toTest)
	return err == nil && u.Scheme != "" && u.Host != ""
}
