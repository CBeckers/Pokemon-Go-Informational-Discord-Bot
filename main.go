// main.go
// Author: Cade Beckers
// Written: 08/23/2024
// Updated: 09/10/2024

package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
)

var (
	GuildID = "" // add your GuildID here (server id)
	BotToken = "" // add your BotToken here
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "best",
			Description: "Gives you the best pokemon for specified category.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "search_setting",
					Description: "Search method",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Same Type Moves",
							Value: "sametype",
						},
						{
							Name:  "Mixed Type Moves",
							Value: "mixtype",
						},
						{
							Name:  "Name",
							Value: "name",
						},
					},
				},
				{
					Name:        "sort_by",
					Description: "Sorting method",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Damage per Second",
							Value: "dps",
						},
						{
							Name:  "Total Damage Output",
							Value: "tdo",
						},
						{
							Name:  "Effectiveness Rating",
							Value: "er",
						},
					},
				},
				{
					Name:        "number_of_results",
					Description: "Top # of results needed",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Top 10",
							Value: "10",
						},
						{
							Name:  "Top 5",
							Value: "5",
						},
						{
							Name:  "Top 3",
							Value: "3",
						},
						{
							Name:  "Top 1",
							Value: "1",
						},
					},
				},
				{
					Name:        "name_or_type",
					Description: "Enter pokemon name",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		{
			Name:        "hundo",
			Description: "Gives you the hundo numbers for a specific pokemon.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "pokemon",
					Description: "Pokemon it search for. Examples: mewtwo | charmander | kartana",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		{
			Name:        "eggs",
			Description: "Gives you the breakdown for pokemon in each egg.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "distance",
					Description: "Distance of egg.",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "2 km",
							Value: "2 km",
						},
						{
							Name:  "5 km",
							Value: "5 km",
						},
						{
							Name:  "7 km",
							Value: "7 km",
						},
						{
							Name:  "10 km",
							Value: "10 km",
						},
						{
							Name:  "12 km",
							Value: "12 km",
						},
					},
				},
			},
		},
		{
			Name:        "raids",
			Description: "Gives you all the pokemon in raids.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "raid_tier",
					Description: "tier name. Examples: tier 5 | all | mega",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Tier 1",
							Value: "tier 1",
						},
						{
							Name:  "Tier 3",
							Value: "tier 3",
						},
						{
							Name:  "Tier 5",
							Value: "tier 5",
						},
						{
							Name:  "Mega",
							Value: "mega",
						},
						{
							Name:  "All",
							Value: "all",
						},
					},
				},
			},
		},
		{
			Name:        "xp",
			Description: "Calculate xp to reach xp goals.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "current_xp",
					Description: "your current xp",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
				{
					Name:        "xp_goal",
					Description: "your current xp",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Level 40",
							Value: "20000000",
						},
						{
							Name:  "Level 41",
							Value: "26000000",
						},
						{
							Name:  "Level 42",
							Value: "33500000",
						},
						{
							Name:  "Level 43",
							Value: "42500000",
						},
						{
							Name:  "Level 44",
							Value: "53500000",
						},
						{
							Name:  "Level 45",
							Value: "66500000",
						},
						{
							Name:  "Level 46",
							Value: "82000000",
						},
						{
							Name:  "Level 47",
							Value: "100000000",
						},
						{
							Name:  "Level 48",
							Value: "121000000",
						},
						{
							Name:  "Level 49",
							Value: "146000000",
						},
						{
							Name:  "Level 50",
							Value: "176000000",
						},
					},
				},
			},
		},
	}
)

func main() {
	// pull new data files on start of bot
	pullFiles()

	// connect to the bot
	sess, err := discordgo.New("") //insert bot token in quotes
	if err != nil {
		log.Fatal(err)
	}

	// Add a handler for commands
	sess.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			handleCommand(s, i)
		}
	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	// check if bot is online
	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	// pull all commands
	for _, cmd := range commands {
		_, err := sess.ApplicationCommandCreate(sess.State.User.ID, GuildID, cmd)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", cmd.Name, err)
		}
	}

	println("The bot is online")

	// check for termination signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// add a space after "," for better formating
func addSpace(input string) string {
	input = strings.ReplaceAll(input, ",", ", ")
	return input
}

// convert input from database into yes/no
func convertBool(input string) string {
	if input == "1" {
		return "yes"
	}
	return "no"
}

// func for getting best attackers
func getBest(s *discordgo.Session, search string, sort string, num string, name_type string) string {
	dsn := "root:mysql@tcp(127.0.0.1:3306)/pogodb"

	// open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// build the sql query for all collected data
	query := "SELECT * FROM newdps2 " + search + "WHERE  ORDER BY " + sort + " DESC LIMIT " + num + ";"
	switch search {
	case "name":
		query = "SELECT * FROM newdps2 WHERE name=\"" + name_type + "\" ORDER BY " + sort + " DESC LIMIT " + num + ";"
	case "sametype":
		query = "SELECT * FROM newdps2 WHERE ftype=\"" + name_type + "\" AND ctype=\"" + name_type + "\"  ORDER BY " + sort + " DESC LIMIT " + num + ";"
	case "mixtype":
		query = "SELECT * FROM newdps2 WHERE ctype=\"" + name_type + "\"  ORDER BY " + sort + " DESC LIMIT " + num + ";"
	}

	// execute query
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// count row amount for ranking number and start the message string
	count := 0
	var msg string = ""

	// output each row pulled and format it
	for rows.Next() {
		count++
		var name string
		var fmove string
		var ftype string
		var cmove string
		var ctype string
		var dps string
		var tdo string
		var er string
		var cp string

		err = rows.Scan(&name, &fmove, &ftype, &cmove, &ctype, &dps, &tdo, &er, &cp)
		if err != nil {
			panic(err)
		}
		strcount := strconv.Itoa(count)
		var tt string = "Rank " + strcount + " : **" + name + "**\n**" + fmove + "** (" + ftype + ") / **" + cmove +
			"** (" + ctype + ")\nDPS: **" + dps + "**  |  TDO: **" + tdo + "**  |  Rating: " + er + "  |  CP: " + cp + "\n\n"
		msg = msg + tt
	}

	// close db and send message
	db.Close()
	return msg
}

// func to calculate cp (combat power)
func getCP(mult float64, hp int, attack int, defense int) int {
	// forumla to calculate pokemon cp
	cp := (math.Sqrt(float64(defense+15)) * math.Sqrt(float64(hp+15)) * float64(attack+15) * math.Pow(mult, 2)) / 10
	return int(cp)
}

// func for getting the current pokemon pool for eggs
func getEggs(s *discordgo.Session, egg_distance string) string {
	dsn := "root:mysql@tcp(127.0.0.1:3306)/pogodb"

	// open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// build and execute query
	query := "SELECT * FROM eggs WHERE distance=\"" + egg_distance + "\";"

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// vars for message
	var name string
	var distance string
	var adventure_sync string
	var image string
	var shiny string
	var min_cp string
	var max_cp string
	var regional string

	// add message header and then build the message
	msg := "**Eggs:**\n"

	for rows.Next() {
		err = rows.Scan(&name, &distance, &adventure_sync, &image, &shiny, &min_cp, &max_cp, &regional)
		if err != nil {
			panic(err)
		}
		// create header text and each row of the message
		adventure_sync = convertBool(adventure_sync)
		shiny = convertBool(shiny)
		regional = convertBool(regional)

		msg = msg + "**" + name + " : " + distance + "**   cp: **" + min_cp + "-" + max_cp + "**   Shiny: " + shiny + "   Adventure sync: " + adventure_sync + "\n"
	}

	// close connection and send message
	db.Close()
	return msg
}

// func to get all the relevant hundo numbers for a specific pokemon
func getHundo(s *discordgo.Session, pokemon string) string {
	dsn := "root:mysql@tcp(127.0.0.1:3306)/pogodb"

	// open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// prepare the sql query
	query := "SELECT * FROM pokemon_data WHERE name=\"" + pokemon + "\";"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	//declare all data for message building and constants for calculation
	var name string
	var hp int
	var attack int
	var defense int
	// constants for calculating cp at specific level
	const m15 = 0.51739395
	const m20 = 0.5974
	const m25 = 0.667934
	const m30 = 0.7317
	const m35 = 0.76156384
	const m40 = 0.7903
	const m50 = 0.84029999

	var msg string

	// pull each row of the result
	for rows.Next() {
		err = rows.Scan(&name, &hp, &attack, &defense)
		if err != nil {
			panic(err)
		}
		// create header text and each row of the message
		msg = "Pokemon: **" + name + "**\n"

		msg = msg + "**" + strconv.Itoa(getCP(m15, hp, attack, defense)) + "** - Field Research\n"
		msg = msg + "**" + strconv.Itoa(getCP(m20, hp, attack, defense)) + "** - Eggs / Raid no WB\n"
		msg = msg + "**" + strconv.Itoa(getCP(m25, hp, attack, defense)) + "** - Raid with WB\n"
		msg = msg + "**" + strconv.Itoa(getCP(m30, hp, attack, defense)) + "** - Wild no WB\n"
		msg = msg + "**" + strconv.Itoa(getCP(m35, hp, attack, defense)) + "** - Wild with WB\n"
		msg = msg + "**" + strconv.Itoa(getCP(m40, hp, attack, defense)) + "** - Level 40\n"
		msg = msg + "**" + strconv.Itoa(getCP(m50, hp, attack, defense)) + "** - Level 50\n"
	}
	// close connection and send message
	db.Close()
	return msg
}

// func to get the current raid pool
func getRaids(s *discordgo.Session, raid_tier string) string {
	dsn := "root:mysql@tcp(127.0.0.1:3306)/pogodb"

	// open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	var query string

	// build query
	query = "SELECT * FROM raids WHERE tier=\"" + raid_tier + "\";"
	if raid_tier == "all" {
		query = "SELECT * FROM raids;"
	}

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// vars for message building
	var name string
	var tier string
	var shiny string
	var types string
	var min_cp string
	var max_cp string
	var wb_min_cp string
	var wb_max_cp string
	var boosted_weather string
	var image string

	// create header and then format each raid
	msg := "**Raids:**\n"

	for rows.Next() {
		err = rows.Scan(&name, &tier, &shiny, &types, &min_cp, &max_cp, &wb_min_cp, &wb_max_cp, &boosted_weather, &image)
		if err != nil {
			panic(err)
		}
		// create header text and each row of the message
		shiny = convertBool(shiny)
		types = addSpace(types)
		boosted_weather = addSpace(boosted_weather)

		msg = msg + "**" + name + " : " + tier + "**   Shiny: **" + shiny + "**   cp: **" + min_cp + "-" + max_cp + " | " + wb_min_cp + " - " + wb_max_cp + "**\n"
	}
	// close connection and send message
	db.Close()
	return msg
}

// func to calculate progression towards xp (experience points) landmarks
func getXp(s *discordgo.Session, current_xp int64, goal_xp int64) string {
	var level string = ""
	switch goal_xp {
	case 20000000:
		level = "Level 40"
	case 26000000:
		level = "Level 41"
	case 33500000:
		level = "Level 42"
	case 42500000:
		level = "Level 43"
	case 53500000:
		level = "Level 44"
	case 66500000:
		level = "Level 45"
	case 82000000:
		level = "Level 46"
	case 100000000:
		level = "Level 47"
	case 121000000:
		level = "Level 48"
	case 146000000:
		level = "Level 49"
	case 176000000:
		level = "Level 50"
	}

	// calc percent of completion towards and xp goal
	remaining := goal_xp - current_xp
	percent := roundToDecimal(((float64(current_xp) / float64(goal_xp)) * 100), 2)
	str_percent := fmt.Sprintf("%.2f", percent)

	// build and send message
	var msg string = "Xp to " + level + ": **" + strconv.Itoa(int(remaining)) + "**  |  Percent of xp gained: **" + str_percent + "%**"
	return msg
}

// func to handle and create commands using "/" on the discord end
func handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// switch for each command
	switch i.ApplicationCommandData().Name {
	case "best":
		// Get the user inputs from the options
		search := i.ApplicationCommandData().Options[0].StringValue()
		sort := i.ApplicationCommandData().Options[1].StringValue()
		num := i.ApplicationCommandData().Options[2].StringValue()
		name_type := i.ApplicationCommandData().Options[3].StringValue()

		// build response
		response := getBest(s, search, sort, num, name_type)

		// push message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	case "hundo":
		// Get the user inputs from the options
		pokemon := i.ApplicationCommandData().Options[0].StringValue()

		// build response
		response := getHundo(s, pokemon)

		// push message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	case "eggs":
		// Get the user inputs from the options
		distance := i.ApplicationCommandData().Options[0].StringValue()

		// build response
		response := getEggs(s, distance)

		// push message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	case "raids":
		// Get the user inputs from the options
		raid_tier := i.ApplicationCommandData().Options[0].StringValue()

		// build response
		response := getRaids(s, raid_tier)

		// push message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	case "xp":
		// Get the user inputs from the options
		current_xp := i.ApplicationCommandData().Options[0].IntValue()
		goal_xp := i.ApplicationCommandData().Options[1].IntValue()

		// build response
		response := getXp(s, current_xp, goal_xp)

		// push message
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	}
	// refresh files after sending the response to user
	pullFiles()
}

func pullFiles() {
	// repo url
	repoURL := "https://github.com/bigfoott/ScrapedDuck.git"

	// folder locations
	clonePath := "./data"
	outputPath := "./ScrapedDuck"

	// delete old files
	os.RemoveAll(clonePath)
	os.RemoveAll(outputPath)
	os.Mkdir(outputPath, os.ModePerm)

	// clone repo
	err := CloneRepo(repoURL, clonePath)
	if err != nil {
		log.Fatalf("Error cloning repository: %v", err)
	}

	// copy files from the repo
	err = CopyFilesFromBranch(clonePath, "data", outputPath)
	if err != nil {
		log.Fatalf("Error copying files: %v", err)
	}

	fmt.Println("Files copied successfully to", outputPath)
	// refresh the database with the new pulled data
	refreshDB()
}

// round the given float64 to _ decimal places
func roundToDecimal(f float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals)) // Calculate 10^decimals
	return float64(math.Round(float64(f)*factor) / factor)
}

// send a message to discord
func sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	// send message to channel
	s.ChannelMessageSend(m.ChannelID, msg)
	// refresh files after message
	pullFiles()
}
