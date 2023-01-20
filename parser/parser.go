package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ankalang/anka/ast"
	"github.com/ankalang/anka/lexer"
	"github.com/ankalang/anka/token"
)

const (
	_ int = iota
	LOWEST
	AND         
	EQUALS      
	LESSGREATER 
	SUM         
	PRODUCT     
	RANGE       
	PREFIX      
	CALL        
	INDEX       
	QUESTION    
	DOT         
	HIGHEST     
)

var precedences = map[token.TokenType]int{
	token.AND:           AND,
	token.OR:            AND,
	token.BIT_AND:       AND,
	token.BIT_XOR:       AND,
	token.BIT_RSHIFT:    AND,
	token.BIT_LSHIFT:    AND,
	token.PIPE:          AND,
	token.EQ:            EQUALS,
	token.NOT_EQ:        EQUALS,
	token.TILDE:         EQUALS,
	token.IN:            EQUALS,
	token.NOT_IN:        EQUALS,
	token.COMMA:         EQUALS,
	token.LT:            LESSGREATER,
	token.LT_EQ:         LESSGREATER,
	token.GT:            LESSGREATER,
	token.GT_EQ:         LESSGREATER,
	token.COMBINED_COMP: LESSGREATER,
	token.PLUS:          SUM,
	token.MINUS:         SUM,
	token.SLASH:         PRODUCT,
	token.ASTERISK:      PRODUCT,
	token.EXPONENT:      PRODUCT,
	token.MODULO:        PRODUCT,
	token.COMP_PLUS:     EQUALS,
	token.COMP_MINUS:    EQUALS,
	token.COMP_SLASH:    EQUALS,
	token.COMP_ASTERISK: EQUALS,
	token.COMP_EXPONENT: EQUALS,
	token.COMP_MODULO:   EQUALS,
	token.RANGE:         RANGE,
	token.LPAREN:        CALL,
	token.LBRACKET:      INDEX,
	token.QUESTION:      QUESTION,
	token.DOT:           DOT,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	
	prevIndexExpression *ast.IndexExpression

	
	prevPropertyExpression *ast.PropertyExpression

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NUMBER, p.ParseNumberLiteral)
	p.registerPrefix(token.STRING, p.ParseStringLiteral)
	p.registerPrefix(token.NULL, p.ParseNullLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.PLUS, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TILDE, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.ParseBoolean)
	p.registerPrefix(token.FALSE, p.ParseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.WHILE, p.parseWhileExpression)
	p.registerPrefix(token.FOR, p.parseForExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.LBRACKET, p.ParseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.ParseHashLiteral)
	p.registerPrefix(token.COMMAND, p.parseCommand)
	p.registerPrefix(token.BREAK, p.parseBreak)
	p.registerPrefix(token.CONTINUE, p.parseContinue)
	p.registerPrefix(token.CURRENT_ARGS, p.parseCurrentArgsLiteral)
	p.registerPrefix(token.AT, p.parseDecorator)
	p.registerPrefix(token.DEFER, p.parseDefer)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.QUESTION, p.parseQuestionExpression)
	p.registerInfix(token.DOT, p.parseDottedExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.EXPONENT, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.COMP_PLUS, p.parseCompoundAssignment)
	p.registerInfix(token.COMP_MINUS, p.parseCompoundAssignment)
	p.registerInfix(token.COMP_SLASH, p.parseCompoundAssignment)
	p.registerInfix(token.COMP_EXPONENT, p.parseCompoundAssignment)
	p.registerInfix(token.COMP_MODULO, p.parseCompoundAssignment)
	p.registerInfix(token.COMP_ASTERISK, p.parseCompoundAssignment)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.TILDE, p.parseInfixExpression)
	p.registerInfix(token.IN, p.parseInfixExpression)
	p.registerInfix(token.NOT_IN, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)
	p.registerInfix(token.COMBINED_COMP, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.BIT_AND, p.parseInfixExpression)
	p.registerInfix(token.BIT_XOR, p.parseInfixExpression)
	p.registerInfix(token.PIPE, p.parseInfixExpression)
	p.registerInfix(token.BIT_RSHIFT, p.parseInfixExpression)
	p.registerInfix(token.BIT_LSHIFT, p.parseInfixExpression)
	p.registerInfix(token.RANGE, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	if p.curTokenIs(token.ILLEGAL) {
		msg := fmt.Sprintf(`Bu tokenin kullanımı yasak: '%s'`, p.curToken.Literal)
		p.reportError(msg, p.curToken)
	}
}

func (p *Parser) curTokenIs(typ token.TokenType) bool {
	return p.curToken.Type == typ
}

func (p *Parser) peekTokenIs(typ token.TokenType) bool {
	return p.peekToken.Type == typ
}

func (p *Parser) expectPeek(typ token.TokenType) bool {
	if p.peekTokenIs(typ) {
		p.nextToken()
		return true
	}
	p.peekError(p.curToken)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) reportError(err string, tok token.Token) {
	
	lineNum, column, errorLine := p.l.ErrorLine(tok.Position)
	msg := fmt.Sprintf("%s\n\033[97m\n%d:%d> %s ", err, lineNum, column, errorLine)
	
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekError(tok token.Token) {
	msg := fmt.Sprintf("\033[32m%s \033[97mbeklenilirken, \033[33m%s \033[97mbulundu", tok.Type, p.peekToken.Type)
	p.reportError(msg, tok)
}

func (p *Parser) noPrefixParseFnError(tok token.Token) {
	msg := fmt.Sprintf("'%s' için prefix bulunamadı.", tok.Literal)
	p.reportError(msg, tok)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	if p.curToken.Type == token.RETURN {
		return p.parseReturnStatement()
	}

	statement := p.parseAssignStatement()
	if statement != nil {
		return statement
	}

	return p.parseExpressionStatement()
}








func (p *Parser) Rewind(pos int) {
	p.l.Rewind(0)

	for p.l.CurrentPosition() < pos {
		p.nextToken()
	}
}

func (p *Parser) parseDestructuringIdentifiers() []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(token.ASSIGN) {
		return list
	}

	list = append(list, p.parseIdentifier())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.peekTokenIs(token.ASSIGN) {
		return nil
	}

	return list
}





func (p *Parser) parseAssignStatement() ast.Statement {
	stmt := &ast.AssignStatement{}

	
	if p.peekTokenIs(token.COMMA) {
		lexerPosition := p.l.CurrentPosition()
		
		if !p.curTokenIs(token.IDENT) {
			return nil
		}

		stmt.Names = p.parseDestructuringIdentifiers()

		if !p.peekTokenIs(token.ASSIGN) {
			p.Rewind(lexerPosition)
			return nil
		}
	} else if p.curTokenIs(token.IDENT) {
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.curTokenIs(token.ASSIGN) {
		stmt.Token = p.curToken
		if p.prevIndexExpression != nil {
			
			stmt.Index = p.prevIndexExpression
			p.nextToken()
			stmt.Value = p.parseExpression(LOWEST)
			
			p.prevIndexExpression = nil

			if p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}

			return stmt
		}
		if p.prevPropertyExpression != nil {
			
			stmt.Property = p.prevPropertyExpression
			p.nextToken()
			stmt.Value = p.parseExpression(LOWEST)
			
			p.prevPropertyExpression = nil

			if p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}

			return stmt
		}
	}

	if !p.peekTokenIs(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Token = p.curToken
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}


func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	returnToken := p.curToken

	
	if p.peekTokenIs(token.SEMICOLON) {
		stmt.ReturnValue = &ast.NullLiteral{Token: p.curToken}
	} else if p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) {
		
		stmt.ReturnValue = &ast.NullLiteral{Token: returnToken}
	} else {
		
		p.nextToken()
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}


func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}


func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}


func (p *Parser) ParseNumberLiteral() ast.Expression {
	lit := &ast.NumberLiteral{Token: p.curToken}
	var abbr float64
	var ok bool
	number := p.curToken.Literal

	
	if abbr, ok = token.NumberAbbreviations[strings.ToLower(string(number[len(number)-1]))]; ok {
		number = p.curToken.Literal[:len(p.curToken.Literal)-1]
	}

	value, err := strconv.ParseFloat(number, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as number", number)
		p.reportError(msg, p.curToken)
		return nil
	}

	if abbr != 0 {
		value *= abbr
	}

	lit.Value = value

	return lit
}


func (p *Parser) ParseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}


func (p *Parser) ParseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}


func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	precedence := PREFIX

	
	
	
	
	if p.curTokenIs(token.PLUS) || p.curTokenIs(token.MINUS) {
		precedence = HIGHEST
	}
	p.nextToken()

	expression.Right = p.parseExpression(precedence)

	return expression
}


func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}


func (p *Parser) parseCompoundAssignment(left ast.Expression) ast.Expression {
	expression := &ast.CompoundAssignment{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}


func (p *Parser) parseDottedExpression(object ast.Expression) ast.Expression {
	t := p.curToken
	precedence := p.curPrecedence()
	p.nextToken()

	
	
	
	
	
	
	
	
	if p.peekTokenIs(token.LPAREN) {
		exp := &ast.MethodExpression{Token: t, Object: object}
		exp.Method = p.parseExpression(precedence)
		p.nextToken()
		exp.Arguments = p.parseExpressionList(token.RPAREN)
		return exp
	} else {
		
		exp := &ast.PropertyExpression{Token: t, Object: object}
		exp.Property = p.parseIdentifier()
		p.prevPropertyExpression = exp
		p.prevIndexExpression = nil
		return exp
	}
}




func (p *Parser) parseQuestionExpression(object ast.Expression) ast.Expression {
	p.nextToken()
	exp := p.parseDottedExpression(object)

	switch res := exp.(type) {
	case *ast.PropertyExpression:
		res.Optional = true
		return res
	case *ast.MethodExpression:
		res.Optional = true
		return res
	default:
		return exp
	}
}


func (p *Parser) parseMethodExpression(object ast.Expression) ast.Expression {
	exp := &ast.MethodExpression{Token: p.curToken, Object: object}
	precedence := p.curPrecedence()
	p.nextToken()
	exp.Method = p.parseExpression(precedence)
	p.nextToken()
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}


func (p *Parser) ParseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}








func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	scenarios := []*ast.Scenario{}

	p.nextToken()
	scenario := &ast.Scenario{}
	scenario.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	scenario.Consequence = p.parseBlockStatement()
	scenarios = append(scenarios, scenario)
	
	
	for p.peekTokenIs(token.ELSE) {
		p.nextToken()
		p.nextToken()
		scenario := &ast.Scenario{}

		
		if p.curTokenIs(token.IF) {
			p.nextToken()
			scenario.Condition = p.parseExpression(LOWEST)

			if !p.expectPeek(token.LBRACE) {
				return nil
			}
		} else {
			
			
			
			
			
			
			
			
			tok := &token.Token{Position: -99, Literal: "true", Type: token.LookupIdent(token.TRUE)}
			scenario.Condition = &ast.Boolean{Token: *tok, Value: true}
		}

		scenario.Consequence = p.parseBlockStatement()
		scenarios = append(scenarios, scenario)
	}

	expression.Scenarios = scenarios
	return expression
}




func (p *Parser) parseWhileExpression() ast.Expression {
	expression := &ast.WhileExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()
	return expression
}




func (p *Parser) parseForExpression() ast.Expression {
	expression := &ast.ForExpression{Token: p.curToken}
	p.nextToken()

	if !p.curTokenIs(token.IDENT) {
		return nil
	}

	if !p.peekTokenIs(token.ASSIGN) {
		return p.parseForInExpression(expression)
	}

	expression.Identifier = p.curToken.Literal
	expression.Starter = p.parseAssignStatement()

	if expression.Starter == nil {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if expression.Condition == nil {
		return nil
	}
	p.nextToken()
	p.nextToken()
	expression.Closer = p.parseAssignStatement()
	if expression.Closer == nil {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Block = p.parseBlockStatement()

	return expression
}




func (p *Parser) parseForInExpression(initialExpression *ast.ForExpression) ast.Expression {
	expression := &ast.ForInExpression{Token: initialExpression.Token}

	if !p.curTokenIs(token.IDENT) {
		return nil
	}

	val := p.curToken.Literal
	var key string
	p.nextToken()

	if p.curTokenIs(token.COMMA) {
		p.nextToken()

		if !p.curTokenIs(token.IDENT) {
			return nil
		}

		key = val
		val = p.curToken.Literal
		p.nextToken()
	}

	expression.Key = key
	expression.Value = val

	if !p.curTokenIs(token.IN) {
		return nil
	}
	p.nextToken()

	expression.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Block = p.parseBlockStatement()

	
	
	
	
	
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}


func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}




func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		lit.Name = p.curToken.Literal
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}



func (p *Parser) parseDecorator() ast.Expression {
	dc := &ast.Decorator{Token: p.curToken}
	
	
	
	
	
	
	
	defer (func() {
		p.nextToken()
		exp := p.parseExpressionStatement()

		switch fn := exp.Expression.(type) {
		case *ast.FunctionLiteral:
			if fn.Name == "" {
				p.reportError("isimsiz bir fonksiyonda dekaratör kullanılamaz", dc.Token)
			}

			dc.Decorated = fn
		case *ast.Decorator:
			dc.Decorated = fn
		default:
			p.reportError("isimsiz bir fonksiyonda dekaratör kullanılamaz", dc.Token)
		}
	})()

	p.nextToken()
	exp := p.parseExpressionStatement()
	dc.Expression = exp.Expression

	return dc
}


func (p *Parser) parseDefer() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(0)

	if d, ok := exp.(ast.Deferrable); ok {
		d.SetDeferred(true)
	} else {
		p.reportError("sadece çağrılar beklenilebilir: bekle metod() | bekle `komut` | bekle fonksiyon()", p.curToken)
	}

	return exp
}


func (p *Parser) parseCurrentArgsLiteral() ast.Expression {
	return &ast.CurrentArgsLiteral{Token: p.curToken}
}


func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	parameters := []*ast.Parameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return parameters
	}

	p.nextToken()

	param, foundOptionalParameter := p.parseFunctionParameter()
	parameters = append(parameters, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		param, optional := p.parseFunctionParameter()

		if foundOptionalParameter && !optional {
			p.reportError("zorunlu parametrenin ardından isteğe bağlı parametre bulundu.", p.curToken)
		}

		if optional {
			foundOptionalParameter = true
		}

		parameters = append(parameters, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return parameters
}




func (p *Parser) parseFunctionParameter() (param *ast.Parameter, optional bool) {
	
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	
	
	if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.RPAREN) {
		return &ast.Parameter{Identifier: ident, Default: nil}, false
	}

	
	
	
	
	if !p.peekTokenIs(token.ASSIGN) {
		p.reportError("isteğe bağlı parametre kullanımı hatalı.", p.curToken)
		return &ast.Parameter{Identifier: ident, Default: nil}, false
	}

	
	p.nextToken()
	
	p.nextToken()
	
	
	
	
	
	
	
	
	exp := p.parseExpression(LOWEST)

	return &ast.Parameter{Identifier: ident, Default: exp}, true
}


func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}


func (p *Parser) ParseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}


func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	if p.peekTokenIs(token.COLON) {
		exp.Index = &ast.NumberLiteral{Value: 0, Token: token.Token{Type: token.NUMBER, Position: 0, Literal: "0"}}
		exp.IsRange = true
	} else {
		p.nextToken()
		exp.Index = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.COLON) {
		exp.IsRange = true
		p.nextToken()

		if p.peekTokenIs(token.RBRACKET) {
			exp.End = nil
		} else {
			p.nextToken()
			exp.End = p.parseExpression(LOWEST)
		}
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	
	p.prevIndexExpression = exp
	p.prevPropertyExpression = nil

	return exp
}


func (p *Parser) ParseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) parseCommand() ast.Expression {
	cmd := &ast.CommandExpression{Token: p.curToken, Value: p.curToken.Literal}
	return cmd
}



func (p *Parser) parseComment() ast.Expression {
	return nil
}

func (p *Parser) parseBreak() ast.Expression {
	return &ast.BreakStatement{Token: p.curToken}
}

func (p *Parser) parseContinue() ast.Expression {
	return &ast.ContinueStatement{Token: p.curToken}
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
