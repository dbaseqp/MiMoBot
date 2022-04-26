package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
	AdminRole      = flag.String("admin", "", "Name of the role that can use these commands")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue = 1.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name: "new-eboard",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "Announce a new Eboard member",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "The name of the new eboard member",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "discord-handle",
					Description: "The Discord user of the new Eboard member",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "position",
					Description: "The position of the new eboard member",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "major",
					Description: "The major of the new eboard member",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "year",
					Description: "The year of the new eboard member",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link-to-picture",
					Description: "The picture of the new eboard member",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "announcement-channel",
					Description: "Channel to send embed",
					// Channel type mask
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildText,
					},
					Required: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "website",
					Description: "Website of new Eboard member",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "linkedin",
					Description: "LinkedIn of new Eboard member",
					Required:    false,
				},
			},
		},
		{
			Name: "announce-embed",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "Make an announcement as an embed",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "The title of the announcement",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "content",
					Description: "The content of the announcement",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "announcement-channel",
					Description: "Channel to send embed",
					// Channel type mask
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildText,
					},
					Required: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link-to-picture",
					Description: "The link to a picture for the announcement",
					Required:    false,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"new-eboard": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !HasAdminRole(s, i) {
				return
			}

			options := i.ApplicationCommandData().Options

			// Or convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			// This example stores the provided arguments in an []interface{}
			// which will be used to format the bot's response
			margs := make([]interface{}, 0, len(options))
			msgformat := "**About:**\n"

			// Get the value from the option map.
			// When the option exists, ok = true
			if opt, ok := optionMap["name"]; ok {
				// Option values must be type asserted from interface{}.
				// Discordgo provides utility functions to make this simple.
				margs = append(margs, opt.StringValue())
				msgformat += " •  Name: %s"
			}

			if opt, ok := optionMap["discord-handle"]; ok {
				margs = append(margs, opt.UserValue(nil).ID)
				msgformat += " | <@%s>\n"
			}

			if opt, ok := optionMap["position"]; ok {
				margs = append(margs, opt.StringValue())
				msgformat += " •  Position: %s\n"
			}

			if opt, ok := optionMap["major"]; ok {
				margs = append(margs, opt.StringValue())
				msgformat += " •  Major: %s\n"
			}

			if opt, ok := optionMap["year"]; ok {
				margs = append(margs, opt.StringValue())
				msgformat += " •  Year: %s\n"
			}

			picture := ""
			if opt, ok := optionMap["link-to-picture"]; ok {
				picture = opt.StringValue()
			} else {
				picture = "https://www.calpolyswift.org/assets/images/swift.png"
			}

			links := ""

			if opt, ok := optionMap["website"]; ok {
				margs = append(margs, opt.StringValue())
				links += " •  [Website](%s)\n"
			}

			if opt, ok := optionMap["linkedin"]; ok {
				margs = append(margs, opt.StringValue())
				links += " •  [LinkedIn](%s)\n"
			}

			if links != "" {
				msgformat += "\n**Links:**\n"
				msgformat += links
			}

			embed := []*discordgo.MessageEmbed{
				{
					Type:        "rich",
					Title:       "SWIFT • Meet the New Eboard",
					Description: fmt.Sprintf(msgformat, margs...),
					Color:       16755520,
					Image: &discordgo.MessageEmbedImage{
						URL:    picture,
						Height: 500,
						Width:  500,
					},
				},
			}

			if opt, ok := optionMap["announcement-channel"]; ok {
				log.Printf("Sending embed to channel %s...", opt.Name)
				s.ChannelMessageSendEmbeds(opt.ChannelValue(s).ID, embed)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: embed,
				},
			})
		},
		"announce-embed": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !HasAdminRole(s, i) {
				return
			}

			options := i.ApplicationCommandData().Options

			// Or convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			var embed []*discordgo.MessageEmbed
			if opt, ok := optionMap["link-to-picture"]; ok {
				embed = []*discordgo.MessageEmbed{
					{
						Type:        "rich",
						Title:       fmt.Sprintf("SWIFT • %s", optionMap["title"].StringValue()),
						Description: optionMap["content"].StringValue(),
						Color:       16755520,
						Image: &discordgo.MessageEmbedImage{
							URL:    opt.StringValue(),
							Height: 500,
							Width:  500,
						},
					},
				}
			} else {
				embed = []*discordgo.MessageEmbed{
					{
						Type:        "rich",
						Title:       fmt.Sprintf("SWIFT • %s", optionMap["title"].StringValue()),
						Description: strings.Replace(optionMap["content"].StringValue(), "\\n", "\n", -1),
						Color:       16755520,
					},
				}
			}

			if opt, ok := optionMap["announcement-channel"]; ok {
				log.Printf("Sending embed to channel %s...", opt.ChannelValue(s).ID)
				s.ChannelMessageSendEmbeds(opt.ChannelValue(s).ID, embed)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: embed,
				},
			})
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func FindRole(s *discordgo.Session, i *discordgo.InteractionCreate, rolename string) *discordgo.Role {
	var role *discordgo.Role
	channel, _ := s.Channel(i.ChannelID)
	guild, _ := s.Guild(channel.GuildID)
	for j := 0; j < len(guild.Roles); j++ {
		if guild.Roles[j].Name == rolename {
			role = guild.Roles[j]
		}
	}
	return role
}

func HasAdminRole(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	var admin *discordgo.Role
	admin = FindRole(s, i, *AdminRole)

	allow := false
	for _, role := range i.Member.Roles {
		if role == admin.ID {
			allow = true
		}
	}
	if !allow {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You don't have the permissions to do that!",
			},
		})
	}
	return allow
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")
		// // We need to fetch the commands, since deleting requires the command ID.
		// // We are doing this from the returned commands on line 375, because using
		// // this will delete all the commands, which might not be desirable, so we
		// // are deleting only the commands that we added.
		// registeredCommands, err := s.ApplicationCommands(s.State.User.ID, *GuildID)
		// if err != nil {
		// 	log.Fatalf("Could not fetch registered commands: %v", err)
		// }

		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
