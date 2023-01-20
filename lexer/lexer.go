package lexer

import (
	"strings"
	"unicode"

	"github.com/ankalang/anka/token"
)

type Lexer struct {
	position     int  
	readPosition int  
	ch           rune 
	input        []rune
	
	lineMap [][2]int 
}

func New(in string) *Lexer {
	l := &Lexer{input: []rune(in)}
	
	l.buildLineMap()
	
	l.readChar()
	return l
}


func (l *Lexer) buildLineMap() {
	begin := 0
	idx := 0
	for i, ch := range l.input {
		idx = i
		if ch == '\n' {
			l.lineMap = append(l.lineMap, [2]int{begin, idx})
			begin = idx + 1
		}
	}
	
	l.lineMap = append(l.lineMap, [2]int{begin, idx + 1})
}


func (l *Lexer) CurrentPosition() int {
	return l.position
}


func (l *Lexer) linePosition(pos int) (int, int, int) {
	idx := 0
	begin := 0
	end := 0
	for i, tuple := range l.lineMap {
		idx = i
		begin, end = tuple[0], tuple[1]
		if pos >= begin && pos <= end {
			break
		}
	}
	lineNum := idx + 1
	return lineNum, begin, end
}


func (l *Lexer) ErrorLine(pos int) (int, int, string) {
	lineNum, begin, end := l.linePosition(pos)
	errorLine := l.input[begin:end]
	column := pos - begin + 1
	return lineNum, column, string(errorLine)
}

func (l *Lexer) newToken(tokenType token.TokenType) token.Token {
	return token.Token{
		Type:     tokenType,
		Position: l.position,
		Literal:  string(l.ch)}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			tok = l.newToken(token.EQ)
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok.Literal = literal
		} else {
			tok = l.newToken(token.ASSIGN)
		}
	case '+':
		if l.peekChar() == '=' {
			tok.Type = token.COMP_PLUS
			tok.Position = l.position
			tok.Literal = "+="
			l.readChar()
		} else {
			tok = l.newToken(token.PLUS)
		}
	case '-':
		if l.peekChar() == '=' {
			tok.Type = token.COMP_MINUS
			tok.Position = l.position
			tok.Literal = "-="
			l.readChar()
		} else {
			tok = l.newToken(token.MINUS)
		}
	case '%':
		if l.peekChar() == '=' {
			tok.Type = token.COMP_MODULO
			tok.Position = l.position
			tok.Literal = "%="
			l.readChar()
		} else {
			tok = l.newToken(token.MODULO)
		}
	case '!':
		if l.peekChar() == '=' {
			tok = l.newToken(token.NOT_EQ)
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok.Literal = literal
		} else if l.peekChars(3) == "in " {
			tok = l.newToken(token.NOT_IN)
			l.readChar()
			l.readChar()
			tok.Literal = "!in"
		} else {
			tok = l.newToken(token.BANG)
		}
	case '/':
		if l.peekChar() == '/' {
			
			_ = l.readLine()
			l.readChar()
			return l.NextToken()
		} else if l.peekChar() == '=' {
			tok.Type = token.COMP_SLASH
			tok.Position = l.position
			tok.Literal = "/="
			l.readChar()
		} else {
			tok = l.newToken(token.SLASH)
		}
	case '#':
		
		_ = l.readLine()
		l.readChar()
		return l.NextToken()
	case '&':
		if l.peekChar() == '&' {
			tok.Type = token.AND
			tok.Position = l.position
			tok.Literal = l.readLogicalOperator()
		} else {
			tok = l.newToken(token.BIT_AND)
		}
	case '^':
		tok = l.newToken(token.BIT_XOR)
	case '@':
		tok = l.newToken(token.AT)
	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			if l.peekChar() == '=' {
				tok.Type = token.COMP_EXPONENT
				tok.Position = l.position
				tok.Literal = "**="
				l.readChar()
			} else {
				tok.Type = token.EXPONENT
				tok.Position = l.position
				tok.Literal = "**"
			}
		} else if l.peekChar() == '=' {
			tok.Type = token.COMP_ASTERISK
			tok.Position = l.position
			tok.Literal = "*="
			l.readChar()
		} else {
			tok = l.newToken(token.ASTERISK)
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()

			if l.peekChar() == '>' {
				tok.Type = token.COMBINED_COMP
				tok.Position = l.position
				tok.Literal = "<=>"
				l.readChar()
			} else {
				tok.Type = token.LT_EQ
				tok.Position = l.position
				tok.Literal = "<="
			}
		} else if l.peekChar() == '<' {
			tok.Type = token.BIT_LSHIFT
			tok.Position = l.position
			tok.Literal = "<<"
			l.readChar()
		} else {
			tok = l.newToken(token.LT)
		}
	case '>':
		if l.peekChar() == '=' {
			tok.Type = token.GT_EQ
			tok.Position = l.position
			tok.Literal = ">="
			l.readChar()
		} else if l.peekChar() == '>' {
			tok.Type = token.BIT_RSHIFT
			tok.Position = l.position
			tok.Literal = ">>"
			l.readChar()
		} else {
			tok = l.newToken(token.GT)
		}
	case ';':
		tok = l.newToken(token.SEMICOLON)
	case ':':
		tok = l.newToken(token.COLON)
	case ',':
		tok = l.newToken(token.COMMA)
	case '.':
		position := l.position

		if l.peekChar() == '.' {
			l.readChar()

			if l.peekChar() == '.' {
				tok.Type = token.CURRENT_ARGS
				tok.Position = position
				tok.Literal = "..."
				l.readChar()
			} else {
				tok.Type = token.RANGE
				tok.Position = position
				tok.Literal = ".."
			}
		} else {
			tok = l.newToken(token.DOT)
		}
	case '?':
		tok = l.newToken(token.QUESTION)
	case '|':
		if l.peekChar() == '|' {
			tok.Type = token.OR
			tok.Position = l.position
			tok.Literal = l.readLogicalOperator()
		} else {
			tok = l.newToken(token.PIPE)
		}
	case '{':
		tok = l.newToken(token.LBRACE)
	case '}':
		tok = l.newToken(token.RBRACE)
	case '~':
		tok = l.newToken(token.TILDE)
	case '(':
		tok = l.newToken(token.LPAREN)
	case ')':
		tok = l.newToken(token.RPAREN)
	case '"':
		tok.Type = token.STRING
		tok.Position = l.position
		tok.Literal = l.readString('"')
	case '\'':
		tok.Type = token.STRING
		tok.Position = l.position
		tok.Literal = l.readString('\'')
	case '`':
		tok.Type = token.COMMAND
		tok.Position = l.position
		tok.Literal = l.readString('`')
	case '$':
		if l.peekChar() == '(' {
			tok.Type = token.COMMAND
			tok.Position = l.position
			tok.Literal = l.readCommand()
		} else {
			tok.Type = token.ILLEGAL
			tok.Position = l.position
			tok.Literal = l.readLine()
		}
	case '[':
		tok = l.newToken(token.LBRACKET)
	case ']':
		tok = l.newToken(token.RBRACKET)
	case 0:
		tok.Type = token.EOF
		tok.Position = l.position
		tok.Literal = ""
	default:
		if isLetter(l.ch) {
			tok.Position = l.position
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Position = l.position
			literal, kind := l.readNumber()
			tok.Type = kind
			tok.Literal = literal
			return tok
		} else {
			tok = l.newToken(token.ILLEGAL)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}


func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) Rewind(pos int) {
	l.ch = l.input[0]
	l.position = 0
	l.readPosition = l.position + 1

	for l.position < pos {
		l.NextToken()
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) prevChar(steps int) rune {
	prevPosition := l.readPosition - steps
	if prevPosition < 1 {
		return 0
	}
	return l.input[prevPosition]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || unicode.IsDigit(l.ch) {
		l.readChar()
	}
	return string(l.input[position:l.position])
}

func (l *Lexer) peekChars(amount int) string {
	if l.readPosition+amount >= len(l.input) {
		return ""
	}
	return string(l.input[l.readPosition : l.readPosition+amount])
}








func isCharAllowedInNumber(c rune) bool {
	lowchar := unicode.ToLower(c)
	_, isAbbr := token.NumberAbbreviations[string(lowchar)]

	return isDigit(lowchar) || lowchar == '.' || lowchar == '+' || lowchar == '-' || c == 'e' || lowchar == token.NumberSeparator || isAbbr
}






func (l *Lexer) readNumber() (number string, kind token.TokenType) {
	position := l.position
	kind = token.NUMBER
	var hasDot bool
	var hasExponent bool

	for isCharAllowedInNumber(l.ch) {
		
		
		
		if (l.ch == '+' || l.ch == '-') && !hasExponent {
			return string(l.input[position:l.position]), kind
		}

		
		
		if l.ch == 'e' {
			hasExponent = true
		}

		
		
		
		if _, isAbbr := token.NumberAbbreviations[string(l.ch)]; isAbbr {
			l.readChar()
			return string(l.input[position:l.position]), kind
		}

		
		
		if l.ch == '.' && (l.peekChar() == '.' || !isDigit(l.peekChar())) {
			return string(l.input[position:l.position]), kind
		}

		if l.ch == '.' {
			
			if hasDot {
				return string(l.input[position : l.position+1]), token.ILLEGAL
			}

			hasDot = true
		}
		l.readChar()
	}

	
	
	if l.input[l.position-1] == 'e' {
		return string(l.input[position:l.position]), token.ILLEGAL
	}

	return strings.ReplaceAll(string(l.input[position:l.position]), "_", ""), kind
}




func (l *Lexer) readLogicalOperator() string {
	l.readChar()
	return string(l.input[l.position-1 : l.position+1])
}






func (l *Lexer) readString(quote byte) string {
	var chars []string
	esc := rune('\\')
	doubleEscape := false
	for {
		l.readChar()

		if l.ch == esc && l.peekChar() == esc {
			chars = append(chars, string(esc))
			l.readChar()
			
			if l.peekChar() == rune(quote) {
				doubleEscape = true
			} else {
				
				chars = append(chars, string(esc))
			}
			continue
		}
		
		
		
		if l.ch == esc && l.peekChar() == rune(quote) {
			chars = append(chars, string(quote))
			l.readChar()
			continue
		}
		
		
		
		if quote == '"' {
			if l.ch == esc && l.peekChar() == 'n' {
				chars = append(chars, "\n")
				l.readChar()
				continue
			} else if l.ch == esc && l.peekChar() == 'r' {
				chars = append(chars, "\r")
				l.readChar()
				continue
			} else if l.ch == esc && l.peekChar() == 't' {
				chars = append(chars, "\t")
				l.readChar()
				continue
			}
		}
		
		
		
		if (l.ch == rune(quote) && (l.prevChar(2) != esc || doubleEscape)) || l.ch == 0 {
			break
		}
		chars = append(chars, string(l.ch))
		doubleEscape = false
	}
	return strings.Join(chars, "")
}




func (l *Lexer) readLine() string {
	position := l.position
	for {
		l.readChar()
		if l.ch == '\n' || l.ch == '\r' || l.ch == 0 {
			break
		}
	}
	return string(l.input[position:l.position])
}










func (l *Lexer) readCommand() string {
	position := l.position + 2
	subtract := 1
	for {
		l.readChar()

		if l.ch == '\n' || l.ch == '\r' || l.ch == 0 {
			
			if l.prevChar(2) == ';' {
				subtract = 2
			}
			break
		}
	}
	ret := l.input[position : l.position-subtract]

	
	
	if subtract == 2 {
		l.position = l.position - 1
		l.readPosition = l.position
	}

	return string(ret)
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
