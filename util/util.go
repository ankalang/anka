package util

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ankalang/anka/object"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func IsNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)

	return err == nil
}

func ExpandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func GetEnvVar(env *object.Environment, varName, defaultVal string) string {
	var ok bool
	var value string
	valueObj, ok := env.Get(varName)
	if ok {
		value = valueObj.Inspect()
	} else {
		value = os.Getenv(varName)
		if len(value) == 0 {
			value = defaultVal
		}
	}
	return value
}

func InterpolateStringVars(str string, env *object.Environment) string {

	re := regexp.MustCompile("(\\\\)?\\$(\\{)?([a-zA-Z_0-9]{1,})(\\})?")
	str = re.ReplaceAllStringFunc(str, func(m string) string {

		if string(m[0]) == "\\" {
			return m[1:]
		}

		varName := ""
		if m[1] == '{' {

			if m[len(m)-1] != '}' {
				return m
			}

			varName = m[2 : len(m)-1]
		} else {
			varName = m[1:]
		}

		v, ok := env.Get(varName)

		if !ok {
			return ""
		}
		return v.Inspect()
	})

	return str
}

func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func UnaliasPath(path string, packageAlias map[string]string) string {

	parts := strings.Split(path, string(os.PathSeparator))

	if len(parts) < 1 {
		return path
	}

	if packageAlias[parts[0]] != "" {

		p := []string{packageAlias[parts[0]]}
		p = append(p, parts[1:]...)
		path = filepath.Join(p...)
	}
	return appendIndexFile(path)
}


func appendIndexFile(path string) string {
	if filepath.Ext(path) != ".ank" {
		return filepath.Join(path, "index.ank")
	}

	return path
}

func Mapify(list []object.Object) map[string]object.Object {
	m := make(map[string]object.Object)

	for _, v := range list {
		m[object.GenerateEqualityString(v)] = v
	}

	return m
}
