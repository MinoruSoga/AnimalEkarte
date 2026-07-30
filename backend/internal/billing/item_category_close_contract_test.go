package billing

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func declaredItemCategories(t *testing.T) map[string]string {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "../model/accounting.go", nil, 0)
	require.NoError(t, err)

	declarations := make(map[string]string)
	for _, declaration := range file.Decls {
		generic, ok := declaration.(*ast.GenDecl)
		if !ok || generic.Tok != token.CONST {
			continue
		}
		for _, specification := range generic.Specs {
			values, ok := specification.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for index, name := range values.Names {
				if !strings.HasPrefix(name.Name, "ItemCategory") {
					continue
				}
				require.Greater(t, len(values.Values), index, "category constant %s must have an explicit value", name.Name)
				literal, ok := values.Values[index].(*ast.BasicLit)
				require.True(t, ok, "category constant %s must use a string literal", name.Name)
				require.Equal(t, token.STRING, literal.Kind, "category constant %s must use a string literal", name.Name)
				value, err := strconv.Unquote(literal.Value)
				require.NoError(t, err)
				declarations[name.Name] = value
			}
		}
	}
	require.NotEmpty(t, declarations)
	return declarations
}

func validatorItemCategorySelectors(t *testing.T) map[string]struct{} {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "validators_billing_item.go", nil, 0)
	require.NoError(t, err)

	selectors := make(map[string]struct{})
	foundValidator := false
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "validateItemCategory" {
			continue
		}
		foundValidator = true
		ast.Inspect(function.Body, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name == "ItemCategory" || !strings.HasPrefix(selector.Sel.Name, "ItemCategory") {
				return true
			}
			packageName, ok := selector.X.(*ast.Ident)
			if ok && packageName.Name == "model" {
				selectors[selector.Sel.Name] = struct{}{}
			}
			return true
		})
	}

	require.True(t, foundValidator)
	require.NotEmpty(t, selectors)
	return selectors
}

func TestCloseCategoryContract(t *testing.T) {
	t.Run("source declarations helper and validator stay exactly consistent", func(t *testing.T) {
		declarations := declaredItemCategories(t)
		declaredNames := make(map[string]struct{}, len(declarations))
		declaredValues := make(map[string]struct{}, len(declarations))
		for name, value := range declarations {
			declaredNames[name] = struct{}{}
			declaredValues[value] = struct{}{}
		}

		helperValues := make(map[string]struct{}, len(model.AllItemCategories()))
		for _, category := range model.AllItemCategories() {
			helperValues[string(category)] = struct{}{}
		}

		assert.Equal(t, declaredValues, helperValues)
		assert.Equal(t, declaredNames, validatorItemCategorySelectors(t))
	})

	t.Run("canonical categories stay exhaustive and validator-consistent", func(t *testing.T) {
		want := []model.ItemCategory{
			model.ItemCategoryExamination,
			model.ItemCategoryTest,
			model.ItemCategoryProcedure,
			model.ItemCategorySurgery,
			model.ItemCategoryMedicine,
			model.ItemCategoryFood,
			model.ItemCategoryGoods,
			model.ItemCategoryOther,
			model.ItemCategoryVaccine,
			model.ItemCategoryTrimming,
			model.ItemCategoryHotel,
			model.ItemCategoryTraining,
		}

		got := model.AllItemCategories()
		require.Equal(t, want, got)

		seen := make(map[model.ItemCategory]struct{}, len(got))
		for _, category := range got {
			require.NotContains(t, seen, category)
			seen[category] = struct{}{}
			require.NoError(t, validateItemCategory(string(category)))
		}
		for _, unknown := range []string{"", "exam", "診察"} {
			require.Error(t, validateItemCategory(unknown))
		}

		got[0] = model.ItemCategoryOther
		assert.Equal(t, model.ItemCategoryExamination, model.AllItemCategories()[0])
	})

	t.Run("repository converter preserves valid rows and rejects unknown categories", func(t *testing.T) {
		input := []closeCategoryRow{
			{Category: string(model.ItemCategoryExamination), Amount: 1001},
			{Category: string(model.ItemCategoryTest), Amount: 2999},
			{Category: string(model.ItemCategoryExamination), Amount: -50},
		}

		got, err := toCategoryAggregateRows(input)
		require.NoError(t, err)
		assert.Equal(t, []CategoryAggregateRow{
			{Category: string(model.ItemCategoryExamination), Amount: 1001},
			{Category: string(model.ItemCategoryTest), Amount: 2999},
			{Category: string(model.ItemCategoryExamination), Amount: -50},
		}, got)

		got, err = toCategoryAggregateRows([]closeCategoryRow{
			{Category: string(model.ItemCategoryExamination), Amount: 100},
			{Category: "exam", Amount: 200},
		})
		require.EqualError(t, err, `unknown item category in close aggregate: "exam"`)
		assert.Nil(t, got)
	})

	t.Run("preview allocation preserves conservation and rejects unknown categories (#247)", func(t *testing.T) {
		// #247: matrix is payment-net based (not category gross * period ratio).
		// Single category, payments 1+2 → cells equal payment amounts (sum=3).
		matrix := map[string]map[string]int64{
			string(model.ItemCategoryExamination): {
				"現金":  1,
				"カード": 2,
			},
		}
		got, err := buildPreviewCategories(matrix)
		require.NoError(t, err)
		assert.Equal(t, map[string]map[string]int64{
			string(model.ItemCategoryExamination): {
				"現金":  1,
				"カード": 2,
			},
		}, got)
		assert.Equal(t, int64(3), MatrixGrandTotal(got), "preview matrix conserves payment net")

		bad := map[string]map[string]int64{
			string(model.ItemCategoryExamination): {"現金": 1},
			"exam":                                {"現金": 2},
		}
		got, err = buildPreviewCategories(bad)
		require.EqualError(t, err, `unknown item category in close aggregate: "exam"`)
		assert.Nil(t, got)
	})
}
