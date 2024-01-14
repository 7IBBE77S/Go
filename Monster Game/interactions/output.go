package interactions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/common-nighthawk/go-figure"
)

type RoundData struct {
	Action            string
	PlayerAttackDmg   int
	PlayerHealValue   int
	MonsterHealValue  int
	MonsterAttackDmg  int
	PlayerHealth      int
	MonsterHealth     int
	PlayerDefendValue int
}
type Colors struct {
	LightRed string
	Red      string
	Green    string
	Yellow   string
	Blue     string
	Magenta  string
	Purple   string
	Cyan     string
	White    string
	Grey     string
	Clean    string
}

var colors = Colors{
	LightRed: "\033[1;31m",
	Red:      "\x1b[31m",
	Green:    "\x1b[32m",
	Yellow:   "\x1b[33m",

	Blue:    "\x1b[34m",
	Magenta: "\x1b[35m",
	Purple:  "\x1b[36m",
	Cyan:    "\x1b[36m",
	White:   "\x1b[37m",
	Grey:    "\x1b[90m",
	Clean:   "\x1b[0m",
}

func PrintGreetings() {

	startTitle := figure.NewColorFigure("MONSTER SLAYER!", "", "red", true)
	startTitle.Print()

	// fmt.Println("Starting a new game"+colors.Grey+"...", colors.Clean)

}

func ShowAvailableActions(specialAttackAvailable bool) {
	fmt.Println(" ")
	fmt.Println("Available actions")
	fmt.Println("-----------------")
	fmt.Println("1. ", colors.LightRed+"Attack", colors.Clean)
	fmt.Println("2. ", colors.Green+"Heal", colors.Clean)
	if specialAttackAvailable {
		fmt.Println("3. ", colors.Yellow+"Critical Attack", colors.Clean)

	}

	// fmt.Println("3. Exit")
}
func (roundData *RoundData) PrintRoundStats() {
	if roundData.Action == "ATTACK" {

		fmt.Printf("Player attacks monster for %v damage.\n", roundData.PlayerAttackDmg)

	} else if roundData.Action == "SPECIAL ATTACK" {
		if roundData.PlayerAttackDmg <= 15 {
			fmt.Printf("Player lands a critical hit for "+colors.Yellow+"%v"+colors.Clean+" damage!\n", roundData.PlayerAttackDmg)
		} else if roundData.PlayerAttackDmg > 15 {
			fmt.Printf("Player lands a critical hit for "+colors.Red+"%v"+colors.Clean+" damage!\n", roundData.PlayerAttackDmg)
		}
	} else if roundData.Action == "HEAL" {

		fmt.Printf("Player uses heal for %v amount.\n", roundData.PlayerHealValue)

	}
	fmt.Printf("Monster attacks player for %v damage.\n", roundData.MonsterAttackDmg)

	var healthColor string
	//Player
	health := roundData.PlayerHealth
	var healthBars string
	for i := 0; i < health/10; i++ {
		healthBars += "█"

	}

	switch {
	case health >= 70:
		healthColor = colors.White
	case health >= 40:
		healthColor = colors.Yellow
	case health >= 25:
		healthColor = colors.LightRed
	default:
		healthColor = colors.Red
	}
	fmt.Printf("Player Health "+healthColor+"%v "+strings.Repeat(healthBars, 2)+"\n"+colors.Clean, health)
	//Monster
	monsterHealth := roundData.MonsterHealth
	var healthBarsM string
	for i := 0; i < monsterHealth/10; i++ {
		healthBarsM += "█"
	}
	switch {
	case monsterHealth >= 70:
		healthColor = colors.White
	case monsterHealth >= 40:
		healthColor = colors.Yellow
	case monsterHealth >= 25:
		healthColor = colors.LightRed
	default:
		healthColor = colors.Red
	}
	fmt.Printf("Monster Health "+healthColor+"%v "+strings.Repeat(healthBarsM, 2)+"\n"+colors.Clean, monsterHealth)
}
func DeclareWinners(winner string) {
	fmt.Println("---------------------------------------------------------------------------")
	endGame := figure.NewColorFigure("GAME OVER!", "", "yellow", true)
	endGame.Print()
	//	fmt.Println(colors.Grey,endGame,colors.Clean)
	fmt.Println("---------------------------------------------------------------------------")

	fmt.Printf("%v won!\n", winner)
}

func WriteLogFile(rounds *[]RoundData) {
	exPath, err := os.Executable()
	if err != nil {
		fmt.Println("Writing to log file failed. Exiting...")
		return
	}
	exPath = filepath.Dir(exPath)

	executableFile, executableErr := os.Create(exPath + "/gamelog.txt")

	file, err := os.Create("gamelog.txt")

	if err != nil {
		fmt.Println("Saving to log file failed")
		return
	}
	if executableErr != nil {
		fmt.Println("Saving to log file failed")
		return
	}
	defer func() {
		err1,err := executableFile.Close(),file.Close()
		if err != nil{ 
			fmt.Println("Closing the file failed!")
		}
		if err1 != nil {
			fmt.Println("Closing the file failed!")
		}
	}()
	for i, v := range *rounds {
		logEntry := map[string]string{
			"Round":                 fmt.Sprintln(i + 1),
			"\n Action":             fmt.Sprintln(v.Action),
			"Player Attack Damage":  fmt.Sprintln(v.PlayerAttackDmg),
			"Player Heal Value":     fmt.Sprintln(v.PlayerHealValue),
			"Monster Attack Damage": fmt.Sprintln(v.MonsterAttackDmg),
			"Player Health":         fmt.Sprintln(v.PlayerHealth),
			"Monster Health":        fmt.Sprintln(v.MonsterHealth), // fmt.Sprintf(" v%\n", v.MonsterHealth) for better log formatting.
		}
		logLine := fmt.Sprintln(logEntry)
		
		
		_, executableErr = executableFile.WriteString(logLine)
		_, err = file.WriteString(logLine)
		if executableErr != nil {
			fmt.Println("Writing to log file failed. Exiting...")
			continue
		}
		if err != nil {
			fmt.Println("Writing to log file failed. Exiting...")
			continue
		}

	}
	

	fmt.Println("Writing to log file complete.")
}
