package main

import (
	//	"fmt"

	// "math/rand"
	// "time"

	"github.com/7IBBE77S/monsterslayer/actions"
	"github.com/7IBBE77S/monsterslayer/interactions"
)
  
var currentRound int = 0        
var gameRounds = []interactions.RoundData{}

func main() {
	startGame()
	winner := "" //player || Monster || ""

	for winner == "" {
		winner = executeRound()
		//winner = getWinners()
	}

	endGame(winner)

}

func startGame() {
	interactions.PrintGreetings()
}

func executeRound() string {
	currentRound++ // currentRound = currentRound + 1 || currentRound += 1
	//randRound := generateRandBetween(0,3)
	isSpecialRound := currentRound%3 == 0
	//isMonsterSpecialRound := currentRound%5 == 0

	interactions.ShowAvailableActions(isSpecialRound)

	userChoice := interactions.GetPlayerChoice(isSpecialRound)

	var playerAttackDmg int
	var monsterAttackDmg int
	var playerHealValue int
	var playerDefendValue int
	var monsterHealValue int

	if userChoice == "ATTACK" {
		playerAttackDmg = actions.AttackMonsters(false)

	} else if userChoice == "SPECIAL ATTACK" {
		playerAttackDmg = actions.AttackMonsters(true)
	} else if userChoice == "HEAL" {
		playerHealValue = actions.HealPlayer()

	}


	//monsterHealValue = actions.HealMonster()

	monsterAttackDmg = actions.AttackPlayer()

	//monsterHealValue = actions.HealMonster()

	playerHealth, monsterHealth := actions.GetHealthAmount()

	roundData := interactions.RoundData{
		Action:           userChoice,
		PlayerAttackDmg:  playerAttackDmg,
		PlayerHealValue:  playerHealValue,
		PlayerDefendValue: playerDefendValue,
		MonsterHealValue: monsterHealValue,
		MonsterAttackDmg: monsterAttackDmg,
		PlayerHealth:     playerHealth,
		MonsterHealth:    monsterHealth,
	}
	//interactions.PrintRoundStats(&roundData)

	roundData.PrintRoundStats()
	gameRounds = append(gameRounds, roundData)
	if playerHealth <= 0 {
		return "Monster"
	} else if monsterHealth <= 0 {
		return "Player"
	}
	return ""
}
func endGame(winner string) {
	interactions.DeclareWinners(winner)
	interactions.WriteLogFile(&gameRounds)

}
