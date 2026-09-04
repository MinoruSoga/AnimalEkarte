package textsearch

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeKana(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "hiragana unchanged", input: "ぴーたー", want: "ぴーたー"},
		{name: "katakana converted", input: "ピーター", want: "ぴーたー"},
		{name: "mixed script", input: "山田ハナコ", want: "山田はなこ"},
		{name: "small katakana", input: "ァィゥェォ", want: "ぁぃぅぇぉ"},
		{name: "extended katakana", input: "ヴヵヶ", want: "ゔゕゖ"},
		{name: "symbols unchanged", input: `abc123!@#%_\`, want: `abc123!@#%_\`},
		{name: "empty", input: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, NormalizeKana(test.input))
		})
	}
}

func TestNormalizeQuerySpaces(t *testing.T) {
	assert.Equal(t, "山田 花子", NormalizeQuerySpaces("山田　花子"))
	assert.Equal(t, "山田 花子", NormalizeQuerySpaces("  山田   花子  "))
	assert.Equal(t, "", NormalizeQuerySpaces("　　"))
}

func TestEscapeLike(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "wildcards", input: `%_`, want: `\%\_`},
		{name: "backslash before wildcards", input: `a\b%c_d`, want: `a\\b\%c\_d`},
		{name: "plain text", input: "山田", want: "山田"},
		{name: "empty", input: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, EscapeLike(test.input))
		})
	}
}

func TestKanaTranslationConstantsCoverExpectedRange(t *testing.T) {
	assert.Equal(t, []rune(KanaSourceChars), []rune("ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ"))
	assert.Equal(t, []rune(KanaTargetChars), []rune("ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ"))
	assert.Len(t, []rune(KanaSourceChars), len([]rune(KanaTargetChars)))
	assert.NotContains(t, KanaSourceChars, "\u3000")
	assert.NotContains(t, KanaTargetChars, "\u3000")
}

func TestSpaceTranslationConstantsFoldIdeographicSpace(t *testing.T) {
	assert.Equal(t, "\u3000", SpaceSourceChars)
	assert.Equal(t, " ", SpaceTargetChars)
	assert.Len(t, []rune(SpaceSourceChars), len([]rune(SpaceTargetChars)))
}

func TestKanaAndSpaceTranslationConstantsComposeWithoutMutatingKana(t *testing.T) {
	assert.Equal(t, KanaSourceChars+SpaceSourceChars, KanaAndSpaceSourceChars)
	assert.Equal(t, KanaTargetChars+SpaceTargetChars, KanaAndSpaceTargetChars)
	assert.Len(t, []rune(KanaAndSpaceSourceChars), len([]rune(KanaAndSpaceTargetChars)))
	assert.True(t, strings.HasPrefix(KanaAndSpaceSourceChars, KanaSourceChars))
	assert.True(t, strings.HasSuffix(KanaAndSpaceSourceChars, "\u3000"))
	assert.True(t, strings.HasPrefix(KanaAndSpaceTargetChars, KanaTargetChars))
	assert.True(t, strings.HasSuffix(KanaAndSpaceTargetChars, " "))
	assert.Equal(t, []rune(KanaSourceChars), []rune("ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ"))
	assert.Equal(t, []rune(KanaTargetChars), []rune("ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ"))
}
