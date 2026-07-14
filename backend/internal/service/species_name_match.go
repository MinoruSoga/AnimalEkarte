package service

import "strings"

// isDogSpeciesName は free-text 種名が犬を示すかを部分一致で判定する（L-STEP マーケティングタグ用途）。
// 部分一致（マーケティングタグ用途）。投薬計算の doseSpeciesAliases（完全一致・fail-closed）とは契約が異なり、統合してはならない。
func isDogSpeciesName(name string) bool {
	return strings.Contains(name, "犬")
}

// isCatSpeciesName は free-text 種名が猫を示すかを部分一致で判定する（L-STEP マーケティングタグ用途）。
// 部分一致（マーケティングタグ用途）。投薬計算の doseSpeciesAliases（完全一致・fail-closed）とは契約が異なり、統合してはならない。
func isCatSpeciesName(name string) bool {
	return strings.Contains(name, "猫")
}
