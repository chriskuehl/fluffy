import pygments.lexers
import pytest

from fluffy.highlighting import guess_lexer
from fluffy.highlighting import UI_LANGUAGES_MAP


@pytest.mark.parametrize('language', UI_LANGUAGES_MAP)
def test_ui_language_exists(language):
    """Ensure a lexer exists for each language we advertise."""
    assert pygments.lexers.get_lexer_by_name('python') is not None


@pytest.fixture
def example_c():
    return '''
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


def test_guess_lexer_uses_valid_lang(example_c):
    assert guess_lexer(example_c, 'ruby').name == 'Ruby'


@pytest.mark.parametrize('invalid_lang', ['herpderp', '', None, 'autodetect'])
def test_guess_lexer_autodetects_with_invalid_lang(example_c, invalid_lang):
    assert guess_lexer(example_c, invalid_lang).name == 'C'


def test_guess_lexer_falls_back_to_python():
    assert guess_lexer('what language even is this', None).name == 'Python'
