package posix

import (
	"fmt"
	"strings"
	"unicode"
)

// Parser handles POSIX shell command parsing
type Parser struct {
	input string
	pos   int
}

// ASTNode represents a node in the abstract syntax tree
type ASTNode interface {
	String() string
}

// CommandNode represents a command with arguments and redirections
type CommandNode struct {
	Command   string
	Args      []ASTNode
	Redirs    []Redirection
	Background bool
}

// String implements fmt.Stringer interface
func (n *CommandNode) String() string {
	args := make([]string, len(n.Args))
	for i, arg := range n.Args {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s %s", n.Command, strings.Join(args, " "))
}

// PipelineNode represents a pipeline of commands
type PipelineNode struct {
	Commands []ASTNode
}

// String implements fmt.Stringer interface
func (n *PipelineNode) String() string {
	commands := make([]string, len(n.Commands))
	for i, cmd := range n.Commands {
		commands[i] = cmd.String()
	}
	return strings.Join(commands, " | ")
}

// IfNode represents an if statement
type IfNode struct {
	Condition  ASTNode
	ThenBlock  []ASTNode
	ElseBlock  []ASTNode
}

// String implements fmt.Stringer interface
func (n *IfNode) String() string {
	return "if [condition]"
}

// ForNode represents a for loop
type ForNode struct {
	Variable string
	Values   []ASTNode
	Body     []ASTNode
}

// String implements fmt.Stringer interface
func (n *ForNode) String() string {
	return fmt.Sprintf("for %s in values", n.Variable)
}

// WhileNode represents a while loop
type WhileNode struct {
	Condition ASTNode
	Body      []ASTNode
}

// String implements fmt.Stringer interface
func (n *WhileNode) String() string {
	return "while [condition]"
}

// StringNode represents a literal string
type StringNode struct {
	Value string
}

// String implements fmt.Stringer interface
func (n *StringNode) String() string {
	return fmt.Sprintf(`"%s"`, n.Value)
}

// VariableNode represents a variable reference
type VariableNode struct {
	Name string
}

// String implements fmt.Stringer interface
func (n *VariableNode) String() string {
	return fmt.Sprintf("$%s", n.Name)
}

// SubstitutionNode represents command substitution
type SubstitutionNode struct {
	Command ASTNode
}

// String implements fmt.Stringer interface
func (n *SubstitutionNode) String() string {
	return fmt.Sprintf("$(%s)", n.Command.String())
}

// Redirection represents I/O redirection
type Redirection struct {
	Type     RedirType
	Source   int
	Target   string
}

type RedirType int

const (
	RedirOutput RedirType = iota // >
	RedirAppend                  // >>
	RedirInput                   // <
	RedirError                   // 2>
	RedirAll                     // &>
)

// NewParser creates a new parser
func NewParser(input string) *Parser {
	return &Parser{input: input, pos: 0}
}

// Parse parses the input into an AST
func (p *Parser) Parse() (ASTNode, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("empty input")
	}
	
	return p.parsePipeline()
}

// parsePipeline parses a pipeline of commands
func (p *Parser) parsePipeline() (ASTNode, error) {
	commands := []ASTNode{}
	
	for {
		cmd, err := p.parseCommand()
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
		
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.peek() != '|' {
			break
		}
		p.pos++ // skip '|'
		p.skipWhitespace()
	}
	
	if len(commands) == 1 {
		return commands[0], nil
	}
	
	return &PipelineNode{Commands: commands}, nil
}

// parseCommand parses a single command
func (p *Parser) parseCommand() (ASTNode, error) {
	p.skipWhitespace()
	
	// Check for control structures
	if strings.HasPrefix(p.input[p.pos:], "if ") {
		return p.parseIf()
	}
	if strings.HasPrefix(p.input[p.pos:], "for ") {
		return p.parseFor()
	}
	if strings.HasPrefix(p.input[p.pos:], "while ") {
		return p.parseWhile()
	}
	
	// Parse regular command
	args := []ASTNode{}
	redirs := []Redirection{}
	background := false
	
	for p.pos < len(p.input) {
		p.skipWhitespace()
		
		// Check for background operator
		if p.peek() == '&' {
			background = true
			p.pos++
			break
		}
		
		// Check for redirection
		if redir, ok := p.parseRedirection(); ok {
			redirs = append(redirs, redir)
			continue
		}
		
		// Parse argument
		arg, err := p.parseArgument()
		if err != nil {
			break
		}
		args = append(args, arg)
	}
	
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	
	cmd, ok := args[0].(*StringNode)
	if !ok {
		return nil, fmt.Errorf("command must be a string")
	}
	
	return &CommandNode{
		Command:   cmd.Value,
		Args:      args[1:],
		Redirs:    redirs,
		Background: background,
	}, nil
}

// parseArgument parses a single argument
func (p *Parser) parseArgument() (ASTNode, error) {
	p.skipWhitespace()
	
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end of input")
	}
	
	// Check if we're at a command terminator
	ch := p.peek()
	if unicode.IsSpace(ch) || ch == '|' || ch == '&' || ch == '>' || ch == '<' || ch == ')' {
		return nil, fmt.Errorf("no argument at position %d", p.pos)
	}
	
	switch ch {
	case '"':
		return p.parseQuotedString('"')
	case '\'':
		return p.parseQuotedString('\'')
	case '$':
		return p.parseVariable()
	case '`':
		return p.parseBacktickSubstitution()
	case '(':
		return p.parseParenSubstitution()
	default:
		return p.parseUnquotedString()
	}
}

// parseQuotedString parses a quoted string
func (p *Parser) parseQuotedString(quote rune) (ASTNode, error) {
	p.pos++ // skip opening quote
	var value strings.Builder
	
	for p.pos < len(p.input) && p.peek() != quote {
		ch := p.peek()
		if ch == '\\' {
			p.pos++
			if p.pos < len(p.input) {
				value.WriteRune(p.peek())
				p.pos++
			}
		} else {
			value.WriteRune(ch)
			p.pos++
		}
	}
	
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unterminated quoted string")
	}
	
	p.pos++ // skip closing quote
	return &StringNode{Value: value.String()}, nil
}

// parseUnquotedString parses an unquoted string
func (p *Parser) parseUnquotedString() (ASTNode, error) {
	var value strings.Builder
	
	for p.pos < len(p.input) {
		ch := p.peek()
		if unicode.IsSpace(ch) || ch == '|' || ch == '&' || ch == '>' || ch == '<' || ch == '"' || ch == '\'' || ch == '$' || ch == '`' || ch == '(' || ch == ')' {
			break
		}
		if ch == '\\' {
			p.pos++
			if p.pos < len(p.input) {
				value.WriteRune(p.peek())
				p.pos++
			}
		} else {
			value.WriteRune(ch)
			p.pos++
		}
	}
	
	return &StringNode{Value: value.String()}, nil
}

// parseVariable parses a variable reference
func (p *Parser) parseVariable() (ASTNode, error) {
	p.pos++ // skip '$'
	
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end after $")
	}
	
	if p.peek() == '(' {
		return p.parseParenSubstitution()
	}
	
	// Parse variable name
	var name strings.Builder
	for p.pos < len(p.input) {
		ch := p.peek()
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			name.WriteRune(ch)
			p.pos++
		} else {
			break
		}
	}
	
	if name.Len() == 0 {
		return nil, fmt.Errorf("empty variable name")
	}
	
	return &VariableNode{Name: name.String()}, nil
}

// parseParenSubstitution parses $(command) substitution
func (p *Parser) parseParenSubstitution() (ASTNode, error) {
	p.pos += 2 // skip '$('
	
	var cmd strings.Builder
	nesting := 1
	
	for p.pos < len(p.input) && nesting > 0 {
		ch := p.peek()
		switch ch {
		case '(':
			nesting++
			cmd.WriteRune(ch)
			p.pos++
		case ')':
			nesting--
			if nesting > 0 {
				cmd.WriteRune(ch)
			}
			p.pos++
		default:
			cmd.WriteRune(ch)
			p.pos++
		}
	}
	
	if nesting > 0 {
		return nil, fmt.Errorf("unterminated command substitution")
	}
	
	// Parse the command recursively
	subParser := NewParser(cmd.String())
	subAST, err := subParser.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in command substitution: %v", err)
	}
	
	return &SubstitutionNode{Command: subAST}, nil
}

// parseBacktickSubstitution parses `command` substitution
func (p *Parser) parseBacktickSubstitution() (ASTNode, error) {
	p.pos++ // skip '`'
	
	var cmd strings.Builder
	
	for p.pos < len(p.input) && p.peek() != '`' {
		ch := p.peek()
		if ch == '\\' {
			p.pos++
			if p.pos < len(p.input) {
				cmd.WriteRune(p.peek())
				p.pos++
			}
		} else {
			cmd.WriteRune(ch)
			p.pos++
		}
	}
	
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unterminated backtick substitution")
	}
	
	p.pos++ // skip closing '`'
	
	// Parse the command recursively
	subParser := NewParser(cmd.String())
	subAST, err := subParser.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in backtick substitution: %v", err)
	}
	
	return &SubstitutionNode{Command: subAST}, nil
}

// parseRedirection parses I/O redirection
func (p *Parser) parseRedirection() (Redirection, bool) {
	p.skipWhitespace()
	
	if p.pos >= len(p.input) {
		return Redirection{}, false
	}
	
	// Check for redirection operators
	if strings.HasPrefix(p.input[p.pos:], ">>") {
		p.pos += 2
		target, err := p.parseRedirectionTarget()
		if err != nil {
			return Redirection{}, false
		}
		return Redirection{Type: RedirAppend, Source: 1, Target: target}, true
	}
	
	if strings.HasPrefix(p.input[p.pos:], "2>") {
		p.pos += 2
		target, err := p.parseRedirectionTarget()
		if err != nil {
			return Redirection{}, false
		}
		return Redirection{Type: RedirError, Source: 2, Target: target}, true
	}
	
	if strings.HasPrefix(p.input[p.pos:], "&>") {
		p.pos += 2
		target, err := p.parseRedirectionTarget()
		if err != nil {
			return Redirection{}, false
		}
		return Redirection{Type: RedirAll, Source: 0, Target: target}, true
	}
	
	if p.peek() == '>' {
		p.pos++
		target, err := p.parseRedirectionTarget()
		if err != nil {
			return Redirection{}, false
		}
		return Redirection{Type: RedirOutput, Source: 1, Target: target}, true
	}
	
	if p.peek() == '<' {
		p.pos++
		target, err := p.parseRedirectionTarget()
		if err != nil {
			return Redirection{}, false
		}
		return Redirection{Type: RedirInput, Source: 0, Target: target}, true
	}
	
	return Redirection{}, false
}

// parseRedirectionTarget parses the target of a redirection
func (p *Parser) parseRedirectionTarget() (string, error) {
	p.skipWhitespace()
	
	if p.pos >= len(p.input) {
		return "", fmt.Errorf("expected redirection target")
	}
	
	// Target can be a word or a quoted string
	if p.peek() == '"' || p.peek() == '\'' {
		node, err := p.parseQuotedString(p.peek())
		if err != nil {
			return "", err
		}
		if strNode, ok := node.(*StringNode); ok {
			return strNode.Value, nil
		}
		return "", fmt.Errorf("expected string for redirection target")
	}
	
	var target strings.Builder
	for p.pos < len(p.input) {
		ch := p.peek()
		if unicode.IsSpace(ch) || ch == '|' || ch == '&' || ch == '>' || ch == '<' {
			break
		}
		target.WriteRune(ch)
		p.pos++
	}
	
	return target.String(), nil
}

// parseIf parses an if statement
func (p *Parser) parseIf() (ASTNode, error) {
	// TODO: Implement if statement parsing
	return nil, fmt.Errorf("if statements not yet implemented")
}

// parseFor parses a for loop
func (p *Parser) parseFor() (ASTNode, error) {
	// TODO: Implement for loop parsing
	return nil, fmt.Errorf("for loops not yet implemented")
}

// parseWhile parses a while loop
func (p *Parser) parseWhile() (ASTNode, error) {
	// TODO: Implement while loop parsing
	return nil, fmt.Errorf("while loops not yet implemented")
}

// Helper methods
func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *Parser) peek() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	return rune(p.input[p.pos])
}
