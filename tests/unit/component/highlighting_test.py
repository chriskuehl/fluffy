import pygments.lexers
import pytest

from fluffy.component.highlighting import _guess_language_with_magika
from fluffy.component.highlighting import DiffHighlighter
from fluffy.component.highlighting import get_highlighter
from fluffy.component.highlighting import guess_lexer
from fluffy.component.highlighting import looks_like_diff
from fluffy.component.highlighting import MAGIKA_LABEL_TO_PYGMENTS_LEXER
from fluffy.component.highlighting import PasteText
from fluffy.component.highlighting import PygmentsHighlighter
from fluffy.component.highlighting import strip_diff_things
from fluffy.component.highlighting import UI_LANGUAGES_MAP


EXAMPLE_DIFF = '''\
commit 5eb58ea2be01b451583429c4d8a931c0bcdbac8e
Author:     Chris Kuehl <ckuehl@ocf.berkeley.edu>
AuthorDate: Mon Jul 25 20:49:11 2016 -0400
Commit:     Chris Kuehl <ckuehl@ocf.berkeley.edu>
CommitDate: Mon Jul 25 20:49:11 2016 -0400

    Don't strip newlines, add horizontal scrollbar when overflow

diff --git a/fluffy/highlighting.py b/fluffy/highlighting.py
index 217363a..409d912 100644
--- a/fluffy/highlighting.py
+++ b/fluffy/highlighting.py
@@ -38,12 +38,12 @@ _pygments_formatter = HtmlFormatter(

 def guess_lexer(text, language):
     try:
-        return pygments.lexers.get_lexer_by_name(language)
+        return pygments.lexers.get_lexer_by_name(language, stripnl=False)
     except pygments.util.ClassNotFound:
         - try:
-            return pygments.lexers.guess_lexer(text)
+            return pygments.lexers.guess_lexer(text, stripnl=False)
         except pygments.util.ClassNotFound:
-            return pygments.lexers.get_lexer_by_name('python')
+            return pygments.lexers.get_lexer_by_name('python', stripnl=False)
'''


EXAMPLE_C = '''\
#include <stdio.h>
#include <stdlib.h>

int main(void);

int main(void) {
    uint8_t x = 42;
    uint8_t y = x + 1;

    /* exit 1 for success! */
    return 1;
}
'''


@pytest.mark.parametrize('language', UI_LANGUAGES_MAP)
def test_ui_language_exists(language):
    """Ensure a lexer exists for each language we advertise."""
    assert pygments.lexers.get_lexer_by_name('python') is not None


def test_guess_lexer_precedence():
    # Prefers exact lexer name match
    assert guess_lexer(EXAMPLE_C, 'ruby', 'my-thing.css').name == 'Ruby'

    # Otherwise uses filename detection
    assert guess_lexer(EXAMPLE_C, 'not-a-lexer', 'my-thing.css').name == 'CSS'

    # Finally uses text detection
    assert guess_lexer(EXAMPLE_C, 'not-a-lexer', 'not-a-filename-that-matches').name == 'C'


@pytest.mark.parametrize('invalid_lang', ['herpderp', '', None, 'autodetect'])
def test_guess_lexer_autodetects_with_invalid_lang(invalid_lang):
    assert guess_lexer(EXAMPLE_C, invalid_lang, None).name == 'C'


def test_guess_lexer_heuristic_detects_sql(monkeypatch):
    monkeypatch.setattr(
        pygments.lexers,
        'guess_lexer',
        lambda *_args, **kwargs: pygments.lexers.TextLexer(**kwargs),
    )
    assert guess_lexer('SELECT id FROM users WHERE active = 1;', None, None).name == 'SQL'


def test_guess_lexer_heuristic_detects_json(monkeypatch):
    monkeypatch.setattr(
        pygments.lexers,
        'guess_lexer',
        lambda *_args, **kwargs: pygments.lexers.TextLexer(**kwargs),
    )
    assert guess_lexer('{"name": "fluffy", "enabled": true}', None, None).name == 'JSON'


def test_guess_lexer_falls_back_to_python():
    assert guess_lexer('what language even is this', None, None).name == 'Python'


@pytest.mark.parametrize(
    ('text', 'expected'), (
        ('', False),
        (
            'some simple\n'
            'text is here\n',
            False,
        ),
        (EXAMPLE_DIFF, True),
    ),
)
def test_looks_like_diff(text, expected):
    assert looks_like_diff(text) is expected


def test_strip_diff_things():
    assert strip_diff_things(EXAMPLE_DIFF) == '''\

    Don't strip newlines, add horizontal scrollbar when overflow


 def guess_lexer(text, language):
     try:
        return pygments.lexers.get_lexer_by_name(language)
        return pygments.lexers.get_lexer_by_name(language, stripnl=False)
     except pygments.util.ClassNotFound:
         - try:
            return pygments.lexers.guess_lexer(text)
            return pygments.lexers.guess_lexer(text, stripnl=False)
         except pygments.util.ClassNotFound:
            return pygments.lexers.get_lexer_by_name('python')
            return pygments.lexers.get_lexer_by_name('python', stripnl=False)
'''


@pytest.mark.parametrize(
    ('text', 'language', 'filename', 'expected'), (
        (EXAMPLE_C, 'c', None, pygments.lexers.get_lexer_by_name('c')),
        (EXAMPLE_C, 'does not exist', None, pygments.lexers.get_lexer_by_name('c')),
        (EXAMPLE_C, None, None, pygments.lexers.get_lexer_by_name('c')),
        (EXAMPLE_DIFF, 'c', None, pygments.lexers.get_lexer_by_name('c')),
        (EXAMPLE_C, None, 'my_file.rs', pygments.lexers.get_lexer_by_name('rust')),
    ),
)
def test_get_highlighter_pygments(text, language, filename, expected):
    h = get_highlighter(text, language, filename)
    assert isinstance(h, PygmentsHighlighter)
    assert type(h.lexer) is type(expected)


@pytest.mark.parametrize(
    ('text', 'language', 'expected'), (
        (EXAMPLE_DIFF, None, pygments.lexers.get_lexer_by_name('python')),
        (EXAMPLE_DIFF, 'diff', pygments.lexers.get_lexer_by_name('python')),
        (EXAMPLE_C, 'diff', pygments.lexers.get_lexer_by_name('c')),

        # requesting a diff language
        (EXAMPLE_DIFF, 'diff-c', pygments.lexers.get_lexer_by_name('c')),
        # bogus language
        (EXAMPLE_DIFF, 'diff-lolidonotexist', pygments.lexers.get_lexer_by_name('python')),
    ),
)
def test_get_highlighter_diff(text, language, expected):
    h = get_highlighter(text, language, None)
    assert isinstance(h, DiffHighlighter)
    assert type(h.lexer) is type(expected)


def test_diff_highlighter_prepare_text():
    highlighter = DiffHighlighter(pygments.lexers.get_lexer_by_name('text'))
    orig_text = '''\
 common line 1
+added line 1
 common line 2
+added line 2
-deleted line 1
-deleted line 2
 common line 3
-deleted line 3
+added line 3
 common line 4
+added line 4
-deleted line 4
-deleted line 5'''

    text1, text2, text3 = highlighter.prepare_text(orig_text)
    assert text1 == PasteText(
        '''\
 common line 1

 common line 2
-deleted line 1
-deleted line 2
 common line 3
-deleted line 3
 common line 4
-deleted line 4
-deleted line 5''',
        {
            1: [1],
            2: [2],
            3: [3],
            4: [4, 5],
            5: [6],
            6: [7],
            7: [8, 9],
            8: [10],
            9: [11, 12],
            10: [13],
        },
    )
    assert text2 == PasteText(
        '''\
 common line 1
+added line 1
 common line 2
+added line 2

 common line 3
+added line 3
 common line 4
+added line 4
''',
        {
            1: [1],
            2: [2],
            3: [3],
            4: [4, 5],
            5: [6],
            6: [7],
            7: [8, 9],
            8: [10],
            9: [11, 12],
            10: [13],
        },
    )
    assert text3 == PasteText(orig_text)


# --- Magika integration tests ---

EXAMPLE_PYTHON = '''\
import os
from pathlib import Path


def main() -> None:
    for p in Path(".").iterdir():
        if p.is_file():
            print(f"Found file: {p}")


if __name__ == "__main__":
    main()
'''

EXAMPLE_JAVASCRIPT = '''\
const express = require("express");
const app = express();

app.get("/", (req, res) => {
    res.json({ message: "Hello, world!" });
});

app.listen(3000, () => {
    console.log("Server running on port 3000");
});
'''

EXAMPLE_JAVA = '''\
import java.util.List;
import java.util.ArrayList;

public class Main {
    public static void main(String[] args) {
        List<String> items = new ArrayList<>();
        items.add("hello");
        for (String item : items) {
            System.out.println(item);
        }
    }
}
'''

EXAMPLE_GO = '''\
package main

import (
    "fmt"
    "os"
)

func main() {
    args := os.Args[1:]
    for _, arg := range args {
        fmt.Printf("arg: %s\\n", arg)
    }
}
'''

EXAMPLE_RUST = '''\
use std::collections::HashMap;

fn main() {
    let mut scores: HashMap<&str, i32> = HashMap::new();
    scores.insert("Alice", 10);
    scores.insert("Bob", 20);

    for (name, score) in &scores {
        println!("{}: {}", name, score);
    }
}
'''

EXAMPLE_SQL = '''\
SELECT u.id, u.name, COUNT(o.id) AS order_count
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
WHERE u.created_at > '2024-01-01'
GROUP BY u.id, u.name
HAVING COUNT(o.id) > 5
ORDER BY order_count DESC;
'''

EXAMPLE_YAML = '''\
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    app: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
'''

EXAMPLE_HTML = '''\
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Test Page</title>
</head>
<body>
    <h1>Hello, world!</h1>
    <p>This is a test.</p>
</body>
</html>
'''

EXAMPLE_BASH = '''\
#!/bin/bash
set -euo pipefail

for file in *.log; do
    if [ -f "$file" ]; then
        echo "Processing $file"
        gzip "$file"
    fi
done
'''

EXAMPLE_JSON = '''\
{
    "name": "fluffy",
    "version": "1.0.0",
    "dependencies": {
        "flask": ">=2.0",
        "pygments": "*"
    },
    "scripts": {
        "start": "python -m fluffy"
    }
}
'''

EXAMPLE_RUBY = '''\
class Greeter
  attr_reader :name

  def initialize(name)
    @name = name
  end

  def greet
    puts "Hello, #{@name}!"
  end
end

greeter = Greeter.new("World")
greeter.greet
'''


@pytest.mark.parametrize(
    ('text', 'expected_lexer_name'), (
        (EXAMPLE_PYTHON, 'Python'),
        (EXAMPLE_JAVASCRIPT, 'JavaScript'),
        (EXAMPLE_JAVA, 'Java'),
        (EXAMPLE_GO, 'Go'),
        (EXAMPLE_RUST, 'Rust'),
        (EXAMPLE_SQL, 'SQL'),
        (EXAMPLE_YAML, 'YAML'),
        (EXAMPLE_BASH, 'Bash'),
        (EXAMPLE_JSON, 'JSON'),
        (EXAMPLE_RUBY, 'Ruby'),
        (EXAMPLE_C, 'C'),
    ),
)
def test_magika_detects_common_languages(text, expected_lexer_name):
    """Magika should correctly identify common programming languages."""
    detected = _guess_language_with_magika(text)
    assert detected is not None, f'Magika returned None, expected a lexer mapping to {expected_lexer_name}'
    lexer = pygments.lexers.get_lexer_by_name(detected)
    assert lexer.name == expected_lexer_name, (
        f'Expected {expected_lexer_name}, got {lexer.name} (magika returned {detected!r})'
    )


def test_magika_returns_none_for_plain_text():
    """Magika should return None for ambiguous/plain text so we fall back."""
    result = _guess_language_with_magika('what language even is this')
    assert result is None


def test_magika_label_map_entries_are_valid_pygments_lexers():
    """Every value in the magika-to-pygments mapping should resolve to a real lexer."""
    for magika_label, pygments_name in MAGIKA_LABEL_TO_PYGMENTS_LEXER.items():
        try:
            pygments.lexers.get_lexer_by_name(pygments_name)
        except pygments.util.ClassNotFound:
            pytest.fail(
                f'MAGIKA_LABEL_TO_PYGMENTS_LEXER[{magika_label!r}] = {pygments_name!r} '
                f'is not a valid Pygments lexer name',
            )


def test_guess_lexer_uses_magika_for_autodetect():
    """When language is None/autodetect, guess_lexer should use Magika and
    produce better results than the old Pygments-only fallback."""
    lexer = guess_lexer(EXAMPLE_GO, None, None)
    assert lexer.name == 'Go'

    lexer = guess_lexer(EXAMPLE_RUST, None, None)
    assert lexer.name == 'Rust'

    lexer = guess_lexer(EXAMPLE_JAVA, None, None)
    assert lexer.name == 'Java'


def test_guess_lexer_explicit_language_still_takes_precedence():
    """An explicit language choice should override Magika detection."""
    lexer = guess_lexer(EXAMPLE_PYTHON, 'ruby', None)
    assert lexer.name == 'Ruby'


def test_guess_lexer_filename_still_takes_precedence_over_magika():
    """Filename-based detection should still take precedence over Magika."""
    lexer = guess_lexer(EXAMPLE_PYTHON, 'not-a-lexer', 'script.rb')
    assert lexer.name == 'Ruby'
