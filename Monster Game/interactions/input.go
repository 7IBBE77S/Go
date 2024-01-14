package interactions

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Constants are used here to store the different player choices (attack: 1, heal: 2, special attack: 3)
// in a central place this could avoid errors (types) as these values don't
// have to be repeated in various places throughout the program.
const (
	PlayerChoiceAttack = iota + 1 //iota starts at 0, so we add 1 to it and now we don't need to add values to the constants, further simplifying the code
	PlayersChoiceHeal
	PlayerChoiceSpecialAttack
)

var reader = bufio.NewReader(os.Stdin)

func GetPlayerChoice(specialAttackAvailable bool) string {

	for {
		playerChoice, _ := getPlayerInput()
		if playerChoice == fmt.Sprint(PlayerChoiceAttack) { //if the constant is a string, we can simply just use PlayerChoiceAttack without fmt.Sprint
			return "ATTACK"
		} else if playerChoice == fmt.Sprint(PlayersChoiceHeal) {
			return "HEAL"
		} else if playerChoice == fmt.Sprint(PlayerChoiceSpecialAttack) && specialAttackAvailable {
			return "SPECIAL ATTACK"
		} 
		fmt.Println("User input fetch failed.")
	}

}

func getPlayerInput() (string, error) {
	fmt.Println("Enter your choice: ")
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	userInput = strings.Replace(userInput, "\n", "", -1)

	return userInput, nil
}
