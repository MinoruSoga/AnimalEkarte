package handler

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	trimmingdomain "github.com/animal-ekarte/backend/internal/trimming"
)

func TestTrimmingResponseTygoCarrierMatchesRuntimeDTO(t *testing.T) {
	t.Parallel()

	carrier := responseShape(reflect.TypeOf(TrimmingResponse{}))
	runtime := responseShape(reflect.TypeOf(trimmingdomain.TrimmingResponse{}))
	require.Equal(t, runtime, carrier,
		"legacy tygo carrier must match the runtime trimming DTO until the BE9-2F codegen move")
}

func responseShape(value reflect.Type) string {
	switch value.Kind() {
	case reflect.Pointer:
		return "*" + responseShape(value.Elem())
	case reflect.Slice:
		return "[]" + responseShape(value.Elem())
	case reflect.Struct:
		if value.PkgPath() == "time" && value.Name() == "Time" {
			return "time.Time"
		}
		var shape strings.Builder
		shape.WriteString("struct{")
		for i := range value.NumField() {
			field := value.Field(i)
			shape.WriteString(field.Name)
			shape.WriteByte(':')
			shape.WriteString(responseShape(field.Type))
			shape.WriteByte(':')
			shape.WriteString(string(field.Tag))
			shape.WriteByte(';')
		}
		shape.WriteByte('}')
		return shape.String()
	default:
		return value.String()
	}
}
