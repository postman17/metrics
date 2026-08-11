// Package analyzers содержит пользовательские анализаторы Go-кода
// для использования в staticlint.
//
// # noosexit
//
// Анализатор NoOsExitAnalyzer запрещает прямой вызов os.Exit
// в функции main пакета main.
//
// ## Обоснование
//
// Вызов os.Exit завершает программу немедленно, минуя deferred-функции
// текущей goroutine. Это может привести к потере данных, не закрытым
// файловым дескрипторам и пропущенным отчётам telemetry.
//
// ## Рекомендуемый паттерн
//
// Вместо прямого os.Exit в main следует вынести логику в отдельную функцию
// (обычно run), которая возвращает ошибку:
//
//	func main() {
//	    if err := run(); err != nil {
//	        log.Error("fatal", "err", err)
//	        os.Exit(1) // допустимо: defer в run уже отработал
//	    }
//	}
//
//	func run() error {
//	    // вся логика здесь; deferred-функции сработают при return
//	    ...
//	}
//
// Или, если os.Exit вообще не нужен, завершить программу через return из main.
//
// ## Примечания
//
//   - Анализатор проверяет только пакет main и только функцию с именем "main".
//   - os.Exit в других функциях (включая goroutine внутри main) разрешён.
//   - Файлы из стандартной библиотеки и vendor/ пропускаются.
package analyzers

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "checks that os.Exit is not called directly in main function of main package",
	Run:  run,
}

func isStdlibOrVendor(fset *token.FileSet, pos token.Pos) bool {
	f := fset.File(pos)
	if f == nil {
		return true
	}
	name := f.Name()
	return strings.Contains(name, "/go-build/") ||
		strings.Contains(name, "GOMODCACHE") ||
		(!strings.Contains(name, "/src/") && !strings.Contains(name, "/projects/"))
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		if isStdlibOrVendor(pass.Fset, file.Pos()) {
			continue
		}

		fileScope := pass.TypesInfo.Scopes[file]

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if fn.Name.Name != "main" {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				fnObj, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
				if !ok {
					return true
				}

				if fnObj.FullName() != "os.Exit" {
					return true
				}

				callScope := fileScope.Innermost(call.Pos())
				mainScope := fileScope.Innermost(fn.Pos())
				if callScope == mainScope || isChildScope(callScope, mainScope) {
					pass.Reportf(call.Pos(), "direct call of os.Exit in main function of main package is forbidden")
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}

func isChildScope(child, parent *types.Scope) bool {
	for s := child; s != nil; s = s.Parent() {
		if s == parent {
			return true
		}
	}
	return false
}

