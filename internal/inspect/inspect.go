package inspect

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RichardFlp/portmaster/internal/ports"
	"github.com/RichardFlp/portmaster/internal/processes"
)

type Listener struct {
	Status       string `json:"status"`
	Protocol     string `json:"protocol"`
	Address      string `json:"address"`
	PID          int    `json:"pid,omitempty"`
	Process      string `json:"process,omitempty"`
	Executable   string `json:"executable,omitempty"`
	Command      string `json:"command,omitempty"`
	ParentPID    int    `json:"parent_pid,omitempty"`
	Parent       string `json:"parent,omitempty"`
	Started      string `json:"started,omitempty"`
	LookupFailed bool   `json:"-"`
}

type Result struct {
	Port      int        `json:"port"`
	Listeners []Listener `json:"listeners"`
}

func Build(port int, ls []ports.Listener) Result {
	result := Result{Port: port}
	for _, l := range ls {
		li := Listener{
			Protocol: l.Protocol,
			Address:  l.Address,
			PID:      l.PID,
			Process:  l.Process,
		}
		if l.Protocol == "udp" {
			li.Status = "BOUND"
		} else {
			li.Status = "LISTENING"
		}
		if l.PID > 0 {
			if info, err := processes.Lookup(l.PID); err == nil {
				li.Executable = info.Executable
				li.Command = info.Command
				li.ParentPID = info.ParentPID
				li.Parent = info.ParentName
				if !info.Started.IsZero() {
					li.Started = info.Started.Format("2006-01-02 15:04:05")
				}
			} else {
				li.LookupFailed = true
			}
		}
		result.Listeners = append(result.Listeners, li)
	}
	return result
}

func (r Result) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Port %d\n\n", r.Port)
	for i, l := range r.Listeners {
		if i > 0 {
			b.WriteString("\n")
		}
		writeField(&b, "Status", l.Status)
		writeField(&b, "Protocol", strings.ToUpper(l.Protocol))
		writeField(&b, "Address", fmt.Sprintf("%s:%d", l.Address, r.Port))
		if l.PID > 0 {
			b.WriteString("\n")
			writeField(&b, "PID", strconv.Itoa(l.PID))
			if l.Process != "" {
				writeField(&b, "Process", l.Process)
			}
			if l.LookupFailed {
				b.WriteString("Unable to inspect process " + strconv.Itoa(l.PID) + ". Additional permissions may be required.\n")
			}
			if l.Executable != "" {
				writeField(&b, "Executable", l.Executable)
			}
			if l.Command != "" {
				b.WriteString("Command:\n  " + l.Command + "\n")
			}
			if l.ParentPID > 0 {
				b.WriteString("\n")
				writeField(&b, "Parent PID", strconv.Itoa(l.ParentPID))
			}
			if l.Parent != "" {
				writeField(&b, "Parent", l.Parent)
			}
			if l.Started != "" {
				b.WriteString("\n")
				writeField(&b, "Started", l.Started)
			}
		}
	}
	return b.String()
}

func writeField(b *strings.Builder, label, value string) {
	fmt.Fprintf(b, "%-13s%s\n", label+":", value)
}
