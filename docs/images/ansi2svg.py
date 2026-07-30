#!/usr/bin/env python3
"""Convert ANSI-colored terminal output to SVG. Stdlib only."""
import re
import sys
from html import escape

FONT = "SFMono-Regular,Menlo,Monaco,Consolas,'Liberation Mono','Courier New',monospace"
CHAR_W = 8.4
LINE_H = 20
PAD_X = 20
PAD_Y = 16
TITLE_H = 40
CORNER_R = 10
BG = "#282a36"
PROMPT_COLOR = "#636d83"

BASIC_COLORS = {
    30: "#6272a4", 31: "#ff5555", 32: "#50fa7b", 33: "#f1fa8c",
    34: "#bd93f9", 35: "#ff79c6", 36: "#8be9fd", 37: "#f8f8f2",
    90: "#6272a4", 91: "#ff6e6e", 92: "#69ff94", 93: "#ffffa5",
    94: "#d6acff", 95: "#ff92df", 96: "#a4ffff", 97: "#ffffff",
}
BASIC_BG = {k + 10: v for k, v in BASIC_COLORS.items()}

SGR_RE = re.compile(r'\033\[([0-9;]*)m')
ANSI_ANY = re.compile(r'\033\[[0-9;]*[A-Za-z]')


def parse_sgr(codes_str, state):
    s = dict(state)
    codes = [int(c) if c else 0 for c in codes_str.split(';')] if codes_str else [0]
    i = 0
    while i < len(codes):
        c = codes[i]
        if c == 0:
            s = {"fg": None, "bg": None, "bold": False, "dim": False, "underline": False}
        elif c == 1:
            s["bold"] = True
        elif c == 2:
            s["dim"] = True
        elif c == 4:
            s["underline"] = True
        elif c == 22:
            s["bold"] = False
            s["dim"] = False
        elif c == 24:
            s["underline"] = False
        elif c == 39:
            s["fg"] = None
        elif c == 49:
            s["bg"] = None
        elif c in BASIC_COLORS:
            s["fg"] = BASIC_COLORS[c]
        elif c in BASIC_BG:
            s["bg"] = BASIC_BG[c]
        elif c == 38 and i + 1 < len(codes) and codes[i + 1] == 2 and i + 4 < len(codes):
            s["fg"] = f"#{codes[i+2]:02x}{codes[i+3]:02x}{codes[i+4]:02x}"
            i += 4
        elif c == 48 and i + 1 < len(codes) and codes[i + 1] == 2 and i + 4 < len(codes):
            s["bg"] = f"#{codes[i+2]:02x}{codes[i+3]:02x}{codes[i+4]:02x}"
            i += 4
        i += 1
    return s


def line_has_bg(raw):
    state = {"fg": None, "bg": None, "bold": False, "dim": False, "underline": False}
    for m in SGR_RE.finditer(raw):
        state = parse_sgr(m.group(1), state)
        if state["bg"]:
            return state["bg"]
    return None


def render_line(raw, y, content_w):
    parts = []
    bg_color = line_has_bg(raw)
    ty_base = TITLE_H + PAD_Y + y * LINE_H + 14

    if bg_color:
        parts.append(
            f'<rect x="0" y="{ty_base - 14}" width="{content_w}" '
            f'height="{LINE_H}" fill="{bg_color}"/>'
        )

    state = {"fg": None, "bg": None, "bold": False, "dim": False, "underline": False}
    col = 0
    pos = 0
    for m in SGR_RE.finditer(raw):
        text = raw[pos:m.start()]
        text = ANSI_ANY.sub('', text)
        if text:
            parts.extend(make_spans(text, state, col, y))
            col += len(text)
        state = parse_sgr(m.group(1), state)
        pos = m.end()
    tail = raw[pos:]
    tail = ANSI_ANY.sub('', tail)
    if tail:
        parts.extend(make_spans(tail, state, col, y))
    return parts


def make_spans(text, state, col, y):
    spans = []
    x = PAD_X + col * CHAR_W
    ty = TITLE_H + PAD_Y + y * LINE_H + 14

    attrs = [f'x="{x}"', f'y="{ty}"']
    fill = state["fg"] or "#f8f8f2"
    if state["dim"]:
        attrs.append('opacity="0.5"')
    attrs.append(f'fill="{fill}"')
    if state["bold"]:
        attrs.append('font-weight="bold"')
    deco = []
    if state["underline"]:
        deco.append("underline")
    if deco:
        attrs.append(f'text-decoration="{" ".join(deco)}"')
    spans.append(f'<text {" ".join(attrs)}>{escape(text)}</text>')
    return spans


def main():
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--prompt', help='Command to show as prompt line')
    parser.add_argument('--bg', default=BG, help='Background color')
    args = parser.parse_args()

    raw = sys.stdin.read()
    lines = raw.rstrip('\n').split('\n')

    prompt_line = None
    if args.prompt:
        prompt_line = args.prompt

    max_cols = 0
    for line in lines:
        plain = ANSI_ANY.sub('', line)
        max_cols = max(max_cols, len(plain))
    if prompt_line:
        max_cols = max(max_cols, len(prompt_line) + 2)

    w = PAD_X * 2 + max_cols * CHAR_W + 10
    total_lines = len(lines) + (1 if prompt_line else 0)
    h = TITLE_H + PAD_Y * 2 + total_lines * LINE_H + 4

    svg = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{w}" height="{h}" '
        f'font-family="{FONT}" font-size="13px">',
        f'<rect width="{w}" height="{h}" rx="{CORNER_R}" fill="{args.bg}"/>',
        f'<circle cx="20" cy="20" r="6" fill="#ff5f56"/>',
        f'<circle cx="40" cy="20" r="6" fill="#ffbd2e"/>',
        f'<circle cx="60" cy="20" r="6" fill="#27c93f"/>',
    ]

    line_offset = 0
    if prompt_line:
        ty = TITLE_H + PAD_Y + 14
        svg.append(f'<text x="{PAD_X}" y="{ty}" fill="{PROMPT_COLOR}">$</text>')
        svg.append(f'<text x="{PAD_X + 2 * CHAR_W}" y="{ty}" fill="#f8f8f2">{escape(prompt_line)}</text>')
        line_offset = 1

    for i, line in enumerate(lines):
        svg.extend(render_line(line, i + line_offset, w))

    svg.append('</svg>')
    print('\n'.join(svg))


if __name__ == '__main__':
    main()
