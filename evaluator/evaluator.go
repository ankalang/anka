package evaluator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/ankalang/anka/ast"
	"github.com/ankalang/anka/lexer"
	"github.com/ankalang/anka/object"
	"github.com/ankalang/anka/token"
	"github.com/ankalang/anka/util"
)

var (
	NULL  = object.NULL
	EOF   = object.EOF
	TRUE  = object.TRUE
	FALSE = object.FALSE
	Fns   map[string]*object.Builtin
)


var lex *lexer.Lexer

func init() {
	Fns = getFns()
	if os.Getenv("ANK_COMMAND_EXECUTOR") == "" {
		
		
		os.Setenv("ANK_COMMAND_EXECUTOR", "bash -c")

		if runtime.GOOS == "windows" {
			os.Setenv("ANK_COMMAND_EXECUTOR", "cmd.exe /C")
		}
	}
}

func newError(tok token.Token, format string, a ...interface{}) *object.Error {
	
	lineNum, collumn, errorLine := lex.ErrorLine(tok.Position)

	errorPosition := fmt.Sprintf("\033[97m\n%d:%d> %s", lineNum,collumn,errorLine)
	return &object.Error{Message: fmt.Sprintf(format, a...) + errorPosition}
}

func newBreakError(tok token.Token, format string, a ...interface{}) *object.BreakError {
	return &object.BreakError{Error: *newError(tok, format, a...)}
}

func newContinueError(tok token.Token, format string, a ...interface{}) *object.ContinueError {
	return &object.ContinueError{Error: *newError(tok, format, a...)}
}




func BeginEval(program ast.Node, env *object.Environment, lexer *lexer.Lexer) object.Object {
	
	lex = lexer
	
	return Eval(program, env)
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.AssignStatement:
		err := evalAssignment(node, env)

		if isError(err) {
			return err
		}

		return NULL
	
	case *ast.NumberLiteral:
		return &object.Number{Token: node.Token, Value: node.Value}

	case *ast.NullLiteral:
		return NULL

	case *ast.CurrentArgsLiteral:
		return &object.Array{Token: node.Token, Elements: env.CurrentArgs, IsCurrentArgs: true}

	case *ast.StringLiteral:
		return &object.String{Token: node.Token, Value: util.InterpolateStringVars(node.Value, env)}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Token, node.Operator, right)

	case *ast.InfixExpression:
		return evalInfixExpression(node.Token, node.Operator, node.Left, node.Right, env)

	case *ast.CompoundAssignment:
		return evalCompoundAssignment(node, env)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.WhileExpression:
		return evalWhileExpression(node, env)

	case *ast.ForExpression:
		return evalForExpression(node, env)

	case *ast.ForInExpression:
		return evalForInExpression(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		name := node.Name
		fn := &object.Function{Token: node.Token, Parameters: params, Env: env, Body: body, Name: name, Node: node}

		if name != "" {
			env.Set(name, fn)
		}

		return fn

	case *ast.Decorator:
		return evalDecorator(node, env)

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)

		
		
		
		
		
		if len(args) > 0 {
			firstArg, ok := args[0].(*object.Array)

			if ok && firstArg.IsCurrentArgs {
				newArgs := env.CurrentArgs
				args = append(newArgs, args[1:]...)
			}
		}

		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(node.Token, function, env, args)

	case *ast.MethodExpression:
		o := Eval(node.Object, env)
		if isError(o) {
			return o
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyMethod(node.Token, o, node, env, args)

	case *ast.PropertyExpression:
		return evalPropertyExpression(node, env)

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Token: node.Token, Elements: elements}

	case *ast.IndexExpression:
		return evalIndexExpression(node, env)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.CommandExpression:
		return evalCommandExpression(node.Token, node.Value, env)

	
	
	
	case *ast.BreakStatement:
		return newBreakError(node.Token, "Döngü dışında \"dur\" çağrıldı")
	
	
	
	case *ast.ContinueStatement:
		return newContinueError(node.Token, "Döngü dışında \"devam\" çağrıldı")

	}

	return NULL
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	deferred := []*ast.ExpressionStatement{}

loop:
	for _, statement := range program.Statements {
		x, ok := statement.(*ast.ExpressionStatement)

		if ok {
			if d, ok := x.Expression.(ast.Deferrable); ok && d.IsDeferred() {
				deferred = append(deferred, x)
				continue
			}
		}
		result = Eval(statement, env)

		switch ret := result.(type) {
		case *object.ReturnValue:
			result = ret.Value
			break loop
		case *object.Error:
			break loop
		}
	}

	for _, statement := range deferred {
		Eval(statement, env)
	}

	return result
}





func evalBlockStatement(
	block *ast.BlockStatement,
	env *object.Environment,
) object.Object {
	var result object.Object
	deferred := []*ast.ExpressionStatement{}

	for _, statement := range block.Statements {
		x, ok := statement.(*ast.ExpressionStatement)

		if ok {
			if d, ok := x.Expression.(ast.Deferrable); ok && d.IsDeferred() {
				deferred = append(deferred, x)
				continue
			}
		}
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				break
			}
		}
	}

	for _, statement := range deferred {
		Eval(statement, env)
	}

	return result
}

func evalCompoundAssignment(node *ast.CompoundAssignment, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}
	
	op := node.Operator
	if len(op) >= 2 {
		op = op[:len(op)-1]
	}
	
	expr := evalInfixExpression(node.Token, op, node.Left, node.Right, env)
	if isError(expr) {
		return expr
	}
	switch nodeLeft := node.Left.(type) {
	case *ast.Identifier:
		env.Set(nodeLeft.String(), expr)
		return NULL
	case *ast.IndexExpression:
		
		return evalIndexAssignment(nodeLeft, expr, env)
	case *ast.PropertyExpression:
		
		return evalPropertyAssignment(nodeLeft, expr, env)
	}
	
	env.Set(node.Left.String(), expr)
	return NULL
}

func evalDecorator(node *ast.Decorator, env *object.Environment) object.Object {
	ident, fn, err := doEvalDecorator(node, env)

	if isError(err) {
		return err
	}

	env.Set(ident, fn)
	return object.NULL
}


func doEvalDecorator(node *ast.Decorator, env *object.Environment) (string, object.Object, object.Object) {
	var decorator object.Object

	evaluated := Eval(node.Expression, env)
	switch evaluated.(type) {
	case *object.Function:
		decorator = evaluated
	case *object.Error:
		return "", nil, evaluated
	default:
		return "", nil, newError(node.Token, "'%s' bir dekaratör değil", evaluated.Inspect())
	}

	name, ok := getDecoratedName(node.Decorated)

	if !ok {
		return "", nil, newError(node.Token, "Dekaratör aranırken bir hata yaşandı. Dekore etmek istediğiniz fonksiyon bulunamadı")
	}

	switch decorated := node.Decorated.(type) {
	case *ast.FunctionLiteral:
		
		fn := &object.Function{Token: decorated.Token, Parameters: decorated.Parameters, Env: env, Body: decorated.Body, Name: name, Node: decorated}
		return name, applyFunction(decorated.Token, decorator, env, []object.Object{fn}), nil
	case *ast.Decorator:
		
		
		

		
		fnName, fn, err := doEvalDecorator(decorated, env)

		if isError(err) {
			return "", nil, err
		}

		return fnName, applyFunction(node.Token, decorator, env, append([]object.Object{fn})), nil
	default:
		return "", nil, newError(node.Token, "Bir dekaratör başka bir dekaratörü ya da bir fonksiyonu dekore etmelidir.")
	}
}


func getDecoratedName(decorated ast.Expression) (string, bool) {
	switch d := decorated.(type) {
	case *ast.FunctionLiteral:
		return d.Name, true
	case *ast.Decorator:
		return getDecoratedName(d.Decorated)
	}

	return "", false
}


func evalIndexAssignment(iex *ast.IndexExpression, expr object.Object, env *object.Environment) object.Object {
	leftObj := Eval(iex.Left, env)
	index := Eval(iex.Index, env)
	if leftObj.Type() == object.ARRAY_OBJ {
		arrayObject := leftObj.(*object.Array)
		idx := index.(*object.Number).Int()
		elems := arrayObject.Elements
		if idx < 0 {
			return newError(iex.Token, "Verilen dizin aralık dışı: %d", idx)
		}
		if idx >= len(elems) {
			
			for i := len(elems); i <= idx; i++ {
				elems = append(elems, NULL)
			}
			arrayObject.Elements = elems
		}
		elems[idx] = expr
		return NULL
	}
	if leftObj.Type() == object.HASH_OBJ {
		hashObject := leftObj.(*object.Hash)
		key, ok := index.(object.Hashable)
		if !ok {
			return newError(iex.Token, "harita anahtarları olarak: Sayılar, fonksiyonlar ve değişkenler kullanılamaz.")
			
		}
		hashed := key.HashKey()
		pair := object.HashPair{Key: index, Value: expr}
		hashObject.Pairs[hashed] = pair
		return NULL
	}
	return NULL
}


func evalPropertyAssignment(pex *ast.PropertyExpression, expr object.Object, env *object.Environment) object.Object {
	leftObj := Eval(pex.Object, env)
	if leftObj.Type() == object.HASH_OBJ {
		hashObject := leftObj.(*object.Hash)
		prop := &object.String{Token: pex.Token, Value: pex.Property.String()}
		hashed := prop.HashKey()
		pair := object.HashPair{Key: prop, Value: expr}
		hashObject.Pairs[hashed] = pair
		return NULL
	}
	return newError(pex.Token, "sadece haritaların anahtarlarına değer atamaları yapılabilir")
}

func evalAssignment(as *ast.AssignStatement, env *object.Environment) object.Object {
	val := Eval(as.Value, env)
	if isError(val) {
		return val
	}

	
	if as.Name != nil {
		env.Set(as.Name.Value, val)
		return nil
	}

	
	if len(as.Names) > 0 {
		switch v := val.(type) {
		case *object.Array:
			elements := v.Elements
			for i, name := range as.Names {
				if i < len(elements) {
					env.Set(name.String(), elements[i])
					continue
				}

				env.Set(name.String(), NULL)
			}
		case *object.Hash:
			for _, name := range as.Names {
				x, ok := v.GetPair(name.String())

				if ok {
					env.Set(name.String(), x.Value)
				} else {
					env.Set(name.String(), NULL)
				}
			}
		default:
			return newError(as.Token, "Listeli değişken tanımlaması için liste ya da tanımlayıcı beklendi, fakat başka bir şey bulundu")
		}

		return nil
	}
	
	if as.Index != nil {
		return evalIndexAssignment(as.Index, val, env)
	}
	
	if as.Property != nil {
		return evalPropertyAssignment(as.Property, val, env)
	}

	return nil
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(tok token.Token, operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(tok, right)
	case "+":
		return evalPlusPrefixOperatorExpression(tok, right)
	case "~":
		return evalTildePrefixOperatorExpression(tok, right)
	default:
		return newError(tok, "bilinmeyen operatör: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(
	tok token.Token, operator string,
	leftExpression, rightExpression ast.Expression,
	env *object.Environment,
) object.Object {
	left := Eval(leftExpression, env)
	if isError(left) {
		return left
	}

	
	
	
	
	
	
	if operator == "&&" {
		if !isTruthy(left) {
			return left
		}
		return Eval(rightExpression, env)
	}

	if operator == "||" {
		if isTruthy(left) {
			return left
		}
		return Eval(rightExpression, env)
	}

	right := Eval(rightExpression, env)
	if isError(right) {
		return right
	}

	switch {
	case left.Type() == object.NUMBER_OBJ && right.Type() == object.NUMBER_OBJ:
		return evalNumberInfixExpression(tok, operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(tok, operator, left, right)
	case left.Type() == object.ARRAY_OBJ && right.Type() == object.ARRAY_OBJ:
		return evalArrayInfixExpression(tok, operator, left, right)
	case left.Type() == object.HASH_OBJ && right.Type() == object.HASH_OBJ:
		return evalHashInfixExpression(tok, operator, left, right)
	case operator == "in":
		return evalInExpression(tok, left, right)
	case operator == "!in":
		return evalNotInExpression(tok, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError(tok, "Tip uyuşmazlığı %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError(tok, "Bilinmeyen operatör: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	if isTruthy(right) {
		return FALSE
	}

	return TRUE
}

func evalTildePrefixOperatorExpression(tok token.Token, right object.Object) object.Object {
	switch o := right.(type) {
	case *object.Number:
		return &object.Number{Value: float64(^int64(o.Value))}
	default:
		return newError(tok, "(~) sadece sayılarda kullanılabilir fakat %s bulundu. (%s) ", o.Type(), o.Inspect())
	}
}

func evalMinusPrefixOperatorExpression(tok token.Token, right object.Object) object.Object {
	if right.Type() != object.NUMBER_OBJ {
		return newError(tok, "Bilinmeyen operatör: -%s", right.Type())
	}

	value := right.(*object.Number).Value
	return &object.Number{Value: -value}
}

func evalPlusPrefixOperatorExpression(tok token.Token, right object.Object) object.Object {
	if right.Type() != object.NUMBER_OBJ {
		return newError(tok, "Bilinmeyen operatör: +%s", right.Type())
	}

	return right
}

func evalNumberInfixExpression(
	tok token.Token, operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.Number).Value
	rightVal := right.(*object.Number).Value
	switch operator {
	case "+":
		return &object.Number{Token: tok, Value: leftVal + rightVal}
	case "-":
		return &object.Number{Token: tok, Value: leftVal - rightVal}
	case "*":
		return &object.Number{Token: tok, Value: leftVal * rightVal}
	case "/":
		return &object.Number{Token: tok, Value: leftVal / rightVal}
	case "**":
		
		return &object.Number{Token: tok, Value: math.Pow(leftVal, rightVal)}
	case "%":
		return &object.Number{Token: tok, Value: math.Mod(leftVal, rightVal)}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "<=>":
		i := &object.Number{Token: tok}

		if leftVal == rightVal {
			i.Value = 0
		} else if leftVal > rightVal {
			i.Value = 1
		} else {
			i.Value = -1
		}

		return i
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "&":
		return &object.Number{Token: tok, Value: float64(int64(leftVal) & int64(rightVal))}
	case "|":
		return &object.Number{Token: tok, Value: float64(int64(leftVal) | int64(rightVal))}
	case ">>":
		return &object.Number{Token: tok, Value: float64(uint64(leftVal) >> uint64(rightVal))}
	case "<<":
		return &object.Number{Token: tok, Value: float64(uint64(leftVal) << uint64(rightVal))}
	case "^":
		return &object.Number{Token: tok, Value: float64(int64(leftVal) ^ int64(rightVal))}
	case "~":
		return &object.Boolean{Token: tok, Value: int64(leftVal) == int64(rightVal)}
	
	case "..":
		a := make([]object.Object, 0)

		if leftVal <= rightVal {
			for i := leftVal; i <= rightVal; i++ {
				a = append(a, &object.Number{Token: tok, Value: float64(i)})
			}
		} else {
			for i := leftVal; i >= rightVal; i-- {
				a = append(a, &object.Number{Token: tok, Value: float64(i)})
			}
		}

		return &object.Array{Token: tok, Elements: a}
	default:
		return newError(tok, "Bilinmeyen operatör: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(
	tok token.Token,
	operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	if operator == "+" {
		return &object.String{Token: tok, Value: leftVal + rightVal}
	}

	if operator == "==" {
		return &object.Boolean{Token: tok, Value: leftVal == rightVal}
	}

	if operator == "!=" {
		return &object.Boolean{Token: tok, Value: leftVal != rightVal}
	}

	if operator == "~" {
		return &object.Boolean{Token: tok, Value: strings.ToLower(leftVal) == strings.ToLower(rightVal)}
	}

	if operator == "in" {
		return evalInExpression(tok, left, right)
	}

	if operator == "!in" {
		return evalNotInExpression(tok, left, right).(*object.Boolean)
	}

	if operator == ">" {
		err := writeFile(rightVal, leftVal)

		if err != nil {
			return newError(tok, "%s'e yazarken başarısız olundu: %s", rightVal, err.Error())
		}

		return &object.Boolean{Token: tok, Value: true}
	}

	if operator == ">>" {
		err := appendFile(rightVal, leftVal)

		if err != nil {
			return newError(tok, "%s'e yazarken başarısız olundu: %s", rightVal, err.Error())
		}

		return &object.Boolean{Token: tok, Value: true}
	}

	return newError(tok, "Bilinmeyen operatör: %s %s %s", left.Type(), operator, right.Type())
}

func writeFile(file string, content string) error {
	return ioutil.WriteFile(file, []byte(content), 0644)
}

func appendFile(file string, content string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return err
	}

	return nil
}

func evalArrayInfixExpression(
	tok token.Token,
	operator string,
	left, right object.Object,
) object.Object {
	if operator == "+" {
		leftVal := left.(*object.Array).Elements
		rightVal := right.(*object.Array).Elements
		return &object.Array{Token: tok, Elements: append(leftVal, rightVal...)}
	}

	return newError(tok, "Bilinmeyen operatör: %s %s %s", left.Type(), operator, right.Type())
}

func evalHashInfixExpression(
	tok token.Token,
	operator string,
	left, right object.Object,
) object.Object {
	leftHashObject := left.(*object.Hash)
	rightHashObject := right.(*object.Hash)
	if operator == "+" {
		leftVal := leftHashObject.Pairs
		rightVal := rightHashObject.Pairs
		for _, rightPair := range rightVal {
			key := rightPair.Key
			hashed := key.(object.Hashable).HashKey()
			leftVal[hashed] = object.HashPair{Key: key, Value: rightPair.Value}
		}
		return &object.Hash{Token: tok, Pairs: leftVal}
	}

	return newError(tok, "Bilinmeyen operatör: %s %s %s", left.Type(), operator, right.Type())
}

func evalInExpression(tok token.Token, left, right object.Object) object.Object {
	var found bool

	switch rightObj := right.(type) {
	case *object.Array:
		switch needle := left.(type) {
		case *object.String:
			for _, v := range rightObj.Elements {
				if v.Inspect() == needle.Value && v.Type() == object.STRING_OBJ {
					found = true
					break 
				}
			}
		case *object.Number:
			for _, v := range rightObj.Elements {
				
				
				
				
				
				if v.Inspect() == strconv.Itoa(int(needle.Value)) && v.Type() == object.NUMBER_OBJ {
					found = true
					break 
				}
			}
		}
	case *object.String:
		if left.Type() == object.STRING_OBJ {
			found = strings.Contains(right.Inspect(), left.Inspect())
		}
	case *object.Hash:
		if left.Type() == object.STRING_OBJ {
			_, ok := rightObj.GetPair(left.(*object.String).Value)
			found = ok
		}
	default:
		return newError(tok, "'de' operatörü, %s için geçerli değil", right.Type())
	}

	return &object.Boolean{Token: tok, Value: found}
}

func evalNotInExpression(tok token.Token, left, right object.Object) object.Object {
	obj := evalInExpression(tok, left, right).(*object.Boolean)
	obj.Value = !obj.Value
	return obj
}

func evalIfExpression(
	ie *ast.IfExpression,
	env *object.Environment,
) object.Object {
	for _, scenario := range ie.Scenarios {
		condition := Eval(scenario.Condition, env)

		if isError(condition) {
			return condition
		}

		if isTruthy(condition) {
			return Eval(scenario.Consequence, env)
		}
	}

	return NULL
}

func evalWhileExpression(
	we *ast.WhileExpression,
	env *object.Environment,
) object.Object {
	condition := Eval(we.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		evaluated := Eval(we.Consequence, env)

		if isError(evaluated) {
			return evaluated
		}

		evalWhileExpression(we, env)
	}
	return NULL
}


func evalForExpression(
	fe *ast.ForExpression,
	env *object.Environment,
) object.Object {
	
	
	existingIdentifier, identifierExisted := env.Get(fe.Identifier)

	
	err := Eval(fe.Starter, env)
	if isError(err) {
		return err
	}

	
	holds := true

	defer func() {
		if identifierExisted {
			env.Set(fe.Identifier, existingIdentifier)
		} else {
			env.Delete(fe.Identifier)
		}
	}()

	
	for holds {
		
		evaluated := Eval(fe.Condition, env)
		if isError(evaluated) {
			return evaluated
		}

		
		if isTruthy(evaluated) {
			res := Eval(fe.Block, env)
			if isError(res) {
				
				switch res.(type) {
				case *object.BreakError:
					return NULL
				case *object.ContinueError:

				case *object.Error:
					return res
				}
			}

			
			switch res.(type) {
			case *object.ReturnValue:
				return res
			default:
				
			}

			err = Eval(fe.Closer, env)
			if isError(err) {
				return err
			}

			continue
		}

		
		holds = false
	}

	return NULL
}


func evalForInExpression(
	fie *ast.ForInExpression,
	env *object.Environment,
) object.Object {
	iterable := Eval(fie.Iterable, env)
	
	
	existingKeyIdentifier, okk := env.Get(fie.Key)
	existingValueIdentifier, okv := env.Get(fie.Value)

	
	
	defer func() {
		if okk {
			env.Set(fie.Key, existingKeyIdentifier)
		} else {
			env.Delete(fie.Key)
		}

		if okv {
			env.Set(fie.Value, existingValueIdentifier)
		} else {
			env.Delete(fie.Value)
		}
	}()

	switch i := iterable.(type) {
	case object.Iterable:
		defer func() {
			i.Reset()
		}()

		return loopIterable(i.Next, env, fie, 0)
	case *object.Builtin:
		if i.Next == nil {
			return newError(fie.Token, "yerleşik fonksiyon dögüde kullanılmaz.")
		}

		return loopIterable(i.Next, env, fie, 0)
	default:
		return newError(fie.Token, "'%s' %s tipine sahip ve yenilenebilir değil. ", i.Inspect(), i.Type())
	}
}

func loopIterable(next func() (object.Object, object.Object), env *object.Environment, fie *ast.ForInExpression, index int64) object.Object {
	
	k, v := next()

	
	
	for k != nil && v != EOF {
		
		
		env.Set(fie.Key, k)
		env.Set(fie.Value, v)
		res := Eval(fie.Block, env)

		if isError(res) {
			
			
			
			
			switch res.(type) {
			case *object.BreakError:
				return NULL
			case *object.ContinueError:

			case *object.Error:
				return res
			}
		}

		
		switch res.(type) {
		case *object.ReturnValue:
			return res
		default:
			
		}
		index++
		k, v = next()
	}

	if k == nil || v == EOF {
		
		
		
		if index == 0 && fie.Alternative != nil {
			return Eval(fie.Alternative, env)
		}
	}

	return NULL
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := Fns[node.Value]; ok {
		return builtin
	}

	return newError(node.Token, "Bulunamadı: "+node.Value)
}
func isTruthy(obj object.Object) bool {
	switch v := obj.(type) {
	
	case *object.Null:
		return false
	case *object.Boolean:
		return v.Value
	
	
	case *object.Number:
		return v.Value != v.ZeroValue()
	
	
	case *object.String:
		return v.Value != v.ZeroValue()
	default:
		return true
	}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}
func evalPropertyExpression(pe *ast.PropertyExpression, env *object.Environment) object.Object {
	o := Eval(pe.Object, env)
	if isError(o) {
		return o
	}

	switch obj := o.(type) {
	case *object.String:
		
		if pe.Property.String() == "ok" {
			if obj.Ok != nil {
				return obj.Ok
			}

			return FALSE
		}
		
		if pe.Property.String() == "done" {
			if obj.Done != nil {
				return obj.Done
			}

			return FALSE
		}
	case *object.Hash:
		return evalHashIndexExpression(obj.Token, obj, &object.String{Token: pe.Token, Value: pe.Property.String()})
	}

	if pe.Optional {
		return NULL
	}

	return newError(pe.Token, "'%s' özelliği, %s tipinde geçersizdir.", pe.Property.String(), o.Type())
}

func applyFunction(tok token.Token, fn object.Object, env *object.Environment, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv, err := extendFunctionEnv(fn, args)

		if err != nil {
			return err
		}
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(tok, env, args...)

	default:
		return newError(tok, "bir fonksiyon değil: %s", fn.Type())
	}
}

func applyMethod(tok token.Token, o object.Object, me *ast.MethodExpression, env *object.Environment, args []object.Object) object.Object {
	method := me.Method.String()
	
	
	hash, isHash := o.(*object.Hash)

	
	if isHash && hash.GetKeyType(method) == object.FUNCTION_OBJ {
		pair, _ := hash.GetPair(method)
		return applyFunction(tok, pair.Value.(*object.Function), env, args)
	}

	
	f, ok := Fns[method]

	if !ok {
		if me.Optional {
			return NULL
		}

		return newError(tok, "%s için'%s() metodu mevcut değil'", o.Type(), method)
	}

	
	if !util.Contains(f.Types, string(o.Type())) && len(f.Types) != 0 {
		return newError(tok, "'%s()' metodu, '%s' üzerinde çağrılamaz.", method, o.Type())
	}

	
	args = append([]object.Object{o}, args...)
	return f.Fn(tok, env, args...)
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) (*object.Environment, *object.Error) {
	env := object.NewEnclosedEnvironment(fn.Env, args)

	for paramIdx, param := range fn.Parameters {
		argumentPassed := len(args) > paramIdx

		if !argumentPassed && param.Default == nil {
			return nil, newError(fn.Token, "%s fonksiyonu için %s argümanı bulunamadı.", param.Value, fn.Inspect())
		}

		var arg object.Object
		if argumentPassed {
			arg = args[paramIdx]
		} else {
			arg = Eval(param.Default, env)
		}

		env.Set(param.Value, arg)
	}

	return env, nil
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalIndexExpression(node *ast.IndexExpression, env *object.Environment) object.Object {
	tok := node.Token
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}
	end := Eval(node.End, env)
	if isError(end) {
		return end
	}

	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.NUMBER_OBJ:
		return evalArrayIndexExpression(tok, left, index, end, node.IsRange)
	case left.Type() == object.HASH_OBJ && index.Type() == object.STRING_OBJ:
		return evalHashIndexExpression(tok, left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.NUMBER_OBJ:
		return evalStringIndexExpression(tok, left, index, end, node.IsRange)
	default:
		return newError(tok, "dizin operatörü bu tip için geçersiz")
	}
}

func evalStringIndexExpression(tok token.Token, array, index object.Object, end object.Object, isRange bool) object.Object {
	stringObject := array.(*object.String)
	idx := index.(*object.Number).Int()
	max := len(stringObject.Value) - 1

	if isRange {
		max++
		
		if idx < 0 {
			idx = 0
		}
		endIdx, ok := end.(*object.Number)

		
		if ok {
			
			
			if endIdx.Int() < 0 {
				max = int(math.Max(float64(max+endIdx.Int()), 0))
			} else if endIdx.Int() < max {
				max = endIdx.Int()
			}
		} else if end != NULL {
			
			
			return newError(tok, `dizinler sayı olmalıdır fakat "%s" bulundu. (tip %s)`, end.Inspect(), end.Type())
		}

		
		
		if idx > max {
			return &object.String{Token: tok, Value: ""}
		}

		return &object.String{Token: tok, Value: string(stringObject.Value[idx:max])}
	}

	
	if idx > max {
		return &object.String{Token: tok, Value: ""}
	}

	if idx < 0 {
		length := max + 1

		
		if math.Abs(float64(idx)) > float64(length) {
			return &object.String{Token: tok, Value: ""}
		}

		
		
		
		idx = length + idx
	}

	return &object.String{Token: tok, Value: string(stringObject.Value[idx])}
}

func evalArrayIndexExpression(tok token.Token, array, index object.Object, end object.Object, isRange bool) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Number).Int()
	max := len(arrayObject.Elements) - 1

	if isRange {
		max++
		
		if idx < 0 {
			idx = 0
		}
		endIdx, ok := end.(*object.Number)

		
		if ok {
			
			
			if endIdx.Int() < 0 {
				max = int(math.Max(float64(max+endIdx.Int()), 0))
			} else if endIdx.Int() < max {
				max = endIdx.Int()
			}
		} else if end != NULL {
			
			
			return newError(tok, `dizinler sayı olmalıdır fakat "%s" bulundu. (tip %s)`, end.Inspect(), end.Type())
		}

		
		
		if idx > max {
			return &object.Array{Token: tok, Elements: []object.Object{}}
		}

		return &object.Array{Token: tok, Elements: arrayObject.Elements[idx:max]}
	}

	
	if idx > max {
		return NULL
	}

	if idx < 0 {
		length := max + 1

		
		if math.Abs(float64(idx)) > float64(length) {
			return NULL
		}

		
		
		
		idx = length + idx
	}

	return arrayObject.Elements[idx]
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError(node.Token, "Harita anahtarları olarak; sayılar, değişkenler ve fonksiyonlar kullanılamaz.")
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(tok token.Token, hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError(tok, "Harita anahtarı olarak %s kullanılamaz", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalCommandExpression(tok token.Token, cmd string, env *object.Environment) object.Object {
	cmd = strings.Trim(cmd, " ")

	
	cmd = util.InterpolateStringVars(cmd, env)

	
	background := len(cmd) > 1 && cmd[len(cmd)-1] == '&'
	
	
	
	if background {
		cmd = cmd[:len(cmd)-2]
	}

	
	s := &object.String{}

	parts := strings.Split(os.Getenv("ANK_COMMAND_EXECUTOR"), " ")
	c := exec.Command(parts[0], append(parts[1:], cmd)...)
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	s.Stdout = &stdout
	s.Stderr = &stderr
	s.Cmd = c
	s.Token = tok

	var err error
	if background {
		
		
		
		s.SetRunning()

		err := c.Start()
		if err != nil {
			s.SetCmdResult(FALSE)
			return FALSE
		}

		go evalCommandInBackground(s)
	} else {
		err = c.Run()
	}

	if !background {
		if err != nil {
			s.SetCmdResult(FALSE)
		} else {
			s.SetCmdResult(TRUE)
		}
	}

	return s
}

func evalCommandInBackground(s *object.String) {
	defer s.SetDone()

	err := s.Cmd.Wait()

	if err != nil {
		s.SetCmdResult(FALSE)
		return
	}

	s.SetCmdResult(TRUE)
}
