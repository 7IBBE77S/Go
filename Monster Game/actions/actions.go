package actions

import (
	
	"math/rand"
	"time"
)

var randSource = rand.NewSource(time.Now().UnixNano())
var randGenerator = rand.New(randSource)
var currentMonsterHealth int = MONSTER_HEALTH
var currentPlayerHealth int = PLAYER_HEALTH

func AttackMonsters(isSpecialAttack bool) int {
	minAttack := MIN_ATTACK_DMG
	maxAttack := MAX_ATTACK_DMG
	if isSpecialAttack {
		minAttack = MIN_SPECIAL_ATTACK
		maxAttack = MAX_SPECIAL_ATTACK

	}
	dmgValue := generateRandBetween(minAttack, maxAttack)
	currentMonsterHealth -= dmgValue // currentMonsterHealth = currentMonsterHealth - dmgValue
	return dmgValue

}

func HealPlayer() int{

	healValue := generateRandBetween(MIN_HEAL, MAX_HEAL)

	healthDiff := PLAYER_HEALTH - currentPlayerHealth
	if healthDiff >= healValue {
		currentPlayerHealth += healValue
		return healValue
	} else {
		currentPlayerHealth = PLAYER_HEALTH
		return healthDiff
	}
	// currentMonsterHealth = currentMonsterHealth - dmgValue

}

// func DefendMonster() int {

// 	healValue := generateRandBetween(MIN_HEAL, MAX_HEAL)
// 	healthDiff := MONSTER_HEALTH - currentMonsterHealth
// 	if healthDiff >= healValue {
// 		currentMonsterHealth += healValue + MIN_ATTACK_DMG
// 		return healValue
// 	} else {
// 		currentMonsterHealth = MONSTER_HEALTH
// 		return healthDiff
// 	}
// }
func AttackPlayer() int {
	minAttack := MIN_MONSTER_ATTACK_DMG
	maxAttack := MAX_MONSTER_ATTACK_DMG

	dmgValue := generateRandBetween(minAttack, maxAttack)
	currentPlayerHealth -= dmgValue

	return dmgValue
}
func GetHealthAmount() (int, int) {
	return currentPlayerHealth, currentMonsterHealth
}
func generateRandBetween(min int, max int) int {
	return randGenerator.Intn(max-min) + min
}
