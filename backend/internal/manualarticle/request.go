package manualarticle

// request.go — PUT /manual/articles/{category}/{slug} の厳格 request parse。
//
// UpsertManualArticleRequest（api.yaml SHA 30538608）:
//   operation_key(uuid) / article_id(正数十進文字列) /
//   expected_version(int64>=1) / expected_content_version_id(正数十進文字列) /
//   content{title,order_value,section,body_markdown} 全必須・非NULL。
// 追加 property は schema の additionalProperties:false に従い拒否する。
// trim・NFC 正規化・改行正規化は行わず、文字列は decode 後の raw を保持する。
// order_value は有限・±99999999.99 以内・小数2桁以下のみ受理し、JCS と
// numeric(10,2) 保存の両方に同じ canonical decimal text を使う。

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
)

var errInvalidEditRequest = apperrors.WrapInvalidInput("invalid manual article edit request")

// 既存 binding max を維持する（validator の max は文字列の rune 数）。
const (
	maxTitleRunes   = 255
	maxSectionRunes = 255
	maxBodyRunes    = 100000
)

// editCommand は厳格 parse 済みの不変編集命令。C06 Content・Preconditions・
// domain write ・digest のすべての源になる（participant は req.Content の
// fingerprint 済み値だけを参照し、digest 再計算にのみコマンド値を使う）。
type editCommand struct {
	OperationKey             uuid.UUID
	ArticleID                uint64
	ArticleIDText            string // canonical positive signed-64 decimal
	ExpectedVersion          int64
	ExpectedContentVersionID uint64
	Title                    string
	OrderText                string // canonical decimal text（JCS・numeric 保存の共通源）
	OrderValue               model.ExactNumeric
	Section                  string
	BodyMarkdown             string
	ContentDigestHex         string
	CommandDigestHex         string
}

// Content は C06 Content の 3 member（すべて fingerprint 済み文字列）を返す。
func (c *editCommand) operationContent() (operation.Content, error) {
	return operation.NewContent(map[string]string{
		"article_id":     c.ArticleIDText,
		"command_digest": c.CommandDigestHex,
		"content_digest": c.ContentDigestHex,
	})
}

// Preconditions は fingerprint 対象外の実行前提（CAS fence）を返す。
func (c *editCommand) operationPreconditions() operation.Preconditions {
	return operation.Preconditions{
		"expected_version":            strconv.FormatInt(c.ExpectedVersion, 10),
		"expected_content_version_id": strconv.FormatUint(c.ExpectedContentVersionID, 10),
	}
}

// parseEditRequest は PUT body を厳格 decode し、編集命令と digest を作る。
// category/slug は route-decoded 済みの raw 文字列をそのまま受け取る。
func parseEditRequest(raw []byte, category model.ManualCategory, slug string) (*editCommand, error) {
	if !utf8.Valid(raw) {
		return nil, errInvalidEditRequest
	}
	if hasLoneSurrogateEscape(raw) {
		return nil, errInvalidEditRequest
	}
	outer, err := decodeStrictObject(raw)
	if err != nil {
		return nil, errInvalidEditRequest
	}
	required := []string{"operation_key", "article_id", "expected_version", "expected_content_version_id", "content"}
	for _, key := range required {
		v, ok := outer[key]
		if !ok || isJSONNull(v) {
			return nil, errInvalidEditRequest
		}
	}
	if len(outer) != len(required) {
		return nil, errInvalidEditRequest
	}

	key, err := operation.ParseOperationKey(rawString(outer["operation_key"]))
	if err != nil {
		return nil, errInvalidEditRequest
	}
	articleIDText, articleID, err := positiveDecimalMember(outer["article_id"])
	if err != nil {
		return nil, errInvalidEditRequest
	}
	expectedVersion, err := positiveIntMember(outer["expected_version"])
	if err != nil {
		return nil, errInvalidEditRequest
	}
	_, expectedContentVersionID, err := positiveDecimalMember(outer["expected_content_version_id"])
	if err != nil {
		return nil, errInvalidEditRequest
	}

	content, err := decodeStrictObject(outer["content"])
	if err != nil || len(content) != 4 {
		return nil, errInvalidEditRequest
	}
	for _, key := range []string{"title", "order_value", "section", "body_markdown"} {
		v, ok := content[key]
		if !ok || isJSONNull(v) {
			return nil, errInvalidEditRequest
		}
	}
	title, ok := jsonString(content["title"])
	if !ok || title == "" || utf8.RuneCountInString(title) > maxTitleRunes {
		return nil, errInvalidEditRequest
	}
	section, ok := jsonString(content["section"])
	if !ok || section == "" || utf8.RuneCountInString(section) > maxSectionRunes {
		return nil, errInvalidEditRequest
	}
	body, ok := jsonString(content["body_markdown"])
	if !ok || utf8.RuneCountInString(body) > maxBodyRunes {
		return nil, errInvalidEditRequest
	}
	orderText, orderValue, err := parseOrderValue(content["order_value"])
	if err != nil {
		return nil, errInvalidEditRequest
	}

	contentDigest := manualContentDigestHex(title, orderText, section, body)
	cmd := &editCommand{
		OperationKey:             key,
		ArticleID:                articleID,
		ArticleIDText:            articleIDText,
		ExpectedVersion:          expectedVersion,
		ExpectedContentVersionID: expectedContentVersionID,
		Title:                    title,
		OrderText:                orderText,
		OrderValue:               orderValue,
		Section:                  section,
		BodyMarkdown:             body,
		ContentDigestHex:         contentDigest,
		CommandDigestHex:         manualEditCommandDigestHex(articleIDText, string(category), slug, contentDigest),
	}
	return cmd, nil
}

// decodeStrictObject は 1 JSON object を member 走査で decode し、重複 member・
// 非 object・ trailing content を拒否する。key 昇順は要求しない。
func decodeStrictObject(raw []byte) (map[string]json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("not json")
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("not an object")
	}
	out := make(map[string]json.RawMessage)
	for dec.More() {
		kt, terr := dec.Token()
		if terr != nil {
			return nil, fmt.Errorf("bad member")
		}
		name, ok := kt.(string)
		if !ok {
			return nil, fmt.Errorf("bad member name")
		}
		if _, dup := out[name]; dup {
			return nil, fmt.Errorf("duplicate member")
		}
		var v json.RawMessage
		if terr := dec.Decode(&v); terr != nil {
			return nil, fmt.Errorf("bad member value")
		}
		out[name] = v
	}
	if _, terr := dec.Token(); terr != nil {
		return nil, fmt.Errorf("unclosed object")
	}
	if dec.More() {
		return nil, fmt.Errorf("trailing content")
	}
	return out, nil
}

// hasLoneSurrogateEscape は \uXXXX escape の surrogate 対が壊れていないか
// raw byte 走査で検査する。Go の JSON decoder は孤立 surrogate を U+FFFD へ
// 黙って置き換えるため、wire 上のまま検査しないと内容が変化する。
func hasLoneSurrogateEscape(raw []byte) bool {
	isHex4 := func(i int) bool {
		if i+4 > len(raw) {
			return false
		}
		for j := i; j < i+4; j++ {
			c := raw[j]
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
				return false
			}
		}
		return true
	}
	codeUnit := func(i int) int {
		v, _ := strconv.ParseUint(string(raw[i:i+4]), 16, 32)
		return int(v)
	}
	for i := 0; i+1 < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		if raw[i+1] != 'u' {
			i++ // \\/\"\\ etc. skip escaped char
			continue
		}
		if !isHex4(i + 2) {
			return true // malformed \u escape
		}
		unit := codeUnit(i + 2)
		switch {
		case unit >= 0xD800 && unit <= 0xDBFF:
			// high surrogate: next escape must be a low surrogate
			if i+12 > len(raw) || raw[i+6] != '\\' || raw[i+7] != 'u' || !isHex4(i+8) {
				return true
			}
			low := codeUnit(i + 8)
			if low < 0xDC00 || low > 0xDFFF {
				return true
			}
			i += 11
		case unit >= 0xDC00 && unit <= 0xDFFF:
			return true // unpaired low surrogate
		default:
			i += 5
		}
	}
	return false
}

func isJSONNull(raw json.RawMessage) bool {
	return string(bytes.TrimSpace(raw)) == "null"
}

func jsonString(raw json.RawMessage) (string, bool) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	return s, true
}

func rawString(raw json.RawMessage) string {
	s, _ := jsonString(raw)
	return s
}

var positiveDecimalPattern = func() func(string) bool {
	return func(s string) bool {
		if s == "" || s[0] == '0' {
			return false
		}
		for i := 0; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return false
			}
		}
		return true
	}
}()

// positiveDecimalMember は文字列 member を canonical 正数 signed-64 decimal
// として検査し、元の text と値を返す（文字列 member のみ受理）。
func positiveDecimalMember(raw json.RawMessage) (string, uint64, error) {
	s, ok := jsonString(raw)
	if !ok || !positiveDecimalPattern(s) {
		return "", 0, fmt.Errorf("not a positive decimal string")
	}
	n, err := strconv.ParseUint(s, 10, 63)
	if err != nil {
		return "", 0, fmt.Errorf("out of signed-64 range")
	}
	return s, n, nil
}

// positiveIntMember は整数 member（非文字列 JSON number）を >=1 int64 として検査する。
func positiveIntMember(raw json.RawMessage) (int64, error) {
	var num json.Number
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&num); err != nil {
		return 0, fmt.Errorf("not a number")
	}
	n, err := num.Int64()
	if err != nil || n < 1 {
		return 0, fmt.Errorf("not a positive int64")
	}
	return n, nil
}

// parseOrderValue は order_value member を numeric(10,2) 契約で検査する:
// JSON number・有限・-99999999.99..99999999.99・小数2桁以下。
// canonical decimal text は受け取った有限 double を最短10進へ写したもので、
// 浮動小数点の丸め・許容誤差は使わない。返す text は JCS と DB 保存が共有する。
func parseOrderValue(raw json.RawMessage) (string, model.ExactNumeric, error) {
	var num json.Number
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&num); err != nil {
		return "", model.ExactNumeric{}, fmt.Errorf("not a number")
	}
	f, err := num.Float64()
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return "", model.ExactNumeric{}, fmt.Errorf("not a finite number")
	}
	if f < -99999999.99 || f > 99999999.99 {
		return "", model.ExactNumeric{}, fmt.Errorf("order_value out of numeric(10,2) range")
	}
	text := strconv.FormatFloat(f, 'f', -1, 64)
	if text == "-0" {
		text = "0"
	}
	// 小数部は2桁まで。-1 精度の最短表現は指数を出さない範囲で
	// 受け取った double を厳密に表す（tolerance・丸めはしない）。
	if dot := strings.IndexByte(text, '.'); dot >= 0 && len(text)-dot-1 > 2 {
		return "", model.ExactNumeric{}, fmt.Errorf("order_value has more than 2 decimal places")
	}
	var numeric pgtype.Numeric
	if err := numeric.Scan(text); err != nil {
		return "", model.ExactNumeric{}, fmt.Errorf("order_value cannot be stored as numeric")
	}
	return text, model.ExactNumeric{Numeric: numeric}, nil
}
