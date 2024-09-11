// getFiles.go
// Author: Cade Beckers
// Written: 08/23/2024
// Updated: 09/10/2024

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// struct to map eggs to json
type Egg struct {
	Name            string `json:"name"`
	EggType         string `json:"eggType"`
	IsAdventureSync bool   `json:"isAdventureSync"`
	Image           string `json:"image"`
	CanBeShiny      bool   `json:"canBeShiny"`
	CombatPower     struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"combatPower"`
	IsRegional bool `json:"isRegional"`
}

// struct to map events to json (not in use)
type Event struct {
	EventID   string `json:"eventID"`
	Name      string `json:"name"`
	EventType string `json:"eventType"`
	Heading   string `json:"heading"`
	Link      string `json:"link"`
	Image     string `json:"image"`
	Start     string `json:"start"`
	End       string `json:"end"`
}

// struct to map raids to json
type Raid struct {
	Name       string `json:"name"`
	Tier       string `json:"tier"`
	CanBeShiny bool   `json:"canBeShiny"`
	Types      []struct {
		Name  string `json:"name"`
		Image string `json:"image"`
	} `json:"types"`
	CombatPower struct {
		Normal struct {
			Min int `json:"min"`
			Max int `json:"max"`
		} `json:"normal"`
		Boosted struct {
			Min int `json:"min"`
			Max int `json:"max"`
		} `json:"boosted"`
	} `json:"combatPower"`
	BoostedWeather []struct {
		Name  string `json:"name"`
		Image string `json:"image"`
	} `json:"boostedWeather"`
	Image string `json:"image"`
}

// struct to map rewards to json (not in use)
type Reward struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	CanBeShiny  bool   `json:"canBeShiny"`
	CombatPower struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"combatPower"`
}

// struct to map researches to json (not in use)
type ResearchTask struct {
	Text    string   `json:"text"`
	Type    string   `json:"type"`
	Rewards []Reward `json:"rewards"`
}

// func to clone given github repo
func CloneRepo(repoURL, clonePath string) error {
	// Run the git clone command
	cmd := exec.Command("git", "clone", repoURL, clonePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	return nil
}

// func to copy files from a given branch
func CopyFilesFromBranch(clonePath, branchName, outputPath string) error {
	// get the branch
	cmd := exec.Command("git", "-C", clonePath, "checkout", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	// copy files from the repo
	err = filepath.Walk(clonePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		// copy files into directory
		destPath := filepath.Join(outputPath, info.Name())
		input, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		err = os.WriteFile(destPath, input, 0644)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}

// func to refresh the values in the database
func refreshDB() {
	clearTables()
	readEgg()
	readEvent()
	readRaid()
	readResearches()
}

// func to read the egg.json data
func readEgg() {
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

	// read json file
	jsonFile, err := os.ReadFile("./data/eggs.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	// export data into array
	var pokemon []Egg
	err = json.Unmarshal(jsonFile, &pokemon)
	if err != nil {
		fmt.Println(err)
		return
	}

	// loop each element and generate query
	for _, p := range pokemon {
		if strings.Contains(p.Name, "'") {
			p.Name = strings.ReplaceAll(p.Name, "'", "")
		}

		query := fmt.Sprintf(`INSERT INTO eggs (name, distance, adventure_sync, image, shiny, min_cp, max_cp, regional) 
			VALUES ('%s', '%s', %t, '%s', %t, %d, %d, %t);`,
			p.Name, p.EggType, p.IsAdventureSync, p.Image, p.CanBeShiny, p.CombatPower.Min, p.CombatPower.Max, p.IsRegional)
		rows, err := db.Query(query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
	}
}

// func to read events.json data (not in use)
func readEvent() {
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

	// read json file
	jsonFile, err := os.ReadFile("./data/events.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// export data into array
	var events []Event
	err = json.Unmarshal(jsonFile, &events)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// loop each element and generate query
	for _, e := range events {
		query := fmt.Sprintf(`INSERT INTO events (event_id, name, event_type, heading, link, image) 
			VALUES ('%s', '%s', '%s', '%s', '%s', '%s');`,
			e.EventID, e.Name, e.EventType, e.Heading, e.Link, e.Image)
		rows, err := db.Query(query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
	}
}

// func to read raids.json data
func readRaid() {
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

	// read json file
	jsonFile, err := os.ReadFile("./data/raids.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// export data into array
	var raids []Raid
	err = json.Unmarshal(jsonFile, &raids)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// loop each element and generate query
	for _, raid := range raids {
		// get each raid type
		var typesStr string
		for _, t := range raid.Types {
			typesStr += t.Name + ","
		}
		typesStr = typesStr[:len(typesStr)-1] // remove trailing comma

		// each weather condition
		var boostedWeatherStr string
		for _, bw := range raid.BoostedWeather {
			boostedWeatherStr += bw.Name + ","
		}
		boostedWeatherStr = boostedWeatherStr[:len(boostedWeatherStr)-1] // remove trailing comma

		// build query and execute
		query := fmt.Sprintf(`INSERT INTO raids (name, tier, shiny, types, min_cp, max_cp, wb_min_cp, wb_max_cp, boosted_weather, image) 
			VALUES ('%s', '%s', %t, '%s', %d, %d, %d, %d, '%s', '%s');`,
			raid.Name, raid.Tier, raid.CanBeShiny, typesStr, raid.CombatPower.Normal.Min, raid.CombatPower.Normal.Max, raid.CombatPower.Boosted.Min, raid.CombatPower.Boosted.Max, boostedWeatherStr, raid.Image)

		rows, err := db.Query(query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
	}
}

// func to read researches.json data (not in use)
func readResearches() {
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

	// read json file
	jsonFile, err := os.ReadFile("./data/research.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// export data into array
	var tasks []ResearchTask
	err = json.Unmarshal(jsonFile, &tasks)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// loop each element and generate query
	for _, task := range tasks {
		for _, reward := range task.Rewards {
			// combine task with reward data
			query := fmt.Sprintf(`INSERT INTO researches (text, type, reward, shiny, min_cp, max_cp, image) 
				VALUES ('%s', '%s', '%s', %t, %d, %d, '%s');`,
				task.Text, task.Type, reward.Name, reward.CanBeShiny, reward.CombatPower.Min, reward.CombatPower.Max, reward.Image)
			rows, err := db.Query(query)
			if err != nil {
				panic(err)
			}
			defer rows.Close()
		}
	}
}

// func to clear tables
func clearTables() {
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

	// clear tables
	commands := [4]string{"DELETE FROM eggs", "DELETE FROM events", "DELETE FROM raids", "DELETE FROM researches"}
	for i := 0; i < len(commands); i++ {
		rows, err := db.Query(commands[i])
		if err != nil {
			panic(err)
		}
		defer rows.Close()
	}
}
