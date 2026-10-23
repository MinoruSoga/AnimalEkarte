package clinic

// 共通医院時間の保存行（clinic_operating_hours）を検証済み Settings へ復号する strict loader（A03 §2.1/§2.2）。
// DB の JSONB は構造しか保証しないため（weekly_hours は長さ 7 の配列、closed_dates は配列まで）、
// 未知 key・重複 key・null・型違い・末尾データ・区間不正は、ここで破損として拒否する。
// 破損は ErrOperatingHoursStoredDataCorrupt で返し、未設定（行なし）や「全日休診」へ化けさせない。

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/animal-ekarte/backend/internal/model"
)

// ErrOperatingHoursStoredDataCorrupt は保存行が契約の形に合わない（スキーマ版・版番号・JSON 形・値の不正）ことを表す。
// 入力不正（apperrors.ErrInvalidInput）ではなくサーバ側の保存データ欠陥で、handler は 503 へ写像する。
var ErrOperatingHoursStoredDataCorrupt = errors.New("clinic operating hours stored data is corrupt")

// storedJSONMaxDepth は保存 JSON の最大ネスト。正規の形（配列→オブジェクト→配列→オブジェクト）は 4 段で収まる。
const storedJSONMaxDepth = 8

func corruptOperatingHours(detail string) error {
	return fmt.Errorf("%w: %s", ErrOperatingHoursStoredDataCorrupt, detail)
}

// LoadOperatingHoursSettings は保存行から検証済み Settings を作る。
// schema_version は model.ClinicOperatingHoursSchemaV1 と完全一致、version は正、clinic_id は非 0 が必要。
// NewOperatingHoursSettings が拒否するものはすべて破損として扱う（原因文字列は含めるが、
// apperrors.ErrInvalidInput は連鎖させず、呼出側が 400 へ誤写像しないようにする）。
func LoadOperatingHoursSettings(row model.ClinicOperatingHours) (OperatingHoursSettings, error) {
	if row.ClinicID == 0 {
		return OperatingHoursSettings{}, corruptOperatingHours("clinic_id is zero")
	}
	if row.SchemaVersion != model.ClinicOperatingHoursSchemaV1 {
		return OperatingHoursSettings{}, corruptOperatingHours("unsupported schema_version")
	}
	if row.Version <= 0 {
		return OperatingHoursSettings{}, corruptOperatingHours("version must be positive")
	}

	weeklyNode, err := parseStoredJSON(row.WeeklyHours)
	if err != nil {
		return OperatingHoursSettings{}, corruptOperatingHours("weekly_hours: " + err.Error())
	}
	weekly, err := storedWeeklyHours(weeklyNode)
	if err != nil {
		return OperatingHoursSettings{}, corruptOperatingHours("weekly_hours: " + err.Error())
	}

	closedNode, err := parseStoredJSON(row.ClosedDates)
	if err != nil {
		return OperatingHoursSettings{}, corruptOperatingHours("closed_dates: " + err.Error())
	}
	closed, err := storedClosedDates(closedNode)
	if err != nil {
		return OperatingHoursSettings{}, corruptOperatingHours("closed_dates: " + err.Error())
	}

	holidayClosed := row.NationalHolidayClosed
	settings, err := NewOperatingHoursSettings(OperatingHoursFields{
		WeeklyHours:           &weekly,
		ClosedDates:           &closed,
		NationalHolidayClosed: &holidayClosed,
	})
	if err != nil {
		return OperatingHoursSettings{}, corruptOperatingHours(err.Error())
	}
	return settings, nil
}

func storedWeeklyHours(node storedJSONNode) ([]WeeklyHoursInput, error) {
	elems, err := node.array()
	if err != nil {
		return nil, err
	}
	out := make([]WeeklyHoursInput, 0, len(elems))
	for _, elem := range elems {
		fields, err := elem.exactObject("weekday", "intervals")
		if err != nil {
			return nil, err
		}
		weekday, err := fields[0].integer()
		if err != nil {
			return nil, fmt.Errorf("weekday: %w", err)
		}
		rawIntervals, err := fields[1].array()
		if err != nil {
			return nil, fmt.Errorf("intervals: %w", err)
		}
		intervals := make([]IntervalInput, 0, len(rawIntervals))
		for _, rawInterval := range rawIntervals {
			bounds, err := rawInterval.exactObject("start", "end")
			if err != nil {
				return nil, fmt.Errorf("interval: %w", err)
			}
			start, err := bounds[0].stringValue()
			if err != nil {
				return nil, fmt.Errorf("start: %w", err)
			}
			end, err := bounds[1].stringValue()
			if err != nil {
				return nil, fmt.Errorf("end: %w", err)
			}
			intervals = append(intervals, IntervalInput{Start: start, End: end})
		}
		out = append(out, WeeklyHoursInput{Weekday: weekday, Intervals: intervals})
	}
	return out, nil
}

func storedClosedDates(node storedJSONNode) ([]ClosedDateInput, error) {
	elems, err := node.array()
	if err != nil {
		return nil, err
	}
	out := make([]ClosedDateInput, 0, len(elems))
	for _, elem := range elems {
		fields, err := elem.exactObject("date", "reason")
		if err != nil {
			return nil, err
		}
		date, err := fields[0].stringValue()
		if err != nil {
			return nil, fmt.Errorf("date: %w", err)
		}
		reason, err := fields[1].stringValue()
		if err != nil {
			return nil, fmt.Errorf("reason: %w", err)
		}
		out = append(out, ClosedDateInput{Date: date, Reason: reason})
	}
	return out, nil
}

type storedJSONKind int

const (
	storedJSONNull storedJSONKind = iota
	storedJSONBool
	storedJSONNumber
	storedJSONString
	storedJSONArray
	storedJSONObject
)

func (k storedJSONKind) String() string {
	switch k {
	case storedJSONNull:
		return "null"
	case storedJSONBool:
		return "boolean"
	case storedJSONNumber:
		return "number"
	case storedJSONString:
		return "string"
	case storedJSONArray:
		return "array"
	case storedJSONObject:
		return "object"
	default:
		return "unknown"
	}
}

type storedJSONMember struct {
	key   string
	value storedJSONNode
}

// storedJSONNode は重複 key を拒否して復号した JSON 値。map[string]any を使わず、型付きの木で持つ。
type storedJSONNode struct {
	kind    storedJSONKind
	text    string // string の値、または number のリテラル
	elems   []storedJSONNode
	members []storedJSONMember
}

func (n storedJSONNode) array() ([]storedJSONNode, error) {
	if n.kind != storedJSONArray {
		return nil, fmt.Errorf("expected array, got %s", n.kind)
	}
	return n.elems, nil
}

func (n storedJSONNode) stringValue() (string, error) {
	if n.kind != storedJSONString {
		return "", fmt.Errorf("expected string, got %s", n.kind)
	}
	return n.text, nil
}

// integer は正規の十進整数リテラル（先頭 0・符号・小数・指数を持たない）だけを受ける。
func (n storedJSONNode) integer() (int, error) {
	if n.kind != storedJSONNumber {
		return 0, fmt.Errorf("expected integer, got %s", n.kind)
	}
	v, err := strconv.Atoi(n.text)
	if err != nil || strconv.Itoa(v) != n.text {
		return 0, fmt.Errorf("expected canonical integer, got %q", n.text)
	}
	return v, nil
}

// exactObject は keys をちょうど1回ずつ含むオブジェクトの値を keys の順で返す。未知 key・欠落 key は拒否する。
func (n storedJSONNode) exactObject(keys ...string) ([]storedJSONNode, error) {
	if n.kind != storedJSONObject {
		return nil, fmt.Errorf("expected object, got %s", n.kind)
	}
	if len(n.members) != len(keys) {
		return nil, fmt.Errorf("expected exactly keys %v", keys)
	}
	out := make([]storedJSONNode, len(keys))
	for i, key := range keys {
		found := false
		for _, member := range n.members {
			if member.key == key {
				out[i] = member.value
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("missing key %q", key)
		}
	}
	return out, nil
}

// parseStoredJSON は raw をちょうど1つの JSON 値として復号する。
// 空入力・複数 JSON 値・末尾データ・オブジェクトの重複 key はエラー。
func parseStoredJSON(raw []byte) (storedJSONNode, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	node, err := decodeStoredJSONValue(dec, 0)
	if err != nil {
		return storedJSONNode{}, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return storedJSONNode{}, errors.New("trailing data after JSON value")
	}
	return node, nil
}

func decodeStoredJSONValue(dec *json.Decoder, depth int) (storedJSONNode, error) {
	if depth > storedJSONMaxDepth {
		return storedJSONNode{}, errors.New("JSON nesting is too deep")
	}
	tok, err := dec.Token()
	if err != nil {
		return storedJSONNode{}, fmt.Errorf("invalid JSON: %w", err)
	}
	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			return decodeStoredJSONObject(dec, depth)
		case '[':
			return decodeStoredJSONArray(dec, depth)
		default:
			return storedJSONNode{}, fmt.Errorf("unexpected delimiter %q", v.String())
		}
	case string:
		return storedJSONNode{kind: storedJSONString, text: v}, nil
	case json.Number:
		return storedJSONNode{kind: storedJSONNumber, text: v.String()}, nil
	case bool:
		return storedJSONNode{kind: storedJSONBool}, nil
	case nil:
		return storedJSONNode{kind: storedJSONNull}, nil
	default:
		return storedJSONNode{}, fmt.Errorf("unexpected JSON token %T", tok)
	}
}

func decodeStoredJSONObject(dec *json.Decoder, depth int) (storedJSONNode, error) {
	node := storedJSONNode{kind: storedJSONObject}
	seen := make(map[string]struct{})
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return storedJSONNode{}, fmt.Errorf("invalid JSON: %w", err)
		}
		key, ok := keyTok.(string)
		if !ok {
			return storedJSONNode{}, errors.New("object key is not a string")
		}
		if _, dup := seen[key]; dup {
			return storedJSONNode{}, fmt.Errorf("duplicate key %q", key)
		}
		seen[key] = struct{}{}
		value, err := decodeStoredJSONValue(dec, depth+1)
		if err != nil {
			return storedJSONNode{}, err
		}
		node.members = append(node.members, storedJSONMember{key: key, value: value})
	}
	if _, err := dec.Token(); err != nil { // closing '}'
		return storedJSONNode{}, fmt.Errorf("invalid JSON: %w", err)
	}
	return node, nil
}

func decodeStoredJSONArray(dec *json.Decoder, depth int) (storedJSONNode, error) {
	node := storedJSONNode{kind: storedJSONArray}
	for dec.More() {
		elem, err := decodeStoredJSONValue(dec, depth+1)
		if err != nil {
			return storedJSONNode{}, err
		}
		node.elems = append(node.elems, elem)
	}
	if _, err := dec.Token(); err != nil { // closing ']'
		return storedJSONNode{}, fmt.Errorf("invalid JSON: %w", err)
	}
	return node, nil
}
