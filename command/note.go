package command

import (
	"fmt"
	"strings"
	"trup/db"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

const noteUsage = "note <@user> [text]"

func note(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("not enough arguments.")
		return
	}

	var (
		about = parseMention(args[1])
	)

	if len(args) > 2 {
		content := strings.Join(args[2:], " ")
		note := db.NewNote(ctx.Message.Author.ID, about, content)

		err := note.Save()
		if err != nil {
			ctx.ReportError(fmt.Sprintf("Failed to save note %#v", note), err)
			return
		}

		ctx.Reply("noted.")
		return
	}

	notes, err := db.GetNotes(about)
	if err != nil {
		ctx.Reply("failed to retrieve notes. Error: " + err.Error())
		return
	}

	aboutMember, err := ctx.Session.GuildMember(ctx.Message.GuildID, about)
	if err != nil {
		ctx.ReportError("Failed to fetch member "+about, err)
		return
	}

	embed := discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Notes for %s#%s(%s)", aboutMember.User.Username, aboutMember.User.Discriminator, aboutMember.User.ID),
		Description: "",
		Color:       0,
		Fields:      make([]*discordgo.MessageEmbedField, 0, len(notes)),
	}

	for _, n := range notes {
		takerMember, err := ctx.Session.GuildMember(ctx.Message.GuildID, n.Taker)
		if err != nil {
			ctx.ReportError("Failed to fetch member "+n.Taker, err)
			return
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Moderator: %s#%s(%s)", takerMember.User.Username, takerMember.User.Discriminator, takerMember.User.ID),
			Value:  n.Content + " - " + humanize.Time(n.CreateDate),
			Inline: false,
		})
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed)
}
