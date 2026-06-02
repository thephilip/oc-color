package output

import (
	"strings"
	"testing"
)

func TestHighlightYAML_DocMarker(t *testing.T) {
	th := testTheme()
	got := highlightYAML("---\n", th)
	want := wrapWithTheme("---", "pink", th) + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHighlightYAML_DocEnd(t *testing.T) {
	th := testTheme()
	got := highlightYAML("...\n", th)
	want := wrapWithTheme("...", "pink", th) + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHighlightYAML_Comment(t *testing.T) {
	th := testTheme()
	got := highlightYAML("# this is a comment\n", th)
	want := wrapWithTheme("# this is a comment", "dim", th) + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHighlightYAML_Key_Colored(t *testing.T) {
	th := testTheme()
	got := highlightYAML("name: my-pod\n", th)
	if !strings.Contains(got, wrapWithTheme("name", "key", th)) {
		t.Errorf("key not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_Bool(t *testing.T) {
	th := testTheme()
	got := highlightYAML("active: true\n", th)
	if !strings.Contains(got, wrapWithTheme("true", "info", th)) {
		t.Errorf("boolean value not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_Null(t *testing.T) {
	th := testTheme()
	got := highlightYAML("deletionTimestamp: null\n", th)
	if !strings.Contains(got, wrapWithTheme("null", "dim", th)) {
		t.Errorf("null value not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_Number(t *testing.T) {
	th := testTheme()
	got := highlightYAML("replicas: 3\n", th)
	if !strings.Contains(got, wrapWithTheme("3", "accent", th)) {
		t.Errorf("numeric value not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_QuotedDouble(t *testing.T) {
	th := testTheme()
	got := highlightYAML("name: \"my-pod\"\n", th)
	if !strings.Contains(got, wrapWithTheme("\"my-pod\"", "success", th)) {
		t.Errorf("double-quoted string not colored; got %q", got)
	}
}

func TestHighlightYAML_Value_QuotedSingle(t *testing.T) {
	th := testTheme()
	got := highlightYAML("name: 'my-pod'\n", th)
	if !strings.Contains(got, wrapWithTheme("'my-pod'", "success", th)) {
		t.Errorf("single-quoted string not colored; got %q", got)
	}
}

func TestHighlightYAML_ListBullet_Colored(t *testing.T) {
	th := testTheme()
	got := highlightYAML("- item\n", th)
	if !strings.Contains(got, wrapWithTheme("-", "pink", th)) {
		t.Errorf("list bullet not colored; got %q", got)
	}
}

func TestHighlightYAML_IndentedKey(t *testing.T) {
	th := testTheme()
	got := highlightYAML("  name: my-pod\n", th)
	if !strings.Contains(got, wrapWithTheme("name", "key", th)) {
		t.Errorf("indented key not colored; got %q", got)
	}
	if !strings.HasPrefix(got, "  ") {
		t.Errorf("indentation not preserved; got %q", got)
	}
}

func TestHighlightYAML_EmptyValue(t *testing.T) {
	th := testTheme()
	got := highlightYAML("labels:\n", th)
	if !strings.Contains(got, wrapWithTheme("labels", "key", th)) {
		t.Errorf("key with empty value not colored; got %q", got)
	}
}

func TestHighlightYAML_PreservesNewlines(t *testing.T) {
	th := testTheme()
	input := "---\nname: pod\n"
	got := highlightYAML(input, th)
	if strings.Count(got, "\n") != strings.Count(input, "\n") {
		t.Errorf("newline count changed: input %d, got %d", strings.Count(input, "\n"), strings.Count(got, "\n"))
	}
}

func TestHighlightYAML_BlockScalar_Pipe(t *testing.T) {
	th := testTheme()
	input := "command: |\n  echo hello\n  echo world\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme("|", "dim", th)) {
		t.Errorf("block scalar header '|' not colored dim; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("echo hello", "success", th)) {
		t.Errorf("block scalar content not colored success; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("echo world", "success", th)) {
		t.Errorf("second block scalar line not colored success; got %q", got)
	}
}

func TestHighlightYAML_BlockScalar_Fold(t *testing.T) {
	th := testTheme()
	input := "message: >\n  long line\n  continuation\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme(">", "dim", th)) {
		t.Errorf("block scalar header '>' not colored dim; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("long line", "success", th)) {
		t.Errorf("folded scalar content not colored success; got %q", got)
	}
}

func TestHighlightYAML_BlockScalar_WithChomping(t *testing.T) {
	th := testTheme()
	input := "script: |-\n  set -e\n  echo done\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme("|-", "dim", th)) {
		t.Errorf("block scalar header '|-' not colored dim; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("set -e", "success", th)) {
		t.Errorf("block scalar content not colored success; got %q", got)
	}
}

func TestHighlightYAML_BlockScalar_EndedByReducedIndent(t *testing.T) {
	th := testTheme()
	input := "data: |\n  line one\nnextKey: value\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme("nextKey", "key", th)) {
		t.Errorf("key after block scalar not recognized as key; got %q", got)
	}
}

func TestHighlightYAML_BlockScalar_BlankLineInside(t *testing.T) {
	th := testTheme()
	// A blank line inside a block scalar must not end the scalar
	input := "data: |\n  line one\n\n  line two\nnextKey: val\n"
	got := highlightYAML(input, th)
	if !strings.Contains(got, wrapWithTheme("line two", "success", th)) {
		t.Errorf("line after blank inside block scalar not colored success; got %q", got)
	}
	if !strings.Contains(got, wrapWithTheme("nextKey", "key", th)) {
		t.Errorf("key after block scalar not recognized; got %q", got)
	}
}
