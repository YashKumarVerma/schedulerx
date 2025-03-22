package scheduler

import (
	"github.com/robfig/cron/v3"
)

// Parser handles cron expression parsing
type Parser struct {
	parser cron.Parser
}

// NewParser creates a new cron parser
func NewParser() *Parser {
	return &Parser{
		parser: cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// Parse parses a cron expression
func (p *Parser) Parse(spec string) (cron.Schedule, error) {
	return p.parser.Parse(spec)
}
