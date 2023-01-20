package evaluator

import (
	"bufio"
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	mrand "math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ankalang/anka/ast"
	"github.com/ankalang/anka/lexer"
	"github.com/ankalang/anka/object"
	"github.com/ankalang/anka/parser"
	"github.com/ankalang/anka/token"
	"github.com/ankalang/anka/util"
)

var scanner *bufio.Scanner
var tok token.Token
var scannerPosition int
var requireCache map[string]object.Object

func init() {
	scanner = bufio.NewScanner(os.Stdin)
	requireCache = make(map[string]object.Object)
}


func getFns() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		
		"uzunluk": &object.Builtin{
			Types: []string{object.STRING_OBJ, object.ARRAY_OBJ},
			Fn:    lenFn,
		},
		
		"rast": &object.Builtin{
			Types: []string{object.NUMBER_OBJ},
			Fn:    randFn,
		},
		
		"çıkış": &object.Builtin{
			Types: []string{object.NUMBER_OBJ},
			Fn:    exitFn,
		},
		
		"bayrak": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    flagFn,
		},
		
		"pwd": &object.Builtin{
			Types: []string{},
			Fn:    pwdFn,
		},
		
		"cd": &object.Builtin{
			Types: []string{},
			Fn:    cdFn,
		},
		
		"eko": &object.Builtin{
			Types: []string{},
			Fn:    echoFn,
		},
		
		
		"int": &object.Builtin{
			Types: []string{object.STRING_OBJ, object.NUMBER_OBJ},
			Fn:    intFn,
		},
		
		
		"yuvarla": &object.Builtin{
			Types: []string{object.STRING_OBJ, object.NUMBER_OBJ},
			Fn:    roundFn,
		},
		
		
		"floor": &object.Builtin{
			Types: []string{object.STRING_OBJ, object.NUMBER_OBJ},
			Fn:    floorFn,
		},
		
		
		"ceil": &object.Builtin{
			Types: []string{object.STRING_OBJ, object.NUMBER_OBJ},
			Fn:    ceilFn,
		},
		
		"num": &object.Builtin{
			Types: []string{object.STRING_OBJ, object.NUMBER_OBJ},
			Fn:    numberFn,
		},
		
		"sayımı": &object.Builtin{
			Types: []string{object.STRING_OBJ, object.NUMBER_OBJ},
			Fn:    isNumberFn,
		},
		
		"girdi": &object.Builtin{
			Next:  stdinNextFn,
			Types: []string{},
			Fn:    stdinFn,
		},
		
		"env": &object.Builtin{
			Types: []string{},
			Fn:    envFn,
		},
		
		"arg": &object.Builtin{
			Types: []string{object.NUMBER_OBJ},
			Fn:    argFn,
		},
		
		"args": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    argsFn,
		},
		
		"tip": &object.Builtin{
			Types: []string{},
			Fn:    typeFn,
		},
		
		"ara": &object.Builtin{
			Types: []string{object.FUNCTION_OBJ, object.BUILTIN_OBJ},
			Fn:    callFn,
		},
		
		"chunk": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    chunkFn,
		},
		
		"ayır": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    splitFn,
		},
		
		"satırlar": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    linesFn,
		},
		
		
		"json": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    jsonFn,
		},
		
		"fmt": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    fmtFn,
		},
		
		"toplam": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    sumFn,
		},
		
		"max": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    maxFn,
		},
		
		"min": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    minFn,
		},
		
		"azalt": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    reduceFn,
		},
		
		"sırala": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    sortFn,
		},
		
		"kes": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    intersectFn,
		},
		
		"fark": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    diffFn,
		},
		
		"birleştir": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    unionFn,
		},
		
		"s_fark": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    diffSymmetricFn,
		},
		
		"düzleştir": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    flattenFn,
		},
		
		"d_düzleştir": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    flattenDeepFn,
		},
		
		"böl": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    partitionFn,
		},
		
		"haritala": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    mapFn,
		},
		
		"bazısında": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    someFn,
		},
		
		"hepsinde": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    everyFn,
		},
		
		"bul": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    findFn,
		},
		
		"filtre": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    filterFn,
		},
		
		"eşsiz": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    uniqueFn,
		},
		
		"str": &object.Builtin{
			Types: []string{},
			Fn:    strFn,
		},
		
		"herhangi": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    anyFn,
		},
		
		"arasında": &object.Builtin{
			Types: []string{object.NUMBER_OBJ},
			Fn:    betweenFn,
		},
		
		"önek": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    prefixFn,
		},
		
		"sonek": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    suffixFn,
		},
		
		"tekrarla": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    repeatFn,
		},
		
		"değiştir": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    replaceFn,
		},
		
		"başlık": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    titleFn,
		},
		
		"küçük": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    lowerFn,
		},
		
		"büyük": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    upperFn,
		},
		
		"bekleyerek": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    waitFn,
		},
		"öldür": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    killFn,
		},
		
		"kırp": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    trimFn,
		},
		
		"göre_kırp": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    trimByFn,
		},
		
		"dizin": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    indexFn,
		},
		
		"son_dizin": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    lastIndexFn,
		},
		
		"shift": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    shiftFn,
		},
		
		"tersine": &object.Builtin{
			Types: []string{object.ARRAY_OBJ, object.STRING_OBJ},
			Fn:    reverseFn,
		},
		
		"karıştır": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    shuffleFn,
		},
		
		"it": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    pushFn,
		},
		
		"çıkar": &object.Builtin{
			Types: []string{object.ARRAY_OBJ, object.HASH_OBJ},
			Fn:    popFn,
		},
		
		
		"anahtarlar": &object.Builtin{
			Types: []string{object.ARRAY_OBJ, object.HASH_OBJ},
			Fn:    keysFn,
		},
		
		"değerler": &object.Builtin{
			Types: []string{object.HASH_OBJ},
			Fn:    valuesFn,
		},
		
		"eşyalar": &object.Builtin{
			Types: []string{object.HASH_OBJ},
			Fn:    itemsFn,
		},
		
		"kat": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    joinFn,
		},
		
		"uyu": &object.Builtin{
			Types: []string{object.NUMBER_OBJ},
			Fn:    sleepFn,
		},
		
		"kaynak": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    sourceFn,
		},
		
		"src": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    requireFn,
		},
		
		"uygula": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    execFn,
		},
		
		"eval": &object.Builtin{
			Types: []string{object.STRING_OBJ},
			Fn:    evalFn,
		},
		
		"tsv": &object.Builtin{
			Types: []string{object.ARRAY_OBJ},
			Fn:    tsvFn,
		},
		
		"unix_ms": &object.Builtin{
			Types: []string{},
			Fn:    unixMsFn,
		},
	}
}



func validateArgs(tok token.Token, name string, args []object.Object, size int, types [][]string) object.Object {
	if len(args) == 0 || len(args) > size || len(args) < size {
		return newError(tok, "%s(...) için yanlış sayıda argüman: bulunan=%d, istenilen=%d", name, len(args), size)
	}

	for i, t := range types {
		if !util.Contains(t, string(args[i].Type())) && !util.Contains(t, object.ANY_OBJ) {
			return newError(tok, "%d argümanı %s(...) için destekli değildir (bulunan: %s, desteklenen: %s)", i, name, args[i].Inspect(), strings.Join(t, ", "))
		}
	}

	return nil
}










func validateVarArgs(tok token.Token, name string, args []object.Object, specs [][][]string) (object.Object, int) {
	required := -1
	max := 0

	for _, spec := range specs {
		
		if required == -1 || len(spec) < required {
			required = len(spec)
		}

		
		if len(spec) > max {
			max = len(spec)
		}
	}

	if len(args) < required || len(args) > max {
		return newError(tok, "%s(...) için yanlış argüman sayısı: bulunan=%d, minimum=%d, maksimum=%d", name, len(args), required, max), -1
	}

	for which, spec := range specs {
		
		if len(args) != len(spec) {
			continue
		}

		
		match := true
		for i, types := range spec {
			if i < len(args) && !util.Contains(types, string(args[i].Type())) {
				match = false
				break
			}
		}

		
		if match {
			return nil, which
		}
	}

	
	return newError(tok, usageVarArgs(name, specs)), -1
}

func usageVarArgs(name string, specs [][][]string) string {
	signatures := []string{ name + "'için yanlış sayıda argüman verildi, kullanımı:"}

	for _, spec := range specs {
		args := []string{}

		for _, types := range spec {
			args = append(args, strings.Join(types, " | "))
		}

		signatures = append(signatures, fmt.Sprintf("%s(%s)", name, strings.Join(args, ", ")))
	}

	return strings.Join(signatures, "\n")
}


func lenFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "len", args, 1, [][]string{{object.STRING_OBJ, object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	switch arg := args[0].(type) {
	case *object.Array:
		return &object.Number{Token: tok, Value: float64(len(arg.Elements))}
	case *object.String:
		return &object.Number{Token: tok, Value: float64(len(arg.Value))}
	default:
		return newError(tok, "len içim argüman yok, bulunan %s", args[0].Type())
	}
}


func randFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "rand", args, 1, [][]string{{object.NUMBER_OBJ}})
	if err != nil {
		return err
	}

	arg := args[0].(*object.Number)
	r, e := rand.Int(rand.Reader, big.NewInt(int64(arg.Value)))

	if e != nil {
		return newError(tok, "'rast(%v)' çağırırken hata oluştu: %s", arg.Value, e.Error())
	}

	return &object.Number{Token: tok, Value: float64(r.Int64())}
}



func exitFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	var err object.Object
	var message string

	if len(args) == 2 {
		err = validateArgs(tok, "exit", args, 2, [][]string{{object.NUMBER_OBJ}, {object.STRING_OBJ}})
		message = args[1].(*object.String).Value
	} else {
		err = validateArgs(tok, "exit", args, 1, [][]string{{object.NUMBER_OBJ}})
	}

	if err != nil {
		return err
	}

	if message != "" {
		fmt.Fprintf(env.Writer, message)
	}

	arg := args[0].(*object.Number)
	os.Exit(int(arg.Value))
	return arg
}


func unixMsFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	return &object.Number{Value: float64(time.Now().UnixNano() / 1000000)}
}


func flagFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	
	
	
	

	err := validateArgs(tok, "flag", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	
	name := args[0].(*object.String)
	found := false

	
	
	
	
	for _, v := range os.Args {
		
		
		if found {
			
			
			
			if strings.HasPrefix(v, "-") {
				break
			}

			
			
			return &object.String{Token: tok, Value: v}
		}

		
		parts := strings.SplitN(v, "=", 2)
		
		left := parts[0]

		
		
		
		if (len(left) > 1 && left[1:] == name.Value) || (len(left) > 2 && left[2:] == name.Value) {
			if len(parts) > 1 {
				return &object.String{Token: tok, Value: parts[1]}
			} else {
				found = true
			}
		}
	}

	
	
	
	if found {
		return object.TRUE
	}

	
	return NULL
}


func pwdFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	dir, err := os.Getwd()
	if err != nil {
		return newError(tok, err.Error())
	}
	return &object.String{Token: tok, Value: dir}
}



func cdFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	user, ok := user.Current()
	if ok != nil {
		return newError(tok, ok.Error())
	}
	
	path := user.HomeDir
	if len(args) == 1 {
		
		pathStr := args[0].(*object.String)
		rawPath := pathStr.Value
		path, _ = util.ExpandPath(rawPath)
	}
	
	error := os.Chdir(path)
	if error != nil {
		
		return &object.String{Token: tok, Value: error.Error(), Ok: &object.Boolean{Token: tok, Value: false}}
	}
	
	
	dir, _ := os.Getwd()
	return &object.String{Token: tok, Value: dir, Ok: &object.Boolean{Token: tok, Value: true}}
}




func echoFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	if len(args) == 0 {
		
		fmt.Fprintln(env.Writer, "")
		return NULL
	}
	var arguments []interface{} = make([]interface{}, len(args)-1)
	for i, d := range args {
		if i > 0 {
			arguments[i-1] = d.Inspect()
		}
	}

	fmt.Fprintf(env.Writer, args[0].Inspect(), arguments...)
	fmt.Fprintln(env.Writer, "")

	return NULL
}



func intFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "int", args, 1, [][]string{{object.NUMBER_OBJ, object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return applyMathFunction(tok, args[0], func(n float64) float64 {
		return float64(int64(n))
	}, "int")
}



func roundFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	
	err := validateArgs(tok, "round", args[:1], 1, [][]string{{object.NUMBER_OBJ, object.STRING_OBJ}})
	if err != nil {
		return err
	}

	decimal := float64(1)

	
	if len(args) > 1 {
		err := validateArgs(tok, "round", args[1:], 1, [][]string{{object.NUMBER_OBJ}})
		if err != nil {
			return err
		}

		decimal = float64(math.Pow(10, args[1].(*object.Number).Value))
	}

	return applyMathFunction(tok, args[0], func(n float64) float64 {
		return math.Round(n*decimal) / decimal
	}, "round")
}



func floorFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "floor", args, 1, [][]string{{object.NUMBER_OBJ, object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return applyMathFunction(tok, args[0], math.Floor, "floor")
}



func ceilFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "ceil", args, 1, [][]string{{object.NUMBER_OBJ, object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return applyMathFunction(tok, args[0], math.Ceil, "ceil")
}







func applyMathFunction(tok token.Token, arg object.Object, fn func(float64) float64, fname string) object.Object {
	switch arg := arg.(type) {
	case *object.Number:
		return &object.Number{Token: tok, Value: float64(fn(arg.Value))}
	case *object.String:
		i, err := strconv.ParseFloat(arg.Value, 64)

		if err != nil {
			return newError(tok, "%s(...) sadece yazılarla kullanılabilir, '%s' değil", fname, arg.Value)
		}

		return &object.Number{Token: tok, Value: float64(fn(i))}
	default:
		
		
		return newError(tok, "`%s` için verilen argüman destekli değil. %s", fname, arg.Type())
	}
}


func numberFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "number", args, 1, [][]string{{object.NUMBER_OBJ, object.STRING_OBJ}})
	if err != nil {
		return err
	}

	switch arg := args[0].(type) {
	case *object.Number:
		return arg
	case *object.String:
		i, err := strconv.ParseFloat(arg.Value, 64)

		if err != nil {
			return newError(tok, "number(...) sadece rakam içeren yazılar ile kullanılabilir, '%s' ile değil", arg.Value)
		}

		return &object.Number{Token: tok, Value: i}
	default:
		
		return newError(tok, "garip bir hata oluştu. Kod: 47586665sx7", args[0].Type())
	}
}


func isNumberFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "number", args, 1, [][]string{{object.NUMBER_OBJ, object.STRING_OBJ}})
	if err != nil {
		return err
	}

	switch arg := args[0].(type) {
	case *object.Number:
		return &object.Boolean{Token: tok, Value: true}
	case *object.String:
		return &object.Boolean{Token: tok, Value: util.IsNumber(arg.Value)}
	default:
		
		return newError(tok, "yok, olmadı., bulunan %s", args[0].Type())
	}
}


func stdinFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	v := scanner.Scan()

	if !v {
		return EOF
	}

	return &object.String{Token: tok, Value: scanner.Text()}
}
func stdinNextFn() (object.Object, object.Object) {
	v := scanner.Scan()

	if !v {
		return nil, EOF
	}

	defer func() {
		scannerPosition += 1
	}()
	return &object.Number{Value: float64(scannerPosition)}, &object.String{Token: tok, Value: scanner.Text()}
}


func envFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err, spec := validateVarArgs(tok, "env", args, [][][]string{
		{{object.STRING_OBJ}, {object.STRING_OBJ}},
		{{object.STRING_OBJ}},
	})

	if err != nil {
		return err
	}

	key := args[0].(*object.String)

	if spec == 0 {
		val := args[1].(*object.String)
		os.Setenv(key.Value, val.Value)
	}

	return &object.String{Token: tok, Value: os.Getenv(key.Value)}
}


func argFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "arg", args, 1, [][]string{{object.NUMBER_OBJ}})
	if err != nil {
		return err
	}

	arg := args[0].(*object.Number)
	i := arg.Int()

	if i > len(os.Args)-1 || i < 0 {
		return &object.String{Token: tok, Value: ""}
	}

	return &object.String{Token: tok, Value: os.Args[i]}
}


func argsFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	length := len(os.Args)
	result := make([]object.Object, length, length)

	for i, v := range os.Args {
		result[i] = &object.String{Token: tok, Value: v}
	}

	return &object.Array{Elements: result}
}


func typeFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "type", args, 1, [][]string{})
	if err != nil {
		return err
	}

	return &object.String{Token: tok, Value: string(args[0].Type())}
}


func callFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "call", args, 2, [][]string{{object.FUNCTION_OBJ, object.BUILTIN_OBJ}, {object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	return applyFunction(tok, args[0], env, args[1].(*object.Array).Elements)
}


func chunkFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "chunk", args, 2, [][]string{{object.ARRAY_OBJ}, {object.NUMBER_OBJ}})
	if err != nil {
		return err
	}

	number := args[1].(*object.Number)
	size := int(number.Value)

	if size < 1 || !number.IsInt() {
		return newError(tok, "chunk(...) için sadece sayılar kullanılabilir, bulunan '%s'", number.Inspect())
	}

	var chunks []object.Object
	elements := args[0].(*object.Array).Elements

	for i := 0; i < len(elements); i += size {
		end := i + size

		if end > len(elements) {
			end = len(elements)
		}

		chunks = append(chunks, &object.Array{Elements: elements[i:end]})
	}

	return &object.Array{Elements: chunks}
}


func splitFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err, spec := validateVarArgs(tok, "split", args, [][][]string{
		{{object.STRING_OBJ}, {object.STRING_OBJ}},
		{{object.STRING_OBJ}},
	})

	if err != nil {
		return err
	}

	s := args[0].(*object.String)

	sep := " "
	if spec == 0 {
		sep = args[1].(*object.String).Value
	}

	parts := strings.Split(s.Value, sep)
	length := len(parts)
	elements := make([]object.Object, length, length)

	for k, v := range parts {
		elements[k] = &object.String{Token: tok, Value: v}
	}

	return &object.Array{Elements: elements}
}


func linesFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "lines", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	s := args[0].(*object.String)
	parts := strings.FieldsFunc(s.Value, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\f'
	})
	length := len(parts)
	elements := make([]object.Object, length, length)

	for k, v := range parts {
		elements[k] = &object.String{Token: tok, Value: v}
	}

	return &object.Array{Elements: elements}
}



func jsonFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	
	
	
	
	
	
	
	
	
	

	err := validateArgs(tok, "json", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	s := args[0].(*object.String)
	str := strings.TrimSpace(s.Value)
	env = object.NewEnvironment(env.Writer, env.Dir, env.Version)
	l := lexer.New(str)
	p := parser.New(l)
	var node ast.Node
	ok := false

	
	
	
	
	
	
	
	if len(str) != 0 {
		switch str[0] {
		case '{':
			node, ok = p.ParseHashLiteral().(*ast.HashLiteral)
		case '[':
			node, ok = p.ParseArrayLiteral().(*ast.ArrayLiteral)
		}
	}

	
	
	if len(str) == 0 || (str[0] == '"' && str[len(str)-1] == '"') {
		node, ok = p.ParseStringLiteral().(*ast.StringLiteral)
	}

	if util.IsNumber(str) {
		node, ok = p.ParseNumberLiteral().(*ast.NumberLiteral)
	}

	if str == "false" || str == "true" {
		node, ok = p.ParseBoolean().(*ast.Boolean)
	}

	if str == "null" {
		return NULL
	}

	if ok {
		return Eval(node, env)
	}

	return newError(tok, "`json`'a gelen argüman JSON objesi olmalı, bulunan '%s'", s.Value)

}


func fmtFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	list := []interface{}{}

	for _, s := range args[1:] {
		list = append(list, s.Inspect())
	}

	return &object.String{Token: tok, Value: fmt.Sprintf(args[0].(*object.String).Value, list...)}
}


func sumFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "sum", args, 1, [][]string{{object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	arr := args[0].(*object.Array)
	if arr.Empty() {
		return &object.Number{Token: tok, Value: float64(0)}
	}

	if !arr.Homogeneous() {
		return newError(tok, "ekle(...) sadece homojenik listelerde kullanılabilir, bulunan: %s", arr.Inspect())
	}

	if arr.Elements[0].Type() != object.NUMBER_OBJ {
		return newError(tok, "ekle(...) sadece sayılardan oluşan listelerde kullanılabilir, bulunan: %s", arr.Inspect())
	}

	var sum float64 = 0

	for _, v := range arr.Elements {
		elem := v.(*object.Number)
		sum += elem.Value
	}

	return &object.Number{Token: tok, Value: sum}
}


func maxFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "max", args, 1, [][]string{{object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	arr := args[0].(*object.Array)
	if arr.Empty() {
		return object.NULL
	}

	if !arr.Homogeneous() {
		return newError(tok, "max(...) sadece aynı türden veri içeren bir listelerde kullanılabilir, bulunan %s", arr.Inspect())
	}

	if arr.Elements[0].Type() != object.NUMBER_OBJ {
		return newError(tok, "max(...) sadece sayılardan oluşan listelerde kullanılabilir, bulunan: %s", arr.Inspect())
	}

	max := arr.Elements[0].(*object.Number).Value

	for _, v := range arr.Elements[1:] {
		elem := v.(*object.Number)

		if elem.Value > max {
			max = elem.Value
		}
	}

	return &object.Number{Token: tok, Value: max}
}


func minFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "min", args, 1, [][]string{{object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	arr := args[0].(*object.Array)
	if arr.Empty() {
		return object.NULL
	}

	if !arr.Homogeneous() {
		return newError(tok, "min(...) sadece aynı türden veri içeren bir listelerde kullanılabilir, bulunan %s", arr.Inspect())
	}

	if arr.Elements[0].Type() != object.NUMBER_OBJ {
		return newError(tok, "min(...) sadece sayılardan oluşan listelerde kullanılabilir, bulunan: %s", arr.Inspect())
	}

	min := arr.Elements[0].(*object.Number).Value

	for _, v := range arr.Elements[1:] {
		elem := v.(*object.Number)

		if elem.Value < min {
			min = elem.Value
		}
	}

	return &object.Number{Token: tok, Value: min}
}


func reduceFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "reduce", args, 3, [][]string{{object.ARRAY_OBJ}, {object.FUNCTION_OBJ}, {object.ANY_OBJ}})
	if err != nil {
		return err
	}

	accumulator := args[2]

	for _, v := range args[0].(*object.Array).Elements {
		accumulator = applyFunction(tok, args[1].(*object.Function), env, []object.Object{accumulator, v})
	}

	return accumulator
}


func sortFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "sort", args, 1, [][]string{{object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	arr := args[0].(*object.Array)
	elements := arr.Elements

	if len(elements) == 0 {
		return arr 
	}

	if !arr.Homogeneous() {
		return newError(tok, "'sort'a gelen argüman aynı türden veri içeren bir liste olmalı, bulunan %s", arr.Inspect())
	}

	switch elements[0].(type) {
	case *object.Number:
		a := []float64{}
		for _, v := range elements {
			a = append(a, v.(*object.Number).Value)
		}
		sort.Float64s(a)

		o := []object.Object{}

		for _, v := range a {
			o = append(o, &object.Number{Token: tok, Value: v})
		}
		return &object.Array{Elements: o}
	case *object.String:
		a := []string{}
		for _, v := range elements {
			a = append(a, v.(*object.String).Value)
		}
		sort.Strings(a)

		o := []object.Object{}

		for _, v := range a {
			o = append(o, &object.String{Token: tok, Value: v})
		}
		return &object.Array{Elements: o}
	default:
		return newError(tok, "verilen elementlere sahip bir listeyi sıralayamam (%s)", arr.Inspect())
	}
}


func intersectFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "intersect", args, 2, [][]string{{object.ARRAY_OBJ}, {object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	left := args[0].(*object.Array).Elements
	right := args[1].(*object.Array).Elements
	found := map[string]object.Object{}
	intersection := []object.Object{}

	for _, o := range right {
		found[object.GenerateEqualityString(o)] = o
	}

	for _, o := range left {
		element, ok := found[object.GenerateEqualityString(o)]

		if ok {
			intersection = append(intersection, element)
		}
	}

	return &object.Array{Elements: intersection}
}


func diff(symmetric bool, fnName string, tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, fnName, args, 2, [][]string{{object.ARRAY_OBJ}, {object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	left := args[0].(*object.Array).Elements
	right := args[1].(*object.Array).Elements
	foundRight := map[string]object.Object{}
	difference := []object.Object{}

	for _, o := range right {
		foundRight[object.GenerateEqualityString(o)] = o
	}

	for _, o := range left {
		_, ok := foundRight[object.GenerateEqualityString(o)]

		if !ok {
			difference = append(difference, o)
		}
	}

	if symmetric {
		
		
		difference = append(difference, diff(false, fnName, tok, env, args[1], args[0]).(*object.Array).Elements...)
	}

	return &object.Array{Elements: difference}
}

func diffFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	return diff(false, "diff", tok, env, args...)
}


func diffSymmetricFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	return diff(true, "diff_symmetric", tok, env, args...)
}


func unionFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "union", args, 2, [][]string{{object.ARRAY_OBJ}, {object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	left := args[0].(*object.Array).Elements
	right := args[1].(*object.Array).Elements

	union := []object.Object{}

	for _, v := range left {
		union = append(union, v)
	}

	m := util.Mapify(left)

	for _, v := range right {
		_, found := m[object.GenerateEqualityString(v)]

		if !found {
			union = append(union, v)
		}
	}

	return &object.Array{Elements: union}
}


func flattenFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	return flatten("flatten", false, tok, env, args...)
}


func flattenDeepFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	return flatten("flatten_deep", true, tok, env, args...)
}

func flatten(fnName string, deep bool, tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, fnName, args, 1, [][]string{{object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	originalElements := args[0].(*object.Array).Elements
	elements := []object.Object{}

	for _, v := range originalElements {
		switch e := v.(type) {
		case *object.Array:
			if deep {
				elements = append(elements, flattenDeepFn(tok, env, e).(*object.Array).Elements...)
			} else {
				for _, x := range e.Elements {
					elements = append(elements, x)
				}
			}
		default:
			elements = append(elements, e)
		}
	}

	return &object.Array{Elements: elements}
}

func partitionFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "partition", args, 2, [][]string{{object.ARRAY_OBJ}, {object.FUNCTION_OBJ, object.BUILTIN_OBJ}})
	if err != nil {
		return err
	}

	partitions := map[string][]object.Object{}
	elements := args[0].(*object.Array).Elements
	
	
	
	
	
	
	
	
	
	
	
	partitionOrder := []string{}
	scanned := map[string]bool{}

	for _, v := range elements {
		res := applyFunction(tok, args[1], env, []object.Object{v})
		eqs := object.GenerateEqualityString(res)

		partitions[eqs] = append(partitions[eqs], v)

		if _, ok := scanned[eqs]; !ok {
			partitionOrder = append(partitionOrder, eqs)
			scanned[eqs] = true
		}
	}

	result := &object.Array{Elements: []object.Object{}}
	for _, eqs := range partitionOrder {
		partition := partitions[eqs]
		result.Elements = append(result.Elements, &object.Array{Elements: partition})
	}

	return result
}


func mapFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "map", args, 2, [][]string{{object.ARRAY_OBJ}, {object.FUNCTION_OBJ, object.BUILTIN_OBJ}})
	if err != nil {
		return err
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	newElements := make([]object.Object, length, length)
	copy(newElements, arr.Elements)

	for k, v := range arr.Elements {
		evaluated := applyFunction(tok, args[1], env, []object.Object{v})

		if isError(evaluated) {
			return evaluated
		}
		newElements[k] = evaluated
	}

	return &object.Array{Elements: newElements}
}


func someFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "some", args, 2, [][]string{{object.ARRAY_OBJ}, {object.FUNCTION_OBJ, object.BUILTIN_OBJ}})
	if err != nil {
		return err
	}

	var result bool

	arr := args[0].(*object.Array)

	for _, v := range arr.Elements {
		r := applyFunction(tok, args[1], env, []object.Object{v})

		if isTruthy(r) {
			result = true
			break
		}
	}

	return &object.Boolean{Token: tok, Value: result}
}


func everyFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "every", args, 2, [][]string{{object.ARRAY_OBJ}, {object.FUNCTION_OBJ, object.BUILTIN_OBJ}})
	if err != nil {
		return err
	}

	result := true

	arr := args[0].(*object.Array)

	for _, v := range arr.Elements {
		r := applyFunction(tok, args[1], env, []object.Object{v})

		if !isTruthy(r) {
			result = false
		}
	}

	return &object.Boolean{Token: tok, Value: result}
}


func findFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "find", args, 2, [][]string{{object.ARRAY_OBJ}, {object.FUNCTION_OBJ, object.BUILTIN_OBJ, object.HASH_OBJ}})
	if err != nil {
		return err
	}

	arr := args[0].(*object.Array)

	switch predicate := args[1].(type) {
	case *object.Hash:
		for _, v := range arr.Elements {
			v, ok := v.(*object.Hash)

			if !ok {
				continue
			}

			match := true
			for k, pair := range predicate.Pairs {
				toCompare, ok := v.GetPair(k.Value)
				if !ok {
					match = false
					continue
				}

				if !object.Equal(pair.Value, toCompare.Value) {
					match = false
				}
			}

			if match {
				return v
			}
		}
	default:
		for _, v := range arr.Elements {
			r := applyFunction(tok, predicate, env, []object.Object{v})

			if isTruthy(r) {
				return v
			}
		}
	}

	return NULL
}


func filterFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "filter", args, 2, [][]string{{object.ARRAY_OBJ}, {object.FUNCTION_OBJ, object.BUILTIN_OBJ}})
	if err != nil {
		return err
	}

	result := []object.Object{}
	arr := args[0].(*object.Array)

	for _, v := range arr.Elements {
		evaluated := applyFunction(tok, args[1], env, []object.Object{v})

		if isError(evaluated) {
			return evaluated
		}

		if isTruthy(evaluated) {
			result = append(result, v)
		}
	}

	return &object.Array{Elements: result}
}


func uniqueFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "unique", args, 1, [][]string{{object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	result := []object.Object{}
	arr := args[0].(*object.Array)
	existingElements := map[string]bool{}

	for _, v := range arr.Elements {
		key := object.GenerateEqualityString(v)

		if _, ok := existingElements[key]; !ok {
			existingElements[key] = true
			result = append(result, v)
		}
	}

	return &object.Array{Elements: result}
}


func strFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "str", args, 1, [][]string{})
	if err != nil {
		return err
	}

	return &object.String{Token: tok, Value: args[0].Inspect()}
}


func anyFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "any", args, 2, [][]string{{object.STRING_OBJ}, {object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return &object.Boolean{Token: tok, Value: strings.ContainsAny(args[0].(*object.String).Value, args[1].(*object.String).Value)}
}


func betweenFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "between", args, 3, [][]string{{object.NUMBER_OBJ}, {object.NUMBER_OBJ}, {object.NUMBER_OBJ}})
	if err != nil {
		return err
	}

	n := args[0].(*object.Number)
	min := args[1].(*object.Number)
	max := args[2].(*object.Number)

	if min.Value >= max.Value {
		return newError(tok, "min ve max arası argümanlar min ve max aralığını tutmalı (%s < %s bulundu)", min.Inspect(), max.Inspect())
	}

	return &object.Boolean{Token: tok, Value: ((min.Value <= n.Value) && (n.Value <= max.Value))}
}


func prefixFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "prefix", args, 2, [][]string{{object.STRING_OBJ}, {object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return &object.Boolean{Token: tok, Value: strings.HasPrefix(args[0].(*object.String).Value, args[1].(*object.String).Value)}
}


func suffixFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "suffix", args, 2, [][]string{{object.STRING_OBJ}, {object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return &object.Boolean{Token: tok, Value: strings.HasSuffix(args[0].(*object.String).Value, args[1].(*object.String).Value)}
}


func repeatFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "repeat", args, 2, [][]string{{object.STRING_OBJ}, {object.NUMBER_OBJ}})
	if err != nil {
		return err
	}

	return &object.String{Token: tok, Value: strings.Repeat(args[0].(*object.String).Value, int(args[1].(*object.Number).Value))}
}




func replaceFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	var err object.Object

	
	if len(args) == 3 {
		err = validateArgs(tok, "replace", args, 3, [][]string{{object.STRING_OBJ}, {object.STRING_OBJ, object.ARRAY_OBJ}, {object.STRING_OBJ}})
	} else {
		err = validateArgs(tok, "replace", args, 4, [][]string{{object.STRING_OBJ}, {object.STRING_OBJ, object.ARRAY_OBJ}, {object.STRING_OBJ}, {object.NUMBER_OBJ}})
	}

	if err != nil {
		return err
	}

	original := args[0].(*object.String).Value
	replacement := args[2].(*object.String).Value

	n := -1

	if len(args) == 4 {
		n = int(args[3].(*object.Number).Value)
	}

	if characters, ok := args[1].(*object.Array); ok {
		for _, c := range characters.Elements {
			original = strings.Replace(original, c.Inspect(), replacement, n)
		}

		return &object.String{Token: tok, Value: original}
	}

	return &object.String{Token: tok, Value: strings.Replace(original, args[1].(*object.String).Value, replacement, n)}
}


func titleFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "title", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return &object.String{Token: tok, Value: strings.Title(args[0].(*object.String).Value)}
}


func lowerFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "lower", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return &object.String{Token: tok, Value: strings.ToLower(args[0].(*object.String).Value)}
}


func upperFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "upper", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return &object.String{Token: tok, Value: strings.ToUpper(args[0].(*object.String).Value)}
}


func waitFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "wait", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	cmd := args[0].(*object.String)

	if cmd.Cmd == nil {
		return cmd
	}

	cmd.Wait()
	return cmd
}


func killFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "kill", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	cmd := args[0].(*object.String)

	if cmd.Cmd == nil {
		return cmd
	}

	errCmdKill := cmd.Kill()

	if errCmdKill != nil {
		return newError(tok, "%s komutunu oldürürken bir hata oldu. Hatta: %s", cmd.Inspect(), errCmdKill.Error())
	}
	return cmd
}


func trimFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "trim", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return &object.String{Token: tok, Value: strings.Trim(args[0].(*object.String).Value, " ")}
}


func trimByFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "trim_by", args, 2, [][]string{{object.STRING_OBJ}, {object.STRING_OBJ}})
	if err != nil {
		return err
	}

	return &object.String{Token: tok, Value: strings.Trim(args[0].(*object.String).Value, args[1].(*object.String).Value)}
}


func indexFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "index", args, 2, [][]string{{object.STRING_OBJ}, {object.STRING_OBJ}})
	if err != nil {
		return err
	}

	i := strings.Index(args[0].(*object.String).Value, args[1].(*object.String).Value)

	if i == -1 {
		return NULL
	}

	return &object.Number{Token: tok, Value: float64(i)}
}


func lastIndexFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "last_index", args, 2, [][]string{{object.STRING_OBJ}, {object.STRING_OBJ}})
	if err != nil {
		return err
	}

	i := strings.LastIndex(args[0].(*object.String).Value, args[1].(*object.String).Value)

	if i == -1 {
		return NULL
	}

	return &object.Number{Token: tok, Value: float64(i)}
}




func sliceStartAndEnd(l int, start int, end int) (int, int) {
	if end == 0 {
		end = l
	}

	if start > l {
		start = l
	}

	if start < 0 {
		newStart := l + start
		if newStart < 0 {
			start = 0
		} else {
			start = newStart
		}
	}

	if end > l || start > end {
		end = l
	}

	return start, end
}


func shiftFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "shift", args, 1, [][]string{{object.ARRAY_OBJ}})
	if err != nil {
		return err
	}

	array := args[0].(*object.Array)
	if len(array.Elements) == 0 {
		return NULL
	}
	e := array.Elements[0]
	array.Elements = append(array.Elements[:0], array.Elements[1:]...)

	return e
}


func reverseFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err, spec := validateVarArgs(tok, "reverse", args, [][][]string{
		{{object.ARRAY_OBJ}},
		{{object.STRING_OBJ}},
	})

	if err != nil {
		return err
	}

	if spec == 0 {
		
		elements := args[0].(*object.Array).Elements
		length := len(elements)
		newElements := make([]object.Object, length, length)
		copy(newElements, elements)

		for i, j := 0, len(newElements)-1; i < j; i, j = i+1, j-1 {
			newElements[i], newElements[j] = newElements[j], newElements[i]
		}

		return &object.Array{Elements: newElements}
	} else {
		
		str := []rune(args[0].(*object.String).Value)

		for i, j := 0, len(str)-1; i < j; i, j = i+1, j-1 {
			str[i], str[j] = str[j], str[i]
		}

		return &object.String{Token: tok, Value: string(str)}
	}
}


func shuffleFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "shuffle", args, 1, [][]string{{object.ARRAY_OBJ}})

	if err != nil {
		return err
	}

	array := args[0].(*object.Array)
	length := len(array.Elements)
	newElements := make([]object.Object, length, length)
	copy(newElements, array.Elements)

	mrand.Seed(time.Now().UnixNano())
	mrand.Shuffle(len(newElements), func(i, j int) { newElements[i], newElements[j] = newElements[j], newElements[i] })

	return &object.Array{Elements: newElements}
}


func pushFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "push", args, 2, [][]string{{object.ARRAY_OBJ}, {object.NULL_OBJ,
		object.ARRAY_OBJ, object.NUMBER_OBJ, object.STRING_OBJ, object.HASH_OBJ}})
	if err != nil {
		return err
	}

	array := args[0].(*object.Array)
	array.Elements = append(array.Elements, args[1])

	return array
}



func popFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	
	var err object.Object
	if len(args) > 0 {
		if args[0].Type() == object.ARRAY_OBJ {
			err = validateArgs(tok, "pop", args, 1, [][]string{{object.ARRAY_OBJ}})
		} else if args[0].Type() == object.HASH_OBJ {
			err = validateArgs(tok, "pop", args, 2, [][]string{{object.HASH_OBJ}})
		}
	}
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return NULL
	}
	switch arg := args[0].(type) {
	case *object.Array:
		if len(arg.Elements) > 0 {
			elem := arg.Elements[len(arg.Elements)-1]
			arg.Elements = arg.Elements[0 : len(arg.Elements)-1]
			return elem
		}
	case *object.Hash:
		if len(args) == 2 {
			key := args[1].(object.Hashable)
			hashKey := key.HashKey()
			item, ok := arg.Pairs[hashKey]
			if ok {
				pairs := make(map[object.HashKey]object.HashPair)
				pairs[hashKey] = item
				delete(arg.Pairs, hashKey)
				return &object.Hash{Pairs: pairs}
			}
		}
	}
	return NULL
}



func keysFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "keys", args, 1, [][]string{{object.ARRAY_OBJ, object.HASH_OBJ}})
	if err != nil {
		return err
	}
	switch arg := args[0].(type) {
	case *object.Array:
		length := len(arg.Elements)
		newElements := make([]object.Object, length, length)
		for k := range arg.Elements {
			newElements[k] = &object.Number{Token: tok, Value: float64(k)}
		}
		return &object.Array{Elements: newElements}
	case *object.Hash:
		pairs := arg.Pairs
		keys := []object.Object{}
		for _, pair := range pairs {
			key := pair.Key
			keys = append(keys, key)
		}
		return &object.Array{Elements: keys}
	}
	return NULL
}


func valuesFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "values", args, 1, [][]string{{object.HASH_OBJ}})
	if err != nil {
		return err
	}
	hash := args[0].(*object.Hash)
	pairs := hash.Pairs
	values := []object.Object{}
	for _, pair := range pairs {
		value := pair.Value
		values = append(values, value)
	}
	return &object.Array{Elements: values}
}


func itemsFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "items", args, 1, [][]string{{object.HASH_OBJ}})
	if err != nil {
		return err
	}
	hash := args[0].(*object.Hash)
	pairs := hash.Pairs
	items := []object.Object{}
	for _, pair := range pairs {
		key := pair.Key
		value := pair.Value
		item := &object.Array{Elements: []object.Object{key, value}}
		items = append(items, item)
	}
	return &object.Array{Elements: items}
}

func joinFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err, spec := validateVarArgs(tok, "join", args, [][][]string{
		{{object.ARRAY_OBJ}, {object.STRING_OBJ}},
		{{object.ARRAY_OBJ}},
	})

	if err != nil {
		return err
	}

	glue := ""
	if spec == 0 {
		glue = args[1].(*object.String).Value
	}

	arr := args[0].(*object.Array)
	length := len(arr.Elements)
	newElements := make([]string, length, length)

	for k, v := range arr.Elements {
		newElements[k] = v.Inspect()
	}

	return &object.String{Token: tok, Value: strings.Join(newElements, glue)}
}

func sleepFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "sleep", args, 1, [][]string{{object.NUMBER_OBJ}})
	if err != nil {
		return err
	}

	ms := args[0].(*object.Number)
	time.Sleep(time.Duration(ms.Value) * time.Millisecond)

	return NULL
}


const ANK_SOURCE_DEPTH = "10"

var sourceDepth, _ = strconv.Atoi(ANK_SOURCE_DEPTH)
var sourceLevel = 0

func sourceFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	file, _ := util.ExpandPath(args[0].Inspect())
	return doSource(tok, env, file, args...)
}


var history = make(map[string]string)

var packageAliases map[string]string
var packageAliasesLoaded bool

func requireFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	if !packageAliasesLoaded {
		a, err := ioutil.ReadFile("./paketler.json")

		
		
		if err == nil {
			
			
			
			json.Unmarshal(a, &packageAliases)
		}

		packageAliasesLoaded = true
	}

	file := util.UnaliasPath(args[0].Inspect(), packageAliases)

	if !strings.HasPrefix(file, "@") {
		file = filepath.Join(env.Dir, file)
	}

	if evaluated, ok := requireCache[file]; ok {
		return evaluated
	}

	e := object.NewEnvironment(env.Writer, filepath.Dir(file), env.Version)
	evaluated := doSource(tok, e, file, args...)

	
	
	switch ret := evaluated.(type) {
	case *object.Error:
		return ret
	default:
		requireCache[file] = evaluated
	}

	return evaluated
}

func doSource(tok token.Token, env *object.Environment, fileName string, args ...object.Object) object.Object {
	err := validateArgs(tok, "source", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		
		sourceLevel = 0
		return err
	}

	
	sourceDepthStr := util.GetEnvVar(env, "ANK_SOURCE_DEPTH", ANK_SOURCE_DEPTH)
	sourceDepth, _ = strconv.Atoi(sourceDepthStr)

	
	if sourceLevel >= sourceDepth {
		
		sourceLevel = 0
		
		errObj := newError(tok, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", sourceDepth)
		errObj = &object.Error{Message: errObj.Message}
		return errObj
	}
	
	sourceLevel++

	var code []byte
	var error error

	
	
	if strings.HasPrefix(fileName, "@") {
		code, error = Asset("stdlib/" + fileName[1:])
	} else {
		
		code, error = ioutil.ReadFile(fileName)
	}

	if error != nil {
		
		sourceLevel = 0
		
		return newError(tok, "kaynak dosyası okunamadı: %s:\n%s", fileName, error.Error())
	}
	
	l := lexer.New(string(code))
	p := parser.New(l)
	program := p.ParseProgram()
	errors := p.Errors()
	if len(errors) != 0 {
		
		sourceLevel = 0
		errMsg := fmt.Sprintf("%s", " Ayrıştırıcı hatası:\n")
		for _, msg := range errors {
			errMsg += fmt.Sprintf("%s", "\t"+msg+"\n")
		}
		return newError(tok, "kaynak dosyasında hata çıktı: %s\n%s", fileName, errMsg)
	}
	
	
	
	savedLexer := lex
	evaluated := BeginEval(program, env, l)
	lex = savedLexer
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		
		evalErrMsg := evaluated.(*object.Error).Message
		sourceErrMsg := newError(tok, "eval bloğunda hata bulundu: %s", fileName).Message
		errObj := &object.Error{Message: fmt.Sprintf("%s\n\t%s", sourceErrMsg, evalErrMsg)}
		return errObj
	}
	
	sourceLevel--

	return evaluated
}

func evalFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "eval", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}

	
	l := lexer.New(string(args[0].Inspect()))
	p := parser.New(l)
	program := p.ParseProgram()
	errors := p.Errors()
	if len(errors) != 0 {
		errMsg := fmt.Sprintf("%s", " Ayrıştırıcı hatası:\n")
		for _, msg := range errors {
			errMsg += fmt.Sprintf("%s", "\t"+msg+"\n")
		}
		return newError(tok, "Eval bloğunda hata bulundu: %s\n%s", args[0].Inspect(), errMsg)
	}
	
	
	
	savedLexer := lex
	evaluated := BeginEval(program, env, l)
	lex = savedLexer

	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		
		evalErrMsg := evaluated.(*object.Error).Message
		sourceErrMsg := newError(tok, "Eval bloğunda hata bulundu: %s", args[0].Inspect()).Message
		errObj := &object.Error{Message: fmt.Sprintf("%s\n\t%s", sourceErrMsg, evalErrMsg)}
		return errObj
	}

	return evaluated
}



func tsvFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	
	if len(args) == 3 {
		err := validateArgs(tok, "tsv", args, 3, [][]string{{object.ARRAY_OBJ}, {object.STRING_OBJ}, {object.ARRAY_OBJ}})
		if err != nil {
			return err
		}
	}

	
	if len(args) == 2 {
		err := validateArgs(tok, "tsv", args, 2, [][]string{{object.ARRAY_OBJ}, {object.STRING_OBJ}})
		if err != nil {
			return err
		}
		args = append(args, &object.Array{Elements: []object.Object{}})
	}

	
	if len(args) == 1 {
		err := validateArgs(tok, "tsv", args, 1, [][]string{{object.ARRAY_OBJ}})
		if err != nil {
			return err
		}
		args = append(args, &object.String{Value: "\t"})
		args = append(args, &object.Array{Elements: []object.Object{}})
	}

	array := args[0].(*object.Array)
	separator := args[1].(*object.String).Value

	if len(separator) < 1 {
		return newError(tok, "tsv() işlevinin ayırıcı bağımsız değişkeninin geçerli bir karakter olması gerekir, '%s' değil", separator)
	}
	
	out := &strings.Builder{}
	tsv := csv.NewWriter(out)
	tsv.Comma = rune(separator[0])

	
	var isArray bool
	var isHash bool
	homogeneous := array.Homogeneous()

	if len(array.Elements) > 0 {
		_, isArray = array.Elements[0].(*object.Array)
		_, isHash = array.Elements[0].(*object.Hash)
	}

	
	if !homogeneous || (!isArray && !isHash) {
		return newError(tok, "tsv() işlevinin ayırıcı bağımsız değişkeninin geçerli bir karakter olması gerekir, örneğin [[1, 2, 3], [4, 5, 6]], '%s' değil", array.Inspect())
	}

	headerObj := args[2].(*object.Array)
	header := []string{}

	if len(headerObj.Elements) > 0 {
		for _, v := range headerObj.Elements {
			header = append(header, v.Inspect())
		}
	} else if isHash {
		
		
		for _, rows := range array.Elements {
			for _, pair := range rows.(*object.Hash).Pairs {
				header = append(header, pair.Key.Inspect())
			}
		}

		
		
		
		header = util.UniqueStrings(header)
		sort.Strings(header)
	}

	if len(header) > 0 {
		err := tsv.Write(header)

		if err != nil {
			return newError(tok, err.Error())
		}
	}

	for _, row := range array.Elements {
		
		values := []string{}

		
		
		
		if isArray {
			for _, element := range row.(*object.Array).Elements {
				values = append(values, element.Inspect())
			}

		}

		
		
		
		if isHash {
			for _, key := range header {
				pair, ok := row.(*object.Hash).GetPair(key)
				var value object.Object

				if ok {
					value = pair.Value
				} else {
					value = NULL
				}

				values = append(values, value.Inspect())
			}
		}

		
		
		err := tsv.Write(values)

		if err != nil {
			return newError(tok, err.Error())
		}
	}

	tsv.Flush()
	return &object.String{Value: strings.TrimSpace(out.String())}
}

func execFn(tok token.Token, env *object.Environment, args ...object.Object) object.Object {
	err := validateArgs(tok, "exec", args, 1, [][]string{{object.STRING_OBJ}})
	if err != nil {
		return err
	}
	cmd := args[0].Inspect()
	cmd = strings.Trim(cmd, " ")

	
	cmd = util.InterpolateStringVars(cmd, env)

	
	parts := strings.Split(os.Getenv("ANK_COMMAND_EXECUTOR"), " ")
	c := exec.Command(parts[0], append(parts[1:], cmd)...)
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	
	
	
	runErr := c.Run()

	if runErr != nil {
		return &object.String{Value: runErr.Error()}
	}
	return NULL
}
