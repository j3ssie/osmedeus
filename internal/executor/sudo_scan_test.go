package executor

import (
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestContainsSudoCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		// Positive cases
		{"sudo at start", "sudo nmap -sS 10.0.0.1", true},
		{"sudo after pipe", "cat hosts.txt | sudo nmap -iL -", true},
		{"sudo after semicolon", "echo starting; sudo apt update", true},
		{"sudo after &&", "cd /tmp && sudo rm -rf old", true},
		{"sudo after ||", "test -f /etc/hosts || sudo touch /etc/hosts", true},
		{"sudo after backtick", "`sudo whoami`", true},
		{"sudo after $(", "echo $(sudo cat /etc/shadow)", true},
		{"sudo with extra spaces", "| sudo  nmap", true},

		// Negative cases
		{"empty string", "", false},
		{"no sudo", "nmap -sS 10.0.0.1", false},
		{"pseudo word", "pseudo random generator", false},
		{"sudoku", "play sudoku today", false},
		{"sudo without space (not a command)", "sudoers file", false},
		{"just the word sudo at end", "check sudo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsSudoCommand(tt.cmd)
			assert.Equal(t, tt.want, got, "containsSudoCommand(%q)", tt.cmd)
		})
	}
}

func TestScanStepsForSudo(t *testing.T) {
	t.Run("no sudo", func(t *testing.T) {
		steps := []core.Step{
			{Name: "step1", Command: "nmap -sT 10.0.0.1"},
			{Name: "step2", Commands: []string{"echo hello", "cat /etc/hostname"}},
		}
		assert.False(t, scanStepsForSudo(steps))
	})

	t.Run("sudo in command", func(t *testing.T) {
		steps := []core.Step{
			{Name: "step1", Command: "echo hello"},
			{Name: "step2", Command: "sudo nmap -sS 10.0.0.1"},
		}
		assert.True(t, scanStepsForSudo(steps))
	})

	t.Run("sudo in commands list", func(t *testing.T) {
		steps := []core.Step{
			{Name: "step1", Commands: []string{"echo one", "sudo apt update"}},
		}
		assert.True(t, scanStepsForSudo(steps))
	})

	t.Run("sudo in parallel commands", func(t *testing.T) {
		steps := []core.Step{
			{Name: "step1", ParallelCommands: []string{"sudo iptables -L"}},
		}
		assert.True(t, scanStepsForSudo(steps))
	})
}

func TestScanStepsForSudo_Nested(t *testing.T) {
	t.Run("sudo in parallel steps", func(t *testing.T) {
		steps := []core.Step{
			{
				Name: "parallel",
				ParallelSteps: []core.Step{
					{Name: "inner", Command: "sudo systemctl restart nginx"},
				},
			},
		}
		assert.True(t, scanStepsForSudo(steps))
	})

	t.Run("sudo in foreach step", func(t *testing.T) {
		inner := core.Step{Name: "inner", Command: "sudo nmap -sS [[target]]"}
		steps := []core.Step{
			{
				Name: "foreach",
				Step: &inner,
			},
		}
		assert.True(t, scanStepsForSudo(steps))
	})

	t.Run("no sudo in nested", func(t *testing.T) {
		inner := core.Step{Name: "inner", Command: "echo [[target]]"}
		steps := []core.Step{
			{
				Name: "foreach",
				Step: &inner,
				ParallelSteps: []core.Step{
					{Name: "p1", Command: "cat file.txt"},
				},
			},
		}
		assert.False(t, scanStepsForSudo(steps))
	})
}

func TestScanStepsForSudo_Empty(t *testing.T) {
	assert.False(t, scanStepsForSudo(nil))
	assert.False(t, scanStepsForSudo([]core.Step{}))
}
