package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
)

var repoPath = "d:\\Documents\\Git\\OWB"
var branch = "workshop-build"
var language = "l_english"
var utf8bom = []byte{0xEF, 0xBB, 0xBF}
var yml *ptool.TParser
var locMapTag = make(map[string]map[string]Localisation)
var locMapHead = make(map[string]map[string]Localisation)
var err error

type Localisation struct {
	Key    string
	Number string
	Value  string
}

var ymlRule = `
	entry                = @language#':' {@pair};

	pair                 = @key ':' [@number] @value [@comment];
	comment              = '#'#anyRune#{#!\x0a#!\x0d#!$#anyRune};

	language             = 'l_'#symbol#{#symbol};
	key                  = symbol#{#symbol};
	number               = digit#{#digit};
	value                = '"'#{#anyRune};

	                     = {spaces|@comment};
	spaces               = \x00..\x20;
	anyRune              = \x00..\x09|\x0b..\x0c|\x0e..$;
	digit                = '0'..'9';
	letter               = 'a'..'z'|'A'..'Z';
	symbol               = digit|letter|'_'|'@'|'.'|'-';
	empty                = '';
`

func main() {
	// r, err := git.PlainOpen(repoPath)
	// if err != nil {
	// 	panic(err)
	// }

	// iter, err := r.Tags()
	// if err != nil {
	// 	panic(err)
	// }

	// var latestTag *object.Tag
	// var date time.Time

	// if err := iter.ForEach(func(ref *plumbing.Reference) error {
	// 	obj, err := r.TagObject(ref.Hash())
	// 	switch err {
	// 	case nil:
	// 		if obj.Tagger.When.After(date) {
	// 			latestTag = obj
	// 			date = obj.Tagger.When
	// 		}
	// 	case plumbing.ErrObjectNotFound:
	// 		// skip
	// 	default:
	// 		return err
	// 	}
	// 	return nil
	// }); err != nil {
	// 	panic(err)
	// }

	// w, err := r.Worktree()
	// if err != nil {
	// 	panic(err)
	// }

	// err = w.Reset(&git.ResetOptions{
	// 	Commit: latestTag.Target,
	// 	Mode:   git.HardReset,
	// })
	// if err != nil {
	// 	panic(err)
	// }

	yml, err = ptool.NewBuilder().FromString(ymlRule).Entries("entry").Build()
	if err != nil {
		panic(err)
	}

	err = parseLoc(locMapTag)
	if err != nil {
		panic(err)
	}

	// fmt.Println(locMapTag[language])
}

func readFile(path string) (string, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(f), nil
}

func parseLoc(locMap map[string]map[string]Localisation) error {
	locFiles, err := filepath.Glob(filepath.Join(repoPath, "oldworldblues", "localisation", "*.yml"))
	if err != nil {
		return err
	}

	locReplaceFiles, err := filepath.Glob(filepath.Join(repoPath, "oldworldblues", "localisation", "replace", "*.yml"))
	if err != nil {
		return err
	}

	locFiles = append(locFiles, locReplaceFiles...)

	for _, lPath := range locFiles {
		f, err := readFile(lPath)
		if err != nil {
			return err
		}
		if len(f) > 0 {
			// Remove utf-8 bom if found.
			if bytes.HasPrefix([]byte(f), utf8bom) {
				f = string(bytes.TrimPrefix([]byte(f), utf8bom))
			}
			// Skip file if it contains a wrong language.
			if !strings.HasPrefix(strings.TrimSpace(f), language) {
				continue
			}
			// fmt.Println(lPath)
			node, err := yml.Parse(f)
			if err != nil {
				return err
			}
			_ = node
			// fmt.Println(ptool.TreeToString(node, yml.ByID))
			err = traverseLoc(locMap, node)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func traverseLoc(locMap map[string]map[string]Localisation, root *ptool.TNode) error {
	lang := "l_english"
	for _, node := range root.Links {
		nodeType := yml.ByID(node.Type)
		switch nodeType {
		case "language":
			lang = node.Value
			if _, ok := locMap[lang]; !ok {
				locMap[lang] = make(map[string]Localisation)
			}
		case "pair":
			var l Localisation
			for _, link := range node.Links {
				nodeType := yml.ByID(link.Type)
				switch nodeType {
				case "key":
					l.Key = link.Value
				case "number":
					l.Number = link.Value
				case "value":
					l.Value = trimQuotes(link.Value)
				}
			}
			if _, ok := locMap[lang][l.Key]; ok {
				fmt.Printf("duplicate loc key found: %q\n", l.Key)
			}
			locMap[lang][l.Key] = l
		default:
			err := traverseLoc(locMap, node)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}
