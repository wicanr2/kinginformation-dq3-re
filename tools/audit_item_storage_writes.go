// 物品儲存寫入候選稽核。入口 docs/188；這不是型別或資料流證明。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

type candidate struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	Kind     string `json:"kind"`
	Code     string `json:"code"`
}

type source struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

var storageFields = map[string]bool{
	"inventory": true, "equip": true, "Inventory": true, "Equip": true,
	"Weapon": true, "Armor": true, "Shield": true, "Head": true,
	"heroItems": true,
}

var borrowHelpers = map[string]bool{
	"equipActorInventory": true, "equipActorSlots": true,
	"actorItems": true,
}

func containsStorage(node ast.Node, aliases map[*ast.Object]bool) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.SelectorExpr:
			found = found || storageFields[v.Sel.Name]
		case *ast.Ident:
			found = found || v.Obj != nil && aliases[v.Obj]
		case *ast.CallExpr:
			if fn, ok := v.Fun.(*ast.SelectorExpr); ok {
				found = found || borrowHelpers[fn.Sel.Name]
			}
		}
		return !found
	})
	return found
}

func representation(set *token.FileSet, node ast.Node) string {
	var output bytes.Buffer
	if err := printer.Fprint(&output, set, node); err != nil {
		panic(err)
	}
	return output.String()
}

func scanFunction(set *token.FileSet, file string, fn *ast.FuncDecl) []candidate {
	if fn.Body == nil {
		return nil
	}
	aliases := map[*ast.Object]bool{}
	// 固定點只尋找可能別名；複製的值也可能入選，需逐筆人工分類。
	for changed := true; changed; {
		changed = false
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i, target := range assignment.Lhs {
				id, ok := target.(*ast.Ident)
				if !ok || id.Obj == nil || i >= len(assignment.Rhs) || aliases[id.Obj] {
					continue
				}
				if containsStorage(assignment.Rhs[i], aliases) {
					aliases[id.Obj], changed = true, true
				}
			}
			return true
		})
	}
	var result []candidate
	appendRow := func(node ast.Node, kind string) {
		result = append(result, candidate{file, set.Position(node.Pos()).Line, fn.Name.Name, kind, representation(set, node)})
	}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch v := node.(type) {
		case *ast.AssignStmt:
			for _, target := range v.Lhs {
				if containsStorage(target, aliases) {
					appendRow(v, "assignment_candidate")
					break
				}
				// 所有解參照賦值均保留，即使未找到別名來源。
				if _, ok := target.(*ast.StarExpr); ok {
					appendRow(v, "unresolved_pointer_write")
					break
				}
				itemTable := false
				ast.Inspect(target, func(n ast.Node) bool {
					if selector, ok := n.(*ast.SelectorExpr); ok && selector.Sel.Name == "items" {
						itemTable = true
					}
					return !itemTable
				})
				if itemTable {
					appendRow(v, "item_table_or_battle_assignment")
					break
				}
			}
		case *ast.IncDecStmt:
			if containsStorage(v.X, aliases) {
				appendRow(v, "increment_candidate")
			}
		case *ast.KeyValueExpr:
			if key, ok := v.Key.(*ast.Ident); ok && storageFields[key.Name] {
				appendRow(v, "construction_or_snapshot")
			} else if ok && key.Name == "items" {
				appendRow(v, "item_table_or_battle_construction")
			}
		case *ast.CallExpr:
			if selector, ok := v.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "setEquipActorSlot" {
				appendRow(v, "equipment_setter_call")
			}
		}
		return true
	})
	return result
}

func main() {
	if len(os.Args) != 3 {
		panic("usage: audit_item_storage_writes <repo> <new-receipt.json>")
	}
	repo := os.Args[1]
	directory := filepath.Join(repo, "dq3_remake_ebitan", "game")
	entries, err := os.ReadDir(directory)
	if err != nil {
		panic(err)
	}
	set := token.NewFileSet()
	var inputs []source
	var rows []candidate
	for _, entry := range entries {
		name := entry.Name()
		if !entry.Type().IsRegular() || filepath.Ext(name) != ".go" || len(name) >= 8 && name[len(name)-8:] == "_test.go" {
			continue
		}
		path := filepath.Join(directory, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		hash := sha256.Sum256(raw)
		inputs = append(inputs, source{name, hex.EncodeToString(hash[:])})
		parsed, err := parser.ParseFile(set, path, raw, 0)
		if err != nil {
			panic(err)
		}
		for _, decl := range parsed.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				rows = append(rows, scanFunction(set, name, fn)...)
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].File != rows[j].File {
			return rows[i].File < rows[j].File
		}
		if rows[i].Line != rows[j].Line {
			return rows[i].Line < rows[j].Line
		}
		return rows[i].Kind < rows[j].Kind
	})
	toolSource, err := os.ReadFile(filepath.Join(repo, "tools", "audit_item_storage_writes.go"))
	if err != nil {
		panic(err)
	}
	toolHash := sha256.Sum256(toolSource)
	receipt := struct {
		Scope      string      `json:"scope"`
		GoVersion  string      `json:"go_version"`
		ToolSHA256 string      `json:"tool_sha256"`
		Inputs     []source    `json:"inputs"`
		Candidates []candidate `json:"candidates"`
	}{"non-test regular Go files in game; syntax candidates, not exhaustive alias/type proof or original parity", runtime.Version(), hex.EncodeToString(toolHash[:]), inputs, rows}
	output, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(receipt); err != nil {
		panic(err)
	}
	if err := output.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("%d inputs, %d syntax candidates; semantic review required\n", len(inputs), len(rows))
}
