package output

import (
	"strings"
	"testing"

	"github.com/thephilip/oc-color/theme"
)

func testTheme() theme.Theme {
	return theme.Theme{
		Name: "test",
		Tokens: map[string]theme.TokenStyle{
			"success": {Color: "green"},
			"warning": {Color: "yellow"},
			"error":   {Color: "red", Bold: true},
			"info":    {Color: "cyan"},
			"accent":  {Color: "purple"},
			"pink":    {Color: "pink"},
			"orange":  {Color: "orange"},
			"dim":     {Color: "gray"},
			"header":  {Color: "purple", Bold: true, Underline: true},
			"key":     {Color: "yellow"},
			"value":   {Color: "white"},
		},
	}
}

func TestLooksLikeDescribe(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty",
			input: "",
			want:  false,
		},
		{
			name: "key-value pairs",
			input: `Name:         my-pod
Namespace:    default
Node:         node-1/192.168.1.1`,
			want: true,
		},
		{
			name: "events section",
			input: `Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  12h   default-scheduler  Successfully assigned`,
			want: true,
		},
		{
			name: "conditions section",
			input: `Conditions:
  Type              Status
  ----              ------
  Initialized       True`,
			want: true,
		},
		{
			name: "oc get pods output",
			input: `NAMESPACE     NAME                        READY   STATUS              RESTARTS   AGE
default       web-1                        1/1     Running             0          12h`,
			want: false,
		},
		{
			name: "JSON output",
			input: `{
  "apiVersion": "v1"
}`,
			want: false,
		},
		{
			name: "YAML output",
			input: `---
apiVersion: v1`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeDescribe(tt.input)
			if got != tt.want {
				t.Errorf("looksLikeDescribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessDescribe(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	input := `Name:         my-pod
Namespace:    default
Status:       Running
Node:         node-1/192.168.1.1
Labels:       app=web
Annotations:  <none>

Conditions:
  Type              Status
  ----              ------
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True

Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  12h   default-scheduler  Successfully assigned
  Warning BackOff    5m    kubelet            Back-off restarting failed container
`

	result := p.processDescribe(input)

	if !strings.Contains(result, "\033[") {
		t.Error("expected ANSI escape codes in output")
	}

	if strings.Contains(result, "<none>") && !strings.Contains(result, "\033[") {
		t.Error("expected <none> to be colorized")
	}

	if !strings.Contains(result, "Running") {
		t.Error("expected Running status in output")
	}
}

func TestProcessDescribeSectionHeader(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("Conditions:\n")
	if !strings.Contains(result, "Conditions") {
		t.Error("expected Conditions in output")
	}
}

func TestProcessDescribeKeyValue(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("Name:         my-pod\n")
	if !strings.Contains(result, "Name") {
		t.Error("expected Name in output")
	}
	if !strings.Contains(result, "my-pod") {
		t.Error("expected my-pod in output")
	}
}

func TestProcessDescribeEventNormal(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("  Normal  Scheduled  12h   default-scheduler  Successfully assigned\n")
	if !strings.Contains(result, "Normal") {
		t.Error("expected Normal in output")
	}
}

func TestProcessDescribeEventWarning(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("  Warning BackOff    5m    kubelet            Back-off restarting\n")
	if !strings.Contains(result, "Warning") {
		t.Error("expected Warning in output")
	}
}

func TestProcessDescribeDash(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("  ----    ------     ----  ----               -------\n")
	if !strings.Contains(result, "----") {
		t.Error("expected dashes in output")
	}
}

func TestProcessDescribeNoneValue(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("Annotations:  <none>\n")
	// <none> should be wrapped in ANSI codes
	if !strings.Contains(result, "\033[") {
		t.Error("expected ANSI codes for <none>")
	}
}

func TestProcessDescribeNotSetValue(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("Node-Selectors:  <not set>\n")
	if !strings.Contains(result, "\033[") {
		t.Error("expected ANSI codes for <not set>")
	}
}

func TestProcessDescribeColourOff(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: false}

	result := p.Process("Name:         my-pod\nStatus:       Running\n")
	if strings.Contains(result, "\033[") {
		t.Error("expected no ANSI codes when Colour is false")
	}
}

func TestProcessDescribeFalseCondition(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("  Ready             False\n")
	if !strings.Contains(result, "\033[") {
		t.Error("expected ANSI codes for False condition")
	}
}

func TestProcessDescribeSubSection(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("  web:\n")
	if !strings.Contains(result, "web") {
		t.Error("expected web in output")
	}
}

func TestProcessDescribeSubKeyValue(t *testing.T) {
	th := testTheme()
	p := &Processor{Theme: th, Colour: true}

	result := p.processDescribeLine("    Container ID:   cri-o://abc123\n")
	if !strings.Contains(result, "Container ID") {
		t.Error("expected Container ID in output")
	}
	if !strings.Contains(result, "cri-o://abc123") {
		t.Error("expected cri-o://abc123 in output")
	}
}
